package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
)

// Uso: go run crearUsuario.go usuario password >> /etc/usuariosProyecto.txt
func main() {
	if len(os.Args) != 3 {
		fmt.Println("Uso: crearUsuario usuario password")
		os.Exit(1)
	}
	usuario := os.Args[1]
	password := os.Args[2]

	saltBytes := make([]byte, 8)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	resumen := sha512.Sum512([]byte(salt + password))
	fmt.Println(usuario + ":" + salt + ":" + hex.EncodeToString(resumen[:]))
}
