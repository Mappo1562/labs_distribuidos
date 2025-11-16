package main

import (
	pb "clienteMR/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	brokerAddr = "localhost:50050"
)

func main() {
	conn, _ := grpc.NewClient(brokerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	brokerClient := pb.NewBrokerMRClient(conn)
	_ = brokerClient
}
