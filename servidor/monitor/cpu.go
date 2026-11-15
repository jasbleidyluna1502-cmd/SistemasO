package monitor

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

func leerCPU() (uint64, uint64) {
	archivo, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0
	}
	defer archivo.Close()
	lector := bufio.NewScanner(archivo)
	lector.Scan()
	campos := strings.Fields(lector.Text())[1:]
	var total, idle uint64
	for i, c := range campos {
		valor, _ := strconv.ParseUint(c, 10, 64)
		total += valor
		if i == 3 || i == 4 { // idle + iowait
			idle += valor
		}
	}
	return total, idle
}

func PorcentajeCPU() int {
	total1, idle1 := leerCPU()
	time.Sleep(time.Millisecond * 500)
	total2, idle2 := leerCPU()
	dTotal := total2 - total1
	if dTotal == 0 {
		return 0
	}
	return int(100 * (dTotal - (idle2 - idle1)) / dTotal)
}
