package web

import (
	"google.golang.org/grpc"
	"context"
	"github.com/lukaspj/go-chord/api"
	"github.com/lukaspj/go-chord/chord"
)

type ChordClient struct {
	api api.ChordClient
}

func (client *ChordClient) Ping(ctx context.Context, opts ...grpc.CallOption) (*chord.ContactInfo, error) {
	ci, err := client.api.Ping(context.Background(), &api.Void{}, opts...)
	return NewContactInfoFromAPI(ci), err
}

func (client *ChordClient) FindSuccessor(ctx context.Context, in *chord.NodeID, opts ...grpc.CallOption) (*chord.ContactInfo, error) {
	ci, err := client.api.FindSuccessor(ctx, &api.Id{Hash: in.String()}, opts...)
	return NewContactInfoFromAPI(ci), err
}

func (client *ChordClient) ClosestPrecedingNode(ctx context.Context, in *chord.NodeID, opts ...grpc.CallOption) (*chord.ContactInfo, error) {
	ci, err := client.api.ClosestPrecedingNode(ctx, &api.Id{Hash: in.String()}, opts...)
	return NewContactInfoFromAPI(ci), err
}

func (client *ChordClient) Predecessor(ctx context.Context, opts ...grpc.CallOption) (*chord.ContactInfo, error) {
	ci, err := client.api.Predecessor(ctx, &api.Void{}, opts...)
	return NewContactInfoFromAPI(ci), err
}

func (client *ChordClient) Successor(ctx context.Context, opts ...grpc.CallOption) (*chord.ContactInfo, error) {
	ci, err := client.api.Successor(ctx, &api.Void{}, opts...)
	return NewContactInfoFromAPI(ci), err
}

func (client *ChordClient) Notify(ctx context.Context, in *chord.ContactInfo, opts ...grpc.CallOption) (error) {
	_, err := client.api.Notify(ctx, ContactInfoToAPI(in), opts...)
	return err
}

func NewChordClient(cc *grpc.ClientConn) ChordClient {
	return ChordClient{
		api: api.NewChordClient(cc),
	}
}