# Revisión de cumplimiento de requerimientos — Remote Shell and System Resource Monitoring

**Alcance:** revisión del código en `proyecto-so` (servidor, cliente, conf, systemd) contra los 12 requerimientos enviados.
**Limitación:** no pude leer el contenido de los PDFs (`LabOs.pdf` y `Final Project Requirements (Details)`) porque el entorno no tiene `pdftoppm`. La revisión se basa en la lista de requerimientos que pegaste. Si los PDFs piden algo más (formato de mensajes, puertos, número de intentos, etc.), hay que contrastarlo aparte.
**Leyenda:** ✅ Cumple · ⚠️ Cumple parcialmente / falta verificar · ❌ No cumple / no existe en la carpeta

## Resumen

| # | Requerimiento | Estado |
|---|---|---|
| 1 | Conexión remota y reporte de recursos (memoria, disco, procesador, procesos) | ✅ |
| 2 | Ejecución remota de comandos shell por sockets TCP | ✅ |
| 3 | Autenticación SHA-512 con salt y base de usuarios en archivo plano | ⚠️ |
| 4 | Configuración mediante `/etc/configProyecto.conf` | ✅ |
| 5 | Monitoreo de CPU, memoria, almacenamiento y procesos | ✅ |
| 6 | Envío/recepción simultáneos con goroutines y canales | ✅ |
| 7 | Servidor como servicio systemd con usuario no privilegiado | ⚠️ |
| 8 | Registro de eventos/actividad en archivos | ✅ |
| 9 | Cliente GUI multiplataforma | ✅ |
| 10 | Código fuente y control de versiones Git/GitHub | ❌ |
| 11 | Paper técnico en formato IEEE | ❌ |
| 12 | Presentación técnica y defensa | ❌ |

## Detalle por requerimiento

### 1. Conexión remota y reporte de recursos — ✅
- El cliente se conecta por IP/puerto/periodo (`cliente/main.go`: `clienteOper IP Puerto Periodo`).
- El servidor envía periódicamente `REPORTE|... [CPU=%% - RAM=%% - DD=%% - PROC=n]` (`servidor/red_servidor/handler.go`, `envMensajeS`; `monitor/reporte.go`).
- El cliente lo muestra en una etiqueta de la GUI (`cliente/ui/principal.go`).

### 2. Ejecución remota de comandos sobre TCP — ✅
- El cliente envía líneas de comando; el servidor las ejecuta con `sh -c` (`servidor/shell/ejecutar.go`) y devuelve la salida como líneas `SALIDA|...`.
- Comando `bye` cierra la sesión.
- Observaciones: no hay timeout para comandos (un `sleep 1000` o un comando interactivo bloquea la goroutine que lee comandos); se ejecutan con los permisos del usuario del servicio, lo cual es correcto con el punto 7.

### 3. Autenticación SHA-512 + salt + archivo plano — ⚠️
- ✅ Formato `usuario:salt:hash`, con `hash = SHA512(salt + password)` en hex (`auth/credenciales.go`).
- ✅ Generador de usuarios con salt aleatorio de 8 bytes vía `crypto/rand` (`servidor/crearUsuario/crearUsuario.go`).
- ✅ Límite de intentos fallidos configurable (`intentos_fallidos`) y respuesta `BLOQUEADO`; lista de `usuarios_permitidos`; filtro por `ip_permitida`.
- ⚠️ `conf/usuarios.txt` solo contiene comentarios; **no hay ningún usuario generado** (luna, leidy). Hay que generarlos con `crearUsuario` y copiarlos a `/etc/usuariosProyecto.txt`.
- ⚠️ La comparación del hash no es de tiempo constante (`==`); es menor para un proyecto académico, pero se puede usar `subtle.ConstantTimeCompare`.
- ⚠️ Usuario y contraseña viajan en texto plano por TCP (sin TLS). Suele aceptarse en el alcance del curso, pero conviene mencionarlo como limitación en el paper.
- ⚠️ SHA-512 simple con un solo paso es rápido de atacar por fuerza bruta; si el enunciado solo pide SHA-512 con salt, cumple.

### 4. Configuración en `/etc/configProyecto.conf` — ✅
- La ruta está fija en `servidor/red_servidor/config.go` (`/etc/configProyecto.conf`).
- Claves soportadas: `ip_permitida`, `puerto`, `intentos_fallidos`, `usuarios_permitidos`, `archivo_usuarios`, `archivo_log`.
- La plantilla está en `conf/configProyecto.conf` (hay que copiarla a `/etc/`).
- Observación: `ip_permitida` es una única IP; si el cliente se conecta desde otra máquina, será rechazado (y la IP del ejemplo, `10.1.28.159`, debe coincidir con la del cliente en la demo).

### 5. Monitoreo de CPU, memoria, disco y procesos — ✅
- CPU: diferencia de `/proc/stat` en 500 ms (`monitor/cpu.go`).
- Memoria: `/proc/meminfo`, `(MemTotal − MemAvailable) / MemTotal` (`monitor/ram.go`).
- Disco: `syscall.Statfs("/")` (`monitor/disco.go`).
- Procesos: cuenta directorios numéricos en `/proc` (`monitor/procesos.go`).
- Observaciones: solo funciona en Linux (correcto, el servidor es Linux). Solo mide la partición `/`. Si los PDFs piden valores absolutos (MB/GB) además del porcentaje, no están.

### 6. Goroutines y canales — ✅
- Por cliente se lanzan 3 goroutines: `escritorS` (única que escribe al socket), `envMensajeS` (reportes periódicos) y `recMensajeS` (lectura de comandos), comunicadas por el canal `salida` y el canal `done`.
- Observaciones:
  - El servidor atiende **un cliente a la vez**: `Iniciar` llama `atenderCliente(socketS, cfg)` sin `go`, por lo que un segundo cliente espera hasta que el primero termine. Si el enunciado pide múltiples clientes simultáneos, hay que cambiarlo a `go atenderCliente(...)`.
  - Posible bloqueo menor: tras cerrar `done`, un `salida <- ...` en `recMensajeS` podría quedar bloqueado si el canal (buffer 10) se llena; riesgo bajo.

### 7. Servicio systemd con usuario no privilegiado — ⚠️
- ✅ `systemd/proyecto.service` define `User=proyectoso`, `ExecStart=/usr/local/bin/servidorOper`, `LogsDirectory=proyecto` (crea `/var/log/proyecto` con permisos para ese usuario), `Restart=on-failure`, `WantedBy=multi-user.target`.
- ⚠️ Falta verificar en la máquina: que el usuario `proyectoso` exista (`useradd -r`), que el binario esté en `/usr/local/bin`, y que pueda leer `/etc/configProyecto.conf` y `/etc/usuariosProyecto.txt` (permisos 640, grupo adecuado).
- ⚠️ Puerto 2810 es > 1024, así que no requiere privilegios (bien).
- Opcional: endurecimiento (`NoNewPrivileges=true`, `ProtectSystem=strict`, `PrivateTmp=true`).

### 8. Registro en archivos — ✅
- `log.SetOutput` al archivo configurado (`red_servidor/log.go`); se registran inicio del servidor, conexiones entrantes, IP rechazada, intentos fallidos, bloqueo, usuario autenticado, cada comando ejecutado y desconexión.
- Observación: no se registra la contraseña (bien). El log debe quedar con rotación o tamaño acotado si se quiere más completo.

### 9. Cliente GUI multiplataforma — ✅
- Usa Fyne (`cliente/ui/login.go`, `principal.go`), que compila para Windows, Linux y macOS. Hay un `clienteOper.exe` ya compilado para Windows.
- ⚠️ Para demostrar "multiplataforma" conviene tener compilado/probado al menos en un segundo sistema (por ejemplo Linux) y mencionarlo en el paper.
- Observación: el cliente recibe IP, puerto y periodo por línea de comandos y la GUI pide usuario/contraseña.

### 10. Git/GitHub — ❌
- La carpeta **no es un repositorio Git** (no existe `.git`). Hay que hacer `git init`, agregar `.gitignore` (excluir `clienteOper.exe` y archivos con usuarios reales), un `README.md` y subirlo a GitHub con commits que muestren el avance del equipo.

### 11. Paper técnico en formato IEEE — ❌
- No hay ningún documento en la carpeta del proyecto. Debe incluir: resumen, introducción, arquitectura (cliente/servidor, protocolo `OK/FAIL/BLOQUEADO`, `REPORTE|`, `SALIDA|`), seguridad (SHA-512 + salt), concurrencia con goroutines/canales, despliegue systemd, pruebas, limitaciones y referencias, usando plantilla IEEE.

### 12. Presentación técnica y defensa — ❌
- No hay presentación en la carpeta. Prepararla (arquitectura, demo en vivo: login, reporte, comando remoto, intento fallido/bloqueo, log, estado del servicio systemd).

## Pendientes priorizados

1. Crear los usuarios en el archivo de usuarios (`crearUsuario luna <pass>`, `leidy <pass>`) y desplegar en `/etc/usuariosProyecto.txt`.
2. Inicializar Git y publicar en GitHub con README.
3. Escribir el paper IEEE.
4. Preparar la presentación y la demo.
5. Confirmar si el enunciado pide varios clientes simultáneos y, si es así, usar `go atenderCliente(...)`.
6. Verificar el despliegue completo en Linux: usuario `proyectoso`, permisos, `systemctl enable --now proyecto`, y que `journalctl`/el log muestren actividad.
7. Revisar los PDFs contra este documento para cualquier requisito que no esté en la lista pegada.
