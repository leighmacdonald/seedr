package qbittorrent

import (
	"fmt"
	"github.com/KnutZuidema/go-qbittorrent"
	"github.com/KnutZuidema/go-qbittorrent/pkg/model"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
)

const driverName = "qbittorrent"

var (
	stateMap = map[model.TorrentState]client.State{
		model.StateUnknown:     client.Unknown,
		model.StateAllocating:  client.Allocating,
		model.StateCheckingDL:  client.Checking,
		model.StateCheckingUP:  client.Checking,
		model.StateDownloading: client.Downloading,
		model.StateError:       client.Error,
		model.StateMoving:      client.Moving,
		model.StatePausedDL:    client.Paused,
		model.StatePausedUP:    client.Paused,
		model.StateQueuedDL:    client.Queued,
		model.StateQueuedUP:    client.Queued,
		model.StateUploading:   client.Seeding,
	}
)

type QBittorrent struct {
	cfg client.Config
	qb  *qbittorrent.Client
}

func (driver QBittorrent) Add(name string, torrent io.Reader, path string, label string) error {
	b, err := ioutil.ReadAll(torrent)
	if err != nil {
		return err
	}
	args := map[string][]byte{
		name: b,
	}
	return driver.qb.Torrent.AddFiles(args, &model.AddTorrentsOptions{
		Savepath: path,
		Category: label,
	})
}

func (driver QBittorrent) Announce(hash string) error {
	return driver.qb.Torrent.ReannounceTorrents([]string{hash})
}

func (driver QBittorrent) ClientVersion() (string, error) {
	appVer, err := driver.qb.Application.GetAppVersion()
	if err != nil {
		return "", err
	}
	apiVer, err := driver.qb.Application.GetAppVersion()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s / %s", appVer, apiVer), nil
}

func (driver QBittorrent) Close() error {
	return nil
}

func (driver QBittorrent) Move(hash string, dest string) error {
	return driver.qb.Torrent.SetLocations([]string{hash}, dest)
}

func (driver QBittorrent) Pause(hash string) error {
	return driver.qb.Torrent.StopTorrents([]string{hash})
}

func (driver QBittorrent) getHashes() ([]string, error) {
	var hashes []string
	torrents, err := driver.Torrents()
	if err != nil {
		return nil, err
	}
	for _, torrent := range torrents {
		hashes = append(hashes, torrent.Hash)
	}
	return hashes, nil
}

func (driver QBittorrent) PauseAll() error {
	hashes, err := driver.getHashes()
	if err != nil {
		return err
	}
	return driver.qb.Torrent.StopTorrents(hashes)
}

func (driver QBittorrent) Queue(hash string, position client.QueuePos) error {
	hashes := []string{hash}
	switch position {
	case client.Top:
		fallthrough
	case client.Up:
		return driver.qb.Torrent.IncreasePriority(hashes)
	default:
		return driver.qb.Torrent.DecreasePriority(hashes)
	}
}

func (driver QBittorrent) Remove(hash string, deleteData bool) error {
	return driver.qb.Torrent.DeleteTorrents([]string{hash}, deleteData)
}

func (driver QBittorrent) Start(hash string) error {
	return driver.qb.Torrent.ResumeTorrents([]string{hash})
}

func (driver QBittorrent) StartAll() error {
	hashes, err := driver.getHashes()
	if err != nil {
		return err
	}
	return driver.qb.Torrent.ResumeTorrents(hashes)
}

func (driver QBittorrent) Stop(hash string) error {
	return driver.qb.Torrent.StopTorrents([]string{hash})
}

func (driver QBittorrent) Torrent(hash string, torrent *client.Torrent) error {
	torrents, err := driver.qb.Torrent.GetList(&model.GetTorrentListOptions{
		Filter: model.FilterAll,
	})
	if err != nil {
		return err
	}
	for _, t := range torrents {
		if t.Hash == hash {
			mapTorrentStatus(t, torrent)
			return nil
		}
	}
	return client.ErrUnknownTorrent
}

func (driver QBittorrent) TorrentsWithState(statuses ...client.State) ([]client.Torrent, error) {
	torrents, err := driver.qb.Torrent.GetList(&model.GetTorrentListOptions{
		Filter: model.FilterAll,
	})
	if err != nil {
		return nil, err
	}
	var validStates []model.TorrentState
	for k, v := range stateMap {
		for _, s := range statuses {
			if s == v {
				validStates = append(validStates, k)
			}
		}
	}
	var validTorrents []client.Torrent
	for _, t := range torrents {
		for _, state := range validStates {
			if state == t.State {
				var tor client.Torrent
				mapTorrentStatus(t, &tor)
				validTorrents = append(validTorrents, tor)
				break
			}
		}
	}
	return validTorrents, nil
}

func (driver QBittorrent) Verify(hash string) error {
	return driver.qb.Torrent.RecheckTorrents([]string{hash})
}

func (driver QBittorrent) Login() error {
	if err := driver.qb.Login(driver.cfg.Username, driver.cfg.Password); err != nil {
		return errors.Wrapf(client.ErrAuthFailed, "Error trying to login: %v", err)
	}
	return nil
}

func mapTorrentStatus(status *model.Torrent, torrent *client.Torrent) {
	torrent.Hash = status.Hash
	torrent.Name = status.Name
	torrent.Size = int64(status.Size)
	torrent.Ratio = status.Ratio
	// TODO get path
	// torrent.Path = ???
}

func (driver QBittorrent) Torrents() ([]client.Torrent, error) {
	qTorrents, err := driver.qb.Torrent.GetList(nil)
	if err != nil {
		return nil, errors.Wrapf(client.ErrDriverError, "Failed to fetch torrents: %v", err)
	}
	var torrents []client.Torrent
	for _, t := range qTorrents {
		var torrent client.Torrent
		mapTorrentStatus(t, &torrent)
		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

type Factory struct{}

func (f Factory) New(cfg client.Config) (client.Driver, error) {
	url := fmt.Sprintf("http://%s:%d/api/v2", cfg.Host, cfg.Port)
	c := qbittorrent.NewClient(url, log.WithField("component", "QBitTorrent Client"))
	return QBittorrent{cfg: cfg, qb: c}, nil
}

func init() {
	if err := client.RegisterDriver(driverName, Factory{}); err != nil {
		log.Fatalf("Failed to register qbittorrent driver: %v", err)
	}
}
