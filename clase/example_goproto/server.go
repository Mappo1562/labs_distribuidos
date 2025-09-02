package main

import (
	"context"
	pb "example/example_goproto/pb"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var (
	syncThreads sync.WaitGroup
	dataMutex  sync.Mutex 
	flag bool
)

type server struct {
	pb.UnimplementedCordinationServiceServer
}

func calValue(value int32)(int32) {
	return value*5
}

func (s *server) Cordination(ctx context.Context, req *pb.Request)(*pb.Response,error){	
	fmt.Println("	[°] Mensaje recibido: ",req.Message)
	result:= calValue(req.Message)
	fmt.Println("	[°] Mensaje enviado:  ",result)
	return &pb.Response{Message:result}, nil
}

func serverBackground(){
	defer syncThreads.Done()
	lis, err := net.Listen("tcp", ":50100")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCordinationServiceServer(s, &server{})
	fmt.Printf("[°] Server running on port 50100\n")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}


func main(){
	go serverBackground()
	time.Sleep(5*time.Second)
	for true {
		fmt.Println("[°] Proceso Operativo")
		time.Sleep(10*time.Second)
	}	
}