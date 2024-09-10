package linereader

import (
	"bytes"
	"errors"
	"io"

	armath "github.com/asymmetric-research/go-commons/math"
)

type T struct {
	reader      io.Reader
	readbufbase []byte
	readbuf     []byte
	blocksize   uint
}

func New(reader io.Reader, blockSize uint) *T {
	lr := &T{}
	NewInto(lr, reader, blockSize)
	return lr
}

func NewInto(dst *T, reader io.Reader, blockSize uint) {
	*dst = T{
		reader:      reader,
		readbufbase: make([]byte, blockSize),
		blocksize:   blockSize,
	}
}

func (lr *T) Read(dst []byte) (n int, err error) {
	n, discarded, err := lr.ReadExtra(dst)
	if discarded != 0 {
		return n + discarded, errors.New("line too long")
	}
	return n, err
}

// ReadExtra reads as much as possible into p, until the next newline or EOF is reached.
// Every new call to read starts on a new line. The remainder of the previous line will be discarted.
func (lr *T) ReadExtra(dst []byte) (nread int, ndiscarted int, err error) {
	// copy as much of read buffer as possible to dst
	if len(lr.readbuf) > 0 {
		// fast path: can we get a new line from the read buffer?
		maxread := armath.Min(len(dst), len(lr.readbuf))
		eolidx := bytes.IndexByte(lr.readbuf[:maxread], '\n')
		if eolidx >= 0 && eolidx < len(dst) {
			// yes - copy to dst and return
			copy(dst[:eolidx], lr.readbuf)
			lr.readbuf = lr.readbuf[eolidx+1:]
			return eolidx, 0, nil
		}

		// no - copy as much of the read buffer as possible to dst, and then continue reading from reader
		n := copy(dst, lr.readbuf)
		nread += n
		lr.readbuf = lr.readbuf[n:]
		dst = dst[n:]

	}

	for i := uint(0); ; i++ {
		readOffset := lr.blocksize * i
		readLimit := armath.Min(readOffset+lr.blocksize, uint(len(dst)))

		// dst has been filled and there hasn't been a new line yet
		if readLimit <= readOffset {
			ndiscarted = lr.discardRestOfLine()
			return
		}

		dstClamp := dst[readOffset:readLimit]
		var n int
		n, err = lr.reader.Read(dstClamp)
		dstClamp = dstClamp[:n]
		nread += n

		if err == io.EOF && n == 0 {
			return
		} else if err != nil {
			return
		}

		// is there a end of line in this block?
		eolidx := bytes.IndexByte(dstClamp, '\n')

		if eolidx < 0 {
			continue
		}

		// discard the new line character
		nread -= 1

		// is new line at the end of read?
		if eolidx == int(readLimit)-1 {
			// yes
			return

		}

		// copy the data after the end of line into the read buffer
		cpyn := copy(lr.readbufbase, dstClamp[eolidx+1:])
		lr.readbuf = lr.readbufbase[:cpyn]
		nread -= n - eolidx - 1
		return
	}
}

func (lr *T) discardRestOfLine() int {
	// discard the rest of the line in the read buffer

	if len(lr.readbuf) > 0 {
		if idx := bytes.IndexByte(lr.readbuf, '\n'); idx >= 0 {
			lr.readbuf = lr.readbuf[idx+1:]
			return idx
		} else {
			lr.readbuf = nil
		}
	}

	// discard the rest of the line in the reader

	prevread := 0
	for {
		n, err := lr.reader.Read(lr.readbufbase)
		lr.readbuf = lr.readbufbase[:n]
		if err != nil {
			return n
		}

		eolidx := bytes.IndexByte(lr.readbuf, '\n')

		if eolidx >= 0 {
			lr.readbuf = lr.readbuf[eolidx+1:]
			return eolidx + prevread
		}
		prevread += n
	}
}
