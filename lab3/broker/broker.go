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
	"fmt"
	"log"
	"os"

	pb "broker/proto"

	"google.golang.org/grpc"
)

var flightsFile = "flight_updates.csv"

const (
	PortBroker      = ":50050"
	PortCoordinador = "localhost:50060"
	PortNodoATC1    = "db1ATC:50051"
	PortNodoATC2    = "db2ATC:50052"
	PortNodoATC3    = "db3ATC:50053"
)

type server struct {
	pb.UnimplementedBrokerServer
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
		archivo, err := os.Open(flightsFile)
		if err != nil {
			log.Fatalf("Could not open the file: %v", err)
		}
		defer archivo.Close()

		scanner := bufio.NewScanner(archivo)
		for scanner.Scan() {
			line := scanner.Text()
			sim_time, flight_id, update_type, update_value := parseFlightUpdate(line) 
			fmt.Println("Read line:", sim_time," - ", flight_id, " - ", update_type, " - ", update_value)
			
		}
	}

}
