package main

import (
	pb "clienteMR/proto"
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
	brokerAddr = "broker:50050"
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

func main() {
	time.Sleep(time.Second * 10) // esperar a que esten arriba los demas servicios
	rand.Seed(02122003)

	vuelos, _ := leerVuelos("flight_updates.csv")
	listaVuelos := strings.Split(vuelos, ",")

	conn, err := grpc.NewClient(brokerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("[MR] No se pudo conectar al Broker: %v\n", err)
	}

	defer conn.Close()

	brokerClient := pb.NewBrokerClient(conn)

	clienteID := os.Getenv("CLIENT_ID")

	lastVersionSeen := make(map[string]int64)

	for i := 0; i < 10; i++ {

		vuelo := listaVuelos[rand.Intn(len(listaVuelos))]
		last := lastVersionSeen[vuelo]
		mensaje := &pb.MRReadRequest{
			FlightId:        vuelo,
			ClientId:        clienteID,
			LastVersionSeen: last,
		}

		resp, err := brokerClient.MRRead(context.Background(), mensaje)
		if err != nil {
			log.Fatalf("[MR] Error al obtener la información del vuelo: %v\n", err)
		}

		if resp.Version < last {
			log.Printf("[MR] La version obtenida de la respuesta para el vuelo %s es antigua ", vuelo)
			log.Printf(" version %d < %d", resp.Version, last)
		} else {
			lastVersionSeen[vuelo] = resp.Version
			log.Printf("[MR] Vueñp %s actualizado en el cliente %s. Version: %d, Estado: %s", vuelo, clienteID, resp.Version, resp.Status)
		}

		time.Sleep(2 * time.Second)
	}

}
