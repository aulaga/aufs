package storage

import (
	"fmt"
	aufs "github.com/aulaga/cloud/src/filesystem"
)

type Mount interface {
	Storage() aufs.Storage
	Point() string
}

type mount struct {
	storage aufs.Storage
	point   string
}

func (m mount) Storage() aufs.Storage {
	return m.storage
}

func (m mount) Point() string {
	return m.point
}

type Filesystem struct {
	id     string
	root   aufs.Storage
	mounts []Mount
}

var _ aufs.Storage = &Filesystem{}

func NewFilesystem(id string, root aufs.Storage) *Filesystem {
	return &Filesystem{
		id:   id,
		root: root,
	}
}

func (f *Filesystem) Id() string {
	return f.id
}

func (f *Filesystem) StorageForPath(path string) aufs.Storage {
	return f.root // TODO implement mounted storages
}

func (f *Filesystem) Open(path string) (aufs.File, error) {
	storage := f.StorageForPath(path)
	file, err := storage.Open(path)

	if err != nil {
		fmt.Println(err.Error())
	}

	return file, err
}

func (f *Filesystem) MkDir(path string) (*aufs.NodeInfo, error) {
	storage := f.StorageForPath(path)
	return storage.MkDir(path)
}

func (f *Filesystem) Stat(path string) (*aufs.NodeInfo, error) {
	storage := f.StorageForPath(path)
	return storage.Stat(path)
}

func (f *Filesystem) Delete(path string) error {
	storage := f.StorageForPath(path)
	return storage.Delete(path)
}

func (f *Filesystem) Copy(srcPath string, dstPath string) error {
	srcStorage := f.StorageForPath(srcPath)
	dstStorage := f.StorageForPath(dstPath)

	if srcStorage == dstStorage {
		return srcStorage.Copy(srcPath, dstPath)
	}

	return ManualCopy(srcStorage, dstStorage, srcPath, dstPath)
}

func (f *Filesystem) Move(srcPath string, dstPath string) error {
	srcStorage := f.StorageForPath(srcPath)
	dstStorage := f.StorageForPath(dstPath)

	if srcStorage == dstStorage {
		return srcStorage.Move(srcPath, dstPath)
	}

	err := ManualCopy(srcStorage, dstStorage, srcPath, dstPath)
	if err != nil {
		err2 := dstStorage.Delete(dstPath)
		if err2 != nil {
			return fmt.Errorf(
				"move operation across storages failed dramatically, first failed copying the files (%s), then failed to delete the incomplete move in destination storage (%s)",
				err.Error(),
				err2.Error(),
			)
		}

		return err
	}

	return srcStorage.Delete(srcPath)
}

func (f *Filesystem) ListDir(path string, recursive bool) ([]*aufs.NodeInfo, error) {
	storage := f.StorageForPath(path)
	return storage.ListDir(path, recursive)
}
