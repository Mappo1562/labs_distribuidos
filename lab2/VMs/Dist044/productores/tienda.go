package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	pb "productores/proto"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PortBroker = "dist101.inf.santiago.usm.cl:50050"
)

type Producto struct {
	producto_id     string
	tienda          string
	categoria       string
	producto_nombre string
	precio_base     int64
	stock           int64
}

var catalogo []Producto
var nombreTienda string
var r *rand.Rand

func leerCatalogo(path string) ([]Producto, error) {

	archivo, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)

	filas, err := lector.ReadAll()

	if err != nil {
		return nil, err
	}

	var productos []Producto
	for i, fila := range filas {
		if i == 0 {
			continue
		}

		precio, err := strconv.ParseInt(fila[4], 10, 64)

		if err != nil {
			log.Printf("Error convirtiendo precio en linea %d: %v", i+1, err)
			continue
		}

		unidades, err := strconv.ParseInt(fila[5], 10, 64)

		if err != nil {
			log.Printf("Error convirtiendo stock en linea %d: %v", i+1, err)
		}

		productos = append(productos, Producto{
			producto_id:     fila[0],
			tienda:          fila[1],
			categoria:       fila[2],
			producto_nombre: fila[3],
			precio_base:     precio,
			stock:           unidades,
		})
	}

	return productos, nil
}

func elegirProducto() *Producto {

	i := r.Intn(len(catalogo))
	return &catalogo[i]
}

func aplicarDescuento(precioBase int64) int64 {
	desc := r.Intn(41) + 10
	precio_final := precioBase * int64(100-desc) / 100

	return precio_final
}

func generarStock(stockActual int64) int64 {
	stock_oferta := r.Intn(int(stockActual)) + 1

	return int64(stock_oferta)
}

func generarID() string {
	t := time.Now().UnixNano()
	ro := r.Intn(100000000000)
	tiempo := strconv.FormatInt(t, 36)
	random := strconv.FormatInt(int64(ro), 36)

	return fmt.Sprintf("%s-%s", tiempo, random)
}

func generarOfertaAleatoria() *pb.Oferta {

	if len(catalogo) == 0 {
		return nil
	}

	intentos := len(catalogo) * 2

	for i := 0; i < intentos; i++ {
		p := elegirProducto()

		if p.stock > 0 {
			precio_oferta := aplicarDescuento(p.precio_base)
			stock := generarStock(p.stock)
			id := generarID()

			p.stock -= stock
			return &pb.Oferta{
				OfertaId:  id,
				Tienda:    p.tienda,
				Categoria: p.categoria,
				Producto:  p.producto_nombre,
				Precio:    precio_oferta,
				Stock:     stock,
				Fecha:     time.Now().Format("02/01/2006 15:04:05.000"),
			}
		}

	}

	return nil
}

func nuevoGenerador(nombre string) *rand.Rand {

	seed := int64(0)

	for _, c := range nombre {
		seed += int64(c)
	}

	src := rand.NewSource(seed)
	return rand.New(src)
}

func main() {
	log.SetOutput(os.Stdout)

	nombreTienda = os.Getenv("NOMBRE_TIENDA")
	path := os.Getenv("CATALOGO")

	if nombreTienda == "" || path == "" {
		log.Fatal("Debes definir NOMBRE_TIENDA y CATALOGO como variables de entorno en docker-compose.yml")
	}
	var err error
	catalogo, err = leerCatalogo(path)

	if err != nil {
		log.Fatalf("Error leyendo el catalogo de la tienda %s: %v", nombreTienda, err)
	}

	r = nuevoGenerador(nombreTienda)
	// Conectarse al broker

	conn, err := grpc.NewClient(PortBroker, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()

	client := pb.NewBrokerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	// Registro del productor
	registro := &pb.Registro{
		Nombre: nombreTienda,
		Rol:    0, // 0 = Tienda
	}

	resp, err := client.Registrarse(ctx, registro)
	if err != nil {
		log.Fatalf("Error al registrarse: %v", err)
	}
	if resp.Flag {
		log.Printf("%s registrado en el broker\n", nombreTienda)
	} else {
		log.Println("Falló el registro")
		return
	}

	time.Sleep(time.Second * 2) //permite que todos alcancen a registrarse
	for {
		oferta := generarOfertaAleatoria()
		if oferta == nil {
			log.Printf("%s no tiene más stock", nombreTienda)
			break
		}

		respOferta, err := client.GenerarOferta(context.Background(), oferta)
		if err != nil {
			log.Fatalf("Error al enviar oferta: %v", err)
		}

		if respOferta.Flag {
			log.Println("Oferta enviada correctamente")
		} else {
			log.Println("Broker rechazó la oferta")
		}

		time.Sleep(time.Duration(r.Intn(3)+2) * time.Second)
	}
}
