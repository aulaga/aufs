package storage

import (
	"context"
	"fmt"
	webdav "github.com/aulaga/webdav"
	"go.beyondstorage.io/services/fs/v4"
	"go.beyondstorage.io/v5/pairs"
	"io"
	iofs "io/fs"
	"path/filepath"
	"strings"
	"time"
)

type NodeInfo struct {
	path     string
	size     int64
	modTime  time.Time
	isDir    bool
	mimeType string
	etag     string
}

var _ iofs.FileInfo = &NodeInfo{}
var _ webdav.ContentTyper = &NodeInfo{}

func (n *NodeInfo) Name() string {
	return filepath.Base(n.path)
}

func (n *NodeInfo) Path() string {
	return n.path
}

func (n *NodeInfo) Size() int64 {
	return n.size
}

func (n *NodeInfo) Mode() iofs.FileMode {
	if n.IsDir() {
		return iofs.ModeDir
	}

	return 0
}

func (n *NodeInfo) ModTime() time.Time {
	return n.modTime
}

func (n *NodeInfo) IsDir() bool {
	return n.isDir
}

func (n *NodeInfo) MimeType() string {
	return n.mimeType
}

func (n *NodeInfo) ContentType(ctx context.Context) (string, error) {
	return n.mimeType, nil
}

func (n *NodeInfo) ETag() string {
	return n.etag
}

func (n *NodeInfo) Sys() any {
	return n
}

type Storage interface {
	Id() string
	Open(path string) (File, error)
	Stat(path string) (*NodeInfo, error)
	Delete(path string) error
	Copy(srcPath string, dstPath string) error
	Move(srcPath string, dstPath string) error
	ListDir(path string, recursive bool) ([]*NodeInfo, error)
	MkDir(path string) (*NodeInfo, error)
}

// TODO remove this function, provide proper storage initialisation
func NewFs(path string) Storage {
	storager, err := fs.NewStorager(pairs.WithWorkDir(path))
	if err != nil {
		panic(err.Error())
	}

	return NewStorager("tempId", storager)
}

func CreateFile(storage Storage, path string, reader io.Reader) error {
	file, err := storage.Open(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, reader)
	return err
}

func copyDir(srcStorage Storage, dstStorage Storage, srcPath string, dstPath string) error {
	infos, err := srcStorage.ListDir(srcPath, false)
	if err != nil {
		return err
	}

	dstPath = strings.TrimRight(dstPath, "/\\")
	for _, srcFileInfo := range infos {
		relPathToFile, err := filepath.Rel(srcPath, srcFileInfo.path)
		if err != nil {
			return err
		}

		dstFilePath := fmt.Sprintf("%s/%s", dstPath, relPathToFile)
		err = ManualCopy(srcStorage, dstStorage, srcFileInfo.path, dstFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// ManualCopy manually copies a path from srcStorage to dstStorage. This function is storage-agnostic.
func ManualCopy(srcStorage Storage, dstStorage Storage, srcPath string, dstPath string) error {
	fmt.Println("Manual copy", srcPath, dstPath)
	info, err := srcStorage.Stat(srcPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		_, err := dstStorage.MkDir(dstPath)
		if err != nil {
			return err
		}
		return copyDir(srcStorage, dstStorage, srcPath, dstPath)
	}

	srcFile, err := srcStorage.Open(srcPath)
	if err != nil {
		return err
	}

	return CreateFile(dstStorage, dstPath, srcFile)
}

func ManualDelete(storage Storage, path string) error {
	info, err := storage.Stat(path)
	if err != nil {
		return err // TODO file doesnt exist?
	}

	deleteFn := func(storage Storage, info *NodeInfo) error {
		return storage.Delete(info.Path())
	}

	return walkFs(storage, info, deleteFn)
}

func walkFs(storage Storage, info *NodeInfo, operationFunc func(Storage, *NodeInfo) error) error {
	if info.IsDir() {
		infos, err := storage.ListDir(info.Path(), false)
		if err != nil {
			return err
		}
		for _, info := range infos {
			err := walkFs(storage, info, operationFunc)
			if err != nil {
				return err
			}
		}
	}

	return operationFunc(storage, info)
}
