package internal

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path"
	"runtime"
)

const BLUETOOTH_UUID string = "94f45378-7d6d-437d-973b-fba39e49d4ee"
const MAGIC string = "ONeSaMa_DaiSuKi"

var MACHINE_NAME string
var MACHINE_ID [6]byte

func init() {
	var hostname string
	var err error
	if hostname, err = os.Hostname(); err != nil {
		hostname = "MyMachine"
	}

	MACHINE_NAME = fmt.Sprintf("%s (%s %s)", hostname, runtime.GOOS, runtime.GOARCH)

	var content []byte
	if content, err = fs.ReadFile(os.DirFS("/etc/"), "machine-id"); err != nil {
		panic(fmt.Sprintf("error reading /etc/machineid: %v\n", err))
	}
	sum := sha256.Sum256(content)
	copy(MACHINE_ID[:], sum[:6])
}

var APP_DIR string
var FILE_RECV_DIR string

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	APP_DIR = path.Join(homeDir, "kuroko")
	FILE_RECV_DIR = path.Join(APP_DIR, "FileRecv")
}
