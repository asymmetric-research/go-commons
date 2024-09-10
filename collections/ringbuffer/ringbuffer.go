package ringbuffer

import (
	"fmt"

	armath "github.com/asymmetric-research/go-commons/math"
)

type T[C any] struct {
	buflen uint
	buf    []C

	// head points to the next free slot
	head uint
}

func New[C any](size int) (*T[C], error) {
	ret := &T[C]{}
	return ret, NewInto(ret, make([]C, size))

}

// NewInto assumes that
func NewInto[C any](dst *T[C], buf []C) error {
	if len(buf) <= 0 {
		return fmt.Errorf("backing buffer must have a greater than zero")
	}
	*dst = T[C]{
		buflen: uint(len(buf)),
		buf:    buf,
		head:   0,
	}
	return nil
}

func (r *T[C]) Push(item C) {
	r.buf[r.head%r.buflen] = item
	r.head += 1
}

func (r *T[C]) Last(dst []C) int {
	// how many entries can we write?
	maxWritable := armath.Min(r.head, r.buflen)

	// if the dst is larger than the amount of entries we can write, let's clamp it.
	if len(dst) > int(maxWritable) {
		// only consider the first available slots of dst
		dst = dst[:maxWritable]
	}

	headmod := int(r.head % r.buflen)

	// we must do at most 2 copies
	n := 0
	// copy the head of our internal buffer to the tail of dst
	{
		// end of src is the head slot
		srcend := headmod

		srcstart := armath.Max(0, headmod-len(dst))
		src := r.buf[srcstart:srcend]

		dststart := armath.Max(0, len(dst)-headmod)
		dst := dst[dststart:]

		n += copy(dst, src)
	}

	// if we haven't filled the buffer, copy the tail of our internal buffer to the head of dst
	if n != len(dst) {
		// copy start of src to end of dst
		dst := dst[:len(dst)-n]

		srcstart := int(maxWritable) - len(dst)
		src := r.buf[srcstart:]

		n += copy(dst, src)
	}

	return n
}

func (r *T[C]) Len() uint {
	used := armath.Min(r.buflen, r.head)
	return used
}

type SeqMode int

const (
	SEQ_MODE_FIFO SeqMode = iota
	SEQ_MODE_FILO
)

func (r *T[C]) Seq(seqMode SeqMode) func(yield func(uint, C) bool) {
	return func(yield func(uint, C) bool) {
		if r.buflen == 0 {
			return
		}

		// how many entries can we write?
		maxWritable := armath.Min(r.head, r.buflen)

		if seqMode == SEQ_MODE_FIFO {
			start := (((r.head - 1) % r.buflen) - maxWritable) % r.buflen

			for i := range maxWritable {
				idx := (start + i) % r.buflen
				if !yield(i, r.buf[idx]) {
					return
				}
			}
			return
		}
		if seqMode == SEQ_MODE_FILO {
			start := r.head - 1
			for i := range maxWritable {
				idx := (start - i) % r.buflen
				if !yield(i, r.buf[idx]) {
					return
				}
			}
			return
		}
	}
}