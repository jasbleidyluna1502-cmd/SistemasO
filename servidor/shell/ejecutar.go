package shell

import (
	"context"
	"os/exec"
	"time"
)

// Ejecuta el comando en Linux y devuelve su salida.
// Si tarda mas de 30 segundos se cancela para no bloquear la sesion.
func Ejecutar(comando string) string {
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	respuesta, err := exec.CommandContext(ctx, "sh", "-c", comando).CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(respuesta) + "\n(comando cancelado: supero 30 segundos)"
	}
	if err != nil && len(respuesta) == 0 {
		return "error: " + err.Error()
	}
	return string(respuesta)
}
