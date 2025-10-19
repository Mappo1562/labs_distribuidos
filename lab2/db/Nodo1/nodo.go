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
	"strings"
	"sync"
	"time"

	pb "Nodo1/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const PortNodo1 = "50051"
const PortNodo2 = "50052"
const PortNodo3 = "50053"

var globalListener net.Listener

type Oferta = pb.Oferta

var LastOffer Oferta

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
			LastOffer = o
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
	LastOffer = *o
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

func (s *DBServer) Registrarse(ctx context.Context, req *pb.Registro) (*pb.Bool, error) {
	return &pb.Bool{Flag: true}, nil
}

func (s *DBServer) Filter(ctx context.Context, req *pb.FilterRequest) (*pb.FilterResponse, error) {
	var filterOffer = req.Oferta
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := &pb.FilterResponse{}
	for _, o := range s.store {
		if strings.EqualFold(o.Categoria, filterOffer.Categoria) &&
			strings.EqualFold(o.Tienda, filterOffer.Tienda) &&
			strings.EqualFold(o.Producto, filterOffer.Producto) {
			res.Ofertas = append(res.Ofertas, o)
		}
	}
	return res, nil
}

// Sincronizar devuelve las ofertas añadidas o modificadas
// después de la oferta pasada en el request.
func (s *DBServer) Sincronizar(ctx context.Context, req *pb.SincronizarRequest) (*pb.SincronizarResponse, error) {
	var lastOffer = req.Oferta
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := &pb.SincronizarResponse{}
	var flagFound bool = false
	for _, o := range s.store {
		if flagFound {
			res.Ofertas = append(res.Ofertas, o)
			continue
		}

		if strings.EqualFold(o.OfertaId, lastOffer.OfertaId) && strings.EqualFold(o.FechaModificacion, lastOffer.FechaModificacion) {
			flagFound = true
		}
	}
	return res, nil
}

// syncTrigger inicia la sincronización con Nodo2 y Nodo3
func (s *DBServer) SyncTrigger() (bool, error) {
	// Conectar con Nodo2
	connNodo2, err := grpc.Dial("localhost:"+PortNodo2, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to Nodo2: %v", err)
	}
	defer connNodo2.Close()
	clientNodo2 := pb.NewDBNodeClient(connNodo2)
	resNodo2, err := clientNodo2.Sincronizar(context.Background(), &pb.SincronizarRequest{Oferta: &LastOffer})
	if err != nil {
		log.Printf("Error sincronizando con Nodo2: %v", err)
	} else {
		log.Printf("Sincronización con Nodo2: recibidas %d ofertas", len(resNodo2.Ofertas))
	}

	// Conectar con Nodo3
	connNodo3, err := grpc.Dial("localhost:"+PortNodo3, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to Nodo3: %v", err)
	}
	defer connNodo3.Close()
	clientNodo3 := pb.NewDBNodeClient(connNodo3)
	resNodo3, err := clientNodo3.Sincronizar(context.Background(), &pb.SincronizarRequest{Oferta: &LastOffer})
	if err != nil {
		log.Printf("Error sincronizando con Nodo3: %v", err)
	} else {
		log.Printf("Sincronización con Nodo3: recibidas %d ofertas", len(resNodo3.Ofertas))
	}

	if len(resNodo2.Ofertas) != len(resNodo3.Ofertas) {
		log.Printf("Inconsistencia detectada entre Nodo2 y Nodo3")
		return false, errors.New("Inconsistencia entre nodos")
	}
	for _, oferta := range resNodo2.Ofertas {
		s.mu.Lock()
		s.store[oferta.OfertaId] = oferta
		s.mu.Unlock()
		// Persistir en disco (escribe al final de data.jsonl)
		if err := s.persistOferta(oferta); err != nil {
			log.Printf("Error al persistir oferta %s: %v", oferta.OfertaId, err)
			return false, err
		}

		log.Printf("Oferta sincronizada: %s (%s - %s, $%d, stock=%d)", oferta.OfertaId, oferta.Tienda, oferta.Producto, oferta.Precio, oferta.Stock)
	}
	return true, nil
}

func main() {
	port := flag.String("port", PortNodo1, "port")
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
	// Guardar el listener globalmente para simulación de error
	globalListener = lis

	grpcSrv := grpc.NewServer()
	pb.RegisterDBNodeServer(grpcSrv, srv)

	// Iniciar la simulación de error en segundo plano
	go func() {
		time.Sleep(10 * time.Second)
		log.Println("Simulando caída del nodo...")
		// Detener operaciones del nodo
		if globalListener != nil {
			globalListener.Close()
		}
		log.Println("SimularError finalizado.")

		// Esperar un tiempo antes de "recuperarse"
		time.Sleep(20 * time.Second)
		log.Println("Nodo recuperado, reanudando operaciones...")

		lis, err := net.Listen("tcp", ":"+PortNodo1)
		if err != nil {
			log.Fatal(err)
		}
		globalListener = lis
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("DB node listening on %s (data=%s)", *port, *data)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Printf("gRPC server stopped: %v", err)
	}

	select {}

}
