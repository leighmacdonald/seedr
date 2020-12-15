package transmission

import (
	"github.com/hekmon/transmissionrpc"
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
)

const driverName = "transmission"

type Transmission struct {
	cfg    client.Config
	client *transmissionrpc.Client
}

func (d Transmission) Announce(hash string) error {
	panic("implement me")
}

func (d Transmission) ClientVersion() (string, error) {
	panic("implement me")
}

func (d Transmission) Close() error {
	panic("implement me")
}

func (d Transmission) Pause(hash string) error {
	panic("implement me")
}

func (d Transmission) PauseAll() error {
	panic("implement me")
}

func (d Transmission) Queue(hash string, position client.QueuePos) error {
	panic("implement me")
}

func (d Transmission) Start(hash string) error {
	panic("implement me")
}

func (d Transmission) StartAll() error {
	panic("implement me")
}

func (d Transmission) Stop(hash string) error {
	panic("implement me")
}

func (d Transmission) TorrentsWithState(statuses ...client.State) ([]client.Torrent, error) {
	panic("implement me")
}

func (d Transmission) Verify(hash string) error {
	panic("implement me")
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

func (d Transmission) Torrents() ([]client.Torrent, error) {
	panic("implement me")
}

func (d Transmission) Torrent(hash string, torrent *client.Torrent) error {
	panic("implement me")
}

func (d Transmission) Move(hash string, dest string) error {
	panic("implement me")
}

func (d Transmission) Remove(hash string, deleteData bool) error {
	panic("implement me")
}

func (d Transmission) Add(filename string, torrent io.Reader, path string, label string) error {
	panic("implement me")
}

type Factory struct{}

func (f Factory) New(cfg client.Config) (client.Driver, error) {
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
