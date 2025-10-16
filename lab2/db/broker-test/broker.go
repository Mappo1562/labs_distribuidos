package main

import (
	"context"
	"log"
	"time"

	pb "broker/proto"

	"google.golang.org/grpc"
)

func main() {
	// Dirección del nodo DB (ajusta si es otra)
	addr := "localhost:50051"

	// Crear conexión gRPC
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
	if err != nil {
		log.Fatalf("No se pudo conectar al nodo DB (%s): %v", addr, err)
	}
	defer conn.Close()

	client := pb.NewDBNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 🟢 1. Enviar una oferta de prueba
	oferta := &pb.Oferta{
		OfertaId:  "test-007",
		Tienda:    "Falabellox",
		Categoria: "Home",
		Producto:  "aaaa bbb",
		Precio:    20000,
		Stock:     3,
		Fecha:     time.Now().Format("2006-01-02"),
	}

	resp, err := client.Store(ctx, &pb.StoreRequest{Oferta: oferta})
	if err != nil {
		log.Fatalf("Error al enviar oferta: %v", err)
	}
	log.Printf("Store -> OK=%v, msg=%s", resp.Ok, resp.Message)

	// 🟡 2. Leer la oferta recién guardada
	getResp, err := client.Get(ctx, &pb.GetRequest{OfertaId: "test-006"})
	if err != nil {
		log.Fatalf("Error al leer oferta: %v", err)
	}
	if getResp.Found {
		log.Printf("Get -> encontrada: %+v", getResp.Oferta)
	} else {
		log.Printf("Get -> no encontrada")
	}

	// 🔵 3. Listar todas las ofertas
	rangeResp, err := client.RangeSince(ctx, &pb.RangeSinceRequest{SinceUnix: 0})
	if err != nil {
		log.Fatalf("Error en RangeSince: %v", err)
	}
	log.Printf("RangeSince -> total ofertas: %d", len(rangeResp.Ofertas))
	for i, o := range rangeResp.Ofertas {
		log.Printf(" [%d] %s - %s (%d CLP)", i+1, o.Tienda, o.Producto, o.Precio)
	}
}
