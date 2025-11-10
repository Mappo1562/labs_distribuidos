package main

import (
	"fmt"
	"log"
	"net"
	"time"

	pb "Nodo1/proto"

	"google.golang.org/grpc"
)

const (
	PortNodo1  = "50051"
	PortNodo2  = "db2:50052"
	PortNodo3  = "db3:50053"
	PortBroker = "broker:50050"
)

func init() {

}

func main() {
	init()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNodoConsensoServer(grpcServer, &server{})
	fmt.Println("Nodo con consenso activo en el puerto ", port)

	go func() {
		time.Sleep(60 * time.Second)
		fmt.Println("Tiempo cumplido. Deteniendo servidor...")
		grpcServer.GracefulStop()
	}()
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}

}
