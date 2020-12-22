package main

import (
	"github.com/leighmacdonald/seedr/cmd"
	_ "github.com/leighmacdonald/seedr/pkg/client/deluge"
	//_ "github.com/leighmacdonald/seedr/pkg/client/qbittorrent"
	//_ "github.com/leighmacdonald/seedr/pkg/client/rtorrent"
	//_ "github.com/leighmacdonald/seedr/pkg/client/transmission"
)

func main() {
	cmd.Execute()
}
