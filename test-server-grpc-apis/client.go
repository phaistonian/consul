package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/consul/proto-public/pbdataplane"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello I am the grpc client")
	clientConn, err := grpc.Dial("127.0.0.1:8502", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer clientConn.Close()

	c := pbdataplane.NewDataplaneServiceClient(clientConn)

	req := &pbdataplane.SupportedDataplaneFeaturesRequest{}

	res, err := c.SupportedDataplaneFeatures(context.Background(), req)
	if err != nil {
		log.Fatalf("Error! %s", err)
	}
	log.Println("Response", res)

}
