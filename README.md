circbuf
=======

This repository provides the `circbuf` package. This provides a `Buffer` object
which is a circular (or ring) buffer. It has a fixed size, and returns errors on writes once that size is exhausted without having been read. The buffer implements the `io.Writer` and `io.Reader` interfaces.

It is not safe for use in a shared concurrent situation

Documentation
=============

Full documentation can be found on [Godoc](http://godoc.org/github.com/thomaso-mirodin/circbuf)

Usage
=====

The `circbuf` package is very easy to use:

```go
buf, _ := NewBuffer(6)
buf.Write([]byte("hello world"))

if string(buf.Bytes()) != "hello " {
    panic("should only have the first 6 bytes!")
}

```

