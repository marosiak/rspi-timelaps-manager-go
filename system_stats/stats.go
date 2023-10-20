package system_stats

import (
	"errors"
	"fmt"
	"github.com/mackerelio/go-osstat/cpu"
	ram "github.com/mackerelio/go-osstat/memory"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/rs/zerolog/log"
	"math"
	"os"
	"runtime"
	"sort"
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

type MemoryInfo struct {
	Total                     uint64
	Free                      uint64
	TimeRemainingForTimelapse string
}

type StatsResponse struct {
	Ram    *ram.Stats  `json:"ram"`
	Cpu    *CpuInfo    `json:"cpu"`
	Memory *MemoryInfo `json:"memory"`
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *SystemStatsService) calculateTimeRemainingForTimelapse(freeSpace uint64, avgFileSize uint64) string {
	// Calculate how many files can fit with free space
	filesToFit := freeSpace / avgFileSize

	// Calculate total time available in seconds
	totalSeconds := filesToFit * uint64(a.cfg.Delay.Seconds())

	// Calculate time in different units
	w := totalSeconds / (60 * 60 * 24 * 7)
	d := (totalSeconds % (60 * 60 * 24 * 7)) / (60 * 60 * 24)
	h := (totalSeconds % (60 * 60 * 24)) / (60 * 60)
	m := (totalSeconds % (60 * 60)) / 60

	// Generate the time string
	var timeStr string

	if w > 0 {
		timeStr += fmt.Sprintf("%d week%s ", w, pluralize(w))
	}
	if d > 0 {
		timeStr += fmt.Sprintf("%d day%s ", d, pluralize(d))
	}
	if h > 0 {
		timeStr += fmt.Sprintf("%d hour%s ", h, pluralize(h))
	}
	if m > 0 {
		timeStr += fmt.Sprintf("%d minute%s", m, pluralize(m))
	}

	if timeStr == "" {
		return "No time left"
	}

	return timeStr
}

func pluralize(n uint64) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func (a *SystemStatsService) getAveragePhotoSize() (uint64, error) {
	var filesToRead = 20
	var totalSize float64

	// Read all files from directory
	files, err := os.ReadDir(a.cfg.OutputDir)
	if err != nil {
		return 0, fmt.Errorf("read dir: %w", err)
	}

	// Sort files by modification time, newest first
	sort.Slice(files, func(i, j int) bool {
		infoI, errI := files[i].Info()
		infoJ, errJ := files[j].Info()

		if errI != nil || errJ != nil {
			return false // Handle error as you see fit
		}

		return infoI.ModTime().After(infoJ.ModTime())
	})

	files = files[:min(len(files), filesToRead)]

	// Calculate average size of the 20 latest files and inflate it by 20%
	for i, file := range files {
		if i >= filesToRead {
			break
		}

		fInfo, err := file.Info()
		if err != nil {
			return 0, fmt.Errorf("get file info: %w", err)
		}

		totalSize += float64(fInfo.Size())
	}
	return uint64(totalSize / float64(filesToRead)), nil
}

func (a *SystemStatsService) getDiskInfo() (memoryInfo MemoryInfo, err error) {
	total, free, err := getDiskUsage(a.cfg.OutputDir)
	if err != nil {
		return MemoryInfo{}, err
	}

	avgFileSize, err := a.getAveragePhotoSize()
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("get average photo size: %w", err)
	}
	avgFileSize = uint64(float64(avgFileSize) * 1.15) // add 15% to the average file size just for safety

	memoryInfo.Free = free
	memoryInfo.Total = total

	memoryInfo.TimeRemainingForTimelapse = a.calculateTimeRemainingForTimelapse(free, avgFileSize)

	return memoryInfo, nil
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
	ramInfo, err := ram.Get()
	if err != nil {
		return nil, fmt.Errorf("get memory info: %w", err)
	}

	memoryInfo, err := a.getDiskInfo()
	if err != nil {
		return nil, fmt.Errorf("get disk space: %w", err)
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
		Ram:    ramInfo,
		Cpu:    cpuInfo,
		Memory: &memoryInfo,
	}, nil
}
