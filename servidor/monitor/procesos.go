package monitor

import (
	"os"
	"strconv"
)

// Cada carpeta numerica de /proc es un proceso
func ContarProcesos() int {
	entradas, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	cuenta := 0
	for _, e := range entradas {
		if _, err := strconv.Atoi(e.Name()); err == nil {
			cuenta++
		}
	}
	return cuenta
}
