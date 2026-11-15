package ui

import (
	"bufio"
	"net"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"clienteoper/red_cliente"
)

var reporteLabel *widget.Label
var salidaTexto *widget.Entry

func mostrarPrincipal(socketC net.Conn, bufsc *bufio.Reader, miApp fyne.App) {
	reporteLabel = widget.NewLabel("Esperando reporte...")
	salidaTexto = widget.NewMultiLineEntry()
	salidaTexto.Wrapping = fyne.TextWrapWord

	comandoEntry := widget.NewEntry()
	comandoEntry.SetPlaceHolder("Digite el comando a enviar (bye para salir)")
	comandoEntry.OnSubmitted = func(comando string) {
		comando = strings.TrimSpace(comando)
		if comando == "" {
			return
		}
		salidaTexto.SetText(salidaTexto.Text + "$ " + comando + "\n")
		red_cliente.EnviarComando(socketC, comando)
		comandoEntry.SetText("")

		if comando == "bye" {
			socketC.Close()
			miApp.Quit()
		}
	}

	ventana.SetContent(container.NewBorder(
		container.NewVBox(reporteLabel, widget.NewSeparator()),
		comandoEntry,
		nil, nil,
		container.NewScroll(salidaTexto),
	))

	// En paralelo: recibir reportes y respuestas del server
	go red_cliente.RecibirMensajes(bufsc, mostrarReporte, mostrarSalida)
}

func mostrarReporte(texto string) {
	fyne.Do(func() {
		reporteLabel.SetText(texto)
	})
}

func mostrarSalida(texto string) {
	fyne.Do(func() {
		salidaTexto.SetText(salidaTexto.Text + texto + "\n")
	})
}
