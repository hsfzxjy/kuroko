package session

import "io"

type UnsignedNumber interface {
	uint8 | uint16 | uint32 | uint64
}

func (s *Session) Flush() error {
	return s.Rw.Flush()
}

func (s *Session) Close() error {
	s.Rw.Flush()
	return s.id.Close()
}

func ReadUInt[T UnsignedNumber](s *Session, nbytes int) (ret T, err error) {
	var buf [8]byte
	_, err = io.ReadFull(s.Rw, buf[:nbytes])

	if err != nil {
		return
	}
	ret = 0
	for _, x := range buf[:nbytes] {
		ret = (ret << 8) | T(x)
	}
	return
}

func (s *Session) ReadUint8() (uint8, error) {
	return ReadUInt[uint8](s, 1)
}

func (s *Session) ReadUint16() (uint16, error) {
	return ReadUInt[uint16](s, 2)
}

func (s *Session) ReadUint32() (uint32, error) {
	return ReadUInt[uint32](s, 4)
}

func (s *Session) ReadUint64() (uint64, error) {
	return ReadUInt[uint64](s, 8)
}

func (s *Session) ReadBytesInto(dst []byte) (err error) {
	_, err = io.ReadFull(s.Rw, dst)
	return
}

func (s *Session) ReadBytes(nbytes int) (ret []byte, err error) {
	ret = make([]byte, nbytes)
	err = s.ReadBytesInto(ret)
	return
}

func (s *Session) ReadStr() (ret string, err error) {
	var length uint32
	length, err = s.ReadUint32()
	if err != nil {
		return
	}
	var buf []byte
	buf, err = s.ReadBytes(int(length))
	if err != nil {
		return
	}
	ret = string(buf[:])
	return
}

func WriteUInt[T UnsignedNumber](s *Session, val T, nbytes int) error {
	var buf [8]byte
	for i := 0; i < nbytes; i++ {
		buf[nbytes-1-i] = byte(val & 0xFF)
		val >>= 8
	}
	_, err := s.Rw.Write(buf[:nbytes])
	return err
}

func (s *Session) WriteUint8(val uint8) error {
	return WriteUInt(s, val, 1)
}

func (s *Session) WriteUint16(val uint16) error {
	return WriteUInt(s, val, 2)
}

func (s *Session) WriteUint32(val uint32) error {
	return WriteUInt(s, val, 4)
}

func (s *Session) WriteUint64(val uint64) error {
	return WriteUInt(s, val, 8)
}

func (s *Session) WriteBytes(src []byte) error {
	_, err := s.Rw.Write(src)
	return err
}

func (s *Session) WriteFromReader(reader io.Reader) error {
	_, err := s.Rw.ReadFrom(reader)
	return err
}

func (s *Session) WriteStr(str string) error {
	if err := s.WriteUint32(uint32(len(str))); err != nil {
		return err
	}
	if err := s.WriteBytes([]byte(str)); err != nil {
		return err
	}
	return nil
}
