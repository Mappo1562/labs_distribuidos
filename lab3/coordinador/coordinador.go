package main

import (
	"context"
	pb "coordinador/proto"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CoordinadorServer struct {
	pb.UnimplementedCoordinadorRYWServer

	sesiones     map[string]SessionInfo
	brokerClient pb.BrokerClient
	mu           sync.Mutex
}

type SessionInfo struct {
	Datanode string
	Expire   time.Time
}

const (
	brokerAddr = "broker:50050"
	port       = ":50060"
)

func (s *CoordinadorServer) CheckIn(ctx context.Context, in *pb.CheckInRequest) (*pb.CheckInResponse, error) {
	log.Printf("[Coordinador] Recibido CheckIn de: %s asiento: %s\n", in.ClienteId, in.Seat)

	mensaje := &pb.ApplyWriteRequest{
		ClienteId: in.ClienteId,
		FlightId:  in.FlightId,
		Seat:      in.Seat,
		RequestId: in.RequestId,
	}
	resp, err := s.brokerClient.ApplyWrite(ctx, mensaje)
	if err != nil {
		return nil, err
	}

	if resp.Success {
		s.mu.Lock()
		s.sesiones[in.ClienteId] = SessionInfo{
			Datanode: resp.DatonodeId,
			Expire:   time.Now().Add(1 * time.Minute),
		}
		s.mu.Unlock()

		log.Printf("[Coordinador] Sesión guardada %s -> %s\n", in.ClienteId, resp.DatonodeId)
	}

	respuesta := &pb.CheckInResponse{
		Success: resp.Success,
		Msg:     resp.Msg,
	}

	return respuesta, nil
}

func (s *CoordinadorServer) GetBoardingPass(ctx context.Context, in *pb.GetBoardingPassRequest) (*pb.GetBoardingPassResponse, error) {
	log.Printf("[Coordinador] Obteniendo el ticket de embarque de %s\n", in.ClienteId)

	s.mu.Lock()
	sesion, exists := s.sesiones[in.ClienteId]
	s.mu.Unlock()

	if exists && time.Now().After(sesion.Expire) {
		log.Printf("[Coordinador] sesión de %s\n", in.ClienteId)
		s.mu.Lock()
		delete(s.sesiones, in.ClienteId)
		s.mu.Unlock()
		exists = false
	}
	var mensaje *pb.BrokerReadRequest
	//arreglar
	if exists {

		mensaje = &pb.BrokerReadRequest{
			FlightId:   in.FlightId,
			DatanodeId: sesion.Datanode,
		}

	} else {
		mensaje = &pb.BrokerReadRequest{
			FlightId: in.FlightId,
		}
	}

	bResp, err := s.brokerClient.BrokerRead(ctx, mensaje)
	if err != nil {
		log.Printf("[Coordinador] Error llamando al broker: %v", err)
		return nil, err
	}

	resp := &pb.GetBoardingPassResponse{
		FlightId: bResp.FlightId,
		Seat:     bResp.Seat,
		Version:  bResp.Version,
		Success:  bResp.Success,
		Msg:      bResp.Msg,
	}
	return resp, nil
}

func (s *CoordinadorServer) GetSeats(ctx context.Context, in *pb.GetSeatsRequest) (*pb.GetSeatsResponse, error) {

	vuelo := in.FlightId
	bResp, err := s.brokerClient.GetInitialInfo(ctx, &pb.GetInitialInfoRequest{
		FlightId: vuelo,
	})

	if err != nil {
		return &pb.GetSeatsResponse{
			Success: false,
			Msg:     fmt.Sprintf("error al leer broker: %v", err),
		}, nil
	}

	return &pb.GetSeatsResponse{
		FlightId: in.FlightId,
		Seats:    bResp.Seats,
		Success:  true,
		Msg:      "ok",
	}, nil
}
func main() {

	conn, err := grpc.NewClient(brokerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar al Broker: %v", err)
	}

	defer conn.Close()

	server := &CoordinadorServer{
		sesiones:     make(map[string]SessionInfo),
		brokerClient: pb.NewBrokerClient(conn),
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error al iniciar escucha: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinadorRYWServer(grpcServer, server)

	log.Println("Coordinador activo en ", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
