package writers

import (
	"fmt"
	"github.com/google/uuid"
	"go.beyondstorage.io/v5/types"
	"io/fs"
	"os"
)

type tempbuffer struct {
	path       string
	storager   types.Storager
	bufferFile *os.File
	offset     int64
	written    bool
	size       *int64
}

func NewTempBuffer(storager types.Storager, path string) Writer {
	return &tempbuffer{
		path:     path,
		storager: storager,
	}
}

func (f *tempbuffer) Close() error {
	if f.bufferFile == nil {
		return nil
	}

	err := f.bufferFile.Close()
	if err != nil {
		return err
	}

	file, err := os.Open(f.bufferFile.Name())
	if err != nil {
		return err
	}

	defer file.Close()
	defer os.Remove(f.bufferFile.Name())

	_, err = f.storager.Write(f.path, file, f.offset) // FIXME is using offset as length here safe?
	return err
}

func (f *tempbuffer) Write(p []byte) (int, error) {
	if f.bufferFile == nil {
		tempName := fmt.Sprintf("aufs_%s", uuid.New().String())
		file, err := os.CreateTemp("", tempName)
		if err != nil {
			return 0, err
		}
		f.bufferFile = file
	}

	f.offset = f.offset + int64(len(p))
	return f.bufferFile.Write(p)
}

func (f *tempbuffer) Stat() (fs.FileInfo, error) {
	if f.bufferFile != nil {
		return f.bufferFile.Stat()
	}

	return nil, fmt.Errorf("temp-buffer writer did not write to any temp file, no stat available")
}
