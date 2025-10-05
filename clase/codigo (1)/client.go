package main

import (
	"fmt"
	"net"
	"os"
)

const (
	CONN_HOST = "localhost" 
	CONN_TYPE = "tcp"	
)

func main() {	
	CONN_PORT := "8086"
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error al conectar:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	message := "Hello server!"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error al enviar mensaje:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[°] Mensaje enviado al servidor:", message)

	buf := make([]byte, 1024)
	resLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error al leer respuesta:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[°] Respuesta del servidor:", string(buf[:resLen]))
}