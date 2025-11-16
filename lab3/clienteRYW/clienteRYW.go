package main

import (
	pb "clienteRYW/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	coordinadorAddr = "localhost:50060"
)

func main() {

	conn, _ := grpc.NewClient(coordinadorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	coordinadorClient := pb.NewCoordinadorRYWClient(conn)
	_ = coordinadorClient
}
