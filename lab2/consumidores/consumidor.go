package main

import (
	pb "consumidores/proto"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
}

var consumidores []Consumidor

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
func main() {

	conn, err := grpc.NewClient("broker:50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}

	defer conn.Close()

	cliente := pb.NewBrokerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	consumidores, err = leerConsumidores("consumidores.csv")

	if err != nil {
		log.Fatalf("Error leyendo los consumidores: %v", err)
	}

	grupo := 1
	inicio := (grupo - 1) * 4
	fin := inicio + 4
	consumidoresGrupo := consumidores[inicio:fin]
	_ = consumidoresGrupo

	for i := range len(consumidores) {
		c := consumidores[i]

		registro := &pb.Registro{
			Nombre: c.id_consumidor,
			Rol:    1,
		}

		resp, err := cliente.Registrarse(ctx, registro)
		if err != nil {
			log.Fatalf("Error al registrarse: %v", err)
		}
		if resp.Flag {
			fmt.Printf("%s registrado en el broker", c.id_consumidor)
		} else {
			fmt.Println("Falló el registro")
			return
		}
	}
}
