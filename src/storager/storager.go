package storager

import (
	"errors"
	"fmt"
	aufs "github.com/aulaga/aufs/src"
	"github.com/aulaga/aufs/src/internal"
	"github.com/aulaga/aufs/src/storager/writers"
	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/types"
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

func (s *StoragerWrapper) Storager() types.Storager {
	return s.storager
}

func (s *StoragerWrapper) Id() string {
	return s.id
}

func sanitizePath(path string) string {
	return strings.TrimLeft(path, "/\\")
}

func (s *StoragerWrapper) Open(path string) (aufs.File, error) {
	path = sanitizePath(path)

	reader := newDefaultReader(s.storager, path)

	var writer writers.Writer
	appender, isAppender := s.storager.(types.Appender)
	if isAppender {
		writer = writers.NewAppender(appender, path)
	}

	if writer == nil {
		writer = writers.NewTempBuffer(s.storager, path)
	}

	return &FileReadWriter{
		writer:   writer,
		reader:   reader,
		storager: s,
		path:     path,
	}, nil
}

func (s *StoragerWrapper) Stat(path string) (aufs.NodeInfo, error) {
	path = sanitizePath(path)
	obj, err := s.storager.Stat(path)
	if err != nil {
		return nil, err
	}

	return getObjectInfo(obj), err
}

func (s *StoragerWrapper) Delete(path string) error {
	path = sanitizePath(path)

	return s.storager.Delete(path)
}

func (s *StoragerWrapper) Copy(srcPath string, dstPath string) error {
	srcPath = sanitizePath(srcPath)
	dstPath = sanitizePath(dstPath)
	info, err := s.Stat(srcPath)
	if err != nil {
		return err
	}

	// If source node is a folder we need to manually copy, storager does not allow copying file structures.
	if info.IsDir() {
		return internal.ManualCopy(s, s, srcPath, dstPath)
	}

	copier, ok := s.storager.(types.Copier)
	if !ok {
		return fmt.Errorf("CopyOperation failed, storage not a copier")
	}

	return copier.Copy(srcPath, dstPath)
}

// TODO implement Move between different storages
func (s *StoragerWrapper) Move(srcPath string, dstPath string) error {
	srcPath = sanitizePath(srcPath)
	dstPath = sanitizePath(dstPath)
	mover, ok := s.storager.(types.Mover)
	if !ok {
		return fmt.Errorf("MoveOperation failed, storage not a copier")
	}

	return mover.Move(srcPath, dstPath)
}

func (s *StoragerWrapper) ListDir(path string, recursive bool) ([]aufs.NodeInfo, error) {
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

func (s *StoragerWrapper) MkDir(path string) (aufs.NodeInfo, error) {
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
