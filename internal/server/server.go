package server

import (
	"fmt"
	"log"
	"net"
	api "github.com/jscottransom/distributed_godis/api"
	store "github.com/jscottransom/distributed_godis/internal/kvstore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type grpcServer struct {
	api.UnimplementedGodisServiceServer
	store *store.KVstore
}

// Build a new grpc server
func newgrpcServer(store *store.KVstore) (*grpcServer, error) {
	return &grpcServer{store: store}, nil
}

// Initialize a new GRPC server at the given port
func InitGRPCServer(port string, path string) (*grpc.Server, error) {
	
	// Create a listener on a port for incoming TCP connections
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return nil, err
	}

	kvstore, err := store.NewKVstore(path)
	if err != nil {
		log.Fatalf("failed to create store: %s", err)
	}
	grpcsrv := grpc.NewServer()
	srv, err := newgrpcServer(kvstore)
	if err != nil {
		return nil, err
	}

	api.RegisterGodisServiceServer(grpcsrv, srv)

	if err := grpcsrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}


	return grpcsrv, nil

}


func (s *grpcServer) SetKey(ctx context.Context, req *api.SetRequest) (*api.SetResponse, error) {

	// Set the key in the store
	if err := s.store.Set(req.Key, req.Value); err != nil {
		fmt.Printf("Unable to set key: %s", req.Key)
		return nil, err
	}
	fmt.Printf("Setting key: %s to value -> %s", req.Key, req.Value)
	return &api.SetResponse{Key: req.Key}, nil

}
func (s *grpcServer) GetKey(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	
	// Getthe key in the store
	val, err := s.store.Get(req.Key);
	if err != nil {
		fmt.Printf("Unable to get key: %s", req.Key)
		return nil, err
	}
	fmt.Printf("Getting the value for the following key: %s: value -> %s", req.Key, val)

	response_val, err := structpb.NewValue(val)
	if err != nil {
		fmt.Printf("Unable to get key: %s", req.Key)
		return nil, err
	}
	return &api.GetResponse{Key: req.Key, Value: response_val}, nil
	
}