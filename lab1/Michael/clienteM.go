package main

// protoc --go_out=. --go-grpc_out=. .\prueba.proto
// go mod init <nombre_carpeta_cliente>
// go mod tidy
import (
	pb "Michael/proto/grpc-server/proto"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const LesterAddr = "lester:50051"
const FranklinAddr = "franklin:50053"
const TrevorAddr = "trevor:50054"

// //////////////////////////////////////////////////////////////////////////
// ProcesarOferta()
// //////////////////////////////////////////////////////////////////////////
// Esta funcion retorna un booleano, y se usa para revisar si la oferta
// será o no aceptada por Michael
// //////////////////////////////////////////////////////////////////////////

func procesarOferta(respuesta *pb.OperationResponse) bool {
	if respuesta.Oferta == nil || respuesta.ExitoFranklin == nil || respuesta.ExitoTrevor == nil || respuesta.Riesgo == nil || respuesta.Botin == nil {
		return false
	}

	if *respuesta.Riesgo > 80 || (*respuesta.ExitoFranklin < 50 && *respuesta.ExitoTrevor < 50) {
		return false
	}

	return true
}

////////////////////////////////////////////////////////////////////////////
// confirmarPago()
////////////////////////////////////////////////////////////////////////////
// Esta funcion retorna un booleano, y se usa para la fase 4, donde se
// reparte el botin a cada integrante
////////////////////////////////////////////////////////////////////////////

func confirmarPago(addr string, nombre string, botinTotal int64, pago int64) bool {

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to %s: %v", nombre, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c := pb.NewPagoBotinClient(conn)
	respuesta, err := c.ConfirmarPago(ctx, &pb.Pago{BotinTotal: botinTotal, Pago: pago})
	if err != nil {
		log.Fatalf("error en RPC con %s: %v", nombre, err)
	}

	if !respuesta.GetConfirma() {
		log.Printf("Le pague mal a %s!!! Oh nooo", nombre)
		return false
	} else {
		log.Printf("Pago confirmado por %s ", nombre)
	}
	return true
}

////////////////////////////////////////////////////////////////////////////
// generarReporte()
////////////////////////////////////////////////////////////////////////////
// Esta funcion retorna un booleano, y se usa para la generación del
// reporte, sea éxito o fracaso
////////////////////////////////////////////////////////////////////////////

func generarReporte(exito bool, botinBase int64, botinExtra int64, faseFallo string, motivoFallo string, confLes bool, confFran bool, confTre bool) bool {
	file, err := os.Create("/app/reportes/Reporte.txt")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	fmt.Fprintln(file, "=========================================================")
	fmt.Fprintln(file, "== REPORTE FINAL DE LA MISION ==")
	fmt.Fprintln(file, "=========================================================")
	//fmt.Fprintf(file, "Mision : Asalto al Banco # %d\n", misionID)

	botinTotal := botinBase + botinExtra

	if exito {
		pagoFranklin := botinTotal / 4
		pagoTrevor := pagoFranklin
		pagoLester := pagoTrevor

		var respuestaLes string
		var respuestaFran string
		var respuestaTre string

		if confLes {
			respuestaLes = "Un placer hacer negocios"
		} else {
			respuestaLes = "Me estafaste, no me busques para más golpes"
		}

		if confFran {
			respuestaFran = "El pago es el correcto, le voy a comprar algo a Chop"
		} else {
			respuestaFran = "Nunca más te ayudo con un trabajo"
		}

		if confTre {
			respuestaTre = "Buen pago"
		} else {
			respuestaTre = "No es la cantidad que me tocaba, la vas a pagar"
		}

		fmt.Fprintln(file, "Resultado Global : MISION COMPLETADA CON EXITO !")
		fmt.Fprintln(file, "--- REPARTO DEL BOTIN ---")
		fmt.Fprintf(file, "Botin Base : $%d\n", botinBase)
		fmt.Fprintf(file, "Botin Extra ( Habilidad de Chop ): $%d\n", botinExtra)
		fmt.Fprintf(file, "Botin Total : $%d\n", botinTotal)
		fmt.Fprintf(file, "------ ------------ ------------ ------------ ------------ --- \n")
		fmt.Fprintf(file, "Pago a Franklin : $%d\n", pagoFranklin)
		fmt.Fprintf(file, "Respuesta de Franklin : \"%s\" \n", respuestaFran)
		fmt.Fprintf(file, "Pago a Trevor : $%d\n", pagoTrevor)
		fmt.Fprintf(file, "Respuesta de Trevor : \"%s\"  \n", respuestaTre)
		fmt.Fprintf(file, "Pago a Lester : $%d (reparto) + $%d (resto)\n", pagoLester, botinTotal%4)
		fmt.Fprintf(file, "Respuesta de Lester : \"%s\"\n", respuestaLes)
		fmt.Fprintf(file, "------ ------------ ------------ ------------ ------------ --- \n")
		fmt.Fprintf(file, "Saldo Final de la Operacion : $%d\n", botinTotal)
	} else {
		fmt.Fprintln(file, "Resultado Global : MISION FALLIDA")
		fmt.Fprintf(file, "La misión terminó en: %s\n", faseFallo)
		fmt.Fprintf(file, "El motivo del fallo fue: %s\n", motivoFallo)
		fmt.Fprintf(file, "El botin perdido fue de: %d\n", botinTotal)
	}

	fmt.Fprintln(file, "=========================================================")

	return true
}

////////////////////////////////////////////////////////////////////////////
// activar_estrellas()
////////////////////////////////////////////////////////////////////////////
// Esta funcion esta hecha para activar a Lester y que comienze a mandar
// estrellas al involucrado en la fase 3, se llamará en una gorutine
////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////
// main()
////////////////////////////////////////////////////////////////////////////
// Esta funcion orquesta todo el proceso del laboratorio, desde la fase 1
// hasta la 4
////////////////////////////////////////////////////////////////////////////

func main() {
	log.Printf("Iniciando...")
	// Fase 1
	time.Sleep(time.Second * 3)

	conn, err := grpc.NewClient(LesterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPruebaClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	exitoMision := true

	log.Printf("Dame un atraco lester, por favor")
	respuestaOferta, err := c.SolicitarOferta(ctx, &pb.OperationRequest{SolicitudOferta: "Dame un atraco lester, por favor"})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

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

	}

	//Fase 2
	time.Sleep(time.Second * 3)
	var mandarGolpe bool
	botinMision := respuestaOferta.GetBotin()

	if respuestaOferta.GetExitoFranklin() > respuestaOferta.GetExitoTrevor() {
		mandarGolpe = true
	} else {
		mandarGolpe = false
	}

	var mensajeDistraccion string
	var distraccionExito bool

	if mandarGolpe {
		log.Printf("Mando a franklin a la primera parte")

		conn, err := grpc.NewClient(FranklinAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c2 := pb.NewDistraccionClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		turnos := 200 - respuestaOferta.GetExitoFranklin()
		log.Printf("Tiene que trabajar estos turnos: %v", turnos)

		respuestaDistraccion, err := c2.InicioDistraccion(ctx, &pb.Instruccion{NumTurnos: int64(turnos)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Franklin dice: %s", respuestaDistraccion.GetExitoDistraccion())
		mensajeDistraccion = respuestaDistraccion.GetExitoDistraccion()
		distraccionExito = respuestaDistraccion.GetExito()

	} else {
		log.Printf("Mando a trevor a la primera parte")

		conn, err := grpc.NewClient(TrevorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c3 := pb.NewDistraccionClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		turnos := 200 - respuestaOferta.GetExitoTrevor()
		log.Printf("Tiene que trabajar estos turnos: %v", turnos)
		respuestaDistraccion, err := c3.InicioDistraccion(ctx, &pb.Instruccion{NumTurnos: int64(turnos)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Trevor dice: %s", respuestaDistraccion.GetExitoDistraccion())
		mensajeDistraccion = respuestaDistraccion.GetExitoDistraccion()
		distraccionExito = respuestaDistraccion.GetExito()
	}

	if !distraccionExito {
		exitoMision = false
	}

	if !exitoMision {
		generarReporte(exitoMision, botinMision, 0, "Fase distracción", mensajeDistraccion, false, false, false)
		return
	}

	var exitoGolpe bool
	var botinAgregado int64

	//Fase 3
	time.Sleep(time.Second * 3)
	if !mandarGolpe {
		log.Printf("Mando a Franklin al Golpe")

		conn2, err := grpc.NewClient(FranklinAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn2.Close()

		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel2()
		turnos := 200 - respuestaOferta.GetExitoFranklin()
		log.Printf("Tiene que trabajar estos turnos: %v", turnos)

		defer conn.Close()
		c2 := pb.NewGolpeClient(conn2)

		go activar_estrellas()
		respuestaGolpe, err2 := c2.InicioGolpe(ctx2, &pb.Instruccion{NumTurnos: int64(turnos)})

		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		if err2 != nil {
			log.Fatalf("could not greet: %v", err2)
		}

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

		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		if respuestaGolpe.GetExitoGolpe() {
			log.Printf("Franklin trabajo bien, ganamos")
		} else {
			log.Printf("Franklin fracaso, perdimos")
		}
		exitoGolpe = respuestaGolpe.GetExitoGolpe()
		botinAgregado = respuestaGolpe.GetBotinExtra()

	} else {
		log.Printf("Mando a Trevor al Golpe")

		conn2, err := grpc.NewClient(TrevorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn2.Close()

		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel2()
		turnos := 200 - respuestaOferta.GetExitoTrevor()
		log.Printf("Tiene que trabajar estos turnos: %v", turnos)

		defer conn.Close()
		c2 := pb.NewGolpeClient(conn2)

		go activar_estrellas()
		respuestaGolpe, err2 := c2.InicioGolpe(ctx2, &pb.Instruccion{NumTurnos: int64(turnos)})

		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		if err2 != nil {
			log.Fatalf("could not greet: %v", err2)
		}

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

		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		if respuestaGolpe.GetExitoGolpe() {
			log.Printf("Trevor trabajo bien, ganamos")
		} else {
			log.Printf("Trevor fracaso, perdimos")
		}
		exitoGolpe = respuestaGolpe.GetExitoGolpe()
		botinAgregado = respuestaGolpe.GetBotinExtra()
	}

	if !exitoGolpe {
		exitoMision = false
	}

	if !exitoMision {
		generarReporte(exitoMision, botinMision, botinAgregado, "Fase Golpe", "Se llego al limite de estrellas", false, false, false)
		return
	}

	//Fase 4
	time.Sleep(time.Second * 3)
	botinTotal := botinMision + botinAgregado
	pagoLester := botinTotal/4 + botinTotal%4
	pagoFranklin := botinTotal / 4
	pagoTrevor := botinTotal / 4

	confirmacionLester := confirmarPago(LesterAddr, "Lester", botinTotal, pagoLester)
	confirmacionFranklin := confirmarPago(FranklinAddr, "Franklin", botinTotal, pagoFranklin)
	confirmacionTrevor := confirmarPago(TrevorAddr, "Trevor", botinTotal, pagoTrevor)

	generarReporte(exitoMision, botinMision, botinAgregado, "", "", confirmacionLester, confirmacionFranklin, confirmacionTrevor)

}
