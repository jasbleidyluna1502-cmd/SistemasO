package ui

import (
	"bufio"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"clienteoper/red_cliente"
)

func mostrarLogin(socketC net.Conn, bufsc *bufio.Reader, direccion string, periodo string, miApp fyne.App) {
	usuarioEntry := widget.NewEntry()
	usuarioEntry.SetPlaceHolder("Usuario")
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Password")
	mensajeLogin := widget.NewLabel("")

	botonLogin := widget.NewButton("Ingresar", func() {
		respuesta := red_cliente.Autenticar(socketC, bufsc, usuarioEntry.Text, passEntry.Text)

		if respuesta == "OK" {
			red_cliente.EnviarPeriodo(socketC, periodo)
			mostrarPrincipal(socketC, bufsc, miApp)
		} else if respuesta == "BLOQUEADO" {
			mensajeLogin.SetText("Demasiados intentos fallidos. Conexion cerrada.")
			socketC.Close()
		} else {
			mensajeLogin.SetText("Usuario o password incorrectos")
		}
	})

	ventana.SetContent(container.NewVBox(
		widget.NewLabel("Acceso al servidor "+direccion),
		usuarioEntry,
		passEntry,
		botonLogin,
		mensajeLogin,
	))
}
