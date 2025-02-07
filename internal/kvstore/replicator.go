package kvstore

import (
	"context"
	// "fmt"
	"sync"
	"io"
	api "github.com/jscottransom/distributed_godis/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Replicator struct {
	DialOptions []grpc.DialOption
	LocalServer api.GodisServiceClient
	logger      *zap.Logger
	mu          sync.Mutex
	servers     map[string]chan struct{}
	closed      bool
	close       chan struct{}
}

func (r *Replicator) init() {
	if r.logger == nil {
		r.logger = zap.L().Named("replicator")
	}
	if r.servers == nil {
		r.servers = make(map[string]chan struct{})
	}
	if r.close == nil {
		r.close = make(chan struct{})
	}
}

func (r *Replicator) logError(err error, msg, addr string) {
	r.logger.Error(
		msg,
		zap.String("addr", addr),
		zap.Error(err),
	)
}

func (r *Replicator) replicate(addr string, leave chan struct{}) {
	
	sugarlog := r.sugarLogger()
	cc, err := grpc.NewClient(addr, r.DialOptions...)
	sugarlog.Infof("Client Connection target is %s", addr)
	
	if err != nil {
		r.logError(err, "failed to dial", addr)
		return
	}
	defer cc.Close()

	ctx := context.Background()
	client := api.NewGodisServiceClient(cc)

	
	logger := r.sugarLogger()

	// Client gets a list of all the keys from the client,
	// and then sets the value based on get requests for those keys
	apiKeyList, err := client.ListKeys(ctx, &api.ListRequest{})
	

	var keyList []string
	for _, k := range apiKeyList.Key {
		keyList = append(keyList, k)
	}

	multiRequest := api.MultiGetRequest{Keys: keyList}
	stream, err := client.GetStream(ctx, &multiRequest)

	if err != nil {
		r.logError(err, "failed to Get", addr)
		return
	}

	valueList := make(chan *api.GetResponse)

	go func() {
		for {
			recv, err := stream.Recv()
			
			if err == io.EOF {
				break
			}

			if err != nil {
				logger.Info(err)
				r.logError(err, "Failed to Receive", addr)
				return
			}

			response := &api.GetResponse{
				Key:   recv.Key,
				Value: recv.Value,
			}
			valueList <- response
		}
	}()

	for {
		select {
		case <-r.close:
			return
		case <-leave:
			return
		case val := <-valueList:

			logger.Infof("Setting the following Key and Value: %s, %s", val.Key, val.Value)
			_, err = r.LocalServer.SetKey(ctx, &api.SetRequest{
				Key:   val.Key,
				Value: val.Value,
			})
			if err != nil {
				r.logError(err, "failed to Set Key", addr)
				return
			}
		}
	}

}

func (r *Replicator) Join(name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}

	if _, ok := r.servers[name]; ok {
		return nil
	}

	r.servers[name] = make(chan struct{})
	go r.replicate(addr, r.servers[name])

	return nil
}


func (r *Replicator) Leave(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	if _, ok := r.servers[name]; !ok {
		return nil
	}
	close(r.servers[name])
	delete(r.servers, name)
	return nil
}



func (r *Replicator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}
	r.closed = true
	close(r.close)
	return nil
}



func (r *Replicator) sugarLogger() *zap.SugaredLogger {
	tempLogger := zap.Must(zap.NewDevelopment())
	sugar := tempLogger.Sugar()

	return sugar
}
