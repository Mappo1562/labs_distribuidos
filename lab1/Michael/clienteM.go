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
const FranklinAddr = "localhost:50053"
const TrevorAddr = "localhost:50054"

func procesarOferta(respuesta *pb.OperationResponse) bool {
	var flag bool = true

	//Revisión campos vacios
	if respuesta.Oferta == nil {
		flag = false
	}
	if respuesta.ExitoFranklin == nil {
		flag = false
	}
	if respuesta.ExitoTrevor == nil {
		flag = false
	}
	if respuesta.Riesgo == nil {
		flag = false
	}
	if respuesta.Botin == nil {
		flag = false
	}

	//Revisión condición de aceptación
	if *respuesta.Riesgo > 0.8 || (*respuesta.ExitoFranklin < 0.5 && *respuesta.ExitoTrevor < 0.5) {
		flag = false
	}

	return flag
}
func main() {
	// Set up a connection to the server.
	conn, err := grpc.NewClient(LesterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPruebaClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//Fase La negociación
	//var ofertaValida bool = false
	respuestaOferta, err := c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "Dame un atraco lester, por favor"})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("respuesta: %s", respuestaOferta.GetOferta())

	ofertaValida := procesarOferta(respuestaOferta)

	if ofertaValida {
		c.AceptarOferta(ctx, &pb.Vacio{})
	}

	for !ofertaValida {
		respuestaOferta, err = c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "Dame otro atraco lester, por favor"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		ofertaValida = procesarOferta(respuestaOferta)
		if ofertaValida {
			c.AceptarOferta(ctx, &pb.Vacio{})
		}
		log.Printf("respuesta: %s", respuestaOferta.GetOferta())
	}

	//Fase La Distracción
	var mandarGolpe bool

	if respuestaOferta.GetExitoFranklin() > respuestaOferta.GetExitoTrevor() {
		mandarGolpe = true
	} else {
		mandarGolpe = false
	}

	if mandarGolpe {
		log.Printf("Mando a franklin a la primera parte")
		// Set up a connection to the server.
		conn, err := grpc.NewClient(FranklinAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c2 := pb.NewDistraccionClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		turnos := 200 - respuestaOferta.GetExitoFranklin()*100
		respuestaDistraccion, err := c2.InicioDistraccion(ctx, &pb.Instruccion{NumTurnos: int64(turnos)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Franklin dice: %s", respuestaDistraccion.GetExitoDistraccion())

	} else {
		log.Printf("Mando a trevor a la primera parte")
		// Set up a connection to the server.
		conn, err := grpc.NewClient(TrevorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c3 := pb.NewDistraccionClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		turnos := 200 - respuestaOferta.GetExitoTrevor()*100
		respuestaDistraccion, err := c3.InicioDistraccion(ctx, &pb.Instruccion{NumTurnos: int64(turnos)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Trevor dice: %s", respuestaDistraccion.GetExitoDistraccion())
	}

	//Fase El Golpe
	if !mandarGolpe {
		log.Printf("Mando a franklin a la segunda parte")
	} else {
		log.Printf("Mando a trevor a la segunda parte")
	}
}
