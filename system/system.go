package system

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"os"
	"time"
)

var persistentDiskPath = "/var/vcap/data"

type Stats struct {
	CpuUsed            float64 `json:"cpu_used"`
	MemoryUsed         float64 `json:"memory_used"`
	PersistentDiskUsed float64 `json:"disk_used,omitempty"`
	Load15             float64 `json:"load15"`
	Uptime             uint64  `json:"uptime"`
}

func GetStats() (Stats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return Stats{}, err
	}

	l, err := load.Avg()
	if err != nil {
		return Stats{}, err
	}

	c, err := cpu.Percent(time.Second, false)
	if err != nil {
		return Stats{}, err
	}

	u, err := host.Uptime()
	if err != nil {
		return Stats{}, err
	}

	var persistentDiskUsed float64

	if fileExists(persistentDiskPath) {
		d, err := disk.Usage(persistentDiskPath)
		if err != nil {
			return Stats{}, err
		}
		persistentDiskUsed = d.UsedPercent
	}

	return Stats{
		Load15:             l.Load15,
		CpuUsed:            avgFloats(c),
		MemoryUsed:         v.UsedPercent,
		PersistentDiskUsed: persistentDiskUsed,
		Uptime:             u,
	}, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	if err == nil {
		return true
	}

	return false
}

func avgFloats(args []float64) float64 {
	var total float64
	for _, arg := range args {
		total += arg
	}
	return total / float64(len(args))
}
