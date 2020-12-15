package client

import (
	"github.com/pkg/errors"
	"io"
	"sync"
	"time"
)

var (
	drivers            map[string]DriverFactory
	driversMu          *sync.RWMutex
	ErrInvalidDriver   = errors.New("Invalid driver")
	ErrUnknownTorrent  = errors.New("Unknown torrent")
	ErrDuplicateDriver = errors.New("Duplicate driver name")
	ErrAuthFailed      = errors.New("Authentication failed")
	ErrDriverError     = errors.New("Backend driver error")
)

type State int

const (
	Unknown State = iota
	Active
	Allocating
	Checking
	Downloading
	Seeding
	Paused
	Error
	Queued
	Moving
	Any
)

type QueuePos int

const (
	Top QueuePos = iota
	Up
	Down
	Bottom
)

// Driver defines our common interface for interacting with the backend torrent clients
type Driver interface {
	Add(filename string, torrent io.Reader, path string, label string) error
	Announce(hash string) error
	ClientVersion() (string, error)
	Close() error
	Login() error
	Move(hash string, dest string) error
	Pause(hash string) error
	PauseAll() error
	Queue(hash string, position QueuePos) error
	Remove(hash string, deleteData bool) error
	Start(hash string) error
	StartAll() error
	Stop(hash string) error
	Torrent(hash string, torrent *Torrent) error
	Torrents() ([]Torrent, error)
	TorrentsWithState(statuses ...State) ([]Torrent, error)
	Verify(hash string) error
}

type DriverFactory interface {
	New(cfg Config) (Driver, error)
}

type Config struct {
	Driver   string `mapstructure:"driver"`
	Username string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     uint16 `mapstructure:"port"`
	TLS      bool   `mapstructure:"tls"`
}

// Torrent is a common data container for the backend Driver
type Torrent struct {
	Name    string
	Hash    string
	Path    string
	Ratio   float64
	Tracker string
	Label   string
	Size    int64
	AddedOn time.Time
	// Completed = Seedtime > 0 ?
	SeedTime   time.Duration
	Seeds      int
	Peers      int
	SpeedUP    int64
	SpeedDN    int64
	Uploaded   int64
	Downloaded int64
	StatusMsg  string
}

func New(cfg Config) (Driver, error) {
	driversMu.RLock()
	defer driversMu.RUnlock()
	factory, found := drivers[cfg.Driver]
	if !found {
		return nil, ErrInvalidDriver
	}
	return factory.New(cfg)
}

func RegisterDriver(name string, factory DriverFactory) error {
	driversMu.Lock()
	defer driversMu.Unlock()
	_, found := drivers[name]
	if found {
		return ErrDuplicateDriver
	}
	drivers[name] = factory
	return nil
}

func init() {
	drivers = make(map[string]DriverFactory)
	driversMu = &sync.RWMutex{}
}
