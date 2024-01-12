package aufs

import (
	"context"
	"github.com/aulaga/webdav"
	"io/fs"
	"path/filepath"
	"time"
)

// TODO remove from here

type SampleFs struct {
}

func (s SampleFs) Root() StorageSpec {
	return StorageSpec{
		Uri: "local:///aulaga/files",
	}
}

func (s SampleFs) Mounts() []MountSpec {
	return nil
}

type StorageProvider interface {
	ProvideFileSystem(FileSystemSpec) (Storage, error)
	ProvideStorage(StorageSpec) (Storage, error)
}

// TODO remove above

type StorageSpec struct {
	Uri string
}

type MountSpec struct {
	StorageSpec
	MountPoint string
}

type FileSystemSpec interface {
	Root() StorageSpec
	Mounts() []MountSpec
}

type Node interface {
	Path() string
	Storage() Storage
}

type File interface {
	Node
	webdav.File
	webdav.CustomPropsHolder
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

type NodeInfo struct {
	path     string
	size     int64
	modTime  time.Time
	isDir    bool
	mimeType string
	etag     string
}

var _ fs.FileInfo = &NodeInfo{}
var _ webdav.ContentTyper = &NodeInfo{}

func NewNodeInfo(path string, size int64, modTime time.Time, isDir bool, mimeType string, etag string) *NodeInfo {
	return &NodeInfo{
		path:     path,
		size:     size,
		modTime:  modTime,
		isDir:    isDir,
		mimeType: mimeType,
		etag:     etag,
	}
}

func (n *NodeInfo) Name() string {
	return filepath.Base(n.path)
}

func (n *NodeInfo) Path() string {
	return n.path
}

func (n *NodeInfo) Size() int64 {
	return n.size
}

func (n *NodeInfo) Mode() fs.FileMode {
	if n.IsDir() {
		return fs.ModeDir
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
