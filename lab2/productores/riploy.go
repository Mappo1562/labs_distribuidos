package main

import (
	"context"
	pb "example/example_goproto/pb"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
)

func cordinationProcess(client pb.CordinationServiceClient, idNode int32) {
	var value int32 = 123
	response, err := client.Cordination(context.Background(), &pb.Request{Message: value})
	fmt.Println("[°] Mensaje enviado.")
	if err != nil {
		fmt.Println("[°] El cliente falló ", err)
		return
	}
	fmt.Println("[°] Mensaje recibido del servidor:", response.Message)

}

func calculateValue(value1 int32, value2 int32) {
	conn, err := grpc.Dial("localhost:50100", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("[°] El cliente no se pudo conectar %v", err)
		return
	}
	defer conn.Close()
	client := pb.NewCordinationServiceClient(conn)
	cordinationProcess(client, 1)
}

func main() {
	time.Sleep(5 * time.Second)
	fmt.Println("[°] Proceso operativo ")
	calculateValue(5, 8)

}
