package internal

import (
	"fmt"
	aufs "github.com/aulaga/aufs/src"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type Filesystem struct {
	id              string
	root            aufs.Storage
	mounts          []aufs.Mount
	eventPropagator *EventPropagator
}

var _ aufs.Storage = &Filesystem{}

func NewFilesystem(id string, root aufs.Storage, mounts []aufs.Mount) *Filesystem {
	return &Filesystem{
		id:              id,
		root:            root,
		mounts:          mounts,
		eventPropagator: &EventPropagator{events: []Event{}},
	}
}

func (f *Filesystem) Id() string {
	return f.id
}

func (f *Filesystem) AddEventListener(listener aufs.EventListener) {
	f.eventPropagator.AddEventListener(listener)
}

func (f *Filesystem) FlushEvents() {
	f.eventPropagator.Publish()
}

func (f *Filesystem) StorageForPath(filePath string) (aufs.Storage, string) {
	filePath = filepath.Clean(filePath)

	for _, mount := range f.mounts {
		// FIXME clean-up mount-point evaluation
		relPath, err := filepath.Rel(mount.Point(), filePath)
		isMountPath := err == nil && !strings.HasPrefix(relPath, "../") && relPath != ".."
		if isMountPath {
			return mount.Storage(), relPath
		}
	}

	return f.root, filePath
}

func (f *Filesystem) Open(path string) (file aufs.File, err error) {
	isRootPath := strings.TrimLeft(path, "/") == ""
	if isRootPath {
		return &fsFile{fs: f}, nil
	}

	storage, relPath := f.StorageForPath(path)
	file, err = storage.Open(relPath)
	if err != nil {
		return nil, err
	}

	file = &EventFile{file: file, propagator: f.eventPropagator, path: path}

	return file, nil
}

func (f *Filesystem) MkDir(path string) (info aufs.NodeInfo, err error) {
	defer func() {
		if err == nil {
			f.eventPropagator.AddEvent(ChangedEvent(path))
		}
	}()

	storage, path := f.StorageForPath(path)
	return storage.MkDir(path)
}

func (f *Filesystem) Stat(path string) (info aufs.NodeInfo, err error) {
	storage, path := f.StorageForPath(path)
	return storage.Stat(path)
}

func (f *Filesystem) Delete(path string) (err error) {
	defer func() {
		if err == nil {
			f.eventPropagator.AddEvent(DeletedEvent(path))
		}
	}()

	storage, path := f.StorageForPath(path)
	return ManualDelete(storage, path)
}

func (f *Filesystem) Copy(srcPath string, dstPath string) (err error) {
	defer func() {
		if err == nil {
			f.eventPropagator.AddEvent(ChangedEvent(dstPath))
		}
	}()
	srcStorage, srcRelPath := f.StorageForPath(srcPath)
	dstStorage, dstRelPath := f.StorageForPath(dstPath)

	if srcStorage == dstStorage {
		return srcStorage.Copy(srcRelPath, dstRelPath)
	}

	return ManualCopy(srcStorage, dstStorage, srcRelPath, dstRelPath)
}

func (f *Filesystem) Move(srcPath string, dstPath string) (err error) {
	defer func() {
		if err == nil {
			f.eventPropagator.AddEvent(MovedEvent(srcPath, dstPath))
		}
	}()

	srcStorage, relSrcPath := f.StorageForPath(srcPath)
	dstStorage, relDstPath := f.StorageForPath(dstPath)

	if srcStorage == dstStorage {
		return srcStorage.Move(relSrcPath, relDstPath)
	}

	err = ManualCopy(srcStorage, dstStorage, relSrcPath, relDstPath)
	if err != nil {
		err2 := dstStorage.Delete(relDstPath)
		if err2 != nil {
			return fmt.Errorf(
				"move operation across storages failed dramatically, first failed copying the files (%s), then failed to delete the incomplete move in destination storage (%s)",
				err.Error(),
				err2.Error(),
			)
		}

		return err
	}

	return ManualDelete(srcStorage, relSrcPath)
}

func (f *Filesystem) ListDir(path string, recursive bool) (infos []aufs.NodeInfo, err error) {
	storage, path := f.StorageForPath(path)
	list, err := storage.ListDir(path, recursive)
	if err != nil {
		return nil, err
	}

	return list, err
}

// Filesystem act as file
type fsFile struct {
	fs *Filesystem
}

func (f fsFile) Path() string {
	return "/"
}

func (f fsFile) Storage() aufs.Storage {
	return f.fs
}

func (f fsFile) Close() error {
	return nil
}

func (f fsFile) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("cannot read filesystem")
}

func (f fsFile) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("cannot seek filesystem")
}

func (f fsFile) Readdir(count int) ([]fs.FileInfo, error) {
	infos, err := f.fs.ListDir("/", false)
	if err != nil {
		return nil, err
	}

	for _, mount := range f.fs.mounts {
		info := aufs.NewNodeInfo(mount.Point(), 0, time.Time{}, true, "", "")
		infos = append(infos, info)
	}

	if count <= 0 || count > len(infos) {
		count = len(infos)
	}

	fsInfos := make([]fs.FileInfo, count)
	for i := 0; i < count; i++ {
		fsInfos[i] = infos[i]
	}

	return fsInfos, err
}

func (f fsFile) Stat() (fs.FileInfo, error) {
	return aufs.NewNodeInfo("/", 0, time.Time{}, true, "", ""), nil
}

func (f fsFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("cannot write to filesystem")
}
