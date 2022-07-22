package kmux

type TransportType byte

const (
	TT_BLUETOOTH TransportType = 1 + iota
	TT_LANv4
)
