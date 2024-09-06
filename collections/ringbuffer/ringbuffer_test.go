package ringbuffer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRange(t *testing.T) {
	rb, err := New[string](5)
	require.NoError(t, err)

	for i := range 7 {
		rb.Push(fmt.Sprintf("%d", i))
	}

	expected := []string{"2", "3", "4", "5", "6"}
	for i, v := range rb.Seq(SEQ_MODE_FIFO) {
		require.Equal(t, expected[i], v)
	}

	expected = []string{"6", "5", "4", "3", "2"}
	for i, v := range rb.Seq(SEQ_MODE_FILO) {
		require.Equal(t, expected[i], v)
	}
}

func TestCycledRingBuffer(t *testing.T) {
	rb, err := New[string](5)
	require.NoError(t, err)

	for i := range 7 {
		rb.Push(fmt.Sprintf("%d", i))
	}

	// ask for 3 last items
	lastN := make([]string, 3)
	n := rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"4", "5", "6"}, lastN)

	// ask for 5 last
	lastN = make([]string, 5)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"2", "3", "4", "5", "6"}, lastN)

	// ask for 2 last
	lastN = make([]string, 2)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"5", "6"}, lastN)

	// ask for 1 last
	lastN = make([]string, 1)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"6"}, lastN)
}

func TestNotFilledRingBuffer(t *testing.T) {
	rb, err := New[string](5)
	require.NoError(t, err)

	for i := range 3 {
		rb.Push(fmt.Sprintf("%d", i))
	}

	// ask for 3 last items
	lastN := make([]string, 3)
	n := rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"0", "1", "2"}, lastN)

	// ask for 5 last
	lastN = make([]string, 5)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"0", "1", "2"}, lastN)

	// ask for 2 last
	lastN = make([]string, 2)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"1", "2"}, lastN)

	// ask for 1 last
	lastN = make([]string, 1)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{"2"}, lastN)

	expected := []string{"0", "1", "2"}
	for i, v := range rb.Seq(SEQ_MODE_FIFO) {
		require.Equal(t, expected[i], v)
	}
}

func TestEmptyRingBuffer(t *testing.T) {
	rb, err := New[string](5)
	require.NoError(t, err)

	// ask for 3 last items
	lastN := make([]string, 3)
	n := rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{}, lastN)

	// ask for 5 last
	lastN = make([]string, 5)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{}, lastN)

	// ask for 2 last
	lastN = make([]string, 2)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{}, lastN)

	// ask for 1 last
	lastN = make([]string, 1)
	n = rb.Last(lastN)
	lastN = lastN[:n]
	require.Equal(t, []string{}, lastN)
}

func TestRingbufferWithoutRoom(t *testing.T) {
	_, err := New[string](0)
	require.Error(t, err)
}
