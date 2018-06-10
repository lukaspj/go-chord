package chord

import (
	"net/rpc"
	"net"
	"net/http"
	"fmt"
	"errors"
	"time"
)

const gearDownPeriod = time.Minute
const stabilizationIntervalStart = time.Second
const stabilizationIntervalEnd = 5 * time.Minute
const fixFingersIntervalStart = time.Second
const fixFingersIntervalEnd = 5 * time.Minute

type ContactInfo struct {
	Address string
	Id      NodeID
	Payload []byte
}

type Peer struct {
	Info                     ContactInfo
	Port                     int
	network                  chordNetwork
	stabilizationFunction    tickingFunction
	fixFingersFunction       tickingFunction
	checkPredecessorFunction tickingFunction
}

func NewPeer(info ContactInfo, port int) (peer Peer) {
	logger.Info("Creating new peer, with id: %s", info.Id.String())
	peer.Port = port
	peer.Info = info
	peer.network = NewChordNetwork(&peer.Info)
	return
}

func (peer *Peer) HandleRPC(request, response *RPCHeader) error {
	if !request.ReceiverId.IsZero() && !request.ReceiverId.Equals(peer.Info.Id) {
		return errors.New(fmt.Sprintf("Expected network ID %s, got %s",
			peer.Info.Id, request.ReceiverId))
	}

	response.Sender = peer.Info
	response.ReceiverId = request.Sender.Id
	return nil
}

func (peer *Peer) Listen() {
	logger.Info("Listening on port: %d", peer.Port)

	peer.network.predecessor = nil
	if peer.network.successors.SetSuccessor(0, peer.Info) {
		peer.network.lastDirtyTime = time.Now()
	}

	rpc.Register(&ChordApi{peer})

	rpc.HandleHTTP()
	if l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", peer.Port)); err == nil {
		go http.Serve(l, nil)
	}

	peer.stabilizationFunction = StartTickingFunction(func() int {
		interpolate := cubic(float64(stabilizationIntervalStart), float64(stabilizationIntervalEnd))
		peer.network.Stabilize()
		return int(interpolate(float64(peer.network.TimeSinceChange()) / float64(gearDownPeriod)))
	})

	peer.fixFingersFunction = StartTickingFunction(func() int {
		interpolate := cubic(float64(fixFingersIntervalStart), float64(fixFingersIntervalEnd))
		peer.network.FixFingers()
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
	var info ContactInfo
	logger.Info("Connecting to: %s", address)
	if info, err = peer.network.Ping(address); err == nil {
		logger.Info("Connection successful, remote peer is: %s", info.Id.String())
		var successor ContactInfo
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

func (peer *Peer) FindSuccessor(id NodeID) (info ContactInfo, err error) {
	successor := peer.network.successors.GetSuccessor(0)

	// if (id ∈ (n, successor] )
	if id.Between(peer.Info.Id, successor.Id) {
		// return successor;
		info = successor
	} else {
		// forward the query around the circle
		// n0 = successor.closest_preceding_node(id);
		var n0 ContactInfo
		n0, err = peer.network.ClosestPrecedingNode(successor, id)

		// return n0.find_successor(id);
		if err == nil {
			info, err = peer.network.FindSuccessor(n0, id)
		}
	}
	return
}

func (peer *Peer) GetSuccessor() (info ContactInfo) {
	info = peer.network.successors.GetSuccessor(0)
	return
}

func (peer *Peer) Poke() {
	go func() { peer.stabilizationFunction.tick <- true }()
	go func() { peer.fixFingersFunction.tick <- true }()
}
