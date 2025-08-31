package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Franklin/proto/grpc-server/proto"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

const Michael = "M_container:50052"

const port = ":50053"

type server struct {
	pb.UnimplementedDistraccionServer
}

func (s *server) InicioDistraccion(ctx context.Context, in *pb.Instruccion) (*pb.Resultado, error) {
	log.Printf("Tengo que trabajar %v", in.GetNumTurnos())
	return &pb.Resultado{
		ExitoDistraccion: "Consegui terminar, sigue con la siguiente fase", Exito: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDistraccionServer(grpcServer, &server{})
	fmt.Println("server en ", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}
}
