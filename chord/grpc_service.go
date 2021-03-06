package chord

import (
	"context"
	"fmt"
	"github.com/lukaspj/go-chord/api"
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
	return ContactInfoToAPI(c), err
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
	return ContactInfoToAPI(c), err
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
	return ContactInfoToAPI(c), err
}

func (w *ServiceWrapper) Predecessor(ctx context.Context, v *api.Void) (*api.ContactInfo, error) {
	c, err := w.service.Predecessor(ctx)
	if c == nil {
		return &api.ContactInfo{}, err
	}
	return ContactInfoToAPI(c), err
}

func (w *ServiceWrapper) Successor(ctx context.Context, v *api.Void) (*api.ContactInfo, error) {
	c, err := w.service.Successor(ctx)
	if c == nil {
		return &api.ContactInfo{}, err
	}
	return ContactInfoToAPI(c), err
}

func (w *ServiceWrapper) Notify(ctx context.Context, ci *api.ContactInfo) (*api.Void, error) {
	return &api.Void{},  w.service.Notify(ctx, NewContactInfoFromAPI(ci))
}