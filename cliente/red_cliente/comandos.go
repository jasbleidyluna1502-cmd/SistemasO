package red_cliente

import (
	"bufio"
	"net"
)

func EnviarComando(socketC net.Conn, comando string) {
	bufsc := bufio.NewWriter(socketC)
	bufsc.WriteString(comando + "\n")
	bufsc.Flush()
}
