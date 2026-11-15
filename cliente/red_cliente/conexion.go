package red_cliente

import (
	"bufio"
	"net"
	"strings"
)

func Conectar(ip string, puerto string) (net.Conn, error) {
	dirServerTCP, _ := net.ResolveTCPAddr("tcp4", ip+":"+puerto)
	socketTCP, err := net.DialTCP("tcp", nil, dirServerTCP)
	if err != nil {
		return nil, err
	}
	return socketTCP, nil
}

// Envia usuario y password. El server responde: OK, FAIL o BLOQUEADO
func Autenticar(socketC net.Conn, bufsc *bufio.Reader, usuario string, password string) string {
	bufw := bufio.NewWriter(socketC)
	bufw.WriteString(usuario + "\n")
	bufw.WriteString(password + "\n")
	bufw.Flush()

	respuesta, _ := bufsc.ReadString('\n')
	return strings.TrimSpace(respuesta)
}

// Cada cuantos segundos quiere el cliente el reporte
func EnviarPeriodo(socketC net.Conn, periodo string) {
	bufw := bufio.NewWriter(socketC)
	bufw.WriteString(periodo + "\n")
	bufw.Flush()
}
