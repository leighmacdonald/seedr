module github.com/leighmacdonald/seedr

go 1.15

require (
	github.com/KnutZuidema/go-qbittorrent v0.0.0-20190814183140-292286ded47f
	github.com/anacrolix/torrent v1.18.1
	github.com/dustin/go-humanize v1.0.0
	github.com/gdm85/go-libdeluge v0.5.4
	github.com/hekmon/transmissionrpc v1.1.0
	github.com/leighmacdonald/golib v1.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mrobinsn/go-rtorrent v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil/v3 v3.20.11
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
)

replace (
	github.com/gdm85/go-libdeluge v0.5.4 => github.com/leighmacdonald/go-libdeluge v0.5.4
	github.com/mrobinsn/go-rtorrent v1.5.0 => github.com/leighmacdonald/go-rtorrent v1.5.1-0.20201220050726-3e0ef1d34434
)
