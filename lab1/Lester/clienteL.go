package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Lester/proto/grpc-server/proto"
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
)

const (
	entrada = "ofertas_pequeno.csv"
	Michael = "M_container:50052"
	port    = ":50051"
)

type server struct {
	pb.UnimplementedPruebaServer // se define el servidor de prueba.proto
}

var (
	contador int = 0
	file     *os.File
	scanner  *bufio.Scanner
)

func (s *server) SolicitarOferta(ctx context.Context, in *pb.OperationRequest) (*pb.OperationResponse, error) {
	if contador%3 == 0 && contador != 0 {
		time.Sleep(10 * time.Second)
	}
	if rand.Float64() > 0.9 {
		return &pb.OperationResponse{Oferta: "No hay ofertas actualmente... \n"}, nil
	}
	contador += 1

	var (
		msg           string
		exitoTrevor   int64
		exitoFranklin int64
		riesgo        int64
		botin         int64
		err           error
	)
	if scanner.Scan() {
		fmt.Println("Propuesta obtenida:", scanner.Text())
		arr := strings.Split(scanner.Text(), ",")
		exitoTrevor, err = strconv.ParseInt(arr[2], 10, 64)
		exitoFranklin, err = strconv.ParseInt(arr[1], 10, 64)
		riesgo, err = strconv.ParseInt(arr[3], 10, 64)
		botin, err = strconv.ParseInt(arr[0], 10, 64)
	} else {
		e := scanner.Err()
		if e != nil {
			fmt.Println("Error al leer el archivo:", e)
			msg = "Error al buscar atracos"
		} else {
			fmt.Println("Fin del archivo")
			msg = "No quedan atracos disponibles"
		}
		exitoTrevor = 0
		exitoFranklin = 0
		riesgo = 0
		botin = 0
	}
	// extras :v
	lugares := [5]string{"Liberty city", "Los santos", "San Andreas", "Vice City", "Cayo Perico"}
	objetivo := [4]string{"un banco", "una joyeria", "una mansión", "un barco"}

	msg = fmt.Sprintf("Encontré una opción, es en %v y se trata de %v, el botín esperado sería %v, pero hay un riesgo asociado de %v, si va trevor las probabilidades de exito son %v y si va franklin son %v \n", lugares[rand.Intn(5)], objetivo[rand.Intn(4)], fmt.Sprintf("%d", botin), fmt.Sprintf("%.2f", riesgo), fmt.Sprintf("%.2f", exitoTrevor), fmt.Sprintf("%.2f", exitoFranklin))

	return &pb.OperationResponse{Oferta: msg, ExitoTrevor: exitoTrevor, ExitoFranklin: exitoFranklin, Riesgo: riesgo, Botin: int64(botin)}, nil
}

func (s *server) AceptarOferta(ctx context.Context, in *pb.Vacio) (*pb.Vacio, error) {
	fmt.Println("oferta ", contador+1, " aceptada")
	contador = 0
	return &pb.Vacio{}, nil
}

func aumento_estrellas() {
	// RabitMQ pa cualquiera
}

func recibir_pago() {

}

func verificar_pago() {

}

func main() {
	// abrir archivo
	path := "../entradas/" + entrada
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("no se pudo abrir el archivo:\n %v", err)
	}

	// iniciar scanner
	scanner = bufio.NewScanner(file)

	scanner.Scan()
	fmt.Println("Línea:", scanner.Text())

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
