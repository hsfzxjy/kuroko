package session

import (
	"bytes"

	"kuroko-linux/internal"
	"kuroko-linux/internal/exc"

	"github.com/jmoiron/sqlx"
)

func (s *Session) RequireIdentity(tx *sqlx.Tx) {
	peerId, _ := ReadUInt[uint64](s, 6)
	s.Bridge.UpdateSelf(tx, &peerId)
}

func (s *Session) WriteIdentity() {
	s.WriteBytes(internal.MACHINE_ID[:])
}

func (s *Session) RequireMagic() {
	magic, _ := s.ReadBytes(len(internal.MAGIC))
	if !bytes.Equal(magic, []byte(internal.MAGIC)) {
		panic(exc.Fields("magic", magic).New("bad magic"))
	}
}

func (s *Session) WriteMagic() {
	s.WriteBytes([]byte(internal.MAGIC))
}

func (s *Session) ReadOpcode() uint32 {
	opcode, _ := s.ReadUint32()
	return opcode
}

func (s *Session) WriteOpcode(opcode uint32) {
	s.WriteUint32(opcode)
}
