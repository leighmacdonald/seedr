package internal

import (
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/shirou/gopsutil/v3/disk"
	log "github.com/sirupsen/logrus"
	"sort"
)

// TODO add finished_time to status map
func sortAge(torrents []client.Torrent) {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[j].AddedOn.Before(torrents[i].AddedOn)
	})
}

func sortRatio(torrents []client.Torrent) {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Ratio > torrents[j].Ratio
	})
}

func removeTorrents(hashes []string, set []client.Torrent) {
	var s []client.Torrent
	for _, t := range set {
		isRemoved := false
		for _, removed := range hashes {
			if removed == t.Hash {
				isRemoved = true
				break
			}
		}
		if !isRemoved {
			s = append(s, t)
		}
	}
}

func checkStatus(torrents []client.Torrent) {
	pathConfigs := pathsByPriority()
	for i, pc := range pathConfigs {
		// Last tier has special meaning, because there is nowhere else to go, this is what will trigger deletions.
		lastTier := i == len(pathConfigs)-1
		diskInfo, err := disk.Usage(pc.Path)
		if err != nil {
			log.Errorf("Failed to get disk info; %v", err)
			return
		}
		var removed []string
		// Disk space related triggers should happen first as they are the bigger blocker for keeping
		// us competitive.
		if pc.MaxUsed > 0 {
			if diskInfo.UsedPercent > pc.MaxUsed {
				log.Debugf("Path use triggered: %v", pc.Path)
				// Get oldest first
				sortAge(torrents)

				// move to next tier if exists
				// newFree is our expected free space after the operation completes
				// TODO The deluge API will return right away, so we need to track in progress long operations like this
				newFree := diskInfo.Free
				for _, t := range torrents {
					if lastTier {
						if err := driver.Remove(t.Hash, true); err != nil {
							log.Errorf("Failed to delete torrent (disk used): %v", err)
							continue
						}
					} else {
						if err := driver.Move(t.Hash, pathConfigs[i+1].Path); err != nil {
							log.Errorf("Failed to move torrent to next tier: %v", err)
							continue
						}
					}
					newFree += uint64(t.Size)
					removed = append(removed, t.Hash)
					if (float64(diskInfo.Total)-float64(newFree))/float64(diskInfo.Total)*100 < pc.MaxUsed {
						log.Debugf("Finished used routine")
						break
					}
				}
			}
		}
		removeTorrents(removed, torrents)
		if pc.MaxRatio > -1 {
			sortRatio(torrents)
			for _, t := range torrents {
				if t.Ratio > pc.MaxRatio {
					if lastTier {
						if err := driver.Remove(t.Hash, true); err != nil {
							log.Errorf("Failed to delete torrent (ratio): %v", err)
							continue
						}
					} else {
						if err := driver.Move(t.Hash, pathConfigs[i+1].Path); err != nil {
							log.Errorf("Failed to move torrent to next tier (ratio): %v", err)
							continue
						}
					}
				}
			}
		}
	}
}
