package server

import (
	"fmt"
	"log"
	"net"
	api "github.com/jscottransom/distributed_godis/api"
	store "github.com/jscottransom/distributed_godis/internal/kvstore"
	kmap "github.com/jscottransom/distributed_godis/internal/keymap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type grpcServer struct {
	api.UnimplementedGodisServiceServer
	store *store.KVstore
	keymap kmap.SafeMap
}



// Build a new grpc server
func newgrpcServer(store *store.KVstore) (*grpcServer, error) {
	
	mapobj := make(kmap.KeyMap, 0)
	return &grpcServer{store: store,
						keymap: kmap.SafeMap{Map: mapobj}}, nil
}

func NewGRPCServer(store *store.KVstore) (*grpc.Server, error) {
	gsrv := grpc.NewServer()
	srv, err := newgrpcServer(store)
	if err != nil {
		return nil, err
	}

	api.RegisterGodisServiceServer(gsrv, srv)
	return gsrv, nil
}

// Initialize a new GRPC server at the given port
func InitGRPCServer(port string, dir string, filename string, uid uint64) (*grpc.Server, error) {
	
	// Create a listener on a port for incoming TCP connections
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return nil, err
	}

	kvstore, err := store.NewKVstore(dir, filename)
	if err != nil {
		log.Fatalf("failed to create store: %s", err)
	}
	
	gsrv, nil := NewGRPCServer(kvstore)

	if err := gsrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	return gsrv, nil

}


func (s *grpcServer) SetKey(ctx context.Context, req *api.SetRequest) (*api.SetResponse, error) {

	// Set the key in the store
	record := store.Record{Key: req.Key,
						   Value: req.Value}
	
	
	lastOffset, err := s.store.Set(record)
	if err != nil {
		fmt.Printf("Unable to set key: %s", req.Key)
		return nil, err
	}

	// Get the number of bytes for the value
	valueLen := uint64(len(req.Value))
	
	// Update the key in the keymap, and save the map
	keyinfo := kmap.KeyInfo{Size: valueLen,
						   Offset: uint64(lastOffset) - valueLen}
	s.keymap.Lock()
	defer s.keymap.Unlock()
	s.keymap.Map[record.Key] = &keyinfo
	s.keymap.SaveMap("keymap", 1)

	// Set the satisfactory message
	msg := "OK"
	return &api.SetResponse{Response: msg}, nil

}
func (s *grpcServer) GetKey(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	s.keymap.RLock()
	defer s.keymap.RUnlock()
	s.keymap.LoadMap("keymap", 1)

	keyInfo := s.keymap.Map[req.Key]

	// Get the key in the store
	val, err := s.store.Get(keyInfo.Offset, keyInfo.Size);
	if err != nil {
		fmt.Printf("Unable to get key: %s", req.Key)
		return nil, err
	}

	return &api.GetResponse{Value: val}, nil
	
}

func (s *grpcServer) ListKeys(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	keylist := []string{}

	// Iterate through the list of keys
	for k := range s.keymap.Map {
		keylist = append(keylist, k)
	}

	return &api.ListResponse{Key: keylist}, nil
	
}