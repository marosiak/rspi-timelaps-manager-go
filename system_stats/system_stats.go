package system_stats

import (
	"errors"
	"fmt"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/rs/zerolog/log"
	"math"
	"runtime"
	"time"
)

type SystemStatsService struct {
	cfg          *config.Config
	lastCpuStats *cpu.Stats
}

type CpuInfo struct {
	User   float64
	System float64
	Idle   float64
}

type StatsResponse struct {
	Memory *memory.Stats `json:"memory"`
	Cpu    *CpuInfo      `json:"cpu"`
}

func NewSystemStats() *SystemStatsService {
	var currentCpuStats *cpu.Stats
	var err error
	systemStatsSrv := &SystemStatsService{
		cfg: config.New(),
	}

	currentCpuStats, err = systemStatsSrv.getCpuStats()
	if err != nil {
		log.Err(err).Msg("get cpu info")
	}

	systemStatsSrv.lastCpuStats = currentCpuStats
	time.Sleep(1 * time.Second)
	return systemStatsSrv
}

func (a *SystemStatsService) getCpuStats() (*cpu.Stats, error) {
	if runtime.GOOS == "windows" {
		return &cpu.Stats{
			User:   1,
			System: 2,
			Idle:   3,
			Nice:   1,
			Total:  1,
		}, nil
	}

	currentCpuStats, err := cpu.Get()
	if err != nil {
		return nil, fmt.Errorf("get cpu info: %w", err)
	}
	return currentCpuStats, nil
}

var CpuInfoIsNaN = errors.New("cpu info is NaN")

func (a *SystemStatsService) getCpuInfo() (*CpuInfo, error) {
	if a.lastCpuStats == nil {
		stats, err := a.getCpuStats()
		if err != nil {
			return nil, err
		}
		a.lastCpuStats = stats
		time.Sleep(1 * time.Second)
	}
	currentCpuStats, err := a.getCpuStats()
	if err != nil {
		return nil, fmt.Errorf("get cpu info: %w", err)
	}

	total := float64(currentCpuStats.Total - a.lastCpuStats.Total)
	cpuInfo := &CpuInfo{
		User:   float64(currentCpuStats.User-a.lastCpuStats.User) / total * 100,
		System: float64(currentCpuStats.System-a.lastCpuStats.System) / total * 100,
		Idle:   float64(currentCpuStats.Idle-a.lastCpuStats.Idle) / total * 100,
	}

	a.lastCpuStats = currentCpuStats
	if math.IsNaN(cpuInfo.User) || math.IsNaN(cpuInfo.System) || math.IsNaN(cpuInfo.Idle) {
		return nil, CpuInfoIsNaN
	}
	return cpuInfo, nil
}

func (a *SystemStatsService) GetSystemUsageInfo() (*StatsResponse, error) {
	memInfo, err := memory.Get()
	if err != nil {
		return nil, fmt.Errorf("get memory info: %w", err)
	}

	cpuInfo, err := a.getCpuInfo()
	if err != nil && !errors.Is(err, CpuInfoIsNaN) {
		return nil, err
	}

	if cpuInfo == nil {
		cpuInfo = &CpuInfo{
			User:   0,
			System: 0,
			Idle:   0,
		}
	}

	return &StatsResponse{
		Memory: memInfo,
		Cpu:    cpuInfo,
	}, nil
}
