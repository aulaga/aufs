package writers

import (
	"io"
	"io/fs"
)

type Writer interface {
	io.WriteCloser
	Stat() (fs.FileInfo, error)
}
