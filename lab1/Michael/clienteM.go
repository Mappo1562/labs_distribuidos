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
	"google.golang.org/grpc/credentials/insecure"
)

const LesterAddr = "localhost:50051"

func main() {
	// Set up a connection to the server.
	conn, err := grpc.NewClient(LesterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPruebaClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	r, err := c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "1dame un atraco lester, por favor"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("respuesta: %s", r.GetOferta())
	time.Sleep(time.Second)

	r, err = c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "2dame un atraco lester, por favor"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("respuesta: %s", r.GetOferta())
	time.Sleep(time.Second)

	r, err = c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "3dame un atraco lester, por favor"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("respuesta: %s", r.GetOferta())
	time.Sleep(time.Second)

	r, err = c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "4dame un atraco lester, por favor"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("respuesta: %s", r.GetOferta())

	c.AceptarOferta(ctx, &pb.Vacio{})

}
