package main

import (
	"fmt"
	server "github.com/jscottransom/distributed_godis/internal/server"
)

func main() {
	srv, err := server.InitGRPCServer(":0", "/godis", 1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(srv.GetServiceInfo())
}