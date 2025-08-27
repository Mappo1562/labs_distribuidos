package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	"context"
	"fmt"
	"net"

	pb "Lester/proto/grpc-server/proto"
	"log"
	"time"

	"google.golang.org/grpc"
)

const Michael = "M_container:50052"

type server struct {
	pb.UnimplementedPruebaServer // se define el servidor de prueba.proto
}

func (s *server) solicitar_oferta(ctx context.Context, req *pb.OperationRequest) error {
	conn, err := grpc.Dial(Michael, grpc.WithInsecure()) // abrir el socket
	if err != nil {
		log.Printf("conexión fallida con michael D:\n %v", err)
		return err
	}
	defer conn.Close()
	log.Printf("creo que recibí\n %v", conn)
	return nil
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
	lis, err := net.Listen("tcp", "50051...")
	if err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}

	grpcServer := grpc.NewServer()
	fmt.Println("server en 50051..")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}
}
