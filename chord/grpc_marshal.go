package chord

import (
	"github.com/lukaspj/go-chord/api"
)

func ContactInfoToAPI(ci *ContactInfo) *api.ContactInfo {
	return &api.ContactInfo{
		Address: ci.Address,
		Id: NodeIDToAPI(&ci.Id),
		Payload: ci.Payload,
	}
}

func NewContactInfoFromAPI(info *api.ContactInfo) *ContactInfo {
	if info.Id == nil {
		return nil
	}

	nid := NewNodeIDFromAPI(info.Id)
	return &ContactInfo{
		Address: info.Address,
		Id: *nid,
		Payload: info.Payload,
	}
}

func NodeIDToAPI(node *NodeID) *api.NodeId {
	return &api.NodeId{
		Val: node.Val,
	}
}

func NewNodeIDFromAPI(id *api.NodeId) *NodeID {
	if id.Val == nil {
		return nil
	}

	return &NodeID{
		Val: id.Val,
	}
}

func NewNodeIDFromAPIId(id *api.Id) *NodeID {
	if id.Id == "" && id.Hash == "" {
		return nil
	}

	var ret NodeID
	if id.Id != "" {
		ret = NewNodeIDFromHash(id.Id)
	} else {
		ret = NewNodeIDFromString(id.Hash)
	}
	return &ret
}