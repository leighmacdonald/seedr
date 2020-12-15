package qbittorrent

import (
	"fmt"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superturkey650/go-qbittorrent/qbt"
	"io"
	"time"
)

const driverName = "qbittorrent"

type QBittorrent struct {
	cfg client.Config
	qb  *qbt.Client
}

func (driver QBittorrent) Add(name string, torrent io.Reader, path string, label string) error {
	panic("implement me")
}

func (driver QBittorrent) Announce(hash string) error {
	panic("implement me")
}

func (driver QBittorrent) ClientVersion() (string, error) {
	panic("implement me")
}

func (driver QBittorrent) Close() error {
	panic("implement me")
}

func (driver QBittorrent) Move(hash string, dest string) error {
	panic("implement me")
}

func (driver QBittorrent) Pause(hash string) error {
	panic("implement me")
}

func (driver QBittorrent) PauseAll() error {
	panic("implement me")
}

func (driver QBittorrent) Queue(hash string, position client.QueuePos) error {
	panic("implement me")
}

func (driver QBittorrent) Remove(hash string, deleteData bool) error {
	panic("implement me")
}

func (driver QBittorrent) Start(hash string) error {
	panic("implement me")
}

func (driver QBittorrent) StartAll() error {
	panic("implement me")
}

func (driver QBittorrent) Stop(hash string) error {
	panic("implement me")
}

func (driver QBittorrent) Torrent(hash string, torrent *client.Torrent) error {
	panic("implement me")
}

func (driver QBittorrent) TorrentsWithState(statuses ...client.State) ([]client.Torrent, error) {
	panic("implement me")
}

func (driver QBittorrent) Verify(hash string) error {
	panic("implement me")
}

func (driver QBittorrent) Login() error {
	success, err := driver.qb.Login(driver.cfg.Username, driver.cfg.Password)
	if err != nil {
		return errors.Wrapf(client.ErrDriverError, "Error trying to login: %v", err)
	}
	if !success {
		return client.ErrAuthFailed
	}
	return nil
}

func (driver QBittorrent) Torrents() ([]client.Torrent, error) {
	qTorrents, err := driver.qb.Torrents(nil)
	if err != nil {
		return nil, errors.Wrapf(client.ErrDriverError, "Failed to fetch torrents: %v", err)
	}
	var torrents []client.Torrent
	for _, t := range qTorrents {
		torrent := client.Torrent{
			Name:       t.Name,
			Hash:       t.Hash,
			Path:       t.SavePath,
			Ratio:      float64(t.Ratio),
			Tracker:    "",
			Label:      t.Category,
			Size:       0,
			AddedOn:    time.Time{},
			SeedTime:   0,
			Seeds:      0,
			Peers:      0,
			SpeedUP:    int64(t.Upspeed),
			SpeedDN:    int64(t.Dlspeed),
			Uploaded:   0,
			Downloaded: 0,
			StatusMsg:  "",
		}
		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

type Factory struct{}

func (f Factory) New(cfg client.Config) (client.Driver, error) {
	url := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
	c := qbt.NewClient(url)
	return QBittorrent{cfg: cfg, qb: c}, nil
}

func init() {
	if err := client.RegisterDriver(driverName, Factory{}); err != nil {
		log.Fatalf("Failed to register qbittorrent driver: %v", err)
	}
}
