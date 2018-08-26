package chord

import (
	"net"
	"fmt"
	"time"
	"context"
	"github.com/lukaspj/go-chord/api"
	"google.golang.org/grpc"
)

const gearDownPeriod = time.Minute
const stabilizationIntervalStart = time.Second
const stabilizationIntervalEnd = 5 * time.Minute
const fixFingersIntervalStart = time.Second
const fixFingersIntervalEnd = 5 * time.Minute

type ContactInfo struct {
	Address string `json:"address"`
	Id      NodeID `json:"id"`
	Payload []byte `json:"payload"`
}

type Peer struct {
	Info                     *ContactInfo
	Port                     int
	network                  *chordNetwork
	stabilizationFunction    tickingFunction
	fixFingersFunction       tickingFunction
	checkPredecessorFunction tickingFunction
}

func NewPeer(info *ContactInfo, port int) (peer Peer) {
	logger.Info("Creating new peer, with id: %s", info.Id.String())
	peer.Port = port
	peer.Info = info
	peer.network = NewChordNetwork(peer.Info)

	return
}

func (peer *Peer) Listen() {
	logger.Info("Listening on port: %d", peer.Port)

	peer.network.predecessor = nil
	if peer.network.successors.SetSuccessor(0, peer.Info) {
		peer.network.lastDirtyTime = time.Now()
	}

	grpcServer := grpc.NewServer()
	api.RegisterChordServer(grpcServer, &ServiceWrapper{service: peer})

	if l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", peer.Port)); err == nil {
		go grpcServer.Serve(l)
	}

	peer.stabilizationFunction = StartTickingFunction(func() int {
		interpolate := cubic(float64(stabilizationIntervalStart), float64(stabilizationIntervalEnd))
		err := peer.network.Stabilize()
		if err != nil {
			logger.Error("error when stabilizing: %v", err)
		}
		return int(interpolate(float64(peer.network.TimeSinceChange()) / float64(gearDownPeriod)))
	})

	peer.fixFingersFunction = StartTickingFunction(func() int {
		interpolate := cubic(float64(fixFingersIntervalStart), float64(fixFingersIntervalEnd))
		err := peer.network.FixFingers()
		if err != nil {
			logger.Error("error when fixing fingers: %v", err)
		}
		return int(interpolate(float64(peer.network.TimeSinceChange()) / float64(gearDownPeriod)))
	})

	peer.checkPredecessorFunction = StartTickingFunction(func() int {
		err := peer.network.CheckPredecessor()
		if err != nil {
			logger.Error("error when checking predecessor: %v", err)
		}
		return int(time.Second * 20)
	})
}

func (peer *Peer) Connect(address string) (err error) {
	var info *ContactInfo
	logger.Info("Connecting to: %s", address)
	if info, err = peer.network.Ping(address); err == nil {
		logger.Info("Connection successful, remote peer is: %s", info.Id.String())
		var successor *ContactInfo
		logger.Info("Looking up successor to: %s", peer.Info.Id.String())
		successor, err = peer.network.FindSuccessor(info, peer.Info.Id)
		if err != nil {
			logger.Error("Failed to lookup successor: %v", err)
		}
		if peer.network.successors.SetSuccessor(0, successor) {
			peer.network.lastDirtyTime = time.Now()
		}
	} else {
		logger.Error("Failed to connect: %v", err)
	}
	return
}

func (peer *Peer) GetSuccessor() (info *ContactInfo) {
	return peer.network.successors.GetSuccessor(0)
}

func (peer *Peer) GetPredecessor() (info *ContactInfo) {
	return peer.network.predecessor
}

func (peer *Peer) ResponsibleFor(id NodeID) bool {
	return id.Between(peer.network.predecessor.Id, peer.Info.Id)
}

func (peer *Peer) Poke() {
	go func() { peer.stabilizationFunction.tick <- true }()
	go func() { peer.fixFingersFunction.tick <- true }()
}

func (peer *Peer) Ping(ctx context.Context) (info *ContactInfo, err error) {
	logger.Debug("Ping")
	info = peer.Info
	if info == nil {
		logger.Error("Error")
	}
	return
}

func (peer *Peer) FindSuccessor(ctx context.Context, id *NodeID) (info *ContactInfo, err error) {
	successor := peer.network.successors.GetSuccessor(0)

	logger.Debug("FindSuccessor to: %s", id.String())

	// if (id ∈ (n, successor] )
	if id.Between(peer.Info.Id, successor.Id) {
		// return successor;
		info = successor
		logger.Debug("returning: %v", info)
	} else {
		// forward the query around the circle
		// n0 = successor.closest_preceding_node(id);
		var n0 *ContactInfo
		n0, err = peer.network.ClosestPrecedingNode(successor, *id)

		if err != nil {
			peer.network.UpdateSuccessorList()
			successor = peer.network.successors.GetSuccessor(0)
			n0, err = peer.network.ClosestPrecedingNode(successor, *id)
		}

		// return n0.find_successor(id);
		if err == nil {
			successor, err = peer.network.FindSuccessor(n0, *id)
			if err != nil {
				logger.Error("successor's FindSuccessor call failed: %v", err)
				return
			}
			info = successor
		}
		logger.Debug("returning: %v", info)
	}

	return
}

func (peer *Peer) ClosestPrecedingNode(ctx context.Context, id *NodeID) (info *ContactInfo, err error) {
	logger.Debug("ClosestPrecedingNode to: %s", id.String())

	for i := fingerCount - 1; i >= 0; i-- {
		finger := peer.network.fingerTable.fingers[i]
		if finger != nil && finger.Id.Between(peer.Info.Id, *id) {
			info = finger
			return
		}
	}

	info = peer.Info

	return
}

func (peer *Peer) Predecessor(ctx context.Context) (info *ContactInfo, err error) {
	logger.Debug("Predecessor")
	info = peer.GetPredecessor()

	return
}

func (peer *Peer) Successor(ctx context.Context) (info *ContactInfo, err error) {
	logger.Debug("Successor")
	info = peer.GetSuccessor()

	return
}

func (peer *Peer) Notify(ctx context.Context, sender *ContactInfo) (err error) {
	logger.Debug("Notify: %s", sender.Address)

	if peer.network.predecessor == nil || sender.Id.Between(peer.network.predecessor.Id, sender.Id) {
		peer.network.predecessor = sender
		peer.Poke()
	}

	return
}