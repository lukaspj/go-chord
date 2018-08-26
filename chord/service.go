package chord

import (
	"context"
	"github.com/lukaspj/go-chord/api"
	"fmt"
)

type Service interface {
	Ping(ctx context.Context) (*ContactInfo, error)
	FindSuccessor(ctx context.Context, id *NodeID) (*ContactInfo, error)
	ClosestPrecedingNode(ctx context.Context, id *NodeID) (*ContactInfo, error)
	Predecessor(ctx context.Context) (*ContactInfo, error)
	Successor(ctx context.Context) (*ContactInfo, error)
	Notify(ctx context.Context, id *ContactInfo) error
}

type ServiceWrapper struct {
	service Service
}

func (w *ServiceWrapper) Ping(ctx context.Context, v *api.Void) (*api.ContactInfo, error) {
	c, err := w.service.Ping(ctx)
	if c == nil {
		return &api.ContactInfo{}, err
	}
	return c.ToAPI(), err
}

func (w *ServiceWrapper) FindSuccessor(ctx context.Context, id *api.Id) (*api.ContactInfo, error) {
	nid := NewNodeIDFromAPIId(id)
	if nid == nil {
		return &api.ContactInfo{}, fmt.Errorf("FindSuccessor id argument must not be nil.")
	}
	c, err := w.service.FindSuccessor(ctx, nid)
	if c == nil {
		return &api.ContactInfo{}, err
	}
	return c.ToAPI(), err
}

func (w *ServiceWrapper) ClosestPrecedingNode(ctx context.Context, id *api.Id) (*api.ContactInfo, error) {
	nid := NewNodeIDFromAPIId(id)
	if nid == nil {
		return &api.ContactInfo{}, fmt.Errorf("ClosestPrecedingNode id argument must not be nil.")
	}
	c, err := w.service.ClosestPrecedingNode(ctx, nid)
	if c == nil {
		return &api.ContactInfo{}, err
	}
	return c.ToAPI(), err
}

func (w *ServiceWrapper) Predecessor(ctx context.Context, v *api.Void) (*api.ContactInfo, error) {
	c, err := w.service.Predecessor(ctx)
	if c == nil {
		return &api.ContactInfo{}, err
	}
	return c.ToAPI(), err
}

func (w *ServiceWrapper) Successor(ctx context.Context, v *api.Void) (*api.ContactInfo, error) {
	c, err := w.service.Successor(ctx)
	if c == nil {
		return &api.ContactInfo{}, err
	}
	return c.ToAPI(), err
}

func (w *ServiceWrapper) Notify(ctx context.Context, ci *api.ContactInfo) (*api.Void, error) {
	return &api.Void{},  w.service.Notify(ctx, NewContactInfoFromAPI(ci))
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

	nid := NewNodeIDFromAPI(info.Id)
	return &ContactInfo{
		Address: info.Address,
		Id: *nid,
		Payload: info.Payload,
	}
}

func (node *NodeID) ToAPI() *api.NodeId {
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