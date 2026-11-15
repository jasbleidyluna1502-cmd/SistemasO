package ui

import (
	"bufio"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var ventana fyne.Window

func Iniciar(socketC net.Conn, direccion string, periodo string) {
	miApp := app.New()
	ventana = miApp.NewWindow("Cliente oper")
	ventana.Resize(fyne.NewSize(700, 450))

	bufsc := bufio.NewReader(socketC)
	mostrarLogin(socketC, bufsc, direccion, periodo, miApp)
	ventana.ShowAndRun()
}
