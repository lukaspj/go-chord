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

func (ci *ContactInfo) ToAPI() *api.ContactInfo {
	return &api.ContactInfo{
		Address: ci.Address,
		Id: ci.Id.ToAPI(),
		Payload: ci.Payload,
	}
}

func NewContactInfoFromAPI(info *api.ContactInfo) *ContactInfo {
	if info.Id == nil {
		return nil
	}
	return &ContactInfo{
		Address: info.Address,
		Id: NewNodeIDFromAPI(info.Id),
		Payload: info.Payload,
	}
}

type Peer struct {
	Info                     *ContactInfo
	Port                     int
	network                  *chordNetwork
	stabilizationFunction    tickingFunction
	fixFingersFunction       tickingFunction
	checkPredecessorFunction tickingFunction
}

func (peer *Peer) Ping(ctx context.Context, void *api.Void) (info *api.ContactInfo, err error) {
	logger.Debug("Ping")
	info = peer.Info.ToAPI()
	if info == nil {
		logger.Error("Error")
	}
	return
}

func (peer *Peer) FindSuccessor(ctx context.Context, in_id *api.Id) (info *api.ContactInfo, err error) {
	successor := peer.network.successors.GetSuccessor(0)

	id := NewNodeID(in_id)

	logger.Debug("FindSuccessor to: %s", id.String())

	// if (id ∈ (n, successor] )
	if id.Between(peer.Info.Id, successor.Id) {
		// return successor;
		info = successor.ToAPI()
	} else {
		// forward the query around the circle
		// n0 = successor.closest_preceding_node(id);
		var n0 *ContactInfo
		n0, err = peer.network.ClosestPrecedingNode(successor, id)

		// return n0.find_successor(id);
		if err == nil {
			successor, err = peer.network.FindSuccessor(n0, id)
			info = successor.ToAPI()
		}
	}

	return
}

func (peer *Peer) ClosestPrecedingNode(ctx context.Context, in_id *api.Id) (info *api.ContactInfo, err error) {
	id := NewNodeID(in_id)
	logger.Debug("ClosestPrecedingNode to: %s", id.String())

	for i := fingerCount - 1; i >= 0; i-- {
		finger := peer.network.fingerTable.fingers[i]
		if finger.Id.Between(peer.Info.Id, id) {
			info = finger.ToAPI()
			return
		}
	}

	return
}

func (peer *Peer) Predecessor(ctx context.Context, void *api.Void) (info *api.ContactInfo, err error) {
	logger.Debug("Predecessor")
	predecessor := peer.GetPredecessor()
	if predecessor != nil {
		info = peer.GetPredecessor().ToAPI()
	} else {
		info = &api.ContactInfo{}
	}

	return
}

func (peer *Peer) Successor(ctx context.Context, void *api.Void) (info *api.ContactInfo, err error) {
	logger.Debug("Successor")
	successor := peer.GetSuccessor()
	if successor != nil {
		info = peer.GetSuccessor().ToAPI()
	} else {
		info = &api.ContactInfo{}
	}

	return
}

func (peer *Peer) Notify(ctx context.Context, in_sender *api.ContactInfo) (ret *api.Void, err error) {
	sender := NewContactInfoFromAPI(in_sender)

	logger.Debug("Notify: %s", sender.Address)

	if peer.network.predecessor == nil || sender.Id.Between(peer.network.predecessor.Id, sender.Id) {
		peer.network.predecessor = sender
		peer.Poke()
	}
	ret = &api.Void{}

	return
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
	api.RegisterChordServer(grpcServer, peer)

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
