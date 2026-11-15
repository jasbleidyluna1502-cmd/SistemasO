package red_servidor

import (
	"fmt"
	"log"
	"os"
)

// Desde aqui todo log.Println() se escribe en el archivo de log
func IniciarLog(rutaLog string) {
	archLog, err := os.OpenFile(rutaLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("No se pudo abrir el log:", err)
		os.Exit(1)
	}
	log.SetOutput(archLog)
}
