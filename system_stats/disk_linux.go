//go:build linux
// +build linux

package system_stats

import (
	"golang.org/x/sys/unix"
)

func getDiskUsage(path string) (total, free uint64, err error) {
	var stat unix.Statfs_t
	err = unix.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bfree * uint64(stat.Bsize)

	return total, free, nil
}
