package rtorrent

import (
	"fmt"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/mrobinsn/go-rtorrent/rtorrent"
	"time"

	log "github.com/sirupsen/logrus"
	"io"
)

const driverName = "rtorrent"

type RTorrent struct {
	cfg       client.Config
	conn      *rtorrent.RTorrent
	connected bool
}

func (d RTorrent) Announce(hash string) error {
	panic("implement me")
}

func (d RTorrent) ClientVersion() (string, error) {
	panic("implement me")
}

func (d RTorrent) Close() error {
	panic("implement me")
}

func (d RTorrent) Pause(hash string) error {
	panic("implement me")
}

func (d RTorrent) PauseAll() error {
	panic("implement me")
}

func (d RTorrent) Queue(hash string, position client.QueuePos) error {
	panic("implement me")
}

func (d RTorrent) Start(hash string) error {
	panic("implement me")
}

func (d RTorrent) StartAll() error {
	panic("implement me")
}

func (d RTorrent) Stop(hash string) error {
	panic("implement me")
}

func (d RTorrent) TorrentsWithState(statuses ...client.State) ([]client.Torrent, error) {
	panic("implement me")
}

func (d RTorrent) Verify(hash string) error {
	panic("implement me")
}

func (d RTorrent) Login() error {
	if d.connected {
		log.Warn("Already connected")
		return nil
	}
	var url string
	if d.cfg.Username != "" {
		url = fmt.Sprintf("http://%s:%s@%s:%d/RPC2", d.cfg.Username, d.cfg.Password, d.cfg.Host, d.cfg.Port)
	} else {
		url = fmt.Sprintf("http://%s:%d/RPC2", d.cfg.Host, d.cfg.Port)
	}
	newConn := rtorrent.New(url, true)
	name, err := newConn.Name()
	if err != nil {
		return err
	}
	log.Debugf("Connected to rtorrent: %s", name)
	d.conn = newConn
	d.connected = true
	return nil
}

func (d RTorrent) Torrents() ([]client.Torrent, error) {
	rt, err := d.conn.GetTorrents(rtorrent.ViewMain)
	if err != nil {
		return nil, err
	}
	var torrents []client.Torrent
	for _, t := range rt {
		status, err := d.conn.GetStatus(t)
		if err != nil {
			log.Warnf("Failed to get status for torrent: %v", err)
			continue
		}
		torrent := client.Torrent{
			Name:       t.Name,
			Hash:       t.Hash,
			Path:       t.Path,
			Ratio:      t.Ratio,
			Label:      t.Label,
			Size:       int64(t.Size),
			AddedOn:    time.Time{},
			SeedTime:   0,
			Seeds:      0,
			Peers:      0,
			SpeedUP:    int64(status.UpRate),
			SpeedDN:    int64(status.DownRate),
			Uploaded:   0,
			Downloaded: int64(status.CompletedBytes),
			StatusMsg:  "",
		}
		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

func (d RTorrent) Torrent(hash string, torrent *client.Torrent) error {
	panic("implement me")
}

func (d RTorrent) Move(hash string, dest string) error {
	panic("implement me")
}

func (d RTorrent) Remove(hash string, deleteData bool) error {
	panic("implement me")
}

func (d RTorrent) Add(name string, torrent io.Reader, path string, label string) error {
	panic("implement me")
}

type Factory struct{}

func (f Factory) New(cfg client.Config) (client.Driver, error) {
	return RTorrent{cfg: cfg}, nil
}

func init() {
	if err := client.RegisterDriver(driverName, Factory{}); err != nil {
		log.Fatalf("Failed to register rtorrent driver: %v", err)
	}
}
