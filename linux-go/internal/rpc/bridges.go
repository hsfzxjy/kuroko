package rpc

import (
	"kuroko-linux/models"
)

func (Core) GetBridgeInfoList(arg int, ret *[]*models.Bridge) (err error) {
	*ret, err = models.GetBridgeList()
	return
}
