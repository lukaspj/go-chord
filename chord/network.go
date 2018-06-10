package chord

import (
	"net/rpc"
	"math/big"
	"time"
)

type chordNetwork struct {
	fingerTable   fingerTable
	successors    successorList
	predecessor   *ContactInfo
	localInfo     *ContactInfo
	lastDirtyTime time.Time
}

func NewChordNetwork(info *ContactInfo) (network chordNetwork) {
	network.fingerTable = fingerTable{
		fingers: [10]ContactInfo{},
		next:    0,
	}
	network.localInfo = info

	return
}

func (network *chordNetwork) Call(contact ContactInfo, method string, args, reply interface{}) (err error) {
	logger.Info("%s -> %s", method, contact.Address)
	var client *rpc.Client
	if client, err = rpc.DialHTTP("tcp", contact.Address); err == nil {
		err = client.Call(method, args, reply)
		client.Close()
	}
	return
}

func (network *chordNetwork) NewRpcHeader(id NodeID) RPCHeader {
	return RPCHeader{
		Sender:     *network.localInfo,
		ReceiverId: id,
	}
}

func (network *chordNetwork) Ping(address string) (info ContactInfo, err error) {
	info = ContactInfo{
		Address: address,
		Id:      NewEmptyNodeID(),
	}

	args := PingRequest{
		RPCHeader: network.NewRpcHeader(NewEmptyNodeID()),
	}

	reply := PingResponse{}

	err = network.Call(info, "ChordApi.Ping", &args, &reply)
	info = reply.Sender
	return
}

func (network *chordNetwork) FindSuccessor(info ContactInfo, id NodeID) (res ContactInfo, err error) {
	request := FindSuccessorRequest{
		RPCHeader: network.NewRpcHeader(info.Id),
		Id:        id,
	}

	response := FindSuccessorResponse{}

	err = network.Call(info, "ChordApi.FindSuccessor", request, &response)

	res = response.Info
	return
}

func (network *chordNetwork) ClosestPrecedingNode(info ContactInfo, id NodeID) (res ContactInfo, err error) {
	request := ClosestPrecedingNodeRequest{
		RPCHeader: network.NewRpcHeader(info.Id),
		Id:        id,
	}

	response := ClosestPrecedingNodeResponse{}

	err = network.Call(info, "ChordApi.ClosestPrecedingNode", &request, &response)
	res = response.Info
	return
}

func (network *chordNetwork) Predecessor(info ContactInfo) (res *ContactInfo, err error) {
	request := PredecessorRequest{
		RPCHeader: network.NewRpcHeader(info.Id),
	}

	response := PredecessorResponse{}

	err = network.Call(info, "ChordApi.Predecessor", &request, &response)
	res = response.Info
	return
}

func (network *chordNetwork) Successor(info ContactInfo) (res ContactInfo, err error) {
	request := SuccessorRequest{
		RPCHeader: network.NewRpcHeader(info.Id),
	}

	response := SuccessorResponse{}

	err = network.Call(info, "ChordApi.Successor", &request, &response)
	res = response.Info
	return
}

func (network *chordNetwork) Notify(info ContactInfo) (err error) {
	request := NotifyRequest{
		RPCHeader: network.NewRpcHeader(info.Id),
	}

	response := NotifyResponse{}

	err = network.Call(info, "ChordApi.Notify", &request, &response)
	return
}

func (network *chordNetwork) Stabilize() (err error) {
	var x *ContactInfo

	// Update succlist
	for _, succ := range network.successors{
		if succ.Id.IsZero() {
			continue
		}

		succ, err = network.Ping(succ.Address)
		if err != nil {
			logger.Error("unresponsive successor, trying to rebuild successorlist from the next successor")
			continue
		}

		// Found a stable successor, build list
		dirty := false
		dirty = network.successors.SetSuccessor(0, succ) || dirty

		var prev, curr ContactInfo
		for i := 1; i < len(network.successors); i++ {
			prev = network.successors.GetSuccessor(i - 1)
			curr, err = network.Successor(prev)
			dirty = network.successors.SetSuccessor(i, curr) || dirty
		}

		if dirty {
			network.lastDirtyTime = time.Now()
		}
		break
	}

	successor := network.successors.GetSuccessor(0)
	x, err = network.Predecessor(successor)

	if err != nil {
		logger.Error("an error happened during stabilize: %v", err)
	}

	if x != nil && x.Id.Equals(network.localInfo.Id) {
		// Nothing has changed
		return
	}

	if x != nil && x.Id.Between(network.localInfo.Id, successor.Id) {
		if network.successors.SetSuccessor(0, *x) {
			network.lastDirtyTime = time.Now()
		}
	}
	network.Notify(successor)
	return
}

func (network *chordNetwork) FixFingers() (err error) {
	network.fingerTable.next++
	if network.fingerTable.next >= fingerCount {
		network.fingerTable.next = 0
	}
	fingerId := NewEmptyNodeID()

	// TODO move to NodeID file
	// fingerId.Val = peer.Info.Id.Val + 2^(next-1) mod 2^(20*8)
	var a, b, e big.Int
	tmp := network.localInfo.Id.BigInt()
	a.Add(tmp,
		e.Lsh(big.NewInt(2), uint(network.fingerTable.next)))
	b.Exp(big.NewInt(2), big.NewInt(20*8), nil)

	fingerId.Val = tmp.Mod(&a, &b).Bytes()

	var successor ContactInfo
	successor, err = network.FindSuccessor(*network.localInfo, fingerId)

	if network.fingerTable.SetFinger(network.fingerTable.next, successor) {
		network.lastDirtyTime = time.Now()
	}
	return
}

func (network *chordNetwork) CheckPredecessor() (err error) {
	if network.predecessor != nil {
		if _, err = network.Ping(network.predecessor.Address); err != nil {
			logger.Warn("Connection to predecessor has been lost")
			network.predecessor = nil
			network.lastDirtyTime = time.Now()
		}
	}
	return
}

func (network *chordNetwork) TimeSinceChange() time.Duration {
	return time.Now().Sub(network.lastDirtyTime)
}
