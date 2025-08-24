package main

import (
    "fmt"
)
func main() {

    // --- Imprimir en consola ---
    fmt.Println("Hola")        // imprime: Hola
    fmt.Print("Hola")          // imprime sin salto de línea
    fmt.Printf("Hola %s\n", "Mundo") // imprime con formato (como printf en C)

    // --- Variables ---
    var x int = 10             // declaración explícita
    y := 20                    // declaración corta (Go infiere el tipo)
    fmt.Println(x, y)          // imprime: 10 20

    // --- Constantes ---
    const Pi = 3.1416
    fmt.Println(Pi)            // imprime: 3.1416

    // --- Arreglos y slices ---
    arr := [3]int{1, 2, 3}     // array de tamaño fijo
    slice := []int{4, 5, 6}    // slice (dinámico)
    fmt.Println(arr, slice)    // imprime: [1 2 3] [4 5 6]

    // --- Mapas (diccionarios) ---
    m := map[string]int{"a": 1, "b": 2}
    fmt.Println(m["a"])        // imprime: 1

    // --- Condicionales ---
    if y > x {
        fmt.Println("y es mayor que x")
    }

    // --- Bucles ---
    for i := 0; i < 3; i++ {
        fmt.Println("i:", i)
    }

    // --- Funciones ---
    fmt.Println(sumar(5, 7))   // imprime: 12

    // --- Goroutines (concurrencia) ---
    go fmt.Println("Hola desde goroutine") // se ejecuta en paralelo

    // --- Canales ---
    c := make(chan string)
    go func() { c <- "mensaje por canal" }()
    msg := <-c
    fmt.Println(msg)           // imprime: mensaje por canal
}

// Función que suma dos enteros
func sumar(a int, b int) int {
    return a + b
}