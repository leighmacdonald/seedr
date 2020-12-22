package transmission

import (
	"encoding/base64"
	"fmt"
	"github.com/hekmon/transmissionrpc"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
)

const driverName = "transmission"

var (
	stateMap = map[transmissionrpc.TorrentStatus]client.State{
		transmissionrpc.TorrentStatusStopped:      client.Paused,
		transmissionrpc.TorrentStatusCheckWait:    client.Checking,
		transmissionrpc.TorrentStatusCheck:        client.Checking,
		transmissionrpc.TorrentStatusDownloadWait: client.Queued,
		transmissionrpc.TorrentStatusDownload:     client.Downloading,
		transmissionrpc.TorrentStatusSeedWait:     client.Queued,
		transmissionrpc.TorrentStatusSeed:         client.Seeding,
		transmissionrpc.TorrentStatusIsolated:     client.Unknown,
	}
)

type Transmission struct {
	cfg    *client.Config
	client *transmissionrpc.Client
}

func (d Transmission) FreeSpace(path string) (int64, error) {
	panic("implement me")
}

func (d Transmission) Announce(hash string) error {
	return d.client.TorrentReannounceHashes([]string{hash})
}

func (d Transmission) ClientVersion() (string, error) {
	ok, verA, verB, err := d.client.RPCVersion()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("RPC version too low")
	}
	return fmt.Sprintf("Transmission %d/%d", verA, verB), nil
}

func (d Transmission) Close() error {
	return d.client.SessionClose()
}

func (d Transmission) Pause(hash string) error {
	return d.client.TorrentStopHashes([]string{hash})
}

func (d Transmission) PauseAll() error {
	torrents, err := d.Torrents()
	if err != nil {
		return err
	}
	var hashes []string
	for _, torrent := range torrents {
		hashes = append(hashes, torrent.Hash)
	}
	return d.client.TorrentStopHashes(hashes)
}

func (d Transmission) getIDs(hashes []string) ([]int64, error) {
	var ids []int64
	all, err := d.client.TorrentGetAll()
	if err != nil {
		return nil, err
	}
	for _, torrent := range all {
		for _, hash := range hashes {
			if *torrent.HashString == hash {
				ids = append(ids, *torrent.ID)
				break
			}
		}
	}
	return ids, nil
}

func (d Transmission) Queue(hash string, position client.QueuePos) error {
	ids, err := d.getIDs([]string{hash})
	if err != nil {
		return err
	}
	switch position {
	case client.Top:
		return d.client.QueueMoveTop(ids)
	case client.Up:
		return d.client.QueueMoveUp(ids)
	case client.Down:
		return d.client.QueueMoveDown(ids)
	default:
		return d.client.QueueMoveBottom(ids)
	}
}

func (d Transmission) Start(hash string) error {
	return d.client.TorrentStartHashes([]string{hash})
}

func (d Transmission) StartAll() error {
	return d.client.TorrentStartHashes(nil)
}

func (d Transmission) Stop(hash string) error {
	return d.client.TorrentStopHashes([]string{hash})
}

func (d Transmission) TorrentsWithState(statuses ...client.State) ([]client.Torrent, error) {
	var validStatuses []transmissionrpc.TorrentStatus
	for k, v := range stateMap {
		for _, status := range statuses {
			if v == status {
				validStatuses = append(validStatuses, k)
			}
		}
	}
	torrents, err := d.client.TorrentGetAll()
	if err != nil {
		return nil, err
	}
	var validTorrents []client.Torrent
	for _, torrent := range torrents {
		for _, validState := range validStatuses {
			if *torrent.Status == validState {
				var t client.Torrent
				mapTorrentStatus(torrent, &t)
				validTorrents = append(validTorrents, t)
				break
			}
		}
	}
	return validTorrents, nil
}

func (d Transmission) Verify(hash string) error {
	return d.client.TorrentVerifyHashes([]string{hash})
}

func (d Transmission) Login() error {
	ok, serverVersion, serverMinimumVersion, err := d.client.RPCVersion()
	if err != nil {
		return errors.Wrapf(client.ErrDriverError, "Failed client checks: %v", err)
	}
	if !ok {
		return errors.Wrapf(client.ErrAuthFailed, "Remote transmission RPC version (v%d) is incompatible with the transmission library (v%d): remote needs at least v%d",
			serverVersion, transmissionrpc.RPCVersion, serverMinimumVersion)
	}
	log.Debugf("Remote transmission RPC version (v%d) is compatible with our transmissionrpc library (v%d)\n",
		serverVersion, transmissionrpc.RPCVersion)
	return nil
}

func mapTorrentStatus(status *transmissionrpc.Torrent, torrent *client.Torrent) {
	torrent.Hash = *status.HashString
	torrent.Name = *status.Name
	torrent.Path = *status.DownloadDir

}

func (d Transmission) Torrents() ([]client.Torrent, error) {
	all, err := d.client.TorrentGetAll()
	if err != nil {
		return nil, err
	}
	var torrents []client.Torrent
	for _, t := range all {
		var torrent client.Torrent
		mapTorrentStatus(t, &torrent)
		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

func (d Transmission) Torrent(hash string, torrent *client.Torrent) error {
	torrents, err := d.client.TorrentGetAllForHashes([]string{hash})
	if err != nil {
		return err
	}
	if len(torrents) != 1 {
		return client.ErrUnknownTorrent
	}
	mapTorrentStatus(torrents[0], torrent)
	return nil
}

func (d Transmission) Move(hash string, dest string) error {
	return d.client.TorrentSetLocationHash(hash, dest, true)
}

func (d Transmission) Remove(hash string, deleteData bool) error {
	torrents, err := d.client.TorrentGetAllForHashes([]string{hash})
	if err != nil {
		return errors.Wrapf(client.ErrDriverError, "Failed to get torrent to delete: %v", err)
	}
	if len(torrents) != 1 {
		return client.ErrUnknownTorrent
	}
	return d.client.TorrentRemove(&transmissionrpc.TorrentRemovePayload{
		IDs:             []int64{*torrents[0].ID},
		DeleteLocalData: deleteData,
	})
}

func (d Transmission) Add(filename string, torrent io.Reader, path string, _ string) error {
	b, err := ioutil.ReadAll(torrent)
	if err != nil {
		return err
	}
	b64 := base64.StdEncoding.EncodeToString(b)
	paused := false
	if _, err := d.client.TorrentAdd(&transmissionrpc.TorrentAddPayload{
		Paused:      &paused,
		DownloadDir: &path,
		//Filename:    &filename,
		MetaInfo: &b64,
	}); err != nil {
		return errors.Wrapf(err, "Failed to upload new torrent")
	}
	return nil
}

type Factory struct{}

func (f Factory) New(cfg *client.Config) (client.Driver, error) {
	c, err := transmissionrpc.New(cfg.Host, cfg.Username, cfg.Password,
		&transmissionrpc.AdvancedConfig{
			HTTPS: cfg.TLS,
			Port:  cfg.Port,
		})
	if err != nil {
		return nil, errors.Wrapf(client.ErrDriverError, "Failed to create driver instance: %v", err)
	}
	return Transmission{cfg: cfg, client: c}, nil
}

func init() {
	if err := client.RegisterDriver(driverName, Factory{}); err != nil {
		log.Fatalf("Failed to register transmission driver: %v", err)
	}
}
