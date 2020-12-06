package internal

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/disk"
	log "github.com/sirupsen/logrus"
	"sort"
)

// TODO add finished_time to status map
func sortAge(torrents torrentSet) {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].TimeAdded > torrents[j].TimeAdded
	})
}

func sortRatio(torrents torrentSet) {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Ratio > torrents[j].Ratio
	})
}

func removeTorrents(hashes []string, set torrentSet) {
	var s torrentSet
	for _, t := range set {
		isRemoved := false
		for _, removed := range hashes {
			if removed == t.InfoHash {
				isRemoved = true
				break
			}
		}
		if !isRemoved {
			s = append(s, t)
		}
	}
}

func deleteTorrent(t *Torrent) error {
	if config.General.DryRunMode {
		log.Infof("Would have deleted: %s", t.Name)
		return nil
	}
	ok, err := client.RemoveTorrent(t.InfoHash, true)
	if err != nil {
		return errors.Wrapf(err, "Error trying to delete torrent")
	}
	if !ok {
		return errors.Wrapf(err, "Bad status trying to delete torrent")
	}
	log.Infof("Torrent deleted (free space): %s", t.Name)
	return nil
}

func moveTorrent(t *Torrent, dest string) error {
	if config.General.DryRunMode {
		log.Infof("Would have moved: %s → %s", t.DownloadLocation, dest)
		return nil
	}
	if err := client.MoveStorage([]string{t.InfoHash}, dest); err != nil {
		return errors.Wrapf(err, "Failed to move torrent to new location")

	}
	log.Infof("Torrent moved (free space): %s", t.Name)
	return nil
}

func checkStatus(torrents torrentSet) {
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
						if err := deleteTorrent(t); err != nil {
							log.Errorf("Failed to delete torrent (disk used): %v", err)
							continue
						}
					} else {
						if err := moveTorrent(t, pathConfigs[i+1].Path); err != nil {
							log.Errorf("Failed to move torrent to next tier: %v", err)
							continue
						}
					}
					newFree += uint64(t.TotalSize)
					removed = append(removed, t.InfoHash)
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
						if err := deleteTorrent(t); err != nil {
							log.Errorf("Failed to delete torrent (ratio): %v", err)
							continue
						}
					} else {
						if err := moveTorrent(t, pathConfigs[i+1].Path); err != nil {
							log.Errorf("Failed to move torrent to next tier (ratio): %v", err)
							continue
						}
					}
				}
			}
		}
	}
}
