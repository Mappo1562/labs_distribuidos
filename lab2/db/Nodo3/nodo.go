package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"sync"

	pb "Nodo1/proto"

	"google.golang.org/grpc"
)

type Oferta = pb.Oferta

type DBServer struct {
	pb.UnimplementedDBNodeServer
	mu          sync.RWMutex
	store       map[string]*Oferta
	dataFile    *os.File
	backlogFile *os.File
}

func NewDBServer(dataPath string) (*DBServer, error) {
	f, err := os.OpenFile(dataPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	srv := &DBServer{
		store:    make(map[string]*Oferta),
		dataFile: f,
	}

	// load existing data
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var o Oferta
		if err := json.Unmarshal(scanner.Bytes(), &o); err == nil {
			srv.store[o.OfertaId] = &o
		}
	}
	log.Printf("Guardando datos en: %s", dataPath)
	return srv, nil
}

func (s *DBServer) persistOferta(o *Oferta) error {
	b, _ := json.Marshal(o)
	s.dataFile.Write(b)
	s.dataFile.Write([]byte("\n"))
	s.dataFile.Sync()
	return nil
}

func (s *DBServer) Store(ctx context.Context, req *pb.StoreRequest) (*pb.StoreResponse, error) {
	o := req.Oferta
	if o == nil {
		return &pb.StoreResponse{Ok: false, Message: "nil Oferta"}, errors.New("nil")
	}

	// Validaciones mínimas
	if o.Stock <= 0 {
		return &pb.StoreResponse{Ok: false, Message: "stock<=0"}, nil
	}
	if o.Categoria == "" || o.Tienda == "" || o.OfertaId == "" {
		return &pb.StoreResponse{Ok: false, Message: "missing fields"}, nil
	}

	// Guardar en memoria (protege acceso concurrente)
	s.mu.Lock()
	s.store[o.OfertaId] = o
	s.mu.Unlock()

	// Persistir en disco (escribe al final de data.jsonl)
	if err := s.persistOferta(o); err != nil {
		log.Printf("Error al persistir oferta %s: %v", o.OfertaId, err)
		return &pb.StoreResponse{Ok: false, Message: "persist error"}, err
	}

	log.Printf("Oferta guardada: %s (%s - %s, $%d, stock=%d)", o.OfertaId, o.Tienda, o.Producto, o.Precio, o.Stock)
	return &pb.StoreResponse{Ok: true, Message: "stored"}, nil
}

func (s *DBServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if o, ok := s.store[req.OfertaId]; ok {
		return &pb.GetResponse{Oferta: o, Found: true}, nil
	}
	return &pb.GetResponse{Found: false}, nil
}

func (s *DBServer) RangeSince(ctx context.Context, req *pb.RangeSinceRequest) (*pb.RangeSinceResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := &pb.RangeSinceResponse{}
	for _, o := range s.store {
		res.Ofertas = append(res.Ofertas, o)
	}
	return res, nil
}

func main() {
	port := flag.String("port", "50051", "port")
	data := flag.String("data", os.Getenv("DATA_PATH"), "data file") // <--- lee la variable del contenedor
	flag.Parse()

	if *data == "" {
		*data = "data.jsonl" // fallback por si no hay variable
	}

	srv, err := NewDBServer(*data)
	if err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatal(err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterDBNodeServer(grpcSrv, srv)
	log.Printf("DB node listening on %s (data=%s)", *port, *data)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
