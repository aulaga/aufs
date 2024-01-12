package storage

import (
	"fmt"
	aufs "github.com/aulaga/cloud/src/filesystem"
	"go.beyondstorage.io/services/fs/v4"
	"go.beyondstorage.io/v5/pairs"
	"net/url"
	"path/filepath"
)

func CreateFilesystemFromAuFS(fs aufs.FileSystemSpec) (aufs.Storage, error) {
	rootAuStorage := fs.Root()
	rootStorage, err := CreateStorageFromAuFs(rootAuStorage)
	if err != nil {
		return nil, err
	}

	return NewFilesystem("", rootStorage), nil
}

func CreateStorageFromAuFs(storage aufs.StorageSpec) (aufs.Storage, error) {
	storageUrl, err := url.Parse(storage.Uri)
	if err != nil {
		return nil, err
	}

	scheme := storageUrl.Scheme

	storageBuilder, err := builderFromScheme(scheme)
	if err != nil {
		return nil, err
	}

	return storageBuilder.build(storageUrl)
}

type builder interface {
	build(*url.URL) (aufs.Storage, error)
}

type localBuilder struct {
}

func (l localBuilder) build(storageUrl *url.URL) (aufs.Storage, error) {
	path := filepath.Clean(storageUrl.Host + storageUrl.Path)

	fmt.Println("building local storage", path)
	storager, err := fs.NewStorager(pairs.WithWorkDir(path))
	if err != nil {
		return nil, err
	}

	return NewStorager("tempId", storager), nil
}

func builderFromScheme(scheme string) (builder, error) {
	switch scheme {
	case "local":
		return &localBuilder{}, nil
	}

	return nil, fmt.Errorf("unknown storage scheme")
}
