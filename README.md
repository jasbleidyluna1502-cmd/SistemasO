# Remote Shell and System Resource Monitoring

Proyecto final de Sistemas Operativos (Universidad de los Llanos). Servidor Linux y cliente gráfico (Fyne, Windows/Linux), ambos en Go.

## Estructura
```
servidor/
  main.go             servidor: socket, configuración, log, autenticación SHA-256 + sal, monitoreo y ejecución de comandos (límite de 30 s)
  go.mod
  servidorOper        binario Linux amd64 ya compilado
  configProyecto.conf copiar a /etc/configProyecto.conf
  usuarios.txt        copiar a /etc/usuariosProyecto.txt (usuario:sal:hash)
  proyecto.service    servicio systemd (usuario proyectoso)
  instalar.sh         instala todo en el servidor Linux
cliente/
  main.go             cliente gráfico: conexión, login, reporte con barras y consola
  go.mod, go.sum
paper/paper_IEEE.docx
```

## Usuarios de prueba
| usuario | password |
|---|---|
| luna | luna2026 |
| leidy | leidy2026 |

Formato: `usuario:sal:hash`, con `hash = SHA256(sal + password)` en hexadecimal. Solo entran los usuarios de `usuarios_permitidos` en la configuración.
Crear un usuario nuevo (en Linux): `./servidorOper crearusuario nombre password | sudo tee -a /etc/usuariosProyecto.txt`

## Servidor (Linux)
1. En `servidor/configProyecto.conf`, poner en `ip_permitida` la IP del equipo cliente.
2. En la carpeta `servidor/`: `sudo bash instalar.sh`
3. Estado: `systemctl status proyecto --no-pager`. Log: `sudo tail -n 30 /var/log/proyecto/proyecto.log`

Recompilar: `cd servidor && GOOS=linux GOARCH=amd64 go build -o servidorOper .`

## Cliente (Windows)
```
cd cliente
go build -o clienteOper.exe .
.\clienteOper.exe IP_DEL_SERVIDOR 2810 5
```
Parámetros: `IP Puerto PeriodoReporte(segundos)`. Se abre la ventana, se ingresa usuario y password, y se ven el reporte con barras de CPU, RAM y disco y la consola de comandos. `bye` cierra la sesión.

## Protocolo (líneas de texto)
Cliente → servidor: usuario, password, periodo y luego comandos. Servidor → cliente: `OK` / `FAIL` / `BLOQUEADO`, luego `REPORTE|...` y `SALIDA|...`.

## Seguridad
El servidor corre como el usuario `proyectoso`, sin privilegios: los comandos que necesiten permisos de administrador (borrar archivos del sistema, leer `/etc/shadow`, apagar la máquina) fallan con "permiso denegado".

## Comandos para probar (escribir en el cliente)
Cada comando corre independiente con `sh -c` (un `cd` no se conserva; use `cd /etc && ls`). Máximo 30 segundos por comando. Evite comandos interactivos.

| Prueba | Comandos |
|---|---|
| Sistema | `hostname`, `whoami`, `id`, `date`, `uptime`, `uname -a`, `cat /etc/os-release` |
| Recursos | `free -h`, `df -h`, `top -bn1 \| head -n 15`, `ps aux \| head -n 10`, `ps aux \| wc -l`, `cat /proc/loadavg`, `lscpu \| head -n 10` |
| Red | `ip a`, `ss -tlnp` (debe verse el 2810), `ss -tn` |
| Archivos | `ls -l /etc \| head -n 15`, `cat /etc/configProyecto.conf`, `echo prueba > /tmp/prueba.txt && cat /tmp/prueba.txt` |
| Sin privilegios (deben fallar) | `cat /etc/shadow`, `touch /etc/prueba`, `ls /root` |
| Sin privilegios (más) | `rm /etc/hostname`, `reboot`, `systemctl stop proyecto` (deben fallar por permisos) |
| Comportamiento | `comandoinexistente`, `true` (sin salida), `sleep 3 && echo listo`, `sleep 40` (se cancela a los 30 s) |
| Autenticación | contraseña mala 3 veces: bloqueo; usuario fuera de `usuarios_permitidos`: rechazado; dos clientes a la vez; periodos 1, 5 y 10 |
| En la VM | `systemctl status proyecto --no-pager`, `sudo tail -f /var/log/proyecto/proyecto.log`, `ps -o user,pid,cmd -C servidorOper`, `ss -tlnp \| grep 2810` |

Guion de defensa: estado del servicio, login fallido y bloqueo, login correcto con reporte, `hostname`/`free -h`/`df -h`, `cat /etc/shadow` y `rm /etc/hostname` (permiso denegado), `sleep 40`, dos clientes, `bye` y revisión del log.

## Git
```
git add .
git commit -m "mensaje"
git push
```
