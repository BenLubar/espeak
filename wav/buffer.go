package wav

import (
	"io"
	"os"
)

type buffer struct {
	buf []byte
	pos int
}

func (b *buffer) Write(p []byte) (n int, err error) {
	l := len(b.buf)
	b.buf = append(b.buf[:b.pos], p...)
	if l > b.pos+len(p) {
		b.buf = b.buf[:l]
	}
	b.pos += len(p)
	return len(p), nil
}

func (b *buffer) Seek(offset int64, whence int) (cur int64, err error) {
	switch whence {
	case os.SEEK_CUR:
		b.pos += int(offset)
	case os.SEEK_SET:
		b.pos = int(offset)
	case os.SEEK_END:
		b.pos = len(b.buf) + int(offset)
	}

	return int64(b.pos), nil
}

func (b *buffer) WriteTo(w io.Writer) (n int64, err error) {
	i, err := w.Write(b.buf)
	return int64(i), err
}
