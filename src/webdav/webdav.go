package webdav

import (
	"context"
	"fmt"
	"github.com/aulaga/cloud/src/domain/storage"
	webdav "github.com/aulaga/webdav"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func DebugRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("")
	fmt.Println("BasicHandler", r.Method, r.URL.Path)
	fmt.Println("form", r.Form.Encode())
	fmt.Println("postform", r.PostForm.Encode())
	fmt.Println("query", r.URL.RawQuery)
	fmt.Println("Header values")

	// Unknown request print body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}
	if len(bodyBytes) < 1000 {
		fmt.Println("Body", string(bodyBytes))
	} else {
		fmt.Println("Body too long to print")
	}

	bodyReader := strings.NewReader(string(bodyBytes))
	r.Body = io.NopCloser(bodyReader)
}

type XFileSystem struct {
	storage storage.Storage
}

var _ webdav.FileSystem = &XFileSystem{}

func (f XFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	fmt.Println("Open", name)
	file, err := f.storage.Open(name)
	if err != nil {
		return nil, err
	}

	return file, err
}

func (f XFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fmt.Println("Stat", name)
	fileInfo, err := f.storage.Stat(name)
	if err != nil {
		return nil, os.ErrNotExist
	}

	return fileInfo, nil
}

func (f XFileSystem) RemoveAll(ctx context.Context, name string) error {
	fmt.Println("RemoveAll")
	return storage.ManualDelete(f.storage, name)
}

func (f XFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	fmt.Println("Mkdir")
	_, err := f.storage.MkDir(name)
	return err
}

func (f XFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	fmt.Println("Move")
	err := f.storage.Move(oldName, newName)
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
	h http.Handler
}

func (m MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	DebugRequest(w, r)

	captureW := &CaptureResponseWriter{w: w}
	m.h.ServeHTTP(captureW, r)
	captureW.Print()
}

func NewXwebdavHandler(storage storage.Storage) http.Handler {
	fs := &XFileSystem{
		storage: storage,
	}

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

	return &MyHandler{webdavHandler}
}
