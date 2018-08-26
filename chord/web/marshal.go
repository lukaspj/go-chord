package web

import (
	"github.com/lukaspj/go-chord/chord"
	"github.com/lukaspj/go-chord/api"
)

func ContactInfoToAPI(ci *chord.ContactInfo) *api.ContactInfo {
	return &api.ContactInfo{
		Address: ci.Address,
		Id: ci.Id.ToAPI(),
		Payload: ci.Payload,
	}
}

func NewContactInfoFromAPI(info *api.ContactInfo) *chord.ContactInfo {
	if info.Id == nil {
		return nil
	}

	nid := NewNodeIDFromAPI(info.Id)
	return &chord.ContactInfo{
		Address: info.Address,
		Id: *nid,
		Payload: info.Payload,
	}
}

func NodeIDToAPI(node *chord.NodeID) *api.NodeId {
	return &api.NodeId{
		Val: node.Val,
	}
}

func NewNodeIDFromAPI(id *api.NodeId) *chord.NodeID {
	if id.Val == nil {
		return nil
	}

	return &chord.NodeID{
		Val: id.Val,
	}
}

func NewNodeIDFromAPIId(id *api.Id) *chord.NodeID {
	if id.Id == "" && id.Hash == "" {
		return nil
	}

	var ret chord.NodeID
	if id.Id != "" {
		ret = chord.NewNodeIDFromHash(id.Id)
	} else {
		ret = chord.NewNodeIDFromString(id.Hash)
	}
	return &ret
}