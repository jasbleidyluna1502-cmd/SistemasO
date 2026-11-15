package auth

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

// Archivo plano: usuario:salt:hash   (hash = sha512(salt + password) en hexadecimal)
func ValidarPassword(usuario string, password string, archivoUsuarios string) bool {
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
		resumen := sha512.Sum512([]byte(partes[1] + password))
		return hex.EncodeToString(resumen[:]) == partes[2]
	}
	return false
}
