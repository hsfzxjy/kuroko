module kuroko-linux

go 1.18

require (
	github.com/godbus/dbus/v5 v5.1.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/muka/go-bluetooth v0.0.0-20220604035144-0b043d86dc03
	github.com/pkg/errors v0.9.1
	github.com/rivo/tview v0.0.0-20220307222120-9994674d60a8
	github.com/sevlyar/go-daemon v0.1.5
	golang.org/x/sys v0.0.0-20220622161953-175b2fd9d664
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/jmoiron/sqlx v1.3.5
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20220517211312-f3a8303e98df // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.36.0 // indirect
	modernc.org/ccgo/v3 v3.16.6 // indirect
	modernc.org/libc v1.16.8 // indirect
	modernc.org/mathutil v1.4.1 // indirect
	modernc.org/memory v1.1.1 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.17.3
	modernc.org/strutil v1.1.2 // indirect
	modernc.org/token v1.0.0 // indirect
)

require (
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/hsfzxjy/smux v1.5.17-0.20220716142752-b0d716d03697 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.13 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/xtaci/smux v1.5.16 // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	golang.org/x/text v0.3.7 // indirect
)

require (
	github.com/gdamore/tcell/v2 v2.5.1
	github.com/hsfzxjy/go-srpc v0.1.8
	kmux v0.0.0
)

replace kmux => ../lib/kmux/
