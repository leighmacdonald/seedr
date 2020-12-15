package deluge

import (
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeluge(t *testing.T) {
	f := Factory{}
	c, err := f.New(client.Config{
		Driver:   "deluge",
		Username: "test_user",
		Password: "test_pass",
		Host:     "192.168.0.200",
		Port:     58846,
		TLS:      false,
	})
	require.NoErrorf(t, err, "failed to create deluge client")
	client.DriverTestSuite(t, c)
}
