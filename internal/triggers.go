package internal

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)

var checkFuncs = map[CheckOrder]func(torrents []*client.Torrent, cfg *checkConfig, pathCurrent int, pathTotal int) error{
	MinFree:  checkMinFree,
	MaxRatio: checkRatio,
}

// TODO add finished_time to status map
func sortAge(torrents []*client.Torrent) {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[j].AddedOn.After(torrents[i].AddedOn)
	})
}

func sortRatio(torrents []*client.Torrent) {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Ratio > torrents[j].Ratio
	})
}

func removeTorrents(hashes []string, set []*client.Torrent) []*client.Torrent {
	var s []*client.Torrent
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
	return s
}

func checkMinFree(torrents []*client.Torrent, cfg *checkConfig, pathCurrent int, pathTotal int) error {
	// Last tier has special meaning, because there is nowhere else to go, this is what will trigger deletions.
	lastTier := pathCurrent == pathTotal-1
	bytesFree, err := driver.FreeSpace(cfg.Path)
	if err != nil {
		return errors.Errorf("Failed to get disk info; %v", err)
	}
	var removed []string
	var moved []string
	// Disk space related triggers should happen first as they are the bigger blocker for keeping
	// us competitive.
	if bytesFree < cfg.MinFree {
		log.Debugf("Path use triggered: %v", cfg.Path)
		// Get oldest first
		sortAge(torrents)

		// move to next tier if exists
		// newFree is our expected free space after the operation completes
		// TODO The deluge API will return right away, so we need to track in progress long operations like this
		newFree := bytesFree
		for _, t := range torrents {
			if lastTier {
				if config.General.DryRunMode {
					t.Log().Infof("[DRY] Removed torrent (disk free)")
				} else {
					if err := driver.Remove(t.Hash, true); err != nil {
						t.Log().Errorf("Failed to delete torrent (disk used): %v", err)
						continue
					}
					t.Log().Infof("Removed torrent (disk free)")
				}
			} else {
				if config.General.DryRunMode {
					t.Log().Infof("[DRY] Moved torrent to next storage tier (disk free)")
				} else {
					if err := driver.Move(t.Hash, config.Checks.Paths[pathCurrent+1].Path); err != nil {
						t.Log().Errorf("Failed to move torrent to next tier: %v", err)
						continue
					}
					t.Log().Infof("Moved torrent to next storage tier (disk free)")
				}
				moved = append(moved, t.Hash)
			}
			newFree += t.Size
			removed = append(removed, t.Hash)
			if newFree > cfg.MinFree {
				log.WithFields(log.Fields{
					"cleared": humanize.Bytes(uint64(newFree - bytesFree)),
					"free":    humanize.Bytes(uint64(newFree)),
				}).Info("Free space threshold met")
				break
			}
		}
	}
	// Wait for torrents that are moving to complete before continuing
	for len(moved) > 0 {
		for _, hash := range moved {
			var torrent client.Torrent
			if err := driver.Torrent(hash, &torrent); err != nil {
				log.Errorf("Failed to get moved torrent state: %v", err)
			} else {
				if torrent.State != client.Moving {
					moved = removeString(moved, hash)
					torrent.Log().Infof("Mmove operation completed")
				} else {
					torrent.Log().Infof("Waiting for move operation")
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
	torrents = removeTorrents(removed, torrents)
	return nil
}

func checkRatio(torrents []*client.Torrent, cfg *checkConfig, pathCurrent int, pathTotal int) error {
	// Last tier has special meaning, because there is nowhere else to go, this is what will trigger deletions.
	lastTier := pathCurrent == pathTotal-1
	var removed []string
	if cfg.MaxRatio > -1 {
		sortRatio(torrents)
		for _, t := range torrents {
			if t.Ratio > cfg.MaxRatio {
				l := log.WithFields(log.Fields{"name": t.Name, "ratio": fmt.Sprintf("%.2f", t.Ratio), "hash": t.Hash})
				if lastTier {
					if config.General.DryRunMode {
						l.Infof("[DRY] Removed torrent (ratio): %s ratio: %f", t.Name, t.Ratio)
					} else {
						if err := driver.Remove(t.Hash, true); err != nil {
							l.Errorf("Failed to delete torrent (ratio): %v", err)
							continue
						}
						l.Infof("Removed torrent from last available tier")
					}

				} else {
					if config.General.DryRunMode {
						l.Infof("[DRY] Move torrent to lower storage tier")
					} else {
						if err := driver.Move(t.Hash, config.Checks.Paths[pathCurrent+1].Path); err != nil {
							l.Errorf("Failed to move torrent to next tier")
							continue
						}
						l.Infof("Moved torrent to lower storage tier")
					}
				}
				removed = append(removed, t.Hash)
			}
		}
	}
	torrents = removeTorrents(removed, torrents)
	return nil
}

func checkStatus(torrents []*client.Torrent) {
	checkConfigs := checksByPriority()
	for checkName, checkFn := range checkFuncs {
		for i, pc := range checkConfigs {
			log.Debugf("Perfoming check: %s", checkName)
			if err := checkFn(torrents, pc, i, len(checkConfigs)); err != nil {
				log.Errorf("Failed to perform check func: %v", err)
				return
			}
		}
	}
}
