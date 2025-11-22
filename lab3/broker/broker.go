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
	"fmt"
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

var reporteFile = "Reporte.txt"

const (
	PortBroker      = ":50050"
	PortCoordinador = "coordinador:50060"
	PortNodoATC1    = "dbatc0:50051"
	PortNodoATC2    = "dbatc1:50052"
	PortNodoATC3    = "dbatc2:50053"
	PortDatanode1   = "datanode0:50061"
	PortDatanode2   = "datanode1:50062"
	PortDatanode3   = "datanode2:50063"
)

// Cantidad de pistas = 8
// Asientos por avion = 15 * A o B = 30

type Reporte struct {
	Timestamp time.Time
	Message   string
}

var reporte []Reporte

func AddToReporte(message string) {
	timestamp := time.Now()
	entry := Reporte{
		Timestamp: timestamp,
		Message:   message,
	}
	reporte = append(reporte, entry)
}

func CreateReporteFile() {
	log.Println("Creando archivo de reporte...")
	file, err := os.Create("/app/reportes/Reporte.txt")
	if err != nil {
		fmt.Println("Error al crear el reporte:", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "Reporte de Operaciones Críticas\n===============================\n")
	for _, entry := range reporte {
		fmt.Fprintf(file, "[%s] %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Message)
	}
	log.Println("Reporte guardado en /app/reportes/Reporte.txt")
}

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
	RoundRobinIndex int64
	RoundRobinPista int64
}

func NewBroker() *Broker {
	b := &Broker{RoundRobinIndex: 0, RoundRobinPista: 1}
	b.connectToDatanodes()
	return b
}

var broker = NewBroker()

//	-------------------------
// 	*       Consenso        *
// 	-------------------------

func BroadcastATCs(flight_id string, pista int64) (int64, bool) {
	log.Printf("Broadcasting to ATCs: FlightID=%s, Pista=%d", flight_id, pista)
	var atcClients []pb.ATCClient
	var atcAddresses = []string{PortNodoATC1, PortNodoATC2, PortNodoATC3}
	for _, address := range atcAddresses {
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Error conectando a %s: %v", address, err)
		}
		client := pb.NewATCClient(conn)
		atcClients = append(atcClients, client)
	}

	for _, client := range atcClients {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		req := &pb.Record{
			FlightId: flight_id,
			Pista:    pista,
		}
		res, err := client.Insert(ctx, req)
		if err != nil {
			log.Printf("Error comunicandose con ATC: %v", err)
		}
		if res.Exito && res.Lider {
			log.Printf("ATC %d asignó pista %d para vuelo %s", res.Id, pista, flight_id)
			reporteText := "Operacion Crítica: Pista " + strconv.FormatInt(pista, 10) + " asignada al vuelo " + flight_id
			AddToReporte(reporteText)
			return pista, true
		} else {
			return 0, false
		}
	}
	return 0, false
}

func AsignarPistaConsenso(flight_id string) (int64, bool) {
	pista := broker.RoundRobinPista
	broker.RoundRobinPista = (broker.RoundRobinPista % 8) + 1 // Suponiendo 8 pistas

	res, ok := BroadcastATCs(flight_id, pista)

	return res, ok
}

// -------------------------
// *  CLIENTE MR DATANODE  *
// -------------------------
func RoundRobinIndex(b *Broker) pb.DatanodeClient {
	client := b.datanodeClients[b.RoundRobinIndex]
	b.RoundRobinIndex = (b.RoundRobinIndex + 1) % int64(len(b.datanodeClients))
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
	coordinadorReq := &pb.CoordinadorWriteRequest{
		ClienteId: req.ClienteId,
		Seat:      req.Seat,
	}
	coordinadorRes, err := client.CoordinadorWrite(ctx, coordinadorReq)
	if err != nil {
		log.Printf("Error comunicandose con datanode: %v", err)
		return nil, err
	}
	res := &pb.ApplyWriteResponse{
		Success:    coordinadorRes.Success,
		Msg:        coordinadorRes.Msg,
		DatonodeId: broker.RoundRobinIndex,
		Version:    coordinadorRes.Version,
	}
	return res, nil
}

func (s *server) GetInitialInfo(ctx context.Context, req *pb.GetInitialInfoRequest) (*pb.GetInitialInfoResponse, error) {
	log.Printf("Solicitud de info inicial para vuelo: %s\n", req.FlightId)
	client := RoundRobinIndex(broker)
	res, err := client.GetInitialInfo(ctx, req)
	if err != nil {
		log.Printf("Error comunicandose con datanode: %v", err)
		return nil, err
	}
	return res, nil
}

func (s *server) BrokerRead(ctx context.Context, req *pb.BrokerReadRequest) (*pb.BrokerReadResponse, error) {
	if req.DatanodeId >= 0 && req.DatanodeId < int64(len(broker.datanodeClients)) {
		log.Printf("Solicitud de lectura para vuelo: %s en datanode: %d\n", req.FlightId, req.DatanodeId)
		client := broker.datanodeClients[req.DatanodeId]
		res, err := client.BrokerRead(ctx, req)
		if err != nil {
			log.Printf("Error comunicandose con datanode: %v", err)
			return nil, err
		}
		return res, nil
	}
	log.Printf("Solicitud de lectura para vuelo: %s en datanode seleccionado por round robin\n", req.FlightId)
	client := RoundRobinIndex(broker)
	res, err := client.BrokerRead(ctx, req)
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
		if err != nil {
			log.Fatalf("Error cargando flight updates: %v", err)
		}
		datanodes := broker.datanodeClients
		for node := range datanodes {
			_, err = datanodes[node].CreateFlights(context.Background(), &pb.Flights{Flights: convertFlightsToProto(Flights)})
			if err != nil {
				log.Fatalf("Error enviando vuelos iniciales al datanode %d: %v", node, err)
			}
		}

		time.Sleep(5 * time.Second) // Espera antes de iniciar la simulación

		startTime := time.Now()
		for _, update := range updates {
			simulatedTime := time.Duration(update.SimTime) * time.Second
			time.Sleep(simulatedTime - time.Since(startTime))
			log.Printf("Simulando Actualizacion: FlightID=%s, Type=%s, Value=%s", update.FlightId, update.UpdateType, update.UpdateValue)
			broker.BroadcastDatanodes(update.FlightId, update.UpdateType, update.UpdateValue)
		}
		log.Println("Simulación de actualizaciones de vuelo completada.")
	}()

	go func() {

		time.Sleep(30 * time.Second)
		log.Println("Solicitando Pistas mediante consenso...")
		for _, flight := range Flights {
			pista, ok := AsignarPistaConsenso(flight.FlightId)
			if ok {
				log.Printf("Pista asignada para vuelo %s: %d", flight.FlightId, pista)
			}

		}

		CreateReporteFile()
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
