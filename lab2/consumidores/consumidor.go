package main

import (
	pb "consumidores/proto"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Consumidor struct {
	id_consumidor   string
	categoria       []string
	tiendas         []string
	precio_max      int64
	archivo_ofertas string
}

type servidorConsumidor struct {
	pb.UnimplementedConsumidorServer
	Consumidor Consumidor
	mu         sync.Mutex
	activo     bool
}

var consumidores []Consumidor

var categoriasValidas = map[string]struct{}{
	"Electronica":       {},
	"Moda":              {},
	"Hogar":             {},
	"Deportes":          {},
	"Belleza":           {},
	"Infantil":          {},
	"Computacion":       {},
	"Electrodomesticos": {},
	"Herramientas":      {},
	"Juguetes":          {},
	"Automotriz":        {},
	"Mascotas":          {},
}

func leerConsumidores(path string) ([]Consumidor, error) {

	archivo, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)
	lector.TrimLeadingSpace = true

	filas, err := lector.ReadAll()

	if err != nil {
		return nil, err
	}

	var ayuda []Consumidor

	for i, fila := range filas {
		if i == 0 {
			continue
		}

		id := strings.TrimSpace(fila[0])
		categorias := parseCampoMultiple(fila[1])
		tiendas := parseCampoMultiple(fila[2])
		precio_max := parsePrecio(fila[3])

		c := Consumidor{
			id_consumidor:   id,
			categoria:       categorias,
			tiendas:         tiendas,
			precio_max:      precio_max,
			archivo_ofertas: fmt.Sprintf("consumidor_%s_ofertas.csv", id),
		}

		ayuda = append(ayuda, c)

	}

	return ayuda, nil
}

func parseCampoMultiple(campo string) []string {
	campo = strings.TrimSpace(campo)

	if campo == "" || campo == "null" {
		return nil
	}

	partes := strings.Split(campo, ";")
	for i := range partes {
		partes[i] = strings.TrimSpace(partes[i])
	}

	return partes
}

func parsePrecio(campo string) int64 {
	campo = strings.TrimSpace(campo)

	if campo == "" || campo == "null" {
		return -1
	}

	val, _ := strconv.ParseInt(campo, 10, 64)
	return val
}

func (s *servidorConsumidor) NotificarOferta(ctx context.Context, oferta *pb.Oferta) (*pb.Bool, error) {
	s.mu.Lock()

	//si esta caido no recibe nuevas ofertas
	if !s.activo {
		s.mu.Unlock()
		return nil, fmt.Errorf("el consumidor %s esta caido", s.Consumidor.id_consumidor)
	}

	if !categoriasValida(oferta.Categoria) {
		log.Printf("[%s] Categoria invalida: %s \n", s.Consumidor.id_consumidor, oferta.Categoria)
		s.mu.Unlock()
		return &pb.Bool{Flag: false}, nil
	}

	//Simulación caida
	if rand.Float64() < 0.01 {
		s.mu.Unlock()
		go simularCaida(s)
		return nil, fmt.Errorf("el consumidor %s se cayó", s.Consumidor.id_consumidor)
	}

	log.Printf("Se leyo la oferta para %s\n", s.Consumidor.id_consumidor)
	ok := guardarOferta(s.Consumidor, oferta)
	s.mu.Unlock()
	return &pb.Bool{Flag: ok}, nil
}

func levantarServidor(c Consumidor, puerto int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", puerto))

	if err != nil {
		log.Fatalf("[%s] Error al escuchar en puerto %d: %v\n", c.id_consumidor, puerto, err)
	}

	srv := &servidorConsumidor{
		Consumidor: c,
		activo:     true,
	}

	s := grpc.NewServer()
	pb.RegisterConsumidorServer(s, srv)

	go func() {
		fmt.Printf("[%s] Escuchando en puerto %d\n", c.id_consumidor, puerto)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("[%s] Error sirviendo: %v\n", c.id_consumidor, err)
		}
	}()
}

func guardarOferta(c Consumidor, oferta *pb.Oferta) bool {
	path := fmt.Sprintf("/app/ofertas/%s", c.archivo_ofertas)

	existeArchivo := true
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		existeArchivo = false
	} else if err != nil {
		log.Printf("[%s] Error al revisar archivo: %v\n", c.id_consumidor, err)
		return false
	}

	archivo, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Printf("[%s] Error al abrir el archivo: %v\n", c.id_consumidor, err)
		return false
	}
	defer archivo.Close()

	writer := csv.NewWriter(archivo)
	defer writer.Flush()

	if !existeArchivo || info.Size() == 0 {
		header := []string{
			"ID Oferta",
			"Tienda",
			"Categoria",
			"Producto",
			"Precio",
			"Stock",
			"Fecha",
		}
		if err := writer.Write(header); err != nil {
			log.Printf("[%s] Error al escribir el encabezado CSV: %v\n", c.id_consumidor, err)
			return false
		}
	}

	record := []string{
		oferta.OfertaId,
		oferta.Tienda,
		oferta.Categoria,
		oferta.Producto,
		strconv.FormatInt(oferta.Precio, 10),
		strconv.FormatInt(oferta.Stock, 10),
		oferta.Fecha,
	}

	if err := writer.Write(record); err != nil {
		log.Printf("[%s] Error al escribir CSV: %v\n", c.id_consumidor, err)
		return false
	}

	return true
}

func simularCaida(s *servidorConsumidor) {
	s.mu.Lock()
	s.activo = false
	s.mu.Unlock()

	duracion := time.Duration(rand.Intn(5)+5) * time.Second
	log.Printf("El consumidor %s se cayó\n", s.Consumidor.id_consumidor)
	time.Sleep(duracion)

	s.mu.Lock()
	s.activo = true
	s.mu.Unlock()

	s.recuperarOfertas()
}

func (s *servidorConsumidor) recuperarOfertas() {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, err := grpc.NewClient("broker:50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[%s] No se pudo conectar: %v\n", s.Consumidor.id_consumidor, err)
		return
	}

	defer conn.Close()

	cliente := pb.NewBrokerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	registro := &pb.Registro{
		Nombre: s.Consumidor.id_consumidor,
		Rol:    1,
	}
	resp, err := cliente.GenerarHistorico(ctx, registro)

	if err != nil {
		log.Printf("[%s] Error al recuperar historial: %v\n", s.Consumidor.id_consumidor, err)
	}
	ofertas := resp.Ofertas

	if len(ofertas) == 0 {
		log.Printf("[%s] No hay ofertas que recuperar\n", s.Consumidor.id_consumidor)
		return
	}

	path := fmt.Sprintf("/app/ofertas/%s", s.Consumidor.id_consumidor)
	archivo, err := os.Create(path)

	if err != nil {
		log.Printf("[%s] Error al crear el archivo: %v\n", s.Consumidor.id_consumidor, err)
		return
	}

	defer archivo.Close()

	writer := csv.NewWriter(archivo)
	defer writer.Flush()

	header := []string{
		"ID Oferta",
		"Tienda",
		"Categoria",
		"Producto",
		"Precio",
		"Stock",
		"Fecha",
	}

	if err := writer.Write(header); err != nil {
		log.Printf("[%s] Error al escribir encabezado en recuperación: %v\n", s.Consumidor.id_consumidor, err)
		return
	}

	for _, oferta := range ofertas {
		record := []string{
			oferta.OfertaId,
			oferta.Tienda,
			oferta.Categoria,
			oferta.Producto,
			strconv.FormatInt(oferta.Precio, 10),
			strconv.FormatInt(oferta.Stock, 10),
			oferta.Fecha,
		}

		if err := writer.Write(record); err != nil {
			log.Printf("[%s] Error al escribir oferta recuperada: %v\n", s.Consumidor.id_consumidor, err)
			continue
		}
	}

	log.Printf("[%s] Recuperadas las ofertas", s.Consumidor.id_consumidor)
}

func categoriasValida(cat string) bool {
	_, ok := categoriasValidas[cat]
	return ok
}

func (s *servidorConsumidor) PedirDatos(ctx context.Context, vacio *pb.Vacio) (*pb.DatosFinalesConsumidor, error) {

	cantOfertas, err := contarOfertas(s.Consumidor.archivo_ofertas)

	if err != nil {
		log.Printf("Hubo un error al contar cuantas ofertas le llegaron a %s\n", s.Consumidor.id_consumidor)
	}
	return &pb.DatosFinalesConsumidor{
		Id:               s.Consumidor.id_consumidor,
		OfertasRecibidas: int64(cantOfertas),
		NombreCSV:        s.Consumidor.archivo_ofertas,
	}, nil
}

func contarOfertas(path string) (int, error) {
	archivo, err := os.Open(path)

	if err != nil {
		return 0, err
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)
	records, err := lector.ReadAll()

	if err != nil {
		return 0, err
	}

	if len(records) <= 1 {
		return 0, nil
	}

	return len(records) - 1, nil
}

func main() {

	conn, err := grpc.NewClient("broker:50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}

	defer conn.Close()

	cliente := pb.NewBrokerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	grupo, _ := strconv.Atoi(os.Getenv("GRUPO"))
	consumidores, err = leerConsumidores("consumidores.csv")

	if err != nil {
		log.Fatalf("Error leyendo los consumidores: %v", err)
	}

	inicio := (grupo - 1) * 4
	fin := inicio + 4
	consumidoresGrupo := consumidores[inicio:fin]

	basePort := 60060 + ((grupo - 1) * 10) + 1

	for i := range len(consumidoresGrupo) {
		c := consumidoresGrupo[i]

		registro := &pb.Registro{
			Nombre: c.id_consumidor,
			Rol:    1,
		}

		resp, err := cliente.Registrarse(ctx, registro)
		if err != nil {
			log.Fatalf("Error al registrarse: %v\n", err)
		}
		if resp.Flag {
			fmt.Printf("%s registrado en el broker\n", c.id_consumidor)
		} else {
			fmt.Println("Falló el registro")
			return
		}
	}

	for i := range len(consumidoresGrupo) {
		c := consumidoresGrupo[i]
		puerto := basePort + i
		levantarServidor(c, puerto)
	}

	select {}
}
