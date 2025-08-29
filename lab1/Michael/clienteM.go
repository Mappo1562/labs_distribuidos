package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Michael/proto/grpc-server/proto"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

const container = "localhost:50051"

func main() {
	conn, err := grpc.Dial(container, grpc.WithInsecure()) // conectar con el socket
	if err != nil {
		log.Fatalf("Error al conectar el servidor: %v", err)
	}
	defer conn.Close()

	client := pb.NewPruebaClient(conn) // hacer un cliente con la conexión creada

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := client.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "quiero un atraco"})
	if err != nil {
		log.Fatalf("Error llamando a SolicitarOferta: %v", err)
	}

	// Usar la respuesta
	log.Println("Respuesta del servidor:", resp.Oferta)
}
