package rpc

import (
	"kuroko-linux/endpoints/outward"
	"kuroko-linux/internal/exc"
	"kuroko-linux/internal/session"
	"kuroko-linux/models"
)

type SendArgs struct {
	Filename string
	Bridge   *models.Bridge
}

func (c Core) SendFile(args SendArgs, ret *bool) (err error) {
	sess, err := session.DialSession(args.Bridge)
	if err != nil {
		err = exc.Wrap(err)
		return
	}
	err = outward.SendFile(sess, args.Filename)

	return
}
