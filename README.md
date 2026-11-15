# Remote Shell and System Resource Monitoring

Proyecto final de Sistemas Operativos (Universidad de los Llanos). Servidor Linux (Go) y cliente con interfaz gráfica (Go + Fyne, Windows/Linux).

## Estructura
- `servidor/` servidor Linux: `red_servidor` (socket, config, log), `auth` (SHA-512 + salt), `monitor` (CPU/RAM/disco/procesos), `shell` (ejecución de comandos), `crearUsuario` (genera líneas del archivo de usuarios).
- `cliente/` cliente GUI: `red_cliente` (conexión) y `ui` (login y pantalla principal).
- `conf/configProyecto.conf` y `conf/usuarios.txt` (plantillas de configuración y usuarios).
- `systemd/proyecto.service` servicio systemd. `instalar.sh` despliega todo en el servidor Linux.
- `bin/servidorOper` binario Linux amd64 ya compilado. `paper/` y `presentacion/` documentos de entrega.

## Usuarios de prueba (solo local, para el profesor)
| usuario | password |
|---|---|
| luna | luna2026 |
| leidy | leidy2026 |

Formato del archivo: `usuario:salt:hash`, con `hash = SHA512(salt + password)` en hexadecimal.
Nuevo usuario: `cd servidor && go run ./crearUsuario usuario password >> /etc/usuariosProyecto.txt`

## Despliegue del servidor (en la máquina Linux)
1. Editar `conf/configProyecto.conf`: poner en `ip_permitida` la IP del equipo cliente.
2. Copiar la carpeta del proyecto a Linux y ejecutar: `sudo bash instalar.sh`
3. Ver estado: `systemctl status proyecto` — log: `/var/log/proyecto/proyecto.log`

Compilar el servidor (si se cambia el código): `cd servidor && GOOS=linux GOARCH=amd64 go build -o ../bin/servidorOper .`

## Cliente (Windows)
```
cd cliente
go build -o clienteOper.exe .
clienteOper.exe <IP_servidor> 2810 5
```
Parámetros: `IP Puerto PeriodoReporte(segundos)`. Se abre la ventana, se ingresa usuario/password; se ven el reporte periódico y se pueden escribir comandos Unix. `bye` cierra la sesión.

## Protocolo (texto, una línea por mensaje)
Cliente → servidor: `usuario\n`, `password\n`, `periodo\n`, luego comandos. Servidor → cliente: `OK` / `FAIL` / `BLOQUEADO`, luego `REPORTE|...` y `SALIDA|...`.

## Git
```
git init
git add .
git commit -m "Proyecto final SO"
git branch -M main
git remote add origin <URL_DEL_REPOSITORIO>
git push -u origin main
```
