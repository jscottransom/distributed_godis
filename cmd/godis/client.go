package main

import (
	"context"
	"fmt"
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
	request := &api.GetRequest{Key: "This is from the client"}
								// Value: value}
	response, err := client.GetKey(context.Background(), request)
	if err != nil {
		log.Fatalf("Error when calling SetRequest: %s", err)
	}
	fmt.Printf("Response from server: %s\n", response)
	

}