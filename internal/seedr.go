package internal

import (
	"context"
	"github.com/dustin/go-humanize"
	delugeclient "github.com/gdm85/go-libdeluge"
	"github.com/leighmacdonald/seedr/pkg/client"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

var (
	driver client.Driver
	// The client is not "threadsafe" so we much use locks to not corrupt the responses
	clientMu *sync.RWMutex
)

func torrentsToSlice(torrents map[string]client.Torrent) []client.Torrent {
	var torrentSlice []client.Torrent
	for _, v := range torrents {
		torrentSlice = append(torrentSlice, v)
	}
	return torrentSlice
}

func Start() {
	cl, err := client.New(config.Client)
	if err != nil {
		log.Fatalf("Failed to create client driver: %v", err)
	}
	defer func() {
		if err := cl.Close(); err != nil {
			log.Errorf("Failed to close connection: %v", err)
		}
	}()
	driver = cl

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
			torrents, err := driver.TorrentsWithState(client.Seeding, client.Active, client.Paused)
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

func statWorker(ctx context.Context, interval time.Duration) {
	t0 := time.NewTicker(interval)
	for {
		select {
		case <-t0.C:
			clientMu.Lock()
			torrents, err := driver.TorrentsWithState(client.Any)
			if err != nil {
				log.Errorf("ERROR: could not list all torrents: %v\n", err)
				clientMu.Unlock()
				continue
			}
			for _, t := range torrents {
				log.Debugln(t)
			}
			clientMu.Unlock()
			states := map[client.State]int{
				client.Seeding:     0,
				client.Downloading: 0,
				client.Active:      0,
				client.Allocating:  0,
				client.Error:       0,
				client.Paused:      0,
				client.Checking:    0,
				client.Moving:      0,
				client.Queued:      0,
			}
			var (
				speedUp int64
				speedDn int64
			)
			//for _, v := range torrents {
			//	states[delugeclient.TorrentState(v.State)]++
			//	speedUp += v.UploadPayloadRate
			//	speedDn += v.DownloadPayloadRate
			//}

			log.Infof("↑ %s/s ↓ %s/s Seeding: %d Downloading: %d Active: %d Queued: %d Paused: %d Error: %d M/C/A: %d/%d/%d",
				humanize.Bytes(uint64(speedUp)), humanize.Bytes(uint64(speedDn)),
				states[client.Seeding],
				states[client.Downloading],
				states[client.Active],
				states[client.Queued],
				states[client.Paused],
				states[client.Error],
				states[client.Moving],
				states[client.Checking],
				states[client.Allocating],
			)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	clientMu = &sync.RWMutex{}
}
