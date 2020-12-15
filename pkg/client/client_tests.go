package client

import (
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/leighmacdonald/golib"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

var (
	builtinAnnounceList = [][]string{
		{"udp://tracker.openbittorrent.com:80"},
		{"udp://tracker.publicbt.com:80"},
		{"udp://tracker.istole.it:6969"},
	}
)

// findParent will walk up the directory tree until it find a file. Max depth of 4 or the minRootDir directory
// is matched
func findParent(p string, minRootDir string) string {
	var dots []string
	for i := 0; i < 10; i++ {
		dir := path.Join(dots...)
		fPath := path.Join(dir, p)
		if golib.Exists(fPath) {
			fp, err := filepath.Abs(fPath)
			if err == nil {
				return fp
			}
			return fp
		}
		if strings.HasSuffix(dir, minRootDir) {
			return p
		}
		dots = append(dots, "..")
	}
	return p
}

func generateTorrent() (string, string) {
	rootDir := filepath.Join(findParent("seedr", "seedr"), ".temp")
	if !golib.Exists(rootDir) {
		if err := os.MkdirAll(rootDir, 0755); err != nil {
			log.Fatalf("Failed to create temp work dir: %s", err)
		}
	}
	file, err := ioutil.TempFile(rootDir, "seedr.*.bin")
	if err != nil {
		log.Fatal(err)
	}
	b, err := golib.GenRandomBytes(10000)
	if _, err := file.Write(b); nil != err {
		log.Fatal(err)
	}
	mi := metainfo.MetaInfo{
		AnnounceList: builtinAnnounceList,
		Comment:      "seedr test torrent",
		CreatedBy:    "seedr",
	}
	info := metainfo.Info{
		PieceLength: 256 * 1024,
	}
	if err := info.BuildFromFilePath(file.Name()); err != nil {
		log.Fatalf("Failed to build from file path: %v", err)
	}
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		log.Fatalf("Failed to generate torrent: %v", err)
	}
	root := filepath.Dir(file.Name())
	torrentFileName := filepath.Join(root, strings.Replace(filepath.Base(file.Name()), ".bin", ".torrent", -1))
	fp, err := os.Create(torrentFileName)
	if err != nil {
		log.Fatalf("Failed to create test torrent: %v", err)
	}
	if err := mi.Write(fp); err != nil {
		log.Fatalf("Failed to write test torrent: %v", err)
	}
	return filepath.Base(file.Name()), torrentFileName
}

func DriverTestSuite(t *testing.T, driver Driver) {
	testBinFile, testTorrentFile := generateTorrent()
	defer func() {
		_ = os.Remove(testTorrentFile)
		_ = os.Remove(testBinFile)
	}()
	filename := filepath.Base(testTorrentFile)
	require.NoErrorf(t, driver.Login(), "Failed to login")
	fp, err := os.Open(testTorrentFile)
	require.NoError(t, err, "Failed to open test torrent")
	require.NoError(t, driver.Add(filename, fp, "/downloads", "test"))
}
