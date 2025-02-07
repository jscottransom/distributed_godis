package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	api "github.com/jscottransom/distributed_godis/api"
	kmap "github.com/jscottransom/distributed_godis/internal/keymap"
	store "github.com/jscottransom/distributed_godis/internal/kvstore"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Config struct {
	Store      *store.KVstore
	Keymap     kmap.SafeMap
	Authorizer Authorizer
}

const (
	objectWildCard = "*"
	setgetAction   = "setget"
	listAction     = "list"
)

type Authorizer interface {
	Authorize(subject, object, action string) error
}

type grpcServer struct {
	api.UnimplementedGodisServiceServer
	*Config
}

// Build a new grpc server
func newgrpcServer(config *Config) (*grpcServer, error) {
	return &grpcServer{
		Config: config}, nil
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {

	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{grpc_zap.WithDurationField(
		func(duration time.Duration) zapcore.Field {
			return zap.Int64(
				"grpc.time_ns",
				duration.Nanoseconds(),
			)
		},
	),
	}

	// TODO: Possibly consider updating this to trace only 50% of the time
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, err
	}
	opts = append(opts,
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(logger, zapOpts...),
				grpc_auth.StreamServerInterceptor(authenticate),
			)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
			grpc_auth.UnaryServerInterceptor(authenticate),
		)),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer(config)
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

	mapobj := make(kmap.KeyMap, 0)
	config := &Config{
		Store:  kvstore,
		Keymap: kmap.SafeMap{Map: mapobj},
	}

	gsrv, nil := NewGRPCServer(config)

	if err := gsrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	return gsrv, nil

}

func (s *grpcServer) SetKey(ctx context.Context, req *api.SetRequest) (*api.SetResponse, error) {

	if err := s.Authorizer.Authorize(subject(ctx), objectWildCard, setgetAction); err != nil {
		log.Printf("Error is %s{}\n", err)
		return nil, err
	}

	// Set the key in the store
	record := store.Record{Key: req.Key,
		Value: req.Value}

	lastOffset, err := s.Config.Store.Set(record)
	if err != nil {
		fmt.Printf("Unable to set key: %s", req.Key)
		return nil, err
	}

	// Get the number of bytes for the value
	valueLen := uint64(len(req.Value))

	// Update the key in the keymap, and save the map
	keyinfo := kmap.KeyInfo{Size: valueLen,
		Offset: uint64(lastOffset) - valueLen}
	s.Config.Keymap.Lock()
	defer s.Config.Keymap.Unlock()
	s.Config.Keymap.Map[record.Key] = &keyinfo
	s.Config.Keymap.SaveMap("keymap", 1)

	// Set the satisfactory message
	msg := "OK"
	return &api.SetResponse{Response: msg}, nil

}

func (s *grpcServer) GetKey(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	if err := s.Authorizer.Authorize(subject(ctx), objectWildCard, setgetAction); err != nil {
		log.Printf("Error is %s{}\n", err)
		return nil, err
	}

	s.Config.Keymap.RLock()
	defer s.Config.Keymap.RUnlock()
	s.Config.Keymap.LoadMap("keymap", 1)

	keyInfo := s.Config.Keymap.Map[req.Key]

	// Get the key in the store
	val, err := s.Config.Store.Get(keyInfo.Offset, keyInfo.Size)
	if err != nil {
		fmt.Printf("Unable to get key: %s", req.Key)
		return nil, err
	}

	return &api.GetResponse{
		Key: req.Key,
		Value: val}, nil

}

func (s *grpcServer) ListKeys(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	if err := s.Authorizer.Authorize(subject(ctx), objectWildCard, listAction); err != nil {
		log.Printf("Error is %s{}\n", err)
		return nil, err
	}

	keylist := []string{}

	// Iterate through the list of keys
	for k := range s.Config.Keymap.Map {
		keylist = append(keylist, k)
	}

	return &api.ListResponse{Key: keylist}, nil

}

func (s *grpcServer) SetStream(stream grpc.BidiStreamingServer[api.SetRequest, api.SetResponse]) error {

	var keysSet []string

	for {

		req := new(api.SetRequest)

		err := stream.RecvMsg(req)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Error while receiving message: %v", err)
			return status.Errorf(codes.Internal, "Failed to process stream: %v", err)
		}

		res, err := s.SetKey(stream.Context(), req)
		if err != nil {
			return err
		}
		fmt.Printf("Stored key: %s with value size: %d bytes", res.Response, len(req.Value))
		keysSet = append(keysSet, req.Key)

		fmt.Sprintf("Response %s", res)
		if err != nil {
			return err
		}
		err = stream.SendMsg(res)
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to send response: %v", err)
		}
	}
	return nil
}

func (s *grpcServer) GetStream(req *api.MultiGetRequest, stream grpc.ServerStreamingServer[api.GetResponse]) error {

	fmt.Sprintf("Streaming the following keys: %s", req.Keys)
	
	for _, key := range req.Keys {
		resp, err := s.GetKey(stream.Context(), &api.GetRequest{Key: key})
		
		
		if err != nil {
            return status.Errorf(codes.Internal, "Failed to retrieve key %s: %v", key, err)
        }
        
        // Send the response back to the client
        if err := stream.SendMsg(resp); err != nil {
            return status.Errorf(codes.Internal, "Failed to send message: %v", err)
        }
	}
	return nil
}

func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info").Err()

	}

	if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}

type subjectContextKey struct{}
