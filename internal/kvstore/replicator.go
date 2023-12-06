package kvstore

// import (
// 	"context"
// 	"sync"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc"
// 	api "github.com/jscottransom/distributed_godis/api"
// )

// type Replicator struct {
// 	DialOptions []grpc.DialOption
// 	LocalServer api.GodisServiceClient
// 	logger *zap.Logger
// 	mu sync.Mutex
// 	servers map[string] chan struct{}
// 	closed bool
// 	close chan struct{}
// }

// func (r *Replicator) Join(name, addr string) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	if r.closed {
// 		return nil
// 	}

// 	if _, ok := r.servers[name]; ok {
// 		return nil
// 	}

// 	r.servers[name] = make(chan struct{})
// 	go r.replicate(addr, r.servers[name])
	
// 	return nil
// }

// func (r *Replicator) replicate(addr string, leave chan struct{}) {
// 	cc, err := grpc.Dial(addr, r.DialOptions...)
// 	if err != nil {
// 		r.logError(err, "failed to dial", addr)
// 		return
// 	}
// 	defer cc.Close()

// 	client := api.NewGodisServiceClient(cc)
// 	ctx := context.Background()

// 	// Client (new server) obtains the Key Mapping
// 	request := &api.GetRequest{Key: key}
// 	keyRepl, err := client.GetKey(ctx, request)
	 
// }

