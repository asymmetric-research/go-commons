package readchunkdump

import (
	"fmt"
	"io"
	"os"
	"path"
)

type T struct {
	r         io.Reader
	chunksdir string
	cnt       int
}

func New(r io.Reader, chunksdir string) *T {
	t := &T{
		r:         r,
		chunksdir: chunksdir,
	}

	return t
}

func (t *T) Read(in []byte) (n int, err error) {
	n, err = t.r.Read(in)

	werr := os.WriteFile(
		path.Join(t.chunksdir, fmt.Sprintf("chunk-%d", t.cnt)),
		in[:n],
		0o666,
	)

	if werr != nil {
		panic(werr)
	}

	t.cnt++

	return
}
