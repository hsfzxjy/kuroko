package outward

import (
	"kuroko-linux/internal"
	"kuroko-linux/internal/exc"
	"kuroko-linux/internal/session"
	"kuroko-linux/models"
	"os"
	"path"
)

func run(sess *session.Session, f func()) (err error) {
	sess.SetPanicOnError(true)
	defer func() {
		if p := recover(); p != nil {
			var ok bool
			err, ok = p.(error)
			if !ok {
				err = exc.New("%+v", p)
			}
		}
	}()
	defer sess.Close()
	f()
	return
}

func SendFile(sess *session.Session, filename string) error {
	return run(sess, func() {
		file, err := os.Open(filename)
		if err != nil {
			panic(exc.Wrap(err))
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			panic(exc.Wrap(err))
		}
		sess.WriteMagic()
		sess.WriteOpcode(0x00)
		sess.WriteIdentity()
		sess.WriteStr(path.Base(filename))
		sess.WriteUint64(uint64(stat.Size()))
		sess.WriteFromReader(file)
	})
}

func ExchangeIdentity(sess *session.Session) error {
	return run(sess, func() {
		sess.WriteMagic()
		sess.WriteOpcode(0x01)
		sess.WriteIdentity()
		sess.WriteStr(internal.MACHINE_NAME)
		if err := sess.Flush(); err != nil {
			panic(exc.Wrap(err))
		}
		var peer models.Peer
		peer.Id, _ = session.ReadUInt[uint64](sess, 6)
		peer.Name, _ = sess.ReadStr()
		sess.Bridge.BindPeer(nil, &peer)
	})
}
