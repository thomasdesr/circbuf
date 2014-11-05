package main

import (
	"bytes"
	"io"
	"testing"
)

func TestBuffer_Impl(t *testing.T) {
	var _ io.Writer = &Buffer{}
	var _ io.Reader = &Buffer{}
}

func TestBuffer_ShortWrite(t *testing.T) {
	buf, err := NewBuffer(1024)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	inp := []byte("hello world")

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	if !bytes.Equal(buf.Bytes(), inp) {
		t.Fatalf("bad: %v", buf.Bytes())
	}
}

func TestBuffer_ShortRead(t *testing.T) {
	buf, err := NewBuffer(1024)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	inp := []byte("hello world")

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	out := make([]byte, len(inp)-2)
	buf.Read(out)

	expected := []byte("hello wor")
	if !bytes.Equal(out, expected) {
		t.Fatalf("bad: %v", buf.Bytes())
	}
}

func TestBuffer_FullWrite(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(int64(len(inp)))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	if !bytes.Equal(buf.Bytes(), inp) {
		t.Fatalf("bad: input=\"%v\" output=\"%v\"", inp, buf.Bytes())
	}
}

func TestBuffer_FullRead(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(int64(len(inp)))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	out := make([]byte, len(inp))
	buf.Read(out)

	if !bytes.Equal(out, inp) {
		t.Fatalf("bad: %v", out)
	}
}

func TestBuffer_LongWrite(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(6)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	n, err := buf.Write(inp)
	if err == nil {
		t.Fatalf("err: %v", buf)
	}
	if int64(n) > buf.Size() {
		t.Fatalf("bad: %v", n)
	}

	expect := []byte("hello ")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %s", buf.Bytes())
	}
}

func TestBuffer_LongRead(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(6)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	n, err := buf.Write(inp)
	if err == nil {
		t.Fatalf("err: %v", buf)
	}

	out := make([]byte, len(inp))
	buf.Read(out)

	expect := []byte("hello ")
	if !bytes.Equal(expect, out[:n]) {
		t.Fatalf("bad: expected=\"%v\" got=\"%v\"", expect, out[:n])
	}
}

// func TestSimpleRead(t *testing.T) {
// 	buf, err := NewBuffer(16)

// 	buf.Write([]byte("hello world"))

// 	out := make([]byte, 6)
// 	_, err := buf.Read(out)

// 	if err != nil {

// 	}
// }

func TestReadBeforeWrite(t *testing.T) {
	buf, err := NewBuffer(8)

	out := make([]byte, 8)
	n, err := buf.Read(out)

	if n != 0 {
		t.Fatalf("err: Read %i bytes without any being written first", n)
	}

	if buf.TotalRead() > 0 {
		t.Fatalf("err: readCount > 0 without any bytes being written first")
	}

	if err != nil {
		t.Fatalf("err: Read should never return an error")
	}

}

func TestReadPastWritePointer(t *testing.T) {
	buf, _ := NewBuffer(16)

	length, _ := buf.Write([]byte("Hello World"))

	out := make([]byte, 16)
	n, err := buf.Read(out)

	if n > length {
		t.Fatal("err: Read past the write cursor")
	} else if n < length {
		t.Fatal("err: Didn't read the full length")
	}

	if err != nil {
		t.Fatal("err: buf.Read should never return an error")
	}
}

func TestBuffer_HugeWrite(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	n, err := buf.Write(inp)
	if err == nil {
		t.Fatalf("err: %v", err)
	}
	if int64(n) > buf.Size() {
		t.Fatalf("bad: %v", n)
	}

	expect := []byte("hel")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %s", buf.Bytes())
	}
}

func TestBuffer_ManySmallWrites(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i, b := range inp {
		n, err := buf.Write([]byte{b})

		if int64(i) < buf.Size() {
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			if n != 1 {
				t.Fatalf("bad: %v", n)
			}
		} else {
			if err == nil {
				t.Fatal("err: Write should've failed")
			}

			if n != 0 {
				t.Fatal("bad: Write should've failed")
			}
		}
	}

	expect := []byte("hel")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %v", buf.Bytes())
	}
}

func TestBuffer_MultiPart(t *testing.T) {
	inputs := [][]byte{
		[]byte("hello world\n"),
		[]byte("this is a test\n"),
		[]byte("my cool input\n"),
	}
	total := 0

	buf, err := NewBuffer(16)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i, b := range inputs {
		n, err := buf.Write(b)
		total += n

		if i == 0 {
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if n != len(b) {
				t.Fatalf("bad: %v", n)
			}
		}
	}

	if int64(total) != buf.TotalWritten() {
		t.Fatalf("bad total")
	}

	expect := []byte("hello world\nthis")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: expected=\"%s\" got=\"%s\"", expect, buf.Bytes())
	}
}
