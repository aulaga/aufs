package aufs

import (
	"context"
	"golang.org/x/net/webdav"
	"io/fs"
	"path/filepath"
	"time"
)

type StorageSpec struct {
	Id  string
	Uri string
}

type MountSpec struct {
	Storage    StorageSpec
	MountPoint string
}

type FileSystemSpec interface {
	Root() StorageSpec
	Mounts() []MountSpec
	Listener() EventListener
}

type EventListener interface {
	Moved(src string, dst string)
	Changed(path string) // modified or created
	Deleted(path string)
}

type Node interface {
	Path() string
	Storage() Storage
}

type File interface {
	Node
	webdav.File
}

type Filesystem interface {
	Storage
	AddEventListener(listener EventListener)
	FlushEvents()
}

type Storage interface {
	Id() string
	Open(path string) (File, error)
	Stat(path string) (NodeInfo, error)
	Delete(path string) error
	Copy(srcPath string, dstPath string) error
	Move(srcPath string, dstPath string) error
	ListDir(path string, recursive bool) ([]NodeInfo, error)
	MkDir(path string) (NodeInfo, error)
}

type Mount interface {
	Storage() Storage
	Point() string
}

type NodeInfo interface {
	fs.FileInfo
	webdav.ContentTyper
	MimeType() string
	ETag() string
	Path() string
}

type nodeInfo struct {
	path     string
	size     int64
	modTime  time.Time
	isDir    bool
	mimeType string
	etag     string
}

var _ fs.FileInfo = &nodeInfo{}
var _ webdav.ContentTyper = &nodeInfo{}

func NewNodeInfo(path string, size int64, modTime time.Time, isDir bool, mimeType string, etag string) NodeInfo {
	return &nodeInfo{
		path:     path,
		size:     size,
		modTime:  modTime,
		isDir:    isDir,
		mimeType: mimeType,
		etag:     etag,
	}
}

func (n *nodeInfo) Name() string {
	return filepath.Base(n.path)
}

func (n *nodeInfo) Path() string {
	return n.path
}

func (n *nodeInfo) Size() int64 {
	return n.size
}

func (n *nodeInfo) Mode() fs.FileMode {
	if n.IsDir() {
		return fs.ModeDir
	}

	return 0
}

func (n *nodeInfo) ModTime() time.Time {
	return n.modTime
}

func (n *nodeInfo) IsDir() bool {
	return n.isDir
}

func (n *nodeInfo) MimeType() string {
	return n.mimeType
}

func (n *nodeInfo) ContentType(ctx context.Context) (string, error) {
	return n.mimeType, nil
}

func (n *nodeInfo) ETag() string {
	return n.etag
}

func (n *nodeInfo) Sys() any {
	return n
}
