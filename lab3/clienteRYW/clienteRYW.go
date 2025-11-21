package main

import (
	pb "clienteRYW/proto"
	"context"
	"encoding/csv"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	coordinadorAddr = "localhost:50060"
)

func leerVuelos(path string) (string, error) {

	archivo, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)
	lector.TrimLeadingSpace = true

	filas, err := lector.ReadAll()

	if err != nil {
		return "", err
	}

	var ayuda string

	for i, fila := range filas {
		if i == 0 {
			continue
		}

		flightIDid := strings.TrimSpace(fila[1])
		ayuda += flightIDid + ","
	}

	ayuda = strings.TrimSuffix(ayuda, ",")
	return ayuda, nil
}

func elegirAzar(completo string) string {
	arreglo := strings.Split(completo, ",")
	escogido := strings.TrimSpace(arreglo[rand.Intn(len(arreglo))])
	return escogido
}

func main() {

	rand.Seed(02122003)

	conn, err := grpc.NewClient(coordinadorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[RYW] No se pudo conectar al coordinador: %v\n", err)
	}

	defer conn.Close()

	coordinadorClient := pb.NewCoordinadorRYWClient(conn)

	ayuda, _ := leerVuelos("flight_updates.csv")
	flightID := elegirAzar(ayuda)
	clienteID := os.Getenv("CLIENT_ID")

	respGetSeats, err := coordinadorClient.GetSeats(context.Background(), &pb.GetSeatsRequest{
		FlightId: flightID,
	})

	if err != nil {
		log.Fatalf("[RYW] Error al obtener los asientos disponibles: %v\n", err)
	}

	if !respGetSeats.Success {
		log.Fatalln("[RYW] No se pudo obtener los asientos disponibles")
		//volver a intentar o terminar
	} else {
		log.Printf("[RYW] Se pudieron obtener los asientos disponibles del vuelo %s\n", flightID)
	}

	seat := elegirAzar(respGetSeats.Seats) // elegir un asiento random dentro de los disponibles.

	respChekIn, err := coordinadorClient.CheckIn(context.Background(), &pb.CheckInRequest{
		ClienteId: clienteID,
		FlightId:  flightID,
		Seat:      seat,
		RequestId: "hola",
	})

	if err != nil {
		log.Fatalf("[RYW] Error en CheckIn: %v\n", err)
	}

	if !respChekIn.Success {
		log.Fatalf("[RYW] NO se pudo hacerl el CheckIn por %s\n", respChekIn.Msg)
		//terminar proceso o volver a intentar checkin
	} else {
		log.Println("[RYW] CheckIn exitoso")
	}

	time.Sleep(time.Second * 1)

	respGetBoardingPass, err := coordinadorClient.GetBoardingPass(context.Background(), &pb.GetBoardingPassRequest{
		ClienteId: clienteID,
		FlightId:  flightID,
	})

	if err != nil {
		log.Fatalf("[RYW] Error en GetBoardingPass: %v \n", err)
	}

	if respGetBoardingPass.Seat == seat {
		log.Printf("[RYW] El asiento coincide %s", respGetBoardingPass.Seat)
	} else {
		log.Printf("[RYW] El asiento leido es distinto al escrito %s", respGetBoardingPass.Seat)
	}

}
