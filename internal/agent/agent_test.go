package agent_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	api "github.com/jscottransom/distributed_godis/api"
	"github.com/jscottransom/distributed_godis/internal/agent"
	"github.com/jscottransom/distributed_godis/internal/config"
)

func TestAgent(t *testing.T) {
	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: config.ServerCertFile,
		KeyFile: config.ServerKeyFile,
		CAFile: config.CAFile,
		Server: true,
		ServerAddress: "127.0.0.1",
	})

	require.NoError(t,err)

	peerTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: config.RootClientCertFile,
		KeyFile: config.RootClientKeyFile,
		CAFile: config.CAFile,
		Server: false,
		ServerAddress: "127.0.0.1",
	})

	require.NoError(t,err)

	var agents []*agent.Agent
	for i := 0; i < 3; i++ {
		ports := dynaport.Get(2)
		bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
		rpcPort := ports[1]

		t.Logf("Selected Ports: %d, %d", ports[0], rpcPort)

		dataDir, err := os.MkdirTemp("", "agent-test-log")
		t.Logf("%s", dataDir)
		storeName := "KV_store"
		require.NoError(t, err)

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(startJoinAddrs, agents[0].Config.BindAddr,
			)
		}

		if i == 0 {

			agent, err := agent.New(agent.Config{
				NodeName: fmt.Sprintf("%d", i),
				StartJoinAddrs: startJoinAddrs,
				BindAddr: bindAddr,
				RPCPort: rpcPort,
				DataDir: dataDir,
				StoreName: storeName,
				ACLModelFile: config.ACLModelFile,
				ACLPolicyFile: config.ACLPolicyFile,
				ServerTLSConfig: serverTLSConfig,
				PeerTLSConfig: peerTLSConfig,
			})
			require.NoError(t, err)
			agents = append(agents, agent)

			leaderClient := client(t, agents[0], peerTLSConfig)
			t.Logf("Start Join Addrs: %s, BindAddr: %s, RPCPort: %d, NodeName: %s", agents[0].StartJoinAddrs, agents[0].BindAddr, agents[0].RPCPort, agents[0].NodeName)
			setResponse, err := leaderClient.SetKey(
							context.Background(),
							&api.SetRequest{
								Key: "hello",
								Value: []byte("world"),
							},
							)
			require.NoError(t, err)
			t.Log(setResponse)

			setResponse2, err := leaderClient.SetKey(
				context.Background(),
				&api.SetRequest{
					Key: "strange",
					Value: []byte("fruit"),
				},
				)
			require.NoError(t, err)
			t.Log(setResponse2)
			getResponse, err := leaderClient.GetKey(
							context.Background(),
							&api.GetRequest{
								Key: "hello",
							},
							)
			require.NoError(t, err)
			require.Equal(t, getResponse.Value, []byte("world"))

			apiKeyList, err := leaderClient.ListKeys(context.Background(), &api.ListRequest{})
			var keyList []string
			for _, k := range apiKeyList.Key {
				keyList = append(keyList, k)
			}

			t.Logf("List repsonse is %v", keyList)


			time.Sleep(3 * time.Second)
		} else {

			agent, err := agent.New(agent.Config{
				NodeName: fmt.Sprintf("%d", i),
				StartJoinAddrs: startJoinAddrs,
				BindAddr: bindAddr,
				RPCPort: rpcPort,
				DataDir: dataDir,
				StoreName: storeName,
				ACLModelFile: config.ACLModelFile,
				ACLPolicyFile: config.ACLPolicyFile,
				ServerTLSConfig: serverTLSConfig,
				PeerTLSConfig: peerTLSConfig,
			})
			agents = append(agents, agent)
			require.NoError(t, err)

		}
		time.Sleep(3 * time.Second)
		
		
	}
	defer func() {
		for _, agent := range agents {
			err := agent.Shutdown()
			require.NoError(t, err)
			require.NoError(t,
				os.RemoveAll(agent.Config.DataDir),
			)
		}
	}()
	time.Sleep(3 * time.Second)

	followerClient := client(t, agents[1], peerTLSConfig)
	getResponseFollower, err := followerClient.GetKey(
		context.Background(),
		&api.GetRequest{
			Key: "hello",
		},
	)
	t.Logf("Follower Get repsonse is %s and %s", getResponseFollower.Key, getResponseFollower.Value)
	require.NoError(t, err)
	require.Equal(t, getResponseFollower.Value, []byte("world"))

}


func client(
	t *testing.T,
	agent *agent.Agent,
	tlsConfig *tls.Config,
) api.GodisServiceClient {

	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
	rpcAddr, err := agent.Config.RPCAddr()
	require.NoError(t, err)
	conn, err := grpc.Dial(fmt.Sprintf(
		"%s",
		rpcAddr,
	), opts...)
	require.NoError(t, err)
	client := api.NewGodisServiceClient(conn)
	return client
}