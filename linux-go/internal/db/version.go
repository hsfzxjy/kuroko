package db

import (
	"encoding/binary"
	"kuroko-linux/internal"
	"os"
	"path"
)

const currentDbVersion uint64 = 1

func dbIsOutdated() bool {
	content, err := os.ReadFile(path.Join(internal.APP_DIR, "db_version"))
	if err != nil {
		return true
	}
	ver := binary.BigEndian.Uint64(content)
	return ver != currentDbVersion
}

func dumpDbVersion() error {
	content := make([]byte, 8)
	binary.BigEndian.PutUint64(content, currentDbVersion)
	filename := path.Join(internal.APP_DIR, "db_version")
	return os.WriteFile(filename, content, 0o600)
}
