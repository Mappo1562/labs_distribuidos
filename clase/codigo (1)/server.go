package main

import (
	"fmt"
	"net"
	"os"
)

const (
	CONN_HOST = "0.0.0.0"	
	CONN_TYPE = "tcp"	
)



func main() {
	CONN_PORT := os.Getenv("SERVER_PORT")
	if CONN_PORT == "" {
		CONN_PORT = "8080"
	}
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error al escuchar:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("[°] Servidor escuchando en " + CONN_HOST + ":" + CONN_PORT)

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error al aceptar:", err.Error())
			os.Exit(1)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	fmt.Println("[°] Conexión entrante de:", conn.RemoteAddr().String())

	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error al leer:", err.Error())
		return
	}
	fmt.Println("[°] Mensaje recibido:", string(buf[:reqLen]))

	conn.Write([]byte("Hello client!"))
}