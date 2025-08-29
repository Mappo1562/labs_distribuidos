package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Michael/proto/grpc-server/proto"
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const container = "localhost:50051"

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPruebaClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "dame un atraco lester, por favor"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("respuesta: %s", r.GetOferta())
}
