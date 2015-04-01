package circbuf

import (
	"fmt"
)

func properMod(x, y int64) int64 {
	return ((x % y) + y) % y
}

// Buffer implements a circular buffer. It is a fixed size, but
// new writes will not overwrite unread data
type Buffer struct {
	data []byte
	size int64

	writeCursor int64
	writeCount  int64
	readCursor  int64
	readCount   int64
}

// NewBuffer creates a new buffer of a given size. The size
// must be greater tha n 0
func NewBuffer(size int64) (*Buffer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("Size must be positive")
	}

	b := &Buffer{
		size: size + 1, // +1 to allow for non-overlapping reading & writing
		data: make([]byte, size),
	}
	return b, nil
}

func (b *Buffer) Read(p []byte) (int, error) {
	bytes_read := 0
	for i, pb := range b.nonCopyRead(int64(len(p))) {
		if pb != nil {
			p[i] = *pb
			bytes_read++
		}
	}

	return bytes_read, nil
}

func (b *Buffer) nonCopyRead(n int64) []*byte {
	buf := make([]*byte, n)
	bytesRead := int64(0)

	for r := b.readCursor; r-b.readCursor < n && b.readCount+(r-b.readCursor) < b.writeCount; r++ {
		buf[r-b.readCursor] = &b.data[r%b.size]
		bytesRead++
	}

	b.readCount += bytesRead
	b.readCursor = (b.readCursor + n) % b.size

	return buf
}

// Write writes up to len(buf) bytes to the internal ring,
// overriding older data if necessary.
func (b *Buffer) Write(buf []byte) (int, error) {

	n := int64(len(buf))

	bytesWritten := int64(0)
	for wc := b.writeCursor; bytesWritten < n && wc != properMod((b.readCursor-1), b.size); wc, bytesWritten = (wc+1)%b.size, bytesWritten+1 {
		b.data[wc%b.size] = buf[bytesWritten]
	}

	// Update location of the cursor
	b.writeCount += bytesWritten
	b.writeCursor = ((b.writeCursor + bytesWritten) % b.size)

	if bytesWritten != n {
		return int(bytesWritten), fmt.Errorf("Unable to write all the bytes")
	}
	return int(bytesWritten), nil
}

// Capacity returns the capacity of the buffer
func (b *Buffer) Capacity() int64 {
	return b.size - 1
}

// Bytes provides a slice of the bytes written. This
// slice should not be written to.
func (b *Buffer) Bytes() []byte {
	switch {
	case b.writeCursor < b.readCursor:
		out := make([]byte, b.size)
		copy(out, b.data[b.writeCursor:])
		copy(out[b.size-b.writeCursor:], b.data[:b.readCursor])
		return out
	case b.writeCursor > b.readCursor:
		out := make([]byte, b.writeCursor-b.readCursor)
		copy(out, b.data[b.readCursor:b.writeCursor])
		return out
	default:
		return make([]byte, b.size)
	}
}

// String returns the contents of the buffer as a string
func (b *Buffer) String() string {
	return string(b.Bytes())
}
