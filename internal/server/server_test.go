package server

import (
	"context"
	"flag"
	"net"
	"os"
	"testing"
	"time"
	"io"

	api "github.com/jscottransom/distributed_godis/api"
	"github.com/jscottransom/distributed_godis/internal/auth"
	"github.com/jscottransom/distributed_godis/internal/config"
	store "github.com/jscottransom/distributed_godis/internal/kvstore"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/examples/exporter"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var debug = flag.Bool("debug", false, "Enable observability for debugging.")

func TestMain(m *testing.M) {
	flag.Parse()
	if *debug {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		zap.ReplaceGlobals(logger)
	}
	os.Exit(m.Run())
}

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		rootClient api.GodisServiceClient,
		nobodyClient api.GodisServiceClient,
		config *Config,
	){
		"Set a Key in store succeeds":       testSetGetKey,
		"List all keys from store succeeds": testListKey,
		"Unauthorized Fails":                testUnauthorized,
		"Set and Get Stream": testSetGetStream,
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

	var telemetryExporter *exporter.LogExporter
	if *debug {
		metricsLogFile, err := os.CreateTemp("", "metrics-*.log")
		require.NoError(t, err)
		t.Logf("metrics log file: %s", metricsLogFile.Name())

		tracesLogFile, err := os.CreateTemp("", "traces-*.log")
		require.NoError(t, err)
		t.Logf("traces log file: %s", tracesLogFile.Name())

		telemetryExporter, err = exporter.NewLogExporter(exporter.Options{
			MetricsLogFile:    metricsLogFile.Name(),
			TracesLogFile:     tracesLogFile.Name(),
			ReportingInterval: time.Second,
		})
		require.NoError(t, err)
		err = telemetryExporter.Start()
		require.NoError(t, err)

	}
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
			KeyFile:  keyPath,
			CAFile:   config.CAFile,
			Server:   false,
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
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})

	require.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	dir, err := os.MkdirTemp(cwd, "server-test")
	require.NoError(t, err)

	t.Log(dir)

	kvstore, err := store.NewKVstore(dir, "testStore")
	require.NoError(t, err)

	cfg = &Config{Store: kvstore,
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
		if telemetryExporter != nil {
			time.Sleep(1500 * time.Millisecond)
			telemetryExporter.Stop()
			telemetryExporter.Close()
		}
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

func testSetGetStream(
	t *testing.T, client, _ api.GodisServiceClient, config *Config,
) {
	ctx := context.Background()

	requests := []*api.SetRequest{
		{Key: "hello", Value: []byte("world")},
		{Key: "strange", Value: []byte("fruits")},
	}
	stream, err := client.SetStream(ctx)
	require.NoError(t, err)

	for _, request := range requests {
		err = stream.SendMsg(request)
		require.NoError(t, err)
		
	}

	for i := 0; i < len(requests); i++ {
		resp, err := stream.Recv()
		require.NoError(t, err)
		t.Logf("%s", resp)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	
	keys := []string{"hello", "strange"}
	keysToGet := &api.MultiGetRequest{
		Keys: keys,
	}
	getStream, err := client.GetStream(context.Background(), keysToGet) 
	if err != nil {
		t.Fatalf("Error calling GetStream %v", err)
	}
	var responses []*api.GetResponse
	for  {
		
		resp, err := getStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Error receiving %v", err)
		}
		responses = append(responses, resp)
		require.NoError(t, err)
		t.Logf("%s", resp)
	}

	// require.Equal(t, len(keysToGet), len(responses), "The number of responses should match the number of requested keys")
	
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
