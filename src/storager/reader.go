package storager

import (
	"bytes"
	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/types"
	"io"
	"io/fs"
)

type Reader interface {
	io.ReadSeekCloser
}

type defaultReader struct {
	storager types.Storager
	path     string
	offset   int64
	size     *int64
}

func newDefaultReader(storager types.Storager, path string) Reader {
	return &defaultReader{
		storager: storager,
		path:     path,
	}
}

func (f *defaultReader) Path() string {
	return f.path
}

func (f *defaultReader) Close() error {
	return nil
}

func (f *defaultReader) Read(p []byte) (int, error) {
	var buf bytes.Buffer

	size, err := f.Size()
	if err != nil {
		return 0, err
	}

	readSize := int64(len(p))
	if readSize+f.offset > size {
		readSize = size - f.offset
	}

	n, err := f.storager.Read(f.path, &buf, pairs.WithOffset(f.offset), pairs.WithSize(readSize))
	if err != nil {
		return 0, err
	}
	f.offset += n

	return buf.Read(p)
}

func (f *defaultReader) Seek(offset int64, whence int) (int64, error) {
	size, err := f.Size()
	if err != nil {
		return 0, err
	}

	if whence == io.SeekStart {
		f.offset = offset
	}

	if whence == io.SeekEnd {
		f.offset = size - offset
	}

	if whence == io.SeekCurrent {
		f.offset += offset
	}

	return f.offset, nil
}

func (f *defaultReader) Size() (int64, error) {
	if f.size != nil {
		return *f.size, nil
	}

	obj, err := f.storager.Stat(f.path)
	if err != nil {
		return -1, err
	}

	size, ok := obj.GetContentLength()
	if !ok {
		return 0, nil // FIXME should return error?
	}
	f.size = &size
	return size, nil
}

func (f *defaultReader) Stat() (fs.FileInfo, error) {
	obj, err := f.storager.Stat(f.path)
	if err != nil {
		return nil, err
	}

	return getObjectInfo(obj), nil
}
