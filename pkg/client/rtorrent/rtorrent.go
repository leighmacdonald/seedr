package rtorrent

import (
	"fmt"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/mrobinsn/go-rtorrent/rtorrent"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"io"
)

const driverName = "rtorrent"

// https://github.com/Novik/ruTorrent/blob/44d43229f07212f20b53b6301fb25882125876c3/plugins/httprpc/action.php#L90
const (
	DStop      rtorrent.Field = "d.stop"
	DClose     rtorrent.Field = "d.close"
	DStart     rtorrent.Field = "d.start"
	DOpen      rtorrent.Field = "d.open"
	DDownTotal rtorrent.Field = "d.down.total"
	DDownRate  rtorrent.Field = "d.down.rate"
	DUPTotal   rtorrent.Field = "d.up.total"
	DUPRate    rtorrent.Field = "d.up.rate"
	DUpTotal   rtorrent.Field = "d.get_up_total"
	DSeeders   rtorrent.Field = "d.get_peers_complete"
	DLeechers  rtorrent.Field = "d.get_peers_connected"
	// States
	DHashChecking rtorrent.Field = "d.is_hash_checking"
	DPriority     rtorrent.Field = "d.get_priority"

	DIsActive     rtorrent.Field = "d.is_active"
	DFreeSpace    rtorrent.Field = "d.get_free_diskspace"
	DCreationDate rtorrent.Field = "d.get_creation_date"
	DState        rtorrent.Field = "d.get_state"
	DMessage      rtorrent.Field = "d.get_message"

	// Use custom1 for labels
	DSetLabel rtorrent.Field = "d.set_custom1"
	DGetLabel rtorrent.Field = "d.get_custom1"

	DVerify rtorrent.Field = "d.check_hash"
)

type RTorrent struct {
	cfg       *client.Config
	c         *rtorrent.RTorrent
	connected bool
}

func (d RTorrent) FreeSpace(path string) (int64, error) {
	panic("implement me")
}

func (d RTorrent) Announce(hash string) error {
	log.Debugf("Announce() not implemented")
	return nil
}

func (d RTorrent) ClientVersion() (string, error) {
	name, err := d.c.Name()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rtorrent %s", name), nil
}

func (d RTorrent) Close() error {
	return nil
}

func (d RTorrent) Pause(hash string) error {
	log.Debugf("Pause() not implemented")
	return nil
}

func (d RTorrent) PauseAll() error {
	log.Debugf("Pause() not implemented")
	return nil
}

func (d RTorrent) Queue(hash string, position client.QueuePos) error {
	log.Debugf("Queue() not implemented")
	return nil
}

func (d RTorrent) Start(hash string) error {
	log.Debugf("Start() not implemented")
	return nil
}

func (d RTorrent) StartAll() error {
	log.Debugf("StartAll() not implemented")
	return nil
}

func (d RTorrent) Stop(hash string) error {
	log.Debugf("Stop() not implemented")
	return nil
}

func (d RTorrent) GetTorrents(view rtorrent.View) ([]rtorrent.Torrent, error) {
	args := []interface{}{"", string(view),
		rtorrent.DName.Query(),
		rtorrent.DSizeInBytes.Query(),
		rtorrent.DHash.Query(),
		rtorrent.DLabel.Query(),
		rtorrent.DBasePath.Query(),
		rtorrent.DIsActive.Query(),
		rtorrent.DComplete.Query(),
		rtorrent.DRatio.Query(),
	}
	results, err := d.c.XMLPRCClient().Call("d.multicall2", args...)
	var torrents []rtorrent.Torrent
	if err != nil {
		return torrents, errors.Wrap(err, "d.multicall2 XMLRPC call failed")
	}
	for _, outerResult := range results.([]interface{}) {
		for _, innerResult := range outerResult.([]interface{}) {
			torrentData := innerResult.([]interface{})
			torrents = append(torrents, rtorrent.Torrent{
				Hash:      torrentData[2].(string),
				Name:      torrentData[0].(string),
				Path:      torrentData[4].(string),
				Size:      torrentData[1].(int),
				Label:     torrentData[3].(string),
				Completed: torrentData[6].(int) > 0,
				Ratio:     float64(torrentData[7].(int)) / float64(1000),
			})
		}
	}
	return torrents, nil
}

func (d RTorrent) TorrentsWithState(statuses ...client.State) ([]client.Torrent, error) {
	rTorrents, err := d.c.GetTorrents(rtorrent.ViewMain)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch torrents")
	}
	for _, rt := range rTorrents {
		log.Println(rt)
	}
	var torrents []client.Torrent
	return torrents, nil
}

func (d RTorrent) Verify(hash string) error {
	log.Debugf("Verify() not implemented")
	return nil
}

func (d RTorrent) Login() error {
	if d.connected {
		log.Warn("Already connected")
		return nil
	}
	name, err := d.c.Name()
	if err != nil {
		return err
	}
	log.Debugf("Connected to rtorrent: %s", name)
	d.connected = true
	return nil
}

func (d RTorrent) Torrents() ([]client.Torrent, error) {
	rt, err := d.c.GetTorrents(rtorrent.ViewMain)
	if err != nil {
		return nil, err
	}
	var torrents []client.Torrent
	for _, t := range rt {
		status, err := d.c.GetStatus(t)
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
	b, err := ioutil.ReadAll(torrent)
	if err != nil {
		return err
	}
	if err := d.c.AddTorrent(b,
		rtorrent.DName.SetValue(name),
		rtorrent.DLabel.SetValue(label),
		rtorrent.DBasePath.SetValue(path)); err != nil {
		return err
	}
	return nil
}

type Factory struct{}

func (f Factory) New(cfg *client.Config) (client.Driver, error) {
	var url string
	login := ""
	if cfg.Username != "" && cfg.Password != "" {
		login = fmt.Sprintf("%s:%s@", cfg.Username, cfg.Password)
	}
	if cfg.TLS {
		url = fmt.Sprintf("https://%s%s:%d/RPC2", login, cfg.Host, cfg.Port)
	} else {
		url = fmt.Sprintf("http://%s%s:%d/RPC2", login, cfg.Host, cfg.Port)
	}
	c := rtorrent.New(url, cfg.TLS)
	return RTorrent{cfg: cfg, c: c}, nil
}

func init() {
	if err := client.RegisterDriver(driverName, Factory{}); err != nil {
		log.Fatalf("Failed to register rtorrent driver: %v", err)
	}
}
