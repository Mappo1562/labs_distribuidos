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
	"strconv"
	"strings"
	"sync"
	"time"

	pb "Nodo3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PortNodo1  = "db1:50051"
	PortNodo2  = "db2:50052"
	PortNodo3  = "50053"
	PortBroker = "broker:50050"
)

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

// Sincronizar devuelve las ofertas añadidas o modificadas
// después de la oferta pasada en el request.
func (s *DBServer) Sincronizar(ctx context.Context, req *pb.SincronizarRequest) (*pb.SincronizarResponse, error) {
	log.Println("Iniciando sincronización...")
	var lastOffer = req.Oferta
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := &pb.SincronizarResponse{}

	layout := "02/01/2006 15:04:05.000"

	tLast, err := time.Parse(layout, lastOffer.Fecha)
	if err != nil {
		log.Printf("Error parseando fecha de lastOffer: %v", err)
		return res, err
	}

	for _, o := range s.store {
		tOferta, err := time.Parse(layout, o.Fecha)
		if err != nil {
			log.Printf("Error parseando fecha de oferta %s: %v", o.OfertaId, err)
			continue
		}

		// Solo agregar si la fecha de o es posterior a la de lastOffer
		if tOferta.After(tLast) {
			res.Ofertas = append(res.Ofertas, o)
		}
	}
	return res, nil
}

// syncTrigger inicia la sincronización con Nodo1 y Nodo2
func (s *DBServer) SyncTrigger() (bool, error) {
	log.Println("Iniciando sincronización con Nodo1 y Nodo2...")
	// Conectar con Nodo2
	connNodo2, err := grpc.Dial(PortNodo2, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	// Conectar con Nodo1
	connNodo1, err := grpc.Dial(PortNodo1, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to Nodo1: %v", err)
	}
	defer connNodo1.Close()
	clientNodo1 := pb.NewDBNodeClient(connNodo1)
	resNodo1, err := clientNodo1.Sincronizar(context.Background(), &pb.SincronizarRequest{Oferta: &LastOffer})
	if err != nil {
		log.Printf("Error sincronizando con Nodo1: %v", err)
	} else {
		log.Printf("Sincronización con Nodo1: recibidas %d ofertas", len(resNodo1.Ofertas))
	}

	if len(resNodo2.Ofertas) != len(resNodo1.Ofertas) {
		log.Printf("Inconsistencia detectada entre Nodo2 y Nodo1")
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

func activoNodo() (*pb.Bool, error) {
	conn, err := grpc.Dial(PortBroker, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
	if err != nil {
		log.Printf("No se pudo conectar al broker (%s): %v", PortBroker, err)
		return &pb.Bool{Flag: false}, err
	}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)

	// Creamos el mensaje de registro
	reg := &pb.Registro{
		Nombre: "db3",
		Rol:    2, // 2 = Nodo
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Activo(ctx, reg)
	if err != nil {
		log.Printf("Error al consultar estado en el broker: %v", err)
		return &pb.Bool{Flag: false}, err
	}
	return resp, nil
}

func RegistrarBroker() {
	for {
		log.Println("Intentando registrar nodo en el broker...")
		conn, err := grpc.Dial(PortBroker, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
		if err != nil {
			log.Printf("No se pudo conectar al broker (%s): %v", PortBroker, err)
			time.Sleep(5 * time.Second)
			continue
		}
		client := pb.NewBrokerClient(conn)

		// Creamos el mensaje de registro
		reg := &pb.Registro{
			Nombre: "db3",
			Rol:    2, // 2 = Nodo
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.Registrarse(ctx, reg)
		if err != nil {
			log.Printf("Error al registrarse en el broker: %v", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.Flag {
			log.Printf("✅ Nodo %s registrado correctamente en el broker", reg.Nombre)
			conn.Close()
			break
		} else {
			log.Printf("⚠️ Broker rechazó el registro de %s, reintentando...", reg.Nombre)
			conn.Close()
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *DBServer) Siguesvivo(ctx context.Context, req *pb.Vacio) (*pb.Bool, error) {
	log.Println("Cerrando nodo en 5 segundos...")
	go func() {
		time.Sleep(5 * time.Second)
		eliminarArchivo()
		log.Println("Nodo cerrado.")
		os.Exit(0)
	}()
	return &pb.Bool{Flag: true}, nil
}

func splitFilter(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "null") {
		return nil
	}
	parts := strings.Split(s, ";")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func (s *DBServer) GetHistoric(ctx context.Context, req *pb.Filtro) (*pb.RangeSinceResponse, error) {
	log.Println("Generando reporte histórico con filtros...")
	log.Printf("Filtros recibidos - Categoría: '%s', Tienda: '%s', PrecioMax: '%s'", req.Categoria, req.Tienda, req.PrecioMax)
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := &pb.RangeSinceResponse{}

	// Parsear filtros múltiples
	categorias := splitFilter(req.Categoria)
	tiendas := splitFilter(req.Tienda)

	// Parsear precio máximo
	var precioMax int64
	if !strings.EqualFold(req.PrecioMax, "") && !strings.EqualFold(req.PrecioMax, "null") {
		if p, err := strconv.ParseInt(req.PrecioMax, 10, 64); err == nil {
			precioMax = p
		}
	}

	for _, o := range s.store {
		match := true

		// ---- FILTRO DE CATEGORÍAS ----
		if len(categorias) > 0 {
			ok := false
			for _, c := range categorias {
				if strings.EqualFold(o.Categoria, c) {
					ok = true
					break
				}
			}
			if !ok {
				match = false
			}
		}

		// ---- FILTRO DE TIENDAS ----
		if len(tiendas) > 0 {
			ok := false
			for _, t := range tiendas {
				if strings.EqualFold(o.Tienda, t) {
					ok = true
					break
				}
			}
			if !ok {
				match = false
			}
		}

		// ---- FILTRO DE PRECIO ----
		if precioMax > 0 && o.Precio > precioMax {
			match = false
		}

		// ---- AGREGAR RESULTADO ----
		if match {
			resp.Ofertas = append(resp.Ofertas, o)
		}
	}
	log.Printf("Reporte histórico generado con %d ofertas", len(resp.Ofertas))

	return resp, nil
}

func eliminarArchivo() {
	err := os.Remove("/data/data.jsonl")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("⚠️ El archivo data.jsonl no existe, no hay nada que borrar.")
		} else {
			log.Printf("❌ Error al eliminar data.jsonl: %v", err)
		}
		return
	}
	log.Println("🗑️ Archivo data.jsonl eliminado correctamente.")
}

func main() {
	//Cambiar puerto segun nodo
	port := flag.String("port", PortNodo3, "port")
	data := flag.String("data", os.Getenv("DATA_PATH"), "data file")
	flag.Parse()

	if *data == "" {
		*data = "data.jsonl"
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

	// Registro del nodo en el broker en segundo plano
	go func() {
		RegistrarBroker()
	}()

	log.Printf("DB node listening on %s (data=%s)", *port, *data)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Printf("gRPC server stopped: %v", err)
	}

	select {}

}
