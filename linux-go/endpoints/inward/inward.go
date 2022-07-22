package inward

import (
	"os"

	"kuroko-linux/internal"
	"kuroko-linux/internal/exc"
	"kuroko-linux/internal/log"
	"kuroko-linux/internal/session"
	"kuroko-linux/models"
	"kuroko-linux/util"
)

var logger = log.Name("inward")

func Route(sess *session.Session) {
	sess.SetPanicOnError(true)

	defer func() {
		if err := recover(); err != nil {
			logger.Panic(err)
		}
		sess.Close()
	}()

	sess.RequireMagic()
	opcode := sess.ReadOpcode()

	var endpoint func(*session.Session)

	switch opcode {
	case 0x00:
		endpoint = recieve_file
	case 0x01:
		endpoint = exchange_identity
	default:
		panic(exc.Fields("Opcode", opcode).Msg("Unknown opcode"))
	}
	endpoint(sess)
}

// 0x00
func recieve_file(b *session.Session) {
	b.RequireIdentity(nil)
	filename, _ := b.ReadStr()
	ctlen, _ := b.ReadUint64()

	var file *os.File
	var err error
	var success = false
	file, err = util.OpenUniqueFile(internal.FILE_RECV_DIR, filename)
	if err != nil {
		goto EXCEPT
	}
	defer func() {
		if !success {
			os.Remove(file.Name())
		}
	}()
	defer file.Close()

	_, err = file.ReadFrom(util.NewReaderWithLength(b.Rw, ctlen))
	if err != nil {
		goto EXCEPT
	}

	logger.Fields(
		"PeerId", b.Bridge.Peer.Id,
		"PeerName", b.Bridge.Peer.Name,
		"FileName", file.Name(),
	).Info("New file recieved")

	success = true
	return
EXCEPT:
	panic(exc.Msg("Unable to recieve file").Wrap(err))
}

// 0x01
func exchange_identity(b *session.Session) {
	var peer models.Peer
	peer.Id, _ = session.ReadUInt[uint64](b, 6)
	peer.Name, _ = b.ReadStr()
	b.Bridge.BindPeer(nil, &peer)
	logger.Fields(
		"PeerId", peer.Id,
		"PeerName", peer.Name,
	).Info("New peer paired")
	b.WriteBytes(internal.MACHINE_ID[:])
	b.WriteStr(internal.MACHINE_NAME)
}
