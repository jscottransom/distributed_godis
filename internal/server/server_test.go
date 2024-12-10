package server

import (
	"context"
	"net"
	"os"
	"testing"

	api "github.com/jscottransom/distributed_godis/api"
	"github.com/jscottransom/distributed_godis/internal/config"
	kmap "github.com/jscottransom/distributed_godis/internal/keymap"
	store "github.com/jscottransom/distributed_godis/internal/kvstore"
	"github.com/jscottransom/distributed_godis/internal/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		rootClient api.GodisServiceClient,
		nobodyClient api.GodisServiceClient,
		config *Config,
	){
		"Set a Key in store succeeds": testSetGetKey,
		"List all keys from store succeeds": testListKey,
		"Unauthorized Fails": testUnauthorized,
	} {
		t.Run(scenario, func(t *testing.T) {
			rootClient,
			nobodyClient,
			config,
			teardown := setupTest(t, nil)
			defer teardown()
			fn(t, rootClient, nobodyClient, config)
		})
	
	}
}

func setupTest(t *testing.T, fn func(*Config)) (
	rootClient api.GodisServiceClient,
	nobodyClient api.GodisServiceClient,
	cfg *Config,
	teardown func(),
) {
	
	authorizer := auth.New(config.ACLModelFile, config.ACLPolicyFile)
	// Marks as a test heelper function
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	newClient := func(crtPath, keyPath string) (
		*grpc.ClientConn,
		api.GodisServiceClient,
		[]grpc.DialOption,
	) {
		tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
			CertFile: crtPath,
			KeyFile: keyPath,
			CAFile: config.CAFile,
			Server: false,
		})

		require.NoError(t, err)

		tlsCreds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
		conn, err := grpc.Dial(l.Addr().String(), opts...)
		require.NoError(t, err)
		client := api.NewGodisServiceClient(conn)
		return conn, client, opts
	}

	var rootConn *grpc.ClientConn
	rootConn, rootClient, _ = newClient(
		config.RootClientCertFile,
		config.RootClientKeyFile,
	)

	var nobodyConn *grpc.ClientConn
	nobodyConn, nobodyClient, _ = newClient(
		config.NobodyClientCertFile,
		config.NobodyClientKeyFile,
	)

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: config.ServerCertFile,
		KeyFile: config.ServerKeyFile,
		CAFile: config.CAFile,
		ServerAddress: l.Addr().String(),
		Server: true,
	})

	require.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)
	cwd, err:= os.Getwd()
	require.NoError(t, err)
	dir, err := os.MkdirTemp(cwd, "server-test")
	require.NoError(t, err)

	t.Log(dir)
	

	kvstore, err := store.NewKVstore(dir, "testStore")
	mapobj := make(kmap.KeyMap, 0)
	require.NoError(t, err)

	cfg = &Config{Store: kvstore,
				  Keymap: kmap.SafeMap{Map:mapobj},
				Authorizer: authorizer}

	if fn != nil {
		fn(cfg)
	}

	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()
	return rootClient, nobodyClient, cfg, func() {
		server.Stop()
		rootConn.Close()
		nobodyConn.Close()
		l.Close()
		kvstore.Remove(dir)
	}
}

func testSetGetKey(t *testing.T, client, _ api.GodisServiceClient, config *Config) {
	ctx := context.Background()

	want := &api.SetRequest{Key: "hello", Value: []byte("world")}
	_, err := client.SetKey(ctx, &api.SetRequest{Key: "hello",
								 Value: []byte("world")})
	
	require.NoError(t, err)
	get, err := client.GetKey(ctx, &api.GetRequest{Key: "hello"})

	require.NoError(t, err)
	require.Equal(t, want.Value, get.Value)

}

func testListKey(t *testing.T, client, _ api.GodisServiceClient, config *Config) {
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

func testUnauthorized(t *testing.T, _, client api.GodisServiceClient, config *Config) {

	ctx := context.Background()
	set, err := client.SetKey(ctx, &api.SetRequest{Key: "hello",
								 Value: []byte("world")})

	if set != nil {
		t.Fatalf("Set response should be nil")
	}

	gotCode, wantCode := status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}

	get, err := client.GetKey(ctx, &api.GetRequest{Key: "hello"})
	if get != nil {
		t.Fatalf("Get response should be nil")
	}
	gotCode, wantCode = status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got cod: %d, want: %d", gotCode, wantCode)
	}
}