package main

import (
	"context"
	"fmt"
	"log"
	pb "productores/proto"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// Conectarse al broker
	conn, err := grpc.Dial("broker:50050", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()

	client := pb.NewBrokerClient(conn)

	// Registro del productor
	registro := &pb.Registro{
		Nombre: "Riploy",
		Rol:    0, // 0 = Tienda
	}

	resp, err := client.Registrarse(context.Background(), registro)
	if err != nil {
		log.Fatalf("Error al registrarse: %v", err)
	}
	if resp.Flag {
		fmt.Println("Riploy registrado en el broker")
	} else {
		fmt.Println("Falló el registro")
		return
	}

	// Enviar oferta de ejemplo
	oferta := &pb.Oferta{
		OfertaId:  "123-uuid",
		Tienda:    "Riploy",
		Categoria: "Electronica",
		Producto:  "Laptop X",
		Precio:    499990,
		Stock:     10,
		Fecha:     time.Now().Format("2006-01-02 15:04:05"),
	}

	respOferta, err := client.GenerarOferta(context.Background(), oferta)
	if err != nil {
		log.Fatalf("Error al enviar oferta: %v", err)
	}

	if respOferta.Flag {
		fmt.Println("[✓] Oferta enviada correctamente")
	} else {
		fmt.Println("[✗] Broker rechazó la oferta")
	}
}
