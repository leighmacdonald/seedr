package rtorrent

import (
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient(t *testing.T) {
	driver, err := Factory{}.New(client.Config{
		Driver:   driverName,
		Username: "",
		Password: "",
		Host:     "192.168.0.200",
		Port:     5000,
		TLS:      false,
	})
	require.NoError(t, err, "Failed to setup driver")
	client.DriverTestSuite(t, driver)
}
