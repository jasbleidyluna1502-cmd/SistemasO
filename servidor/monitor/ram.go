package monitor

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func PorcentajeRAM() int {
	archivo, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer archivo.Close()
	var total, disponible uint64
	lector := bufio.NewScanner(archivo)
	for lector.Scan() {
		campos := strings.Fields(lector.Text())
		if len(campos) < 2 {
			continue
		}
		valor, _ := strconv.ParseUint(campos[1], 10, 64)
		if campos[0] == "MemTotal:" {
			total = valor
		} else if campos[0] == "MemAvailable:" {
			disponible = valor
		}
	}
	if total == 0 {
		return 0
	}
	return int(100 * (total - disponible) / total)
}
