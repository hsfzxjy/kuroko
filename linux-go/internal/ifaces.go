package internal

import "kmux"

type BridgeInfo interface {
	GetType() kmux.TransportType
	GetAddr() string
}

type bridgeInfo struct {
	typ  kmux.TransportType
	addr string
}

func (bi *bridgeInfo) GetType() kmux.TransportType { return bi.typ }
func (bi *bridgeInfo) GetAddr() string             { return bi.addr }

func Bi(typ kmux.TransportType, addr string) BridgeInfo {
	return &bridgeInfo{typ, addr}
}
