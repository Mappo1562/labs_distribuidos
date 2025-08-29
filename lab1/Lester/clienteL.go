package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	"context"
	"flag"
	"fmt"
	"net"

	pb "Lester/proto/grpc-server/proto"
	"log"
	"time"

	"google.golang.org/grpc"
)

const Michael = "M_container:50052"

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedPruebaServer // se define el servidor de prueba.proto
}

func (s *server) SolicitarOferta(ctx context.Context, in *pb.OperationRequest) (*pb.OperationResponse, error) {
	log.Printf("creo que recibí\n %v", in.GetSolicitudOferta())
	return &pb.OperationResponse{Oferta: "no hay ofertas actualmente"}, nil
}

func espera() {
	time.Sleep(10 * time.Second)
}

func aumento_estrellas() {
	// RabitMQ pa cualquiera
}

func recibir_pago() {

}

func verificar_pago() {

}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("conexión fallida1:\n %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPruebaServer(grpcServer, &server{})
	fmt.Println("server en 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("conexión fallida2:\n %v", err)
	}
}
