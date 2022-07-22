package util

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

func OpenUniqueFile(baseDir string, filename string) (ret *os.File, err error) {
	now := time.Now()
	nowStr := now.Format("2006-01-02")
	dir := path.Join(baseDir, nowStr)
	if err = os.MkdirAll(dir, 0o700); err != nil {
		return
	}

	ext := path.Ext(filename)
	basename := strings.TrimSuffix(filename, ext)

	var fullpath string
	id := 0
	for {
		fullpath = path.Join(dir, filename)
		if _, err = os.Stat(fullpath); os.IsNotExist(err) {
			break
		} else if err == nil || os.IsExist(err) {
			id++
			filename = fmt.Sprintf("%s_%d%s", basename, id, ext)
			continue
		} else {
			return
		}
	}

	return os.OpenFile(fullpath, os.O_CREATE|os.O_WRONLY, 0x644)
}
