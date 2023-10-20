//go:build windows
// +build windows

package system_stats

func getDiskUsage(path string) (total, free uint64, err error) {
	// Mockowane wartości
	total = 1000000000
	free = 500000000
	return total, free, nil
}
