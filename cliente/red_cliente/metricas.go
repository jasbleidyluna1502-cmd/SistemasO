package red_cliente

import (
	"bufio"
	"strings"
)

// Lee lo que llega del server: lineas REPORTE|... y SALIDA|...
func RecibirMensajes(bufsc *bufio.Reader, alReporte func(string), alSalida func(string)) {
	for {
		linea, err := bufsc.ReadString('\n')
		if err != nil {
			return
		}
		linea = strings.TrimRight(linea, "\r\n")
		if strings.HasPrefix(linea, "REPORTE|") {
			alReporte(strings.TrimPrefix(linea, "REPORTE|"))
		} else if strings.HasPrefix(linea, "SALIDA|") {
			alSalida(strings.TrimPrefix(linea, "SALIDA|"))
		}
	}
}
