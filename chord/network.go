package chord

import (
	"time"
	"math/big"
	"google.golang.org/grpc"
	"context"
)

type chordNetwork struct {
	fingerTable   fingerTable
	successors    successorList
	predecessor   *ContactInfo
	localInfo     *ContactInfo
	lastDirtyTime time.Time
}

func NewChordNetwork(info *ContactInfo) (network *chordNetwork) {
	network = &chordNetwork{
		fingerTable:   fingerTable{
			fingers: [10]*ContactInfo{},
			next:    0,
		},
		localInfo:     info,
	}

	for i := range network.successors {
		network.successors.SetSuccessor(i, info)
	}

	return
}

func (network *chordNetwork) Call(contact *ContactInfo, cb func(client ChordClient) error) (err error) {
	conn, err := grpc.Dial(contact.Address, grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		logger.Error("error communicating with grpc server [%s]: %v", contact.Address, err)
		return
	}
	client := NewChordClient(conn)
	err = cb(client)

	return
}

func (network *chordNetwork) Ping(address string) (info *ContactInfo, err error) {
	info = &ContactInfo{
		Address: address,
		Id:      NewEmptyNodeID(),
	}

	network.Call(info, func(client ChordClient) error {
		info, err = client.Ping(context.Background())
		return err
	})

	return
}

func (network *chordNetwork) FindSuccessor(info *ContactInfo, id NodeID) (res *ContactInfo, err error) {
	network.Call(info, func(client ChordClient) error {
		res, err = client.FindSuccessor(context.Background(), id)
		return err
	})

	return
}

func (network *chordNetwork) ClosestPrecedingNode(info *ContactInfo, id NodeID) (res *ContactInfo, err error) {
	network.Call(info, func(client ChordClient) error {
		res, err = client.ClosestPrecedingNode(context.Background(), id)
		return err
	})
	return
}

func (network *chordNetwork) Predecessor(info *ContactInfo) (res *ContactInfo, err error) {
	network.Call(info, func(client ChordClient) error {
		res, err = client.Predecessor(context.Background())
		return err
	})
	return
}

func (network *chordNetwork) Successor(info *ContactInfo) (res *ContactInfo, err error) {
	network.Call(info, func(client ChordClient) error {
		res, err = client.Successor(context.Background())
		return err
	})
	return
}

func (network *chordNetwork) Notify(info *ContactInfo) (err error) {
	network.Call(info, func(client ChordClient) error {
		err = client.Notify(context.Background(), network.localInfo)
		return err
	})
	return
}

func (network *chordNetwork) Stabilize() (err error) {
	var x *ContactInfo

	network.UpdateSuccessorList()

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
		if network.successors.SetSuccessor(0, x) {
			network.lastDirtyTime = time.Now()
		}
	}
	network.Notify(successor)
	return
}

func (network *chordNetwork) UpdateSuccessorList() {
	var err error
	for i, succ := range network.successors {
		if succ == nil || succ.Id.IsZero() {
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

		var prev, curr *ContactInfo
		for j := i + 1; j < len(network.successors); j++ {
			prev = network.successors.GetSuccessor(j - 1)
			curr, err = network.Successor(prev)
			if err != nil {
				logger.Info("we lost the connection to successor %d while updating the successorlist: %v", j, err)
				curr = prev
			}
			dirty = network.successors.SetSuccessor(j, curr) || dirty
		}

		if dirty {
			network.lastDirtyTime = time.Now()
		}
		break
	}
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

	var successor *ContactInfo
	successor, err = network.FindSuccessor(network.localInfo, fingerId)

	if err == nil {
		if network.fingerTable.SetFinger(network.fingerTable.next, successor) {
			network.lastDirtyTime = time.Now()
		}
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
