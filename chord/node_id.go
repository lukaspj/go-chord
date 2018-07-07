package chord

import (
	"encoding/hex"
	"math/rand"
	"crypto/sha256"
	"math/big"
	"github.com/lukaspj/go-chord/api"
)

const IdLength = 20

type NodeID struct {
	Val []byte `json:"val"`
}

func NewNodeID(id *api.Id) (ret NodeID) {
	if id.Id != "" {
		ret = NewNodeIDFromHash(id.Id)
	} else {
		ret = NewNodeIDFromString(id.Hash)
	}
	return
}

func NewNodeIDFromString(id string) (ret NodeID) {
	decoded, err := hex.DecodeString(id)

	if err != nil {
		panic(err)
	}

	ret.Val = decoded
	return
}

func NewNodeIDFromHash(data string) (ret NodeID) {
	hashbytes := sha256.Sum256([]byte(data))
	ret.Val = hashbytes[:]
	return
}

func NewNodeIDFromAPI(id *api.NodeId) NodeID {
	return NodeID{
		Val: id.Val,
	}
}

func NewRandomNodeID() (ret NodeID) {
	bytes := make([]byte, IdLength)
	for i := 0; i < IdLength; i++ {
		bytes[i] = uint8(rand.Intn(256))
	}
	ret.Val = bytes
	return
}

func NewEmptyNodeID() (ret NodeID) {
	return
}

func (node NodeID) ToAPI() *api.NodeId {
	return &api.NodeId{
		Val: node.Val,
	}
}

func (node NodeID) BigInt() *big.Int {
	return big.NewInt(0).SetBytes(node.Val)
}

func (node NodeID) String() string {
	return node.BigInt().Text(16)
}

func (node NodeID) Equals(other NodeID) bool {
	return node.BigInt().Cmp(other.BigInt()) == 0
}

func (node NodeID) Less(other interface{}) bool {
	return node.BigInt().Cmp(other.(NodeID).BigInt()) == -1
}

func (node NodeID) Xor(other NodeID) (ret NodeID) {
	ret.Val = ret.BigInt().Xor(node.BigInt(), other.BigInt()).Bytes()
	return
}

func (node NodeID) PrefixLen() int {
	tmp := node.BigInt()
	for i := 0; i < tmp.BitLen(); i++ {
		if tmp.Bit(i) != 0 {
			return i
		}
	}
	return tmp.BitLen()
}

func (node NodeID) IsZero() bool {
	if len(node.Val) == 0 {
		return true
	}
	tmp := node.BigInt()
	return tmp.IsInt64() && tmp.Int64() == 0
}

//  n \in (a, b]
// Assuming b is successor to a
func (node NodeID) Between(a, b NodeID) bool {
	return a.Equals(b) || // Handle full-circle case
		node.Equals(b) || // Equality, handle b]
		(a.Less(b) && !node.Less(a) && node.Less(b)) || // Trivially between a and b
		(b.Less(a) && !node.Less(a) || node.Less(b)) // Handle wrap-around case
}
