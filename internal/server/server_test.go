package server 

import (
	"context"
	"io/ioutil"
	"net"
	"testing"
	"os"

	"github.com/stretchr/testify/require"
	api "github.com/jscottransom/distributed_godis/api"
	store "github.com/jscottransom/distributed_godis/internal/kvstore"
	// kmap "github.com/jscottransom/distributed_godis/internal/keymap"
	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client api.GodisServiceClient,
	){
		"Set a Key in store succeeds": testSetGetKey,
		// "Get a Key from store succeeds": testGetKey,
		"List all keys from store succeeds": testListKey,
		// "Get a key that doesn't exist": testFalseKey,
	} {
		t.Run(scenario, func(t *testing.T) {
				client, teardown := setupTest(t, nil)
				defer teardown()
				fn(t, client)
		})
	
	}
}

func setupTest(t *testing.T, fn func()) (
	client api.GodisServiceClient,
	teardown func(),
) {
	// Marks as a test heelper function
	t.Helper()

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	clientOptions := []grpc.DialOption{grpc.WithInsecure()}
	cc, err := grpc.Dial(l.Addr().String(), clientOptions...)
	require.NoError(t, err)


	cwd, err:= os.Getwd()
	require.NoError(t, err)
	dir, err := ioutil.TempDir(cwd, "server-test")
	require.NoError(t, err)

	t.Log(dir)
	

	kvstore, err := store.NewKVstore(dir, "testStore")
	require.NoError(t, err)

	if fn != nil {
		fn()
	}

	server, err := NewGRPCServer(kvstore)
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	client = api.NewGodisServiceClient(cc)

	return client, func() {
		server.Stop()
		cc.Close()
		l.Close()
		kvstore.Remove(dir)
	}
}

func testSetGetKey(t *testing.T, client api.GodisServiceClient) {
	ctx := context.Background()

	want := &api.SetRequest{Key: "hello", Value: []byte("world")}
	_, err := client.SetKey(ctx, &api.SetRequest{Key: "hello",
								 Value: []byte("world")})
	
	require.NoError(t, err)
	get, err := client.GetKey(ctx, &api.GetRequest{Key: "hello"})

	require.NoError(t, err)
	require.Equal(t, want.Value, get.Value)

}

func testListKey(t *testing.T, client api.GodisServiceClient) {
	ctx := context.Background()

	first := &api.SetRequest{Key: "hello", Value: []byte("world")}
	second := &api.SetRequest{Key: "strange", Value: []byte("fruits")}
	
	keyRef := make([]string, 2)

	keyRef = append(keyRef, first.Key)
	keyRef = append(keyRef, second.Key)

	_, err := client.SetKey(ctx, &api.SetRequest{Key: "hello",
								 Value: []byte("world")})
	
	require.NoError(t, err)

	res, err := client.SetKey(ctx, &api.SetRequest{Key: "strange",
								 Value: []byte("fruits")})
	
	require.NoError(t, err)
	t.Log(res)
	
	list, err := client.ListKeys(ctx, &api.ListRequest{})
	
	finalList := make([]string, 2)
	finalList = append(finalList, list.Key...)

	require.NoError(t, err)
	require.Equal(t, keyRef, finalList)

}