package internal

import (
	"github.com/shirou/gopsutil/v3/disk"
	"os"
	"path/filepath"
	"strings"
)

func getFreePct(path string) float64 {
	use, err := disk.Usage(path)
	if err != nil {
		return -1
	}
	return use.UsedPercent
}

func isSubPath(parent string, path string) (bool, error) {
	up := ".." + string(os.PathSeparator)
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true, nil
	}
	return false, nil
}
