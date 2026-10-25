package utils

import "io"

// RWPipe 基于 io.Pipe 的读写管道，支持边写边读。
type RWPipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func NewRWPipe() *RWPipe {
	pr, pw := io.Pipe()
	return &RWPipe{r: pr, w: pw}
}

func (vi *RWPipe) Read(b []byte) (int, error) {
	return vi.r.Read(b)
}

func (vi *RWPipe) Write(b []byte) (int, error) {
	return vi.w.Write(b)
}

// Close 先关写端通知读端 EOF，再关读端。
func (vi *RWPipe) Close() error {
	werr := vi.w.Close()
	rerr := vi.r.Close()
	if werr != nil {
		return werr
	}
	return rerr
}

// CloseWrite 仅关闭写端，触发读端 EOF。
func (vi *RWPipe) CloseWrite() error {
	return vi.w.Close()
}

func (vi *RWPipe) CloseRead() error {
	return vi.r.Close()
}
