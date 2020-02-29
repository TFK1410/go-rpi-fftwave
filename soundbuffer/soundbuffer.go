package soundbuffer

import "fmt"

// SoundBuffer implements an int16 circular buffer. It is a fixed size,
// and new writes overwrite older data, such that for a buffer
// of size N, for any amount of writes, only the last N bytes
// are retained.
type SoundBuffer struct {
	data        []int16
	size        int64
	writeCursor int64
	written     int64
}

// NewBuffer creates a new buffer of a given size. The size
// must be greater than 0.
func NewBuffer(size int64) (*SoundBuffer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("Size must be positive")
	}

	b := &SoundBuffer{
		size: size,
		data: make([]int16, size),
	}
	return b, nil
}

// Write writes up to len(buf) values to the internal ring,
// overriding older data if necessary.
// This method will move the slice around so that newer data is always at the end
func (b *SoundBuffer) Write(buf []int16) (int, error) {
	// Account for total ints written
	n := len(buf)
	b.written += int64(n)

	// If the buffer is larger than ours, then we only care
	// about the last size values anyways
	if int64(n) > b.size {
		buf = buf[int64(n)-b.size:]
	}

	// Copy in place by moving the slice around
	copy(b.data, b.data[n:])
	copy(b.data[b.size-int64(n):], buf)

	// Update location of the cursor
	return n, nil
}

// Size returns the size of the buffer
func (b *SoundBuffer) Size() int64 {
	return b.size
}

// TotalWritten provides the total number of values written
func (b *SoundBuffer) TotalWritten() int64 {
	return b.written
}

// Sound returns the whole slice in the buffer. This
// slice should not be written to.
func (b *SoundBuffer) Sound() []int16 {
	return b.data
}
