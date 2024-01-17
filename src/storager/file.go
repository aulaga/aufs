package storager

import (
	aufs "github.com/aulaga/cloud/src"
	"github.com/aulaga/cloud/src/storager/writers"
	"io/fs"
)

type File interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	aufs.File
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
