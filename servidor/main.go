package main

import (
	"fmt"

	"servidoroper/red_servidor"
)

func main() {
	fmt.Println("************************************")
	fmt.Println("* server oper luna                 *")
	fmt.Println("************************************")

	cfg := red_servidor.LeerConfig()
	red_servidor.IniciarLog(cfg.ArchivoLog)
	red_servidor.Iniciar(cfg)
}
