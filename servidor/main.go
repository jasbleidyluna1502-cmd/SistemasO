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
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const archivoConf = "/etc/configProyecto.conf"

type Config struct {
	IPPermitida     string
	Puerto          string
	Intentos        int
	Usuarios        []string
	ArchivoUsuarios string
	ArchivoLog      string
}

var listaNegra = map[string]bool{
	"su": true, "sudo": true, "passwd": true,
	"rm": true, "rmdir": true, "dd": true, "mkfs": true, "shred": true,
	"shutdown": true, "reboot": true, "halt": true, "poweroff": true, "init": true,
	"systemctl": true, "service": true,
	"kill": true, "killall": true, "pkill": true,
	"chmod": true, "chown": true, "mount": true, "umount": true,
	"useradd": true, "userdel": true, "usermod": true,
	"vi": true, "vim": true, "nano": true,
}

func main() {
	if len(os.Args) == 4 && os.Args[1] == "crearusuario" {
		crearUsuario(os.Args[2], os.Args[3])
		return
	}

	fmt.Println("************************************")
	fmt.Println("* server oper luna                 *")
	fmt.Println("************************************")

	cfg := leerConfig()
	iniciarLog(cfg.ArchivoLog)
	iniciar(cfg)
}

func crearUsuario(usuario string, password string) {
	saltBytes := make([]byte, 8)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	resumen := sha256.Sum256([]byte(salt + password))
	fmt.Println(usuario + ":" + salt + ":" + hex.EncodeToString(resumen[:]))
}

func leerConfig() Config {
	cfg := Config{
		Puerto:          "2810",
		Intentos:        3,
		ArchivoUsuarios: "/etc/usuariosProyecto.txt",
		ArchivoLog:      "/var/log/proyecto/proyecto.log",
	}
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
			cfg.IPPermitida = valor
		case "puerto":
			cfg.Puerto = valor
		case "intentos_fallidos":
			cfg.Intentos, _ = strconv.Atoi(valor)
		case "usuarios_permitidos":
			cfg.Usuarios = strings.Split(valor, ",")
		case "archivo_usuarios":
			cfg.ArchivoUsuarios = valor
		case "archivo_log":
			cfg.ArchivoLog = valor
		}
	}
	return cfg
}

func iniciarLog(rutaLog string) {
	archLog, err := os.OpenFile(rutaLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("No se pudo abrir el log:", err)
		os.Exit(1)
	}
	log.SetOutput(archLog)
}

func iniciar(cfg Config) {
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

func atenderCliente(socketS net.Conn, cfg Config) {
	defer socketS.Close()

	ipCliente := socketS.RemoteAddr().(*net.TCPAddr).IP.String()
	if ipCliente != cfg.IPPermitida {
		log.Println("IP NO permitida:", ipCliente)
		return
	}

	bufsr := bufio.NewReader(socketS)

	usuario := autenticarCliente(socketS, bufsr, cfg)
	if usuario == "" {
		return
	}
	log.Println("Usuario autenticado:", usuario)

	periodoStr, _ := bufsr.ReadString('\n')
	periodo, err := strconv.Atoi(strings.TrimSpace(periodoStr))
	if err != nil || periodo < 1 {
		periodo = 5
	}

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

		if usuarioPermitido(usuario, cfg) && validarPassword(usuario, password, cfg.ArchivoUsuarios) {
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

func validarPassword(usuario string, password string, archivoUsuarios string) bool {
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

func envMensajeS(salida chan string, periodo int, done chan bool) {
	num := 0
	for {
		num++
		reporteEnv := "REPORTE|Reporte (" + strconv.Itoa(num) + "): " + generarReporte() + "\n"
		select {
		case salida <- reporteEnv:
		case <-done:
			return
		}
		select {
		case <-time.After(time.Second * time.Duration(periodo)):
		case <-done:
			return
		}
	}
}

func recMensajeS(bufss *bufio.Reader, salida chan string, done chan bool, usuario string) {
	for {
		comandoRec, err := bufss.ReadString('\n')
		if err != nil {
			close(done)
			return
		}
		comandoRec = strings.TrimSpace(comandoRec)
		if comandoRec == "" {
			continue
		}
		if comandoRec == "bye" {
			log.Println("Usuario", usuario, "envio bye")
			close(done)
			return
		}
		log.Println("Usuario", usuario, "ejecuta:", comandoRec)
		salida <- formatoSalida(ejecutar(comandoRec))
	}
}

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

func comandoBloqueado(comando string) string {
	separadores := func(r rune) bool { return r == ';' || r == '|' || r == '&' || r == '\n' || r == '(' || r == ')' }
	for _, tramo := range strings.FieldsFunc(comando, separadores) {
		palabras := strings.Fields(tramo)
		if len(palabras) == 0 {
			continue
		}
		nombre := filepath.Base(palabras[0])
		if listaNegra[nombre] {
			return nombre
		}
	}
	return ""
}

func ejecutar(comando string) string {
	if bloqueado := comandoBloqueado(comando); bloqueado != "" {
		log.Println("Comando bloqueado por la lista negra:", bloqueado)
		return "comando no permitido: " + bloqueado
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	respuesta, err := exec.CommandContext(ctx, "sh", "-c", comando).CombinedOutput()
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

func leerCPU() (uint64, uint64) {
	archivo, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0
	}
	defer archivo.Close()
	lector := bufio.NewScanner(archivo)
	lector.Scan()
	campos := strings.Fields(lector.Text())[1:]
	var total, idle uint64
	for i, c := range campos {
		valor, _ := strconv.ParseUint(c, 10, 64)
		total += valor
		if i == 3 || i == 4 {
			idle += valor
		}
	}
	return total, idle
}

func porcentajeCPU() int {
	total1, idle1 := leerCPU()
	time.Sleep(time.Millisecond * 500)
	total2, idle2 := leerCPU()
	dTotal := total2 - total1
	if dTotal == 0 {
		return 0
	}
	return int(100 * (dTotal - (idle2 - idle1)) / dTotal)
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
