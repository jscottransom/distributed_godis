package main

import (
	"context"
	"fmt"
	"strings"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	api "github.com/jscottransom/distributed_godis/api"
)

func main(){

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	cc, err := grpc.Dial(":9001", opts)

	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}

	defer cc.Close()
	
	client := api.NewGodisServiceClient(cc)

	// value := []byte("nah")
	// request := &api.SetRequest{Key: "client",
	// 							Value: value}
	// response, err := client.SetKey(context.Background(), request)
	// if err != nil {
	// 	log.Fatalf("Error when calling SetRequest: %s", err)
	// }

	// fmt.Printf("Response from server: %s\n", response)

	// newvalue := []byte("nah2")
	// newrequest := &api.SetRequest{Key: "Test",
	// 							Value: newvalue}
	// newresponse, err := client.SetKey(context.Background(), newrequest)
	// if err != nil {
	// 	log.Fatalf("Error when calling SetRequest: %s", err)
	// }

	// fmt.Printf("Response from server: %s\n", newresponse)

	listRequest := &api.ListRequest{}
	listResponse, err := client.ListKeys(context.Background(), listRequest)
	if err != nil {
		log.Fatalf("Error when calling ListKeys: %s", err)
	}

		
	
	s := strings.Split(listResponse.Key[0], ",")
	for _, k := range s {
		fmt.Printf("Response from server: %s\n", k)
	}

}