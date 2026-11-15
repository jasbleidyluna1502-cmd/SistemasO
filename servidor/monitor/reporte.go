package monitor

import "fmt"

func GenerarReporte() string {
	return fmt.Sprintf("[CPU=%d%% - RAM=%d%% - DD=%d%% - PROC=%d]",
		PorcentajeCPU(), PorcentajeRAM(), PorcentajeDisco(), ContarProcesos())
}
