package ringbuffer

import (
	"fmt"
	"sync/atomic"

	armath "github.com/asymmetric-research/go-commons/math"
)

type T[C any] struct {
	buflen uint64
	buf    []C

	// head points to the next free slot
	head atomic.Uint64
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
		buflen: uint64(len(buf)),
		buf:    buf,
	}
	dst.head.Store(0)
	return nil
}

func (r *T[C]) Push(item C) {
	nextSlot := r.head.Add(1)
	r.buf[(nextSlot-1)%r.buflen] = item
}

func (r *T[C]) Last(dst []C) int {
	// how many entries can we write?
	maxWritable := armath.Min(r.head.Load(), r.buflen)

	// if the dst is larger than the amount of entries we can write, let's clamp it.
	if len(dst) > int(maxWritable) {
		// only consider the first available slots of dst
		dst = dst[:maxWritable]
	}

	headmod := int(r.head.Load() % r.buflen)

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

func (r *T[C]) Len() uint64 {
	used := armath.Min(r.buflen, r.head.Load())
	return used
}

type SeqMode int

const (
	SEQ_MODE_FIFO SeqMode = iota
	SEQ_MODE_FILO
)

func (r *T[C]) Seq(seqMode SeqMode) func(yield func(uint64, C) bool) {
	return func(yield func(uint64, C) bool) {
		if r.buflen == 0 {
			return
		}

		head := r.head.Load()

		// how many entries can we write?
		maxWritable := armath.Min(head, r.buflen)

		if seqMode == SEQ_MODE_FIFO {
			start := (((head - 1) % r.buflen) - maxWritable) % r.buflen

			for i := range maxWritable {
				idx := (start + i) % r.buflen
				if !yield(i, r.buf[idx]) {
					return
				}
			}
			return
		}
		if seqMode == SEQ_MODE_FILO {
			start := head - 1
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
