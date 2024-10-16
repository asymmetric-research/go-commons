package linereader

import "fmt"

type ErrLineTruncated struct {
	Discarded int
}

func (e *ErrLineTruncated) Error() string {
	return fmt.Sprintf("line truncated (discarded %d bytes)", e.Discarded)
}
