package util

import (
	"io"
)

type ReaderWithLength struct {
	r    io.Reader
	rest uint64
}

func NewReaderWithLength(r io.Reader, len uint64) *ReaderWithLength {
	return &ReaderWithLength{
		r:    r,
		rest: len,
	}
}

func (r *ReaderWithLength) Read(p []byte) (n int, err error) {
	if r.rest == 0 {
		err = io.EOF
		return
	}
	var plen = uint64(len(p))
	if plen > r.rest {
		plen = r.rest
	}
	n, err = r.r.Read(p[:plen])
	if err == nil {
		r.rest -= uint64(n)
	}
	return
}
