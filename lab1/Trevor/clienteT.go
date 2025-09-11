package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Trevor/proto/grpc-server/proto"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"google.golang.org/grpc"
)

const MichaelAddr = "M_container:50052"

const port = ":50054"

type server struct {
	pb.UnimplementedDistraccionServer
}

func (s *server) InicioDistraccion(ctx context.Context, in *pb.Instruccion) (*pb.ResultadoDistraccion, error) {
	var turnos int64 = in.GetNumTurnos()
	log.Printf("Tengo que trabajar %v", turnos)
	for i := 0; i < int(turnos); i++ {
		if i == int(turnos/2) {
			var exito bool = FracasoDistraccion()
			if !exito {
				return &pb.ResultadoDistraccion{
					ExitoDistraccion: "Trevor se emborrachó antes del atraco, aborta la misión", Exito: false}, nil
			}
		}
	}
	return &pb.ResultadoDistraccion{
		ExitoDistraccion: "Consegui terminar, sigue con la siguiente fase", Exito: true}, nil
}

func FracasoDistraccion() bool {
	var ranInt int = rand.Intn(100)
	var exito bool = true
	if ranInt < 10 {
		exito = false
	}
	return exito
}

func InicioGolpe(ctx context.Context, in *pb.Instruccion) (*pb.ResultadoGolpe, error) {
	var turnos int = int(in.GetNumTurnos())
	var limiteEstrellas int = 5
	var furia bool = false
	var resultadoGolpe bool = true
	for i := 0; i < int(turnos); i++ {
		var estrellas int = getEstrellas()
		if estrellas > limiteEstrellas {
			if furia {
				resultadoGolpe = false
				break
			}
			limiteEstrellas = 7
		}
	}
	return &pb.ResultadoGolpe{ExitoGolpe: resultadoGolpe, BotinExtra: 0}, nil
}

func getEstrellas() int {
	return 1
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
