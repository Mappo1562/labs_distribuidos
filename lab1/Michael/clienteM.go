package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Michael/proto/grpc-server/proto"
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const LesterAddr = "lester:50051"
const FranklinAddr = "franklin:50053"
const TrevorAddr = "trevor:50054"

func procesarOferta(respuesta *pb.OperationResponse) bool {
	if respuesta.Oferta == nil || respuesta.ExitoFranklin == nil || respuesta.ExitoTrevor == nil || respuesta.Riesgo == nil || respuesta.Botin == nil {
		return false
	}

	if *respuesta.Riesgo > 80 || (*respuesta.ExitoFranklin < 50 && *respuesta.ExitoTrevor < 50) {
		return false
	}

	return true
}

func activar_estrellas() {
	conn, err := grpc.NewClient(LesterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	c := pb.NewEstrellasClient(conn)
	estrellasInicio, err := c.EmpezarMandarEstrellas(ctx, &pb.Stars{Flag: true})
	if err != nil {
		log.Fatalf("error en RPC: %v", err)
	}
	if !estrellasInicio.GetFlag() {
		log.Printf("No se pudo mandar estrellas")
	}
}

func newClientConn(addr string) (*grpc.ClientConn, context.Context, context.CancelFunc) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("no se pudo conectar a %s: %v", addr, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return conn, ctx, cancel
}

func main() {
	time.Sleep(time.Second * 3)
	// Set up a connection to the server.
	conn, err := grpc.NewClient(LesterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPruebaClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	//Fase La negociación
	//var ofertaValida bool = false
	log.Printf("Dame un atraco lester, por favor")
	respuestaOferta, err := c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "Dame un atraco lester, por favor"})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	// log.Printf("respuesta: %s", respuestaOferta.GetOferta())

	ofertaValida := procesarOferta(respuestaOferta)

	if ofertaValida {
		c.AceptarOferta(ctx, &pb.Vacio{})
	}

	for !ofertaValida {
		log.Printf("Dame otro atraco lester, por favor")
		time.Sleep(time.Second)
		respuestaOferta, err = c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "Dame otro atraco lester, por favor"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		ofertaValida = procesarOferta(respuestaOferta)
		if ofertaValida {
			log.Printf("Me quedo con ese")
			c.AceptarOferta(ctx, &pb.Vacio{})
		}
		// log.Printf("respuesta: %s", respuestaOferta.GetOferta())
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
		turnos := 200 - respuestaOferta.GetExitoFranklin()
		log.Printf("Tiene que trabajar estos turnos: %v", turnos)

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
		turnos := 200 - respuestaOferta.GetExitoTrevor()
		log.Printf("Tiene que trabajar estos turnos: %v", turnos)
		respuestaDistraccion, err := c3.InicioDistraccion(ctx, &pb.Instruccion{NumTurnos: int64(turnos)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Trevor dice: %s", respuestaDistraccion.GetExitoDistraccion())
	}

	//Fase El Golpe
	if !mandarGolpe {
		log.Printf("Mando a franklin a la segunda parte")

		connFranklin, ctxF, cancelF := newClientConn(FranklinAddr)
		defer connFranklin.Close()
		defer cancelF()
		golpeClient := pb.NewGolpeClient(connFranklin)

		// Cliente Lester
		connLester, ctxL, cancelL := newClientConn(LesterAddr)
		defer connLester.Close()
		defer cancelL()
		estrellasClient := pb.NewEstrellasClient(connLester)

		// Datos iniciales
		turnos := 200 - respuestaOferta.GetExitoFranklin()
		log.Printf("Franklin debe trabajar estos turnos: %v", turnos)

		// Variables compartidas
		var mu sync.Mutex
		cond := sync.NewCond(&mu)
		var golpeTerminado bool
		var respuestaGolpe *pb.ResultadoGolpe
		var errGolpe, errEstrellas error

		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine Franklin
		go func() {

			defer wg.Done()
			resp, err := golpeClient.InicioGolpe(ctxF, &pb.Instruccion{NumTurnos: int64(turnos)})
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errGolpe = err
			} else {
				respuestaGolpe = resp
				log.Printf("Franklin terminó: %v, botín extra: %v", respuestaGolpe.GetExitoGolpe(), respuestaGolpe.GetBotinExtra())
			}

			golpeTerminado = true
			cond.Broadcast() // avisamos a Lester
		}()

		// Goroutine Lester
		go func() {
			// Empieza inmediatamente
			defer wg.Done()
			inicio, err := estrellasClient.EmpezarMandarEstrellas(ctxL, &pb.Stars{Flag: true})
			if err != nil {
				errEstrellas = err
				return
			}
			if !inicio.GetFlag() {
				log.Printf("No se pudo empezar a mandar estrellas")
			}

			// Espera a que Franklin termine
			mu.Lock()
			for !golpeTerminado && errGolpe == nil {
				cond.Wait()
			}
			if errGolpe != nil {
				errEstrellas = errGolpe
				mu.Unlock()
				return
			}
			mu.Unlock()

			// Ahora termina
			fin, err := estrellasClient.TerminarMandarEstrellas(ctxL, &pb.Stars{Flag: true})
			if err != nil {
				errEstrellas = err
				return
			}
			log.Printf("Estrellas finalizadas: %v", fin.GetFlag())
		}()

		wg.Wait()

		if errGolpe != nil {
			log.Fatalf("Error en Franklin: %v", errGolpe)
		}

		if errEstrellas != nil {
			log.Fatalf("Erro en Lester: %v", errEstrellas)
		}

	} else {
		log.Printf("Mando a trevor a la segunda parte")

		// Set up a connection to the server.
		// conn, err := grpc.NewClient(TrevorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		// if err != nil {
		// 	log.Fatalf("did not connect: %v", err)
		//}
		// defer conn.Close()
		// c3 := pb.NewGolpeClient(conn)
		// Contact the server and print out its response.

		///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

		conn2, err := grpc.NewClient(TrevorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn2.Close()
		// Contact the server and print out its response.
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel2()
		turnos := 200 - respuestaOferta.GetExitoTrevor()
		log.Printf("Tiene que trabajar estos turnos: %v", turnos)

		defer conn.Close()
		c2 := pb.NewGolpeClient(conn2)

		// Contact the server and print out its response.

		go activar_estrellas()
		respuestaGolpe, err2 := c2.InicioGolpe(ctx2, &pb.Instruccion{NumTurnos: int64(turnos)})

		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		if err2 != nil {
			log.Fatalf("could not greet: %v", err2)
		}

		///

		conn, err := grpc.NewClient(LesterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		c := pb.NewEstrellasClient(conn)
		estrellasInicio, err := c.TerminarMandarEstrellas(ctx, &pb.Stars{Flag: true})
		if err != nil {
			log.Fatalf("error en RPC: %v", err)
		}
		if !estrellasInicio.GetFlag() {
			log.Printf("No se pudo mandar estrellas")
		}
		log.Printf("estrellas detenidas")

		/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

		// ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		// defer cancel()
		// turnos := 200 - respuestaOferta.GetExitoTrevor()
		// log.Printf("Tiene que trabajar estos turnos: %v", turnos)
		// respuestaGolpe, err := c3.InicioGolpe(ctx, &pb.Instruccion{NumTurnos: int64(turnos)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Trevor dice: %v", respuestaGolpe.GetExitoGolpe())
	}
}
