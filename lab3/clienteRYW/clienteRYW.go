package main

import (
	pb "clienteRYW/proto"
	"context"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	coordinadorAddr = "localhost:50060"
)

func main() {

	conn, err := grpc.NewClient(coordinadorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[RYW] No se pudo conectar al coordinador: %v\n", err)
	}

	defer conn.Close()

	coordinadorClient := pb.NewCoordinadorRYWClient(conn)

	clienteID := "cliente-1"
	flighID := "LA-500"
	seat := "12A"

	respGetSeats, err := coordinadorClient.GetSeats(context.Background(), &pb.GetSeatsRequest{
		FlightId: flighID,
	})

	if err != nil {
		log.Fatalf("[RYW] Error al obtener los asientos disponibles: %v", err)
	}

	if !respGetSeats.Success {
		log.Printf("[RYW] No se pudo obtener los asientos disponibles")
		//volver a intentar o terminar
	} else {
		log.Printf("[RYW] Se pudieron obtener los asientos disponibles del vuelo %s", flighID)
	}

	asientos := strings.Split(respGetSeats.Seats, ",")

	_ = asientos // elegir un asiento random dentro de los disponibles.

	respChekIn, err := coordinadorClient.CheckIn(context.Background(), &pb.CheckInRequest{
		ClienteId: clienteID,
		FlightId:  flighID,
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
		FlightId:  flighID,
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
