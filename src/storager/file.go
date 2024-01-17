package storager

import (
	"bytes"
	aufs "github.com/aulaga/cloud/src"
	"github.com/aulaga/cloud/src/storager/writers"
	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/types"
	"io"
	"io/fs"
)

type File interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	aufs.File
}

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

type FileReadWriter struct {
	writer writers.Writer
	reader Reader

	storager *StoragerWrapper
	path     string
}

func (f FileReadWriter) Readdir(count int) ([]fs.FileInfo, error) {
	infos, err := f.storager.ListDir(f.Path(), false)
	if err != nil {
		return nil, err
	}

	if count <= 0 || count > len(infos) {
		count = len(infos)
	}

	fsInfos := make([]fs.FileInfo, count)
	for i := 0; i < count; i++ {
		fsInfos[i] = infos[i]
	}

	return fsInfos, nil
}

func (f FileReadWriter) Stat() (fs.FileInfo, error) {
	// If writer offers stat we use that (helpful when file is recently written and may not be available from storager yet)
	info, err := f.writer.Stat()
	if err == nil {
		return info, nil
	}

	return f.storager.Stat(f.path)
}

func (f FileReadWriter) Write(bytes []byte) (int, error) {
	return f.writer.Write(bytes)
}

func (f FileReadWriter) Read(bytes []byte) (int, error) {
	return f.reader.Read(bytes)
}

func (f FileReadWriter) Seek(offset int64, whence int) (int64, error) {
	return f.reader.Seek(offset, whence)
}

func (f FileReadWriter) Path() string {
	return f.path
}

func (f FileReadWriter) Storage() aufs.Storage {
	return f.storager
}

func (f FileReadWriter) Close() error {
	err := f.writer.Close()
	if err != nil {
		return err
	}

	return f.reader.Close()
}
