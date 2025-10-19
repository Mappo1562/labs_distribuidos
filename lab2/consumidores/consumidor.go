package main

import (
	pb "consumidores/proto"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	conn, err := grpc.NewClient("broker:50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pueeo conectar: %v", err)
	}

	defer conn.Close()

	cliente := pb.NewBrokerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	registro := &pb.Registro{
		Nombre: "C1-1",
		Rol:    1,
	}

	resp, err := cliente.Registrarse(ctx, registro)
	if err != nil {
		log.Fatalf("Error al registrarse: %v", err)
	}
	if resp.Flag {
		fmt.Println("C1-1 registrado en el broker")
	} else {
		fmt.Println("Falló el registro")
		return
	}
}
