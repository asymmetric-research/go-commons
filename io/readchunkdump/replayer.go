package readchunkdump

import (
	"fmt"
	"io"
	"os"
	"path"
)

type Replayer struct {
	chunksdir string
	nbChunks  int
	idx       int
}

func NewReplayer(chunksdir string) (*Replayer, error) {
	chunks, err := os.ReadDir(chunksdir)
	if err != nil {
		return nil, err
	}

	return &Replayer{
		chunksdir: chunksdir,
		nbChunks:  len(chunks),
	}, nil
}

func (r *Replayer) Read(dst []byte) (n int, err error) {
	if r.idx >= r.nbChunks {
		err = io.EOF
		return
	}

	var data []byte
	data, err = os.ReadFile(
		path.Join(r.chunksdir, fmt.Sprintf("chunk-%d", r.idx)),
	)
	n = len(data)

	if err != nil {
		return
	}

	copy(dst, data)

	r.idx++
	return
}
