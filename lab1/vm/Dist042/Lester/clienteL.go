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
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

const (
	entrada = "ofertas_grande.csv"
	Rabbit  = "rabbitmq:5672"
	port    = ":50051"
)

type server struct {
	pb.UnimplementedPruebaServer
	pb.UnimplementedEstrellasServer
	pb.UnimplementedPagoBotinServer
}

var (
	contador   int = 0
	scanner    *bufio.Scanner
	riesgo     *int64
	estrellear bool = true
	mu         sync.Mutex
)

//////////////////////////////////////////////////
//												//
//  ███████╗░█████╗░░██████╗███████╗  ░░███╗░░  //
//  ██╔════╝██╔══██╗██╔════╝██╔════╝  ░████║░░  //
//  █████╗░░███████║╚█████╗░█████╗░░  ██╔██║░░  //
//  ██╔══╝░░██╔══██║░╚═══██╗██╔══╝░░  ╚═╝██║░░  //
//  ██║░░░░░██║░░██║██████╔╝███████╗  ███████╗  //
//  ╚═╝░░░░░╚═╝░░╚═╝╚═════╝░╚══════╝  ╚══════╝  //
//												//
//////////////////////////////////////////////////

func parseIntOpt(s string) *int64 {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &val
}

func formatInt64(p *int64) string {
	if p == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", *p)
}

func (s *server) SolicitarOferta(ctx context.Context, in *pb.OperationRequest) (*pb.OperationResponse, error) {
	contador += 1
	if contador%3 == 0 && contador != 0 {
		log.Printf("voy a buscar, espera...")
		time.Sleep(10 * time.Second)
	}
	if rand.Float64() > 0.9 {
		msg := "No hay ofertas actualmente... \n"
		log.Printf("%s", msg)
		return &pb.OperationResponse{Oferta: &msg}, nil
	}

	var (
		msg           string
		exitoTrevor   *int64
		exitoFranklin *int64
		botin         *int64
	)
	if scanner.Scan() {
		fmt.Println("Propuesta obtenida:", scanner.Text())
		arr := strings.Split(scanner.Text(), ",")
		botin = parseIntOpt(arr[0])
		exitoFranklin = parseIntOpt(arr[1])
		exitoTrevor = parseIntOpt(arr[2])
		riesgo = parseIntOpt(arr[3])
	} else {
		e := scanner.Err()
		if e != nil {
			fmt.Println("Error al leer el archivo:", e)
			msg = "Error al buscar atracos"
		} else {
			fmt.Println("Fin del archivo")
			msg = "No quedan atracos disponibles"
		}
		exitoTrevor = nil
		exitoFranklin = nil
		riesgo = nil
		botin = nil
	}
	// extras :v
	lugares := [5]string{"Liberty city", "Los santos", "San Andreas", "Vice City", "Cayo Perico"}
	objetivo := [4]string{"un banco", "una joyeria", "una mansión", "un barco"}

	log.Printf("Encontré una opción, es en %v y se trata de %v, el botín esperado sería %v, pero hay un riesgo asociado de %v, si va trevor las probabilidades de exito son %v y si va franklin son %v \n", lugares[rand.Intn(5)], objetivo[rand.Intn(4)], formatInt64(botin), formatInt64(riesgo), formatInt64(exitoTrevor), formatInt64(exitoFranklin))
	msg = fmt.Sprintf("Encontré una opción, es en %v y se trata de %v, el botín esperado sería %v, pero hay un riesgo asociado de %v, si va trevor las probabilidades de exito son %v y si va franklin son %v \n", lugares[rand.Intn(5)], objetivo[rand.Intn(4)], formatInt64(botin), formatInt64(riesgo), formatInt64(exitoTrevor), formatInt64(exitoFranklin))

	return &pb.OperationResponse{Oferta: &msg, ExitoTrevor: exitoTrevor, ExitoFranklin: exitoFranklin, Riesgo: riesgo, Botin: botin}, nil
}

func (s *server) AceptarOferta(ctx context.Context, in *pb.Vacio) (*pb.Vacio, error) {
	fmt.Println("oferta ", contador+1, " aceptada")
	contador = 0
	return &pb.Vacio{}, nil
}

//////////////////////////////////////////////////
//												//
//  ███████╗░█████╗░░██████╗███████╗  ██████╗░  //
//  ██╔════╝██╔══██╗██╔════╝██╔════╝  ╚════██╗  //
//  █████╗░░███████║╚█████╗░█████╗░░  ░█████╔╝  //
//  ██╔══╝░░██╔══██║░╚═══██╗██╔══╝░░  ░╚═══██╗  //
//  ██║░░░░░██║░░██║██████╔╝███████╗  ██████╔╝  //
//  ╚═╝░░░░░╚═╝░░╚═╝╚═════╝░╚══════╝  ╚═════╝░  //
//												//
//////////////////////////////////////////////////

func (s *server) TerminarMandarEstrellas(ctx context.Context, in *pb.Stars) (*pb.Stars, error) {
	mu.Lock()
	estrellear = false
	mu.Unlock()
	return &pb.Stars{Flag: true}, nil
}

func connectWithRetry(uri string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	const maxRetries = 10
	const delay = 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			log.Println("[°] Conexión a RabbitMQ exitosa!")
			return conn, nil
		}
		log.Printf("Error de conexión a RabbitMQ (intento %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}

	return nil, err
}

func (s *server) EmpezarMandarEstrellas(ctx context.Context, in *pb.Stars) (*pb.Stars, error) {
	log.Printf(" Aumento de estrellas activado ")
	frecuencia := 100 - int(*riesgo)
	// falta ver cual de los dos hace el atraco
	amqpURI := "amqp://guest:guest@" + Rabbit + "/"

	conn, err := connectWithRetry(amqpURI)
	if err != nil {
		log.Fatalf("Se excedió el número máximo de reintentos. No se pudo conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("No se pudo abrir un canal: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"my_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("No se pudo declarar la cola: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	flag := true
	i := 0
	for flag {
		time.Sleep(time.Nanosecond)
		mu.Lock()
		if !estrellear {
			mu.Unlock()
			break
		}
		mu.Unlock()
		if i%frecuencia == 0 && i != 0 {
			body := "Subiste 1 estrella, ten mas cuidado"
			err = ch.PublishWithContext(ctx,
				"",
				q.Name,
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				})
			if err != nil {
				log.Fatalf("No se pudo publicar el mensaje: %v", err)
			}
		}
		i = i + 1
	}
	return &pb.Stars{Flag: true}, nil
}

//////////////////////////////////////////////////
//												//
//  ███████╗░█████╗░░██████╗███████╗  ░░██╗██╗  //
//  ██╔════╝██╔══██╗██╔════╝██╔════╝  ░██╔╝██║  //
//  █████╗░░███████║╚█████╗░█████╗░░  ██╔╝░██║  //
//  ██╔══╝░░██╔══██║░╚═══██╗██╔══╝░░  ███████║  //
//  ██║░░░░░██║░░██║██████╔╝███████╗  ╚════██║  //
//  ╚═╝░░░░░╚═╝░░╚═╝╚═════╝░╚══════╝  ░░░░░╚═╝  //
//												//
//////////////////////////////////////////////////

func (s *server) ConfirmarPago(ctx context.Context, in *pb.Pago) (*pb.ConfirmacionPago, error) {
	var botinTotal int64 = in.BotinTotal
	var pagoRecibido int64 = in.Pago
	var pagoReal int64 = (botinTotal / int64(4)) + (botinTotal % int64(4))
	if pagoRecibido == pagoReal {
		return &pb.ConfirmacionPago{Confirma: true}, nil
	} else {
		return &pb.ConfirmacionPago{Confirma: false}, nil
	}
}

func main() {

	path := "entradas/" + entrada
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("no se pudo abrir el archivo:\n %v", err)
	}
	defer file.Close()

	scanner = bufio.NewScanner(file)

	scanner.Scan()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPruebaServer(grpcServer, &server{})
	pb.RegisterEstrellasServer(grpcServer, &server{})
	pb.RegisterPagoBotinServer(grpcServer, &server{})
	fmt.Println("server en ", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}
}
