package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Franklin/proto/grpc-server/proto"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

const (
	MichaelAddr = "M_container:50052"
	port        = ":50053"
	Rabbit      = "rabbitmq:50056"
)

type server struct {
	pb.UnimplementedDistraccionServer
	pb.UnimplementedGolpeServer
}

//////////////////////////////////////////////////
//												//
//  ███████╗░█████╗░░██████╗███████╗  ██████╗░  //
//  ██╔════╝██╔══██╗██╔════╝██╔════╝  ╚════██╗  //
//  █████╗░░███████║╚█████╗░█████╗░░  ░░███╔═╝  //
//  ██╔══╝░░██╔══██║░╚═══██╗██╔══╝░░  ██╔══╝░░  //
//  ██║░░░░░██║░░██║██████╔╝███████╗  ███████╗  //
//  ╚═╝░░░░░╚═╝░░╚═╝╚═════╝░╚══════╝  ╚══════╝  //
//												//
//////////////////////////////////////////////////

func (s *server) InicioDistraccion(ctx context.Context, in *pb.Instruccion) (*pb.ResultadoDistraccion, error) {
	var turnos int64 = in.GetNumTurnos()
	log.Printf("Tengo que trabajar %v", turnos)
	for i := 0; i < int(turnos); i++ {
		if i == int(turnos/2) {
			var exito bool = FracasoDistraccion()
			if !exito {
				return &pb.ResultadoDistraccion{
					ExitoDistraccion: "Chop ladró y Franklin perdió la concentración, aborta la misión", Exito: false}, nil
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

// Conexion RabbitMQ
func connectWithRetry(uri string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	const maxRetries = 10
	const delay = 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			log.Println("Conexión a RabbitMQ exitosa!")
			return conn, nil
		}
		log.Printf("Error de conexión a RabbitMQ (intento %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}

	return nil, err
}

func (s *server) InicioGolpe(ctx context.Context, in *pb.Instruccion) (*pb.ResultadoGolpe, error) {
	///////////////
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
		"my_queue", // nombre de la cola
		false,      // durable
		false,      // auto-delete cuando no se usa
		false,      // exclusiva
		false,      // no-wait
		nil,        // argumentos
	)
	if err != nil {
		log.Fatalf("No se pudo declarar la cola: %v", err)
	}
	msgs, err := ch.Consume(
		q.Name, // cola
		"",     // consumidor
		true,   // auto-ack
		false,  // exclusivo
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("No se pudo registrar un consumidor: %v", err)
	}
	///////////////
	var turnos int = int(in.GetNumTurnos())
	var limiteEstrellas int = 5
	var chopBonus int64 = 0
	var resultadoGolpe bool = true
	var estrellas int = 0
	for i := 0; i < int(turnos); i++ {
		//////////////////////////////////////////////////
		d := <-msgs
		if len(d.Body) > 0 {
			log.Printf("respuesta obtenida: %s \n", d.Body)
			estrellas++
		}
		//////////////////////////////////////////////////

		if estrellas > limiteEstrellas {
			resultadoGolpe = false
			return &pb.ResultadoGolpe{ExitoGolpe: resultadoGolpe, BotinExtra: chopBonus}, nil
		}
		if estrellas > 3 {
			chopBonus += 1000
		}
	}
	return &pb.ResultadoGolpe{ExitoGolpe: resultadoGolpe, BotinExtra: 0}, nil
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
	var pagoReal int64 = botinTotal / int64(4)
	if pagoRecibido == pagoReal {
		return &pb.ConfirmacionPago{Confirma: true}, nil
	} else {
		return &pb.ConfirmacionPago{Confirma: false}, nil
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDistraccionServer(grpcServer, &server{})
	pb.RegisterGolpeServer(grpcServer, &server{})
	fmt.Println("server en ", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("conexión fallida:\n %v", err)
	}
}
