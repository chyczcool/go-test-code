package thirdlib

import (
	"github.com/shirou/gopsutil/process"
)

func IsProcessRunning(processName string) (bool, error) {
	processes, err := process.Processes()
	if err != nil {
		return false, err
	}

	for _, proc := range processes {
		if p, err := proc.Name(); err == nil {
			//fmt.Println(p)
			if p == processName {
				return true, nil
			} else {
				continue
			}
		}
	}
	return false, err
}
