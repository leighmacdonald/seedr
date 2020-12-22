package deluge

import (
	"encoding/base64"
	"fmt"
	deluge "github.com/gdm85/go-libdeluge"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"strings"
)

const driverName = "deluge"

var (
	stateMap = map[deluge.TorrentState]client.State{
		deluge.StateUnspecified: client.Unknown,
		deluge.StateActive:      client.Active,
		deluge.StateAllocating:  client.Allocating,
		deluge.StateChecking:    client.Checking,
		deluge.StateDownloading: client.Downloading,
		deluge.StateError:       client.Error,
		deluge.StateMoving:      client.Moving,
		deluge.StatePaused:      client.Paused,
		deluge.StateQueued:      client.Queued,
		deluge.StateSeeding:     client.Seeding,
	}
)

func mapStates(states ...client.State) []deluge.TorrentState {
	var dStates []deluge.TorrentState
	for _, state := range states {
		for k, v := range stateMap {
			if v == state {
				dStates = append(dStates, k)
				break
			}
		}
	}
	return dStates
}

func mapTorrentStatus(status *deluge.TorrentStatus, torrent *client.Torrent) {
	torrent.Name = status.Name
	torrent.Size = status.TotalSize
	torrent.Ratio = float64(status.Ratio)
}

func getState(status *deluge.TorrentStatus) client.State {
	s, ok := stateMap[deluge.TorrentState(status.State)]
	if !ok {
		log.Warnf("Got invalid state: %s", status.State)
		return client.Unknown
	}
	return s
}

type Deluge struct {
	cfg    *client.Config
	client deluge.DelugeClient
}

func (d Deluge) FreeSpace(path string) (int64, error) {
	return d.client.GetFreeSpace(path)
}

// TODO set label
func (d Deluge) Add(filename string, torrent io.Reader, path string, label string) error {
	b, err := ioutil.ReadAll(torrent)
	if err != nil {
		return err
	}
	e := base64.StdEncoding.EncodeToString(b)
	hash, err := d.client.AddTorrentFile(filename, e, &deluge.Options{
		DownloadLocation: &path,
		V2:               deluge.V2Options{},
	})
	if err != nil {
		return err
	}
	log.Debugf("Added torrent %s [%s]", filename, hash)
	return nil
}

func (d Deluge) Announce(hash string) error {
	return d.client.ForceReannounce([]string{hash})
}

func (d Deluge) ClientVersion() (string, error) {
	dv, err := d.client.DaemonVersion()
	if err != nil {
		return "", err
	}
	cv, err := d.client.GetLibtorrentVersion()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s / %s", dv, cv), nil
}

func (d Deluge) Login() error {
	if err := d.client.Connect(); err != nil {
		return errors.Wrapf(client.ErrAuthFailed, "failed to connect to client: %v", err)
	}
	plugins, err := d.client.GetEnabledPlugins()
	if err != nil {
		return errors.Wrapf(client.ErrDriverError, "failed to get enabled plugins: %v", err)
	}
	labelEnabled := false
	for _, name := range plugins {
		if strings.ToLower(name) == "label" {
			labelEnabled = true
			break
		}
	}
	if !labelEnabled {
		log.Warnf("Label plugin is not enabled, some functionality will not work correectly")
	}
	return nil
}

func (d Deluge) Move(hash string, dest string) error {
	return d.client.MoveStorage([]string{hash}, dest)
}

func (d Deluge) Pause(hash string) error {
	return d.client.PauseTorrents(hash)
}

func (d Deluge) PauseAll() error {
	hashes, err := d.getAllHashes()
	if err != nil {
		return err
	}
	return d.client.PauseTorrents(hashes...)
}

func (d Deluge) getAllHashes() ([]string, error) {
	torrents, err := d.Torrents()
	if err != nil {
		return nil, err
	}
	var hashes []string
	for _, t := range torrents {
		hashes = append(hashes, t.Hash)
	}
	return hashes, nil
}

func (d Deluge) Queue(hash string, position client.QueuePos) error {
	// TODO need to add support in lib for queue_* calls
	return nil
}

func (d Deluge) Remove(hash string, deleteData bool) error {
	ok, err := d.client.RemoveTorrent(hash, deleteData)
	if err != nil {
		return errors.Wrapf(client.ErrDriverError, "Failed to remove torrent: %v", err)
	}
	if !ok {
		log.Warnf("Torrent not removed, does not exist?: %s", hash)
	} else {
		log.Infof("Removed torrent: %s", hash)
	}
	return nil
}

func (d Deluge) Start(hash string) error {
	return d.client.ResumeTorrents(hash)
}

func (d Deluge) StartAll() error {
	hashes, err := d.getAllHashes()
	if err != nil {
		return err
	}
	return d.client.ResumeTorrents(hashes...)
}

func (d Deluge) Stop(hash string) error {
	return d.Pause(hash)
}

func (d Deluge) Torrent(hash string, torrent *client.Torrent) error {
	status, err := d.client.TorrentStatus(hash)
	if err != nil {
		return err
	}
	torrent.Hash = hash
	mapTorrentStatus(status, torrent)
	return nil

}

func statusToTorrents(states map[string]*deluge.TorrentStatus) ([]*client.Torrent, error) {
	var torrents []*client.Torrent
	for id, meta := range states {
		var t client.Torrent
		t.Hash = id
		mapTorrentStatus(meta, &t)
		torrents = append(torrents, &t)
	}
	return torrents, nil
}

func (d Deluge) Torrents() ([]*client.Torrent, error) {
	states, err := d.client.TorrentsStatus(deluge.StateUnspecified, nil)
	if err != nil {
		return nil, err
	}
	return statusToTorrents(states)
}

func (d Deluge) TorrentsWithState(statuses ...client.State) ([]*client.Torrent, error) {
	torrents, err := d.client.TorrentsStatus(deluge.StateUnspecified, nil)
	if err != nil {
		return nil, err
	}
	valid := make(map[string]*deluge.TorrentStatus)
	for hash, s := range torrents {
		for _, status := range statuses {
			if getState(s) == status {
				valid[hash] = s
				break
			}
		}
	}
	return statusToTorrents(valid)
}

func (d Deluge) Verify(hash string) error {
	// TODO Add force_recheck to library
	return nil
}

func (d Deluge) Close() error {
	return d.client.Close()
}

type Factory struct{}

func (f Factory) New(cfg *client.Config) (client.Driver, error) {
	c := deluge.NewV2(deluge.Settings{
		Hostname: cfg.Host,
		Port:     uint(cfg.Port),
		Login:    cfg.Username,
		Password: cfg.Password,
	})
	return Deluge{cfg: cfg, client: c}, nil
}

func init() {
	if err := client.RegisterDriver(driverName, Factory{}); err != nil {
		log.Fatalf("Failed to register qbittorrent driver: %v", err)
	}
}
