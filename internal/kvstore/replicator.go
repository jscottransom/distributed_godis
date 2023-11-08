package kvstore

import (
	"context"
	"sync"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	api "github.com/jscottransom/distributed_godis/api"
)

type Replicator struct {
	DialOptions []grpc.DialOption
	LocalServer api.GodisServiceClient
	logger *zap.Logger
	mu sync.Mutex
	servers map[string] chan struct{}
	closed bool
	close chan struct{}
}