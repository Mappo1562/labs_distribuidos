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
	"context"
	"encoding/csv"
	"log"
	"math/rand"
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

// Cantidad de pistas = 8
// Asientos por avion = 15 * A o B = 30

type Seat struct {
	SeatId   string
	Occupied bool
}

type Flight struct {
	FlightId string
	Seats    []Seat
}

var Flights []Flight

// Estructura para representar una actualización de vuelo
type FlightStates struct {
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
	b := &Broker{RoundRobinIndex: 0}
	b.connectToDatanodes()
	return b
}

var broker = NewBroker()

// -------------------------
// *  CLIENTE MR DATANODE  *
// -------------------------
func RoundRobinIndex(b *Broker) pb.DatanodeClient {
	client := b.datanodeClients[b.RoundRobinIndex]
	b.RoundRobinIndex = (b.RoundRobinIndex + 1) % len(b.datanodeClients)
	return client
}

func (s *server) MRRead(ctx context.Context, req *pb.MRReadRequest) (*pb.MRReadResponse, error) {
	client := RoundRobinIndex(broker)
	res, err := client.MRRead(ctx, req)
	if err != nil {
		log.Printf("Error comunicandose con datanode: %v", err)
		return nil, err
	}
	return res, nil
}

//	-------------------------
// 	*      Coordinador      *
// 	-------------------------

func (s *server) ApplyWrite(ctx context.Context, req *pb.ApplyWriteRequest) (*pb.ApplyWriteResponse, error) {
	log.Printf("Escritura recibida de: %s asiento: %s\n", req.ClienteId, req.Seat)
	client := RoundRobinIndex(broker)
	res, err := client.ApplyWrite(ctx, req)
	if err != nil {
		log.Printf("Error comunicandose con datanode: %v", err)
		return nil, err
	}
	return res, nil
}

//	-------------------------
//	*  BROADCAST DATANODES  *
//	-------------------------

func DisponibilidadSeats() []Seat {
	var seats []Seat
	for i := 1; i <= 15; i++ {
		seatId := "A" + strconv.Itoa(i)
		occupied := rand.Intn(2) == 0
		seats = append(seats, Seat{SeatId: seatId, Occupied: occupied})
		seatId = "B" + strconv.Itoa(i)
		occupied = rand.Intn(2) == 0
		seats = append(seats, Seat{SeatId: seatId, Occupied: occupied})
	}
	return seats
}

func ExisteFlight(flightId string) bool {
	for _, flight := range Flights {
		if flight.FlightId == flightId {
			return true
		}
	}
	return false
}

// Carga las actualizaciones de vuelo desde un archivo CSV
func loadFlightUpdates(filename string) ([]FlightStates, error) {
	log.Printf("Cargando actualizaciones de vuelo desde %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var updates []FlightStates

	r := csv.NewReader(file)
	_, err = r.Read()
	if err != nil {
		return nil, err
	}

	for {
		record, err := r.Read()
		if err != nil {
			break
		}

		simTime, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, err
		}

		update := FlightStates{
			SimTime:     simTime,
			FlightId:    record[1],
			UpdateType:  record[2],
			UpdateValue: record[3],
		}
		log.Printf("Flight Update Simulation Loaded - Time: %d, FlightID: %s, Type: %s, Value: %s", update.SimTime, update.FlightId, update.UpdateType, update.UpdateValue)
		updates = append(updates, update)

		if !ExisteFlight(update.FlightId) {
			newFlight := Flight{
				FlightId: update.FlightId,
				Seats:    DisponibilidadSeats(),
			}
			Flights = append(Flights, newFlight)
		}
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
		update := &pb.FlightStates{
			FlightId:    flight_id,
			UpdateType:  update_type,
			UpdateValue: update_value,
		}
		_, err := client.FlightUpdate(ctx, update)
		if err != nil {
			log.Printf("Error comunicandose con datanode: %v", err)
		}
	}

}

func convertFlightsToProto(flights []Flight) []*pb.Flight {
	var protoFlights []*pb.Flight
	for _, flight := range flights {
		var protoSeats []*pb.Seat
		for _, seat := range flight.Seats {
			protoSeat := &pb.Seat{
				SeatId:   seat.SeatId,
				Occupied: seat.Occupied,
			}
			protoSeats = append(protoSeats, protoSeat)
		}
		protoFlight := &pb.Flight{
			FlightId: flight.FlightId,
			Seats:    protoSeats,
		}
		protoFlights = append(protoFlights, protoFlight)
	}
	return protoFlights
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
		updates, err := loadFlightUpdates(flightsFile)
		datanode := broker.datanodeClients[0]
		_, err = datanode.CreateFlights(context.Background(), &pb.Flights{Flights: convertFlightsToProto(Flights)})
		if err != nil {
			log.Fatalf("Error loading flight updates: %v", err)
		}

		time.Sleep(5 * time.Second) // Espera antes de iniciar la simulación

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
