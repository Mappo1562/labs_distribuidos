//	    !!!!      .!!!!!!.
//	   !!!!      !!!!  !!!!
//	  !!!!            .!!!!
//	 !!!!            !!!!'
//	 !!!!             '!!!.
//	  !!!!       !!!    !!!
//	   !!!!      !!!!  !!!!
//	    !!!!      '!!!!!!'

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "broker/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Archivo CSV de actualizaciones de vuelo
var flightsFile = "flight_updates.csv"

const (
	PortBroker      = ":50050"
	PortCoordinador = "localhost:50060"
	PortNodoATC1    = "db1ATC:50051"
	PortNodoATC2    = "db2ATC:50052"
	PortNodoATC3    = "db3ATC:50053"
	PortDatanode1   = "db1:50061"
	PortDatanode2   = "db2:50062"
	PortDatanode3   = "db3:50063"
)

// Estructura para representar una actualización de vuelo
type FligthStates struct {
	SimTime     int64
	FlightId    string
	UpdateType  string
	UpdateValue string
}

type server struct {
	pb.UnimplementedBrokerServer
}

type Broker struct {
	// Clientes gRPC para los Datanodes
	datanodeClients []pb.DatanodeClient
	RoundRobinIndex int
}

func NewBroker() *Broker {
	b := &Broker{}
	b.connectToDatanodes()
	return b
}

func RoundRobinIndex(b *Broker) pb.DatanodeClient {
	client := b.datanodeClients[b.RoundRobinIndex]
	b.RoundRobinIndex = (b.RoundRobinIndex + 1) % len(b.datanodeClients)
	return client
}

func (s *server) MRRead(ctx context.Context, req *pb.MRReadRequest) (*pb.MRReadResponse, error) {

	return &pb.MRReadResponse{}, nil
}

// -------------------------
// *  BROADCAST DATANODES  *
// -------------------------

// Parsea una línea del archivo CSV de actualizaciones de vuelo
func parseFlightUpdate(line string) (string, string, string, string) {
	var sim_time, flight_id, update_type, update_value string
	fmt.Sscanf(line, "%[^,],%[^,],%[^,],%s", &sim_time, &flight_id, &update_type, &update_value)
	return sim_time, flight_id, update_type, update_value
}

// Carga las actualizaciones de vuelo desde un archivo CSV
func loadFlightUpdates(filename string) ([]FligthStates, error) {
	log.Printf("Cargando actualizaciones de vuelo desde %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var updates []FligthStates
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		sim_time, flight_id, update_type, update_value := parseFlightUpdate(line)
		sim_time_int, _ := strconv.Atoi(sim_time)
		update := FligthStates{
			SimTime:     int64(sim_time_int),
			FlightId:    flight_id,
			UpdateType:  update_type,
			UpdateValue: update_value,
		}
		updates = append(updates, update)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	log.Printf("Cargadas %d actualizaciones de vuelo", len(updates))
	return updates, nil
}

// Conecta a los Datanodes
func (b *Broker) connectToDatanodes() {
	log.Println("Conectando a Datanodes...")
	var datanodes = []string{PortDatanode1, PortDatanode2, PortDatanode3}
	for _, address := range datanodes {
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Error conectando a %s: %v", address, err)
		}
		client := pb.NewDatanodeClient(conn)
		b.datanodeClients = append(b.datanodeClients, client)
		log.Printf("Conectado a Datanode en %s", address)
	}
}

// BroadcastDatanodes envía las actualizaciones de vuelo a todos los Datanodes
func (b *Broker) BroadcastDatanodes(flight_id string, update_type string, update_value string) {
	log.Printf("Broadcasting to Datanodes: FlightID=%s, Type=%s, Value=%s", flight_id, update_type, update_value)
	for _, client := range b.datanodeClients {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		update := &pb.FligthStates{
			FlightId:    flight_id,
			UpdateType:  update_type,
			UpdateValue: update_value,
		}
		_, err := client.FligthUpdate(ctx, update)
		if err != nil {
			log.Printf("Error comunicandose con datanode: %v", err)
		}
	}

}

func main() {
	lis, err := net.Listen("tcp", PortBroker)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBrokerServer(s, &server{})
	log.Printf("Broker listening at %v", lis.Addr())

	go func() {
		log.Println("Iniciando simulación de actualizaciones de vuelo...")
		broker := NewBroker()
		updates, err := loadFlightUpdates(flightsFile)
		if err != nil {
			log.Fatalf("Error loading flight updates: %v", err)
		}

		startTime := time.Now()
		for _, update := range updates {
			simulatedTime := time.Duration(update.SimTime) * time.Second
			time.Sleep(simulatedTime - time.Since(startTime))
			log.Printf("Broadcasting update: FlightID=%s, Type=%s, Value=%s", update.FlightId, update.UpdateType, update.UpdateValue)
			broker.BroadcastDatanodes(update.FlightId, update.UpdateType, update.UpdateValue)
		}
		log.Println("Simulación de actualizaciones de vuelo completada.")
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
