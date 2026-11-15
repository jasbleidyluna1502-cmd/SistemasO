package red_servidor

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"servidoroper/auth"
	"servidoroper/monitor"
	"servidoroper/shell"
)

func atenderCliente(socketS net.Conn, cfg Config) {
	defer socketS.Close()

	// 1. Solo se acepta la IP permitida
	ipCliente := socketS.RemoteAddr().(*net.TCPAddr).IP.String()
	if ipCliente != cfg.IPPermitida {
		log.Println("IP NO permitida:", ipCliente)
		return
	}

	bufsr := bufio.NewReader(socketS)

	// 2. Autenticacion
	usuario := autenticarCliente(socketS, bufsr, cfg)
	if usuario == "" {
		return
	}
	log.Println("Usuario autenticado:", usuario)

	// 3. El cliente envia cada cuantos segundos quiere el reporte
	periodoStr, _ := bufsr.ReadString('\n')
	periodo, err := strconv.Atoi(strings.TrimSpace(periodoStr))
	if err != nil || periodo < 1 {
		periodo = 5
	}

	// 4. En paralelo: escribir, enviar reportes y recibir comandos
	salida := make(chan string, 10)
	done := make(chan bool)

	go escritorS(&socketS, salida, done)
	go envMensajeS(salida, periodo, done)
	go recMensajeS(bufsr, salida, done, usuario)

	<-done
	log.Println("Cliente desconectado:", usuario)
}

func autenticarCliente(socketS net.Conn, bufsr *bufio.Reader, cfg Config) string {
	for intento := 1; intento <= cfg.Intentos; intento++ {
		usuario, _ := bufsr.ReadString('\n')
		password, _ := bufsr.ReadString('\n')
		usuario = strings.TrimSpace(usuario)
		password = strings.TrimSpace(password)

		if usuarioPermitido(usuario, cfg) && auth.ValidarPassword(usuario, password, cfg.ArchivoUsuarios) {
			socketS.Write([]byte("OK\n"))
			return usuario
		}
		log.Println("Autenticacion fallida (intento", intento, ") usuario:", usuario)
		if intento < cfg.Intentos {
			socketS.Write([]byte("FAIL\n"))
		}
	}
	socketS.Write([]byte("BLOQUEADO\n"))
	log.Println("Cliente bloqueado por exceso de intentos fallidos")
	return ""
}

func usuarioPermitido(usuario string, cfg Config) bool {
	for _, u := range cfg.Usuarios {
		if strings.TrimSpace(u) == usuario {
			return true
		}
	}
	return false
}

// Unica goroutine que escribe en el socket: recibe los mensajes por el canal
func escritorS(socketS *net.Conn, salida chan string, done chan bool) {
	bufss := bufio.NewWriter(*socketS)
	for {
		select {
		case mensaje := <-salida:
			bufss.WriteString(mensaje)
			bufss.Flush()
		case <-done:
			return
		}
	}
}

// Envia el reporte de recursos cada "periodo" segundos
func envMensajeS(salida chan string, periodo int, done chan bool) {
	num := 0
	for {
		num++
		reporteEnv := "REPORTE|Reporte (" + strconv.Itoa(num) + "): " + monitor.GenerarReporte() + "\n"
		select {
		case salida <- reporteEnv:
		case <-done:
			return
		}
		fmt.Println("Mensaje enviado!")
		fmt.Println("======================================")
		select {
		case <-time.After(time.Second * time.Duration(periodo)):
		case <-done:
			return
		}
	}
}

// Recibe los comandos, los ejecuta y devuelve la respuesta
func recMensajeS(bufss *bufio.Reader, salida chan string, done chan bool, usuario string) {
	for {
		comandoRec, err := bufss.ReadString('\n')
		if err != nil {
			close(done)
			return
		}
		comandoRec = strings.TrimSpace(comandoRec)
		fmt.Println("Comando Rec : ", comandoRec)
		if comandoRec == "" {
			continue
		}
		if comandoRec == "bye" {
			log.Println("Usuario", usuario, "envio bye")
			close(done)
			return
		}
		log.Println("Usuario", usuario, "ejecuta:", comandoRec)
		salida <- formatoSalida(shell.Ejecutar(comandoRec))
	}
}

// Cada linea de la respuesta se envia como: SALIDA|linea
func formatoSalida(texto string) string {
	texto = strings.TrimRight(texto, "\n")
	if texto == "" {
		texto = "(sin salida)"
	}
	resultado := ""
	for _, linea := range strings.Split(texto, "\n") {
		resultado += "SALIDA|" + linea + "\n"
	}
	return resultado
}
