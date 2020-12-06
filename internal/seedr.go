package internal

import (
	"context"
	"github.com/dustin/go-humanize"
	delugeclient "github.com/gdm85/go-libdeluge"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	client *delugeclient.ClientV2
	// The client is not "threadsafe" so we much use locks to not corrupt the responses
	clientMu *sync.RWMutex
)

// wrapper struct to store the info hash with the status instead of as a dict key
type Torrent struct {
	InfoHash string
	*delugeclient.TorrentStatus
}

type torrentSet []*Torrent

func torrentsToSlice(torrents map[string]*delugeclient.TorrentStatus) torrentSet {
	var torrentSlice []*Torrent
	for k, v := range torrents {
		torrentSlice = append(torrentSlice, &Torrent{
			InfoHash:      k,
			TorrentStatus: v,
		})
	}
	return torrentSlice
}

func Start() {
	// you can use NewV1 to create a client for Deluge v1.3
	client = delugeclient.NewV2(delugeclient.Settings{
		Hostname: config.Client.Host,
		Port:     config.Client.Port,
		Login:    config.Client.User,
		Password: config.Client.Password,
	})

	// perform connection to Deluge server
	if err := client.Connect(); err != nil {
		log.Errorf("Could not connect: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Errorf("Failed to close connection: %v", err)
		}
	}()
	ver, err := client.DaemonVersion()
	if err != nil {
		log.Errorf("ERROR: daemon version retrieval: %v\n", err)
		os.Exit(4)
	}
	log.Infof("Connected to Deluge daemon version: %v", ver)

	statInterval, err2 := time.ParseDuration(config.General.StatInterval)
	if err2 != nil {
		log.Fatalf("Invalid general.stat_interval: %v", err)
	}
	updateInterval, err3 := time.ParseDuration(config.General.UpdateInterval)
	if err3 != nil {
		log.Fatalf("Invalid general.update_interval: %v", err)
	}
	ctx := context.Background()
	go updateWorker(ctx, updateInterval)
	go statWorker(ctx, statInterval)
	<-ctx.Done()
}

func updateWorker(ctx context.Context, interval time.Duration) {
	t0 := time.NewTimer(interval)
	for {
		select {
		case <-t0.C:
			// Use a timer so that we can ensure we dont overlap any potentially long running
			// operation.
			log.Debugf("Updating...")
			torrents, err := getTorrentsWithStates([]delugeclient.TorrentState{
				delugeclient.StateSeeding,
				delugeclient.StateActive,
				delugeclient.StatePaused,
			})
			if err != nil {
				log.Errorf("Could not update: %v", err)
				continue
			}
			checkStatus(torrents)
			t0 = time.NewTimer(interval)
		case <-ctx.Done():
			return
		}
	}
}

func getTorrentPathConfig(t *delugeclient.TorrentStatus) (pathConfig, bool) {
	for _, c := range config.Paths {
		if strings.HasPrefix(t.DownloadLocation, c.Path) {
			return c, true
		}
	}
	return pathConfig{}, false
}

func getTorrentsWithStates(states []delugeclient.TorrentState) (torrentSet, error) {
	clientMu.Lock()
	torrents, err := client.TorrentsStatus(delugeclient.StateUnspecified, nil)
	if err != nil {
		clientMu.Unlock()
		return nil, errors.Wrapf(err, "Failed to get torrents")
	}
	clientMu.Unlock()
	var valid torrentSet
	for _, t := range torrentsToSlice(torrents) {
		for _, state := range states {
			if delugeclient.TorrentState(t.State) == state {
				valid = append(valid, t)
				break
			}
		}
	}
	return valid, nil
}

func statWorker(ctx context.Context, interval time.Duration) {
	t0 := time.NewTicker(interval)
	for {
		select {
		case <-t0.C:
			clientMu.Lock()
			torrents, err := client.TorrentsStatus(delugeclient.StateUnspecified, nil)
			if err != nil {
				log.Errorf("ERROR: could not list all torrents: %v\n", err)
				clientMu.Unlock()
				continue
			}
			clientMu.Unlock()
			states := map[delugeclient.TorrentState]int{
				delugeclient.StateSeeding:     0,
				delugeclient.StateDownloading: 0,
				delugeclient.StateActive:      0,
				delugeclient.StateAllocating:  0,
				delugeclient.StateError:       0,
				delugeclient.StatePaused:      0,
				delugeclient.StateChecking:    0,
				delugeclient.StateMoving:      0,
				delugeclient.StateQueued:      0,
			}
			var (
				speedUp int64
				speedDn int64
			)
			for _, v := range torrents {
				states[delugeclient.TorrentState(v.State)]++
				speedUp += v.UploadPayloadRate
				speedDn += v.DownloadPayloadRate
			}

			log.Infof("↑ %s/s ↓ %s/s Seeding: %d Downloading: %d Active: %d Queued: %d Paused: %d Error: %d M/C/A: %d/%d/%d",
				humanize.Bytes(uint64(speedUp)), humanize.Bytes(uint64(speedDn)),
				states[delugeclient.StateSeeding],
				states[delugeclient.StateDownloading],
				states[delugeclient.StateActive],
				states[delugeclient.StateQueued],
				states[delugeclient.StatePaused],
				states[delugeclient.StateError],
				states[delugeclient.StateMoving],
				states[delugeclient.StateChecking],
				states[delugeclient.StateAllocating],
			)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	clientMu = &sync.RWMutex{}
}
