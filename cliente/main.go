package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var ventana fyne.Window
var reporteLabel *widget.Label
var barraCPU, barraRAM, barraDisco *widget.ProgressBar
var salidaTexto *widget.Entry

func main() {
	fmt.Println("****************************")
	fmt.Println(" Cliente oper")
	fmt.Println("****************************")

	if len(os.Args) != 4 {
		fmt.Println("Uso: clienteOper IP Puerto PeriodoReporte")
		fmt.Println("Ej: clienteOper 10.1.28.168 2810 5")
		os.Exit(1)
	}
	ipParam := os.Args[1]
	puertoParam := os.Args[2]
	periodoParam := os.Args[3]

	socketC, err := conectar(ipParam, puertoParam)
	if err != nil {
		fmt.Println("No se pudo conectar al server:", err)
		os.Exit(1)
	}
	fmt.Println("Conectado al server ", socketC.RemoteAddr())

	iniciarGUI(socketC, ipParam+":"+puertoParam, periodoParam)
}

func conectar(ip string, puerto string) (net.Conn, error) {
	dirServerTCP, _ := net.ResolveTCPAddr("tcp4", ip+":"+puerto)
	socketTCP, err := net.DialTCP("tcp", nil, dirServerTCP)
	if err != nil {
		return nil, err
	}
	return socketTCP, nil
}

func autenticar(socketC net.Conn, bufsc *bufio.Reader, usuario string, password string) string {
	bufw := bufio.NewWriter(socketC)
	bufw.WriteString(usuario + "\n")
	bufw.WriteString(password + "\n")
	bufw.Flush()

	respuesta, _ := bufsc.ReadString('\n')
	return strings.TrimSpace(respuesta)
}

func enviarPeriodo(socketC net.Conn, periodo string) {
	bufw := bufio.NewWriter(socketC)
	bufw.WriteString(periodo + "\n")
	bufw.Flush()
}

func enviarComando(socketC net.Conn, comando string) {
	bufsc := bufio.NewWriter(socketC)
	bufsc.WriteString(comando + "\n")
	bufsc.Flush()
}

func recibirMensajes(bufsc *bufio.Reader) {
	for {
		linea, err := bufsc.ReadString('\n')
		if err != nil {
			return
		}
		linea = strings.TrimRight(linea, "\r\n")
		if strings.HasPrefix(linea, "REPORTE|") {
			mostrarReporte(strings.TrimPrefix(linea, "REPORTE|"))
		} else if strings.HasPrefix(linea, "SALIDA|") {
			mostrarSalida(strings.TrimPrefix(linea, "SALIDA|"))
		}
	}
}

func iniciarGUI(socketC net.Conn, direccion string, periodo string) {
	miApp := app.New()
	ventana = miApp.NewWindow("Cliente oper")
	ventana.Resize(fyne.NewSize(700, 450))

	bufsc := bufio.NewReader(socketC)
	mostrarLogin(socketC, bufsc, direccion, periodo, miApp)
	ventana.ShowAndRun()
}

func mostrarLogin(socketC net.Conn, bufsc *bufio.Reader, direccion string, periodo string, miApp fyne.App) {
	usuarioEntry := widget.NewEntry()
	usuarioEntry.SetPlaceHolder("Usuario")
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Password")
	mensajeLogin := widget.NewLabel("")

	botonLogin := widget.NewButton("Ingresar", func() {
		respuesta := autenticar(socketC, bufsc, usuarioEntry.Text, passEntry.Text)

		if respuesta == "OK" {
			enviarPeriodo(socketC, periodo)
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

func mostrarPrincipal(socketC net.Conn, bufsc *bufio.Reader, miApp fyne.App) {
	reporteLabel = widget.NewLabel("Esperando reporte...")
	barraCPU = widget.NewProgressBar()
	barraRAM = widget.NewProgressBar()
	barraDisco = widget.NewProgressBar()
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
		bajarAlFinal()
		enviarComando(socketC, comando)
		comandoEntry.SetText("")

		if comando == "bye" {
			socketC.Close()
			miApp.Quit()
		}
	}

	ventana.SetContent(container.NewBorder(
		container.NewVBox(
			reporteLabel,
			container.NewGridWithColumns(2, widget.NewLabel("CPU"), barraCPU),
			container.NewGridWithColumns(2, widget.NewLabel("RAM"), barraRAM),
			container.NewGridWithColumns(2, widget.NewLabel("Disco"), barraDisco),
			widget.NewSeparator(),
		),
		comandoEntry,
		nil, nil,
		salidaTexto,
	))

	go recibirMensajes(bufsc)
}

func mostrarReporte(texto string) {
	var cpu, ram, disco, procesos int
	inicio := strings.Index(texto, "[")
	if inicio >= 0 {
		fmt.Sscanf(texto[inicio:], "[CPU=%d%% - RAM=%d%% - DD=%d%% - PROC=%d]", &cpu, &ram, &disco, &procesos)
	}
	fyne.Do(func() {
		reporteLabel.SetText(texto)
		barraCPU.SetValue(float64(cpu) / 100)
		barraRAM.SetValue(float64(ram) / 100)
		barraDisco.SetValue(float64(disco) / 100)
	})
}

func mostrarSalida(texto string) {
	fyne.Do(func() {
		salidaTexto.SetText(salidaTexto.Text + texto + "\n")
		bajarAlFinal()
	})
}

func irAlFinal() {
	salidaTexto.TypedKey(&fyne.KeyEvent{Name: fyne.KeyPageDown})
}

func bajarAlFinal() {
	irAlFinal()
	go func() {
		time.Sleep(100 * time.Millisecond)
		fyne.Do(irAlFinal)
	}()
}
