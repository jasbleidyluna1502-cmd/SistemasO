package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const archivoConf = "/etc/configProyecto.conf"

// Datos que se leen de /etc/configProyecto.conf (con valores por defecto)
var ipPermitida string
var puerto = "2810"
var intentos = 3
var usuariosPermitidos []string
var archivoUsuarios = "/etc/usuariosProyecto.txt"
var archivoLog = "/var/log/proyecto/proyecto.log"

func main() {
	if len(os.Args) == 4 && os.Args[1] == "crearusuario" {
		crearUsuario(os.Args[2], os.Args[3])
		return
	}

	fmt.Println("************************************")
	fmt.Println("* server oper luna                 *")
	fmt.Println("************************************")

	leerConfig()
	iniciarLog()
	iniciar()
}

// Imprime la linea usuario:sal:hash para agregarla al archivo de usuarios
func crearUsuario(usuario string, password string) {
	saltBytes := make([]byte, 8)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	resumen := sha256.Sum256([]byte(salt + password))
	fmt.Println(usuario + ":" + salt + ":" + hex.EncodeToString(resumen[:]))
}

func leerConfig() {
	archivo, err := os.Open(archivoConf)
	if err != nil {
		fmt.Println("No se pudo leer", archivoConf, ":", err)
		os.Exit(1)
	}
	defer archivo.Close()

	lector := bufio.NewScanner(archivo)
	for lector.Scan() {
		linea := strings.TrimSpace(lector.Text())
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		partes := strings.SplitN(linea, "=", 2)
		if len(partes) != 2 {
			continue
		}
		clave := strings.TrimSpace(partes[0])
		valor := strings.TrimSpace(partes[1])
		switch clave {
		case "ip_permitida":
			ipPermitida = valor
		case "puerto":
			puerto = valor
		case "intentos_fallidos":
			intentos, _ = strconv.Atoi(valor)
		case "usuarios_permitidos":
			usuariosPermitidos = strings.Split(valor, ",")
		case "archivo_usuarios":
			archivoUsuarios = valor
		case "archivo_log":
			archivoLog = valor
		}
	}
}

func iniciarLog() {
	archLog, err := os.OpenFile(archivoLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("No se pudo abrir el log:", err)
		os.Exit(1)
	}
	log.SetOutput(archLog)
}

func iniciar() {
	dirTCP, _ := net.ResolveTCPAddr("tcp4", ":"+puerto)
	oyente, err := net.ListenTCP("tcp", dirTCP)
	if err != nil {
		fmt.Println("No se pudo abrir el puerto", puerto, err)
		log.Println("ERROR no se pudo abrir el puerto", puerto, err)
		os.Exit(1)
	}
	fmt.Println("Esperando conexiones...")
	log.Println("Servidor iniciado en el puerto", puerto)

	for {
		socketS, err := oyente.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Se ha conectado cliente ", socketS.RemoteAddr())
		log.Println("Conexion entrante de", socketS.RemoteAddr())
		go atenderCliente(socketS)
	}
}

func atenderCliente(socketS net.Conn) {
	defer socketS.Close()

	ipCliente := socketS.RemoteAddr().(*net.TCPAddr).IP.String()
	if ipCliente != ipPermitida {
		log.Println("IP NO permitida:", ipCliente)
		return
	}

	bufsr := bufio.NewReader(socketS)

	usuario := autenticarCliente(socketS, bufsr)
	if usuario == "" {
		return
	}
	log.Println("Usuario autenticado:", usuario)

	periodoStr, _ := bufsr.ReadString('\n')
	periodo, err := strconv.Atoi(strings.TrimSpace(periodoStr))
	if err != nil || periodo < 1 {
		periodo = 5
	}

	// En paralelo: un hilo manda el reporte y otro recibe comandos.
	// El canal fin avisa cuando el cliente se va (bye o desconexion).
	fin := make(chan bool)
	go envMensajeS(socketS, periodo)
	go recMensajeS(socketS, bufsr, usuario, fin)

	<-fin
	log.Println("Cliente desconectado:", usuario)
}

func autenticarCliente(socketS net.Conn, bufsr *bufio.Reader) string {
	for intento := 1; intento <= intentos; intento++ {
		usuario, _ := bufsr.ReadString('\n')
		password, _ := bufsr.ReadString('\n')
		usuario = strings.TrimSpace(usuario)
		password = strings.TrimSpace(password)

		if usuarioPermitido(usuario) && validarPassword(usuario, password) {
			socketS.Write([]byte("OK\n"))
			return usuario
		}
		log.Println("Autenticacion fallida (intento", intento, ") usuario:", usuario)
		if intento < intentos {
			socketS.Write([]byte("FAIL\n"))
		}
	}
	socketS.Write([]byte("BLOQUEADO\n"))
	log.Println("Cliente bloqueado por exceso de intentos fallidos")
	return ""
}

func usuarioPermitido(usuario string) bool {
	for _, u := range usuariosPermitidos {
		if strings.TrimSpace(u) == usuario {
			return true
		}
	}
	return false
}

// Busca usuario:sal:hash en el archivo y compara sha256(sal + password)
func validarPassword(usuario string, password string) bool {
	archivo, err := os.Open(archivoUsuarios)
	if err != nil {
		log.Println("ERROR leyendo base de usuarios:", err)
		return false
	}
	defer archivo.Close()

	lector := bufio.NewScanner(archivo)
	for lector.Scan() {
		partes := strings.Split(strings.TrimSpace(lector.Text()), ":")
		if len(partes) != 3 || partes[0] != usuario {
			continue
		}
		resumen := sha256.Sum256([]byte(partes[1] + password))
		return hex.EncodeToString(resumen[:]) == partes[2]
	}
	return false
}

// Manda el reporte cada n segundos. Termina cuando ya no puede escribir
// en el socket (el cliente se fue y el socket se cerro).
func envMensajeS(socketS net.Conn, periodo int) {
	num := 0
	for {
		num++
		reporte := "REPORTE|Reporte (" + strconv.Itoa(num) + "): " + generarReporte() + "\n"
		_, err := socketS.Write([]byte(reporte))
		if err != nil {
			return
		}
		time.Sleep(time.Second * time.Duration(periodo))
	}
}

// Recibe comandos, los ejecuta y devuelve la respuesta por el socket.
func recMensajeS(socketS net.Conn, bufsr *bufio.Reader, usuario string, fin chan bool) {
	for {
		comando, err := bufsr.ReadString('\n')
		if err != nil {
			fin <- true
			return
		}
		comando = strings.TrimSpace(comando)
		if comando == "" {
			continue
		}
		if comando == "bye" {
			log.Println("Usuario", usuario, "envio bye")
			fin <- true
			return
		}
		log.Println("Usuario", usuario, "ejecuta:", comando)
		socketS.Write([]byte(formatoSalida(ejecutar(comando))))
	}
}

// Pone SALIDA| delante de cada linea para que el cliente la distinga del reporte
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

// Ejecuta el comando con sh -c y un limite de 30 segundos
func ejecutar(comando string) string {
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	cmd := exec.CommandContext(ctx, "sh", "-c", comando)
	// El comando corre en su propio grupo de procesos; al vencer el tiempo
	// se mata el grupo completo (sh y sus hijos) y no solo el sh.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = 2 * time.Second
	respuesta, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(respuesta) + "\n(comando cancelado: supero 30 segundos)"
	}
	if err != nil && len(respuesta) == 0 {
		return "error: " + err.Error()
	}
	return string(respuesta)
}

func generarReporte() string {
	return fmt.Sprintf("[CPU=%d%% - RAM=%d%% - DD=%d%% - PROC=%d]",
		porcentajeCPU(), porcentajeRAM(), porcentajeDisco(), contarProcesos())
}

// Lee la primera linea de /proc/stat: tiempo total y tiempo libre de la CPU
func leerCPU() (uint64, uint64) {
	archivo, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0
	}
	defer archivo.Close()
	lector := bufio.NewScanner(archivo)
	lector.Scan()
	campos := strings.Fields(lector.Text())[1:]
	var total, libre uint64
	for i, c := range campos {
		valor, _ := strconv.ParseUint(c, 10, 64)
		total += valor
		if i == 3 || i == 4 {
			libre += valor
		}
	}
	return total, libre
}

func porcentajeCPU() int {
	total1, libre1 := leerCPU()
	time.Sleep(time.Millisecond * 500)
	total2, libre2 := leerCPU()
	dTotal := total2 - total1
	if dTotal == 0 {
		return 0
	}
	return int(100 * (dTotal - (libre2 - libre1)) / dTotal)
}

func porcentajeRAM() int {
	archivo, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer archivo.Close()
	var total, disponible uint64
	lector := bufio.NewScanner(archivo)
	for lector.Scan() {
		campos := strings.Fields(lector.Text())
		if len(campos) < 2 {
			continue
		}
		valor, _ := strconv.ParseUint(campos[1], 10, 64)
		if campos[0] == "MemTotal:" {
			total = valor
		} else if campos[0] == "MemAvailable:" {
			disponible = valor
		}
	}
	if total == 0 {
		return 0
	}
	return int(100 * (total - disponible) / total)
}

func porcentajeDisco() int {
	var st syscall.Statfs_t
	if syscall.Statfs("/", &st) != nil || st.Blocks == 0 {
		return 0
	}
	return int(100 * (st.Blocks - st.Bavail) / st.Blocks)
}

// Cada carpeta con numero dentro de /proc es un proceso
func contarProcesos() int {
	entradas, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	cuenta := 0
	for _, e := range entradas {
		if _, err := strconv.Atoi(e.Name()); err == nil {
			cuenta++
		}
	}
	return cuenta
}
