package monitor

import "syscall"

func PorcentajeDisco() int {
	var st syscall.Statfs_t
	if syscall.Statfs("/", &st) != nil || st.Blocks == 0 {
		return 0
	}
	return int(100 * (st.Blocks - st.Bavail) / st.Blocks)
}
