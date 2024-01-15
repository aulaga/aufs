package storage

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	aufs "github.com/aulaga/cloud/src/filesystem"
	webdav "github.com/aulaga/webdav"
	"github.com/beyondstorage/go-storage/v4/pairs"
	"github.com/beyondstorage/go-storage/v4/types"
	"io"
	"io/fs"
	"os"
	"strings"
)

type Object = types.Object

func getObjectInfo(node *Object) aufs.NodeInfo {
	contentLength, _ := node.GetContentLength()
	lastModified, _ := node.GetLastModified()
	isDir := node.GetMode().IsDir()
	mimeType, _ := node.GetContentType()
	etag, _ := node.GetEtag()

	return aufs.NewNodeInfo(
		node.GetPath(),
		contentLength,
		lastModified,
		isDir,
		mimeType,
		etag,
	)
}

type file struct {
	id      string
	path    string
	storage StoragerWrapper
	offset  int64
}

func (f *file) Props() []xml.Name {
	return []xml.Name{}
}

func (f *file) PropFn(name xml.Name) (func(context.Context, webdav.FileSystem, webdav.LockSystem, string, os.FileInfo) (string, error), bool) {
	o, err := f.storager().Stat(f.Path())
	if err != nil {
		return nil, false
	}
	info := getObjectInfo(o)

	if name.Local == "mimetype" {
		return func(context.Context, webdav.FileSystem, webdav.LockSystem, string, os.FileInfo) (string, error) {
			return info.MimeType(), nil
		}, true
	}

	return nil, false
}

var _ aufs.File = &file{}

func newFile(storage StoragerWrapper, path string) aufs.File {
	return &file{
		id:      path,
		path:    path,
		storage: storage,
		offset:  0,
	}
}

func (f *file) storager() types.Storager {
	return f.storage.storager
}

func (f *file) Path() string {
	return f.id
}

func (f *file) Storage() aufs.Storage {
	return f.storage
}

func (f *file) Write(p []byte) (int, error) {
	buf := bytes.NewBuffer(p)
	n := int64(len(p))
	n64, err := f.storager().Write(f.path, buf, n, pairs.WithOffset(f.offset))
	f.offset += n64

	return int(n64), err
}

func (f *file) Read(p []byte) (int, error) {
	var buf bytes.Buffer
	n, err := f.storager().Read(f.path, &buf, pairs.WithOffset(f.offset), pairs.WithSize(int64(len(p))))
	if err != nil {
		return 0, err
	}
	f.offset += n

	return buf.Read(p)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	stat, err := f.Stat()
	if err != nil {
		return 0, err
	}

	if whence == io.SeekStart {
		f.offset = offset
	}

	if whence == io.SeekEnd {
		f.offset = stat.Size() - offset
	}

	if whence == io.SeekCurrent {
		f.offset += offset
	}

	return f.offset, nil
}

func (f *file) Close() error {
	return nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.storage.Stat(f.path)
}

func (f *file) Readdir(count int) ([]fs.FileInfo, error) {
	infos, err := f.storage.ListDir(f.Path(), false)
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

type StoragerWrapper struct {
	id       string
	storager types.Storager
}

func NewStorager(id string, storager types.Storager) aufs.Storage {
	return &StoragerWrapper{
		id:       id,
		storager: storager,
	}
}

func (s StoragerWrapper) Id() string {
	return s.storager.String()
}

func sanitizePath(path string) string {
	return strings.TrimLeft(path, "/\\")
}

func (s StoragerWrapper) Open(path string) (aufs.File, error) {
	path = sanitizePath(path)
	f := newFile(s, path)
	return f, nil
}

func (s StoragerWrapper) Stat(path string) (aufs.NodeInfo, error) {
	path = sanitizePath(path)
	obj, err := s.storager.Stat(path)
	if err != nil {
		return nil, err
	}

	return getObjectInfo(obj), err
}

func (s StoragerWrapper) Delete(path string) error {
	path = sanitizePath(path)
	return s.storager.Delete(path)
}

func (s StoragerWrapper) Copy(srcPath string, dstPath string) error {
	srcPath = sanitizePath(srcPath)
	dstPath = sanitizePath(dstPath)
	info, err := s.Stat(srcPath)
	if err != nil {
		return err
	}

	// If source node is a folder we need to manually copy, storager does not allow copying file structures.
	if info.IsDir() {
		return ManualCopy(s, s, srcPath, dstPath)
	}

	copier, ok := s.storager.(types.Copier)
	if !ok {
		return fmt.Errorf("CopyOperation failed, storage not a copier")
	}

	return copier.Copy(srcPath, dstPath)
}

// TODO implement Move between different storages
func (s StoragerWrapper) Move(srcPath string, dstPath string) error {
	srcPath = sanitizePath(srcPath)
	dstPath = sanitizePath(dstPath)
	mover, ok := s.storager.(types.Mover)
	if !ok {
		return fmt.Errorf("MoveOperation failed, storage not a copier")
	}

	return mover.Move(srcPath, dstPath)
}

func (s StoragerWrapper) ListDir(path string, recursive bool) ([]aufs.NodeInfo, error) {
	path = sanitizePath(path)
	// TODO for recursion maybe listmode prefix can be used here
	iterator, err := s.storager.List(path, pairs.WithListMode(types.ListModeDir))
	if err != nil {
		return nil, err
	}

	var infos []aufs.NodeInfo
	for obj, err := iterator.Next(); obj != nil; obj, err = iterator.Next() {
		if err != nil && errors.Is(err, types.IterateDone) {
			break
		}
		if err != nil {
			return nil, err
		}

		infos = append(infos, getObjectInfo(obj))

	}

	return infos, nil
}

func (s StoragerWrapper) MkDir(path string) (aufs.NodeInfo, error) {
	path = sanitizePath(path)
	direr, ok := s.storager.(types.Direr)
	if !ok {
		return nil, fmt.Errorf("mkdir operation failed, storage is not direr")
	}

	o, err := direr.CreateDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory '%s', %s", path, err.Error())
	}

	return getObjectInfo(o), nil
}
