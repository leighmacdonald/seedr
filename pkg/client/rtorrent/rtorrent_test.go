package rtorrent

import (
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient(t *testing.T) {
	driver, err := Factory{}.New(client.Config{
		Username: "",
		Password: "",
		Host:     "",
		Port:     0,
	})
	require.NoError(t, err, "Failed to setup driver")
	client.DriverTestSuite(t, driver)
}
