package circbuf

import (
	"errors"
	"fmt"
)

func properMod(x, y int64) int64 {
	return ((x % y) + y) % y
}

var ErrBufferFull = errors.New("Unable to write more data, the buffer is full")
var ErrNoNewData = errors.New("No more data available to read")

// Buffer implements a circular buffer. It is a fixed size, but
// new writes will not overwrite unread data
type Buffer struct {
	data []byte
	size int64

	writeCursor int64
	readCursor  int64
}

// NewBuffer creates a new buffer of a given size. The size
// must be greater tha n 0
func NewBuffer(size int64) (*Buffer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("circbuf.Buffers must have a size > 0")
	}

	b := &Buffer{
		size: size + 1, // +1 to allow for non-overlapping reading & writing
		data: make([]byte, size),
	}
	return b, nil
}

func (b *Buffer) Read(p []byte) (int, error) {
	bytes_read := 0

	switch {
	case b.readCursor < b.writeCursor:
		bytes_read += copy(p, b.data[b.readCursor:b.writeCursor])
	case b.readCursor > b.writeCursor: // We wrapped around the end of the buffer, we need to read around
		bytes_read += copy(p, b.data[b.readCursor:])               // Read to the end
		bytes_read += copy(p[bytes_read:], b.data[:b.writeCursor]) // Copy from the beginning to the last read byte
	default:
		return 0, ErrNoNewData
	}

	b.readCursor += int64(bytes_read)

	return bytes_read, nil
}

// Write writes up to len(buf) bytes to the internal ring,
// overriding older data if necessary.
func (b *Buffer) Write(buf []byte) (int, error) {
	var (
		bytesWritten int
		err          error
	)
	switch {
	case b.Free() >= int64(len(buf)):
		bytesWritten, err = b.writeAround(buf)
	case b.Free() < int64(len(buf)):
		bytesWritten, err = b.writeAround(buf[:b.Free()])
	}

	if err != nil {
		return bytesWritten, err
	}

	if bytesWritten != len(buf) {
		return int(bytesWritten), ErrBufferFull
	}
	return int(bytesWritten), nil
}

// DO NOT pass a buffer with more data to this than you want to
// 		  write, it will write it and destroy data you didn't mean to
func (b *Buffer) writeAround(buf []byte) (int, error) {
	bytes_written := 0

	switch {
	case b.writeCursor < b.readCursor:
		bytes_written += copy(b.data[b.writeCursor:b.readCursor], buf)
	// case b.writeCursor > b.readCursor:
	default:
		bytes_written += copy(b.data[b.writeCursor:], buf)
		bytes_written += copy(b.data, buf[bytes_written:])
	}

	b.writeCursor = (b.writeCursor + int64(bytes_written)) % b.size

	if bytes_written != len(buf) {
		return bytes_written, fmt.Errorf("Failed to write all the data out")
	}
	return bytes_written, nil
}

// Capacity returns the capacity of the buffer
func (b *Buffer) Capacity() int64 {
	return b.size - 1
}

func (b *Buffer) Free() int64 {
	switch {
	case b.readCursor > b.writeCursor:
		return b.writeCursor - b.readCursor
	case b.readCursor < b.writeCursor:
		return (b.Capacity() - b.writeCursor) + (b.readCursor)
	default:
		return b.Capacity()
	}
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
		return make([]byte, 0)
	}
}

// String returns the contents of the buffer as a string
func (b *Buffer) String() string {
	return string(b.Bytes())
}
