package red_servidor

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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

func LeerConfig() Config {
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
