package storager

import (
	"fmt"
	"github.com/aulaga/cloud/src"
	"github.com/aulaga/cloud/src/internal"
	_ "go.beyondstorage.io/services/fs/v4"
	"go.beyondstorage.io/v5/services"
	"strings"
)

type DefaultStorageProvider struct {
	filesystems map[aufs.FileSystemSpec]aufs.Filesystem
	storages    map[aufs.StorageSpec]aufs.Storage
}

func Provider() aufs.StorageProvider {
	return &DefaultStorageProvider{
		filesystems: map[aufs.FileSystemSpec]aufs.Filesystem{},
		storages:    map[aufs.StorageSpec]aufs.Storage{},
	}
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

func (p *DefaultStorageProvider) ProvideFileSystem(spec aufs.FileSystemSpec) (aufs.Filesystem, error) {
	fs, ok := p.filesystems[spec]
	if ok {
		return fs, nil
	}

	rootAuStorage := spec.Root()
	rootStorage, err := p.ProvideStorage(rootAuStorage)
	if err != nil {
		return nil, err
	}

	mountSpecs := spec.Mounts()
	mounts := make([]aufs.Mount, len(mountSpecs))
	for i, mountSpec := range mountSpecs {
		storage, err := p.ProvideStorage(mountSpec.Storage)
		if err != nil {
			return nil, err
		}

		point := strings.Trim(mountSpec.MountPoint, "/")
		point = fmt.Sprintf("/%s/", point)

		mount := &mount{
			storage: storage,
			point:   point,
		}
		mounts[i] = mount
	}

	id := spec.Root().Id
	fs = internal.NewFilesystem(id, rootStorage, mounts)
	listener := spec.Listener()
	if listener != nil {
		fs.AddEventListener(listener)
	}

	p.filesystems[spec] = fs
	return fs, nil
}

func (p *DefaultStorageProvider) ProvideStorage(spec aufs.StorageSpec) (aufs.Storage, error) {
	st, ok := p.storages[spec]
	if ok {
		return st, nil
	}

	uri := strings.TrimLeft(spec.Uri, "@")
	storager, err := services.NewStoragerFromString(uri)
	if err != nil {
		return nil, err
	}

	id := spec.Id

	p.storages[spec] = NewStorager(id, storager)
	return p.storages[spec], nil
}
