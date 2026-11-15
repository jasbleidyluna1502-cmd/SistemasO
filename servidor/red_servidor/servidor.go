package red_servidor

import (
	"fmt"
	"log"
	"net"
	"os"
)

func Iniciar(cfg Config) {
	dirTCP, _ := net.ResolveTCPAddr("tcp4", ":"+cfg.Puerto)
	oyente, err := net.ListenTCP("tcp", dirTCP)
	if err != nil {
		fmt.Println("No se pudo abrir el puerto", cfg.Puerto, err)
		log.Println("ERROR no se pudo abrir el puerto", cfg.Puerto, err)
		os.Exit(1)
	}
	fmt.Println("Esprando conexiones...")
	log.Println("Servidor iniciado en el puerto", cfg.Puerto)

	for {
		var socketS net.Conn
		socketS, err = oyente.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Se ha conectado cliente ", socketS.RemoteAddr())
		log.Println("Conexion entrante de", socketS.RemoteAddr())
		go atenderCliente(socketS, cfg)
	}
}
