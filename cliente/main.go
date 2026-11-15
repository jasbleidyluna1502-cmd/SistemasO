package main

import (
	"fmt"
	"os"

	"clienteoper/red_cliente"
	"clienteoper/ui"
)

func main() {
	fmt.Println("****************************")
	fmt.Println(" Cliente oper")
	fmt.Println("****************************")

	if len(os.Args) != 4 {
		fmt.Println("Uso: clienteOper IP Puerto PeriodoReporte")
		fmt.Println("Ej: clienteOper 10.1.28.168 2810 5")
		os.Exit(1)
	}
	ipParam := os.Args[1]
	puertoParam := os.Args[2]
	periodoParam := os.Args[3]

	socketC, err := red_cliente.Conectar(ipParam, puertoParam)
	if err != nil {
		fmt.Println("No se pudo conectar al server:", err)
		os.Exit(1)
	}
	fmt.Println("Conectado al server ", socketC.RemoteAddr())

	ui.Iniciar(socketC, ipParam+":"+puertoParam, periodoParam)
}
