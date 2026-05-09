package arena

import "io"

type Cursor struct {
	cursor
}

func (c *Cursor) Read(p []byte) (int, error) {
	n := c.read(p)
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (a *Linked) Write(p []byte) uint64    { return a.write(p) }
func (a *Linked) Free(id uint64)           { a.free(id) }
func (a *Linked) Open(id uint64) io.Reader { return &Cursor{cursor: a.open(id)} }

type Reader struct {
	reader
}

func (r *Reader) Read(p []byte) (int, error) {
	n := r.read(p)
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (a *Sized) Bytes(id uint64) []byte   { return a.bytes(id) }
func (a *Sized) Write(p []byte) uint64    { return a.write(p) }
func (a *Sized) Free(id uint64)           { a.free(id) }
func (a *Sized) Open(id uint64) io.Reader { return &Reader{reader: a.open(id)} }
