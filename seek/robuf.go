package seek

import (
	"errors"
	"io"
)

type (
	ReadOnlyBuffer struct {
		idx int
		buf []byte
	}
)

func NewReadOnlyBuffer(b []byte) *ReadOnlyBuffer {
	return &ReadOnlyBuffer{
		idx: 0,
		buf: b,
	}
}

func (r *ReadOnlyBuffer) Read(t []byte) (int, error) {
	if r.idx >= len(r.buf) {
		return 0, io.EOF
	}
	tail := r.buf[r.idx:]
	n := copy(t, tail)
	r.idx += n
	return n, nil
}

func (r *ReadOnlyBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		r.idx += int(offset)
	case io.SeekEnd:
		if offset != 0 {
			return 0, errors.New("invalid whence (buffer cannot expand)")
		}
		r.idx = len(r.buf)
		return int64(r.idx), nil
	case io.SeekStart:
		r.idx = int(offset)
	default:
		return 0, errors.New("invalid whence")
	}
	if r.idx >= len(r.buf) {
		r.idx = len(r.buf)
		return int64(r.idx), errors.New("invalid offset (buffer to short)")
	}
	return int64(r.idx), nil
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
