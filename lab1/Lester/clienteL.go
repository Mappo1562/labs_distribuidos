package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Lester/proto/grpc-server/proto"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"
)

const Michael = "M_container:50052"

const port = ":50051"

type server struct {
	pb.UnimplementedPruebaServer // se define el servidor de prueba.proto
}

var contador int = 0

func (s *server) SolicitarOferta(ctx context.Context, in *pb.OperationRequest) (*pb.OperationResponse, error) {
	// falta alguna algo q resetee el contador cuando michael acepte la oferta
	if contador > 2 {
		contador = 0
		time.Sleep(10 * time.Second)
	}
	if rand.Float64() > 0.9 {
		return &pb.OperationResponse{Oferta: "No hay ofertas actualmente..."}, nil
	}
	contador += 1
	exitoTrevor := rand.Float64()
	exitoFranklin := rand.Float64()
	riesgo := rand.Float64()
	botin := rand.Intn(100000)
	// extras :v
	lugares := [5]string{"Liberty city", "Los santos", "San Andreas", "Vice City", "Cayo Perico"}
	objetivo := [4]string{"un banco", "una joyeria", "una mansión", "un barco"}

	msg := fmt.Sprintf("Encontré una opción, es en %v y se trata de %v, el botín esperado sería %v, pero hay un riesgo asociado de %v, si va trevor las probabilidades de exito son %v y si va franklin son %v", lugares[rand.Intn(5)], objetivo[rand.Intn(4)], fmt.Sprintf("%d", botin), fmt.Sprintf("%.2f", riesgo), fmt.Sprintf("%.2f", exitoTrevor), fmt.Sprintf("%.2f", exitoFranklin))

	return &pb.OperationResponse{Oferta: msg, ExitoTrevor: exitoTrevor, ExitoFranklin: exitoFranklin, Riesgo: riesgo, Botin: int64(botin)}, nil
}

func aumento_estrellas() {
	// RabitMQ pa cualquiera
}

func recibir_pago() {

}

func verificar_pago() {

}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPruebaServer(grpcServer, &server{})
	fmt.Println("server en ", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}
}
