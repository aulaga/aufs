package webdav

import (
	"context"
	"fmt"
	aufs "github.com/aulaga/cloud/src"
	"golang.org/x/net/webdav"
	"log"
	"net/http"
	"os"
)

const CtxAulagaFs = "aulaga_fs"

type FileSystem struct {
	storageProvider aufs.StorageProvider
}

var _ webdav.FileSystem = &FileSystem{}

func newFileSystem(storageProvider aufs.StorageProvider) *FileSystem {
	return &FileSystem{storageProvider: storageProvider}
}

func (f FileSystem) fsFromContext(ctx context.Context) (aufs.Filesystem, error) {
	ctxVal := ctx.Value(CtxAulagaFs)
	fsSpec, ok := ctxVal.(aufs.FileSystemSpec)
	if !ok {
		return nil, fmt.Errorf("aulaga filesystem not in context") // TODO type-wrap this error
	}

	return f.storageProvider.ProvideFileSystem(fsSpec)
}

func (f FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	fs, err := f.fsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	file, err := fs.Open(name)
	if err != nil {
		return nil, err
	}

	return file, err
}

func (f FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fs, err := f.fsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fileInfo, err := fs.Stat(name)
	if err != nil {
		return nil, os.ErrNotExist
	}

	return fileInfo, nil
}

func (f FileSystem) RemoveAll(ctx context.Context, name string) error {
	fs, err := f.fsFromContext(ctx)
	if err != nil {
		return err
	}

	return fs.Delete(name)
}

func (f FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	fs, err := f.fsFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = fs.MkDir(name)
	return err
}

func (f FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	fs, err := f.fsFromContext(ctx)
	if err != nil {
		return err
	}

	err = fs.Move(oldName, newName)
	return err
}

type CaptureResponseWriter struct {
	w     http.ResponseWriter
	bytes []byte
}

func (c *CaptureResponseWriter) Print() {
	if len(c.bytes) > 5000 {
		fmt.Println("Captured response too long to print")
	} else {
		fmt.Println("Captured response", string(c.bytes))
	}
}

func (c *CaptureResponseWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CaptureResponseWriter) Write(bytes []byte) (int, error) {
	c.bytes = append(c.bytes, bytes...)
	return c.w.Write(bytes)
}

func (c *CaptureResponseWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

type MyHandler struct {
	h  http.Handler
	fs *FileSystem
}

func (m MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fs := &aufs.SampleFs{}

	ctx := r.Context()
	ctx = context.WithValue(ctx, CtxAulagaFs, fs)
	r = r.WithContext(ctx)

	captureW := &CaptureResponseWriter{w: w}
	m.h.ServeHTTP(captureW, r)
	captureW.Print()

	aulagaFs, err := m.fs.fsFromContext(ctx)
	if err == nil {
		aulagaFs.FlushEvents()
	}
}

func NewWebdavHandler(provider aufs.StorageProvider) http.Handler {
	fs := newFileSystem(provider)

	webdavHandler := &webdav.Handler{
		Prefix:     "/dav",
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}

	return &MyHandler{webdavHandler, fs}
}
