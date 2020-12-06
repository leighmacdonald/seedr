# seedr

A tool to automatically manage seed boxes for optimal usage using a declarative configuration.

## Features

- [ ] **Tiered storage** This allows you to configure different HDD / SSD for different purposes. For example, you
can have a hot storage drive (Fast, SSD) used for all new / very active downloads. Once it cools down in activity
you can automatically move it to a slower tier (HDD) storage when certain thresholds like age, ratio, space available,
or a combination of several, are met.   

- [ ] **Client support**
    - [x] Deluge (Via built in RPC interface, not the WebUI plugin)
    - [ ] Transmission
    - [ ] rTorrent
    - [ ] qBittorrent
    
- [ ] **Triggers** These are the different strategies employed to decide if a torrent should be moved
    - [ ] Max Ratio
    - [ ] Disk Space Free
    - [ ] Age
    
- [ ] **Notifications**
    - [ ] IRC
    - [ ] Discord
    - [ ] Webhook