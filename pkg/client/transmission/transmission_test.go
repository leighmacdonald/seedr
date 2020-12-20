package transmission

import (
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQBittorrent(t *testing.T) {
	f := Factory{}
	c, err := f.New(client.Config{
		Driver:   "transmission",
		Username: "test_user",
		Password: "test_pass",
		Host:     "192.168.0.200",
		Port:     8090,
		TLS:      false,
	})
	require.NoErrorf(t, err, "failed to create qbittorrent client")
	client.DriverTestSuite(t, c)
}
