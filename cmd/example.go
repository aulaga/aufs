package main

import (
	"fmt"
	aufs "github.com/aulaga/aufs/src"
)

// TODO remove file

type bus struct {
}

func (b bus) Moved(src string, dst string) {
	fmt.Println("[EventListener] Moved", src, dst)
}

func (b bus) Changed(path string) {
	fmt.Println("[EventListener] Changed", path)
}

func (b bus) Deleted(path string) {
	fmt.Println("[EventListener] Deleted", path)
}

type SampleFs struct {
}

func (s SampleFs) Root() aufs.StorageSpec {
	return aufs.StorageSpec{
		Uri: "@fs:///aulaga/files",
	}
}

func (s SampleFs) Listener() aufs.EventListener {
	return &bus{}
}

func (s SampleFs) Mounts() []aufs.MountSpec {
	return []aufs.MountSpec{
		{
			MountPoint: "/tmp/",
			Storage: aufs.StorageSpec{
				Uri: "@fs:///aulaga/tmp",
			},
		},
		{
			MountPoint: "/test/",
			Storage: aufs.StorageSpec{
				Uri: "@fs:///aulaga/test",
			},
		},
		{
			MountPoint: "/test2/",
			Storage: aufs.StorageSpec{
				Uri: "@memory:///",
			},
		},
	}
}
