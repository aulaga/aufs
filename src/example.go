package aufs

import (
	"fmt"
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

func (s SampleFs) Root() StorageSpec {
	return StorageSpec{
		Uri: "@fs:///aulaga/files",
	}
}

func (s SampleFs) Listener() EventListener {
	return &bus{}
}

func (s SampleFs) Mounts() []MountSpec {
	return []MountSpec{
		{
			MountPoint: "/tmp/",
			Storage: StorageSpec{
				Uri: "@fs:///aulaga/tmp",
			},
		},
		{
			MountPoint: "/test/",
			Storage: StorageSpec{
				Uri: "@fs:///aulaga/test",
			},
		},
		{
			MountPoint: "/test2/",
			Storage: StorageSpec{
				Uri: "@memory:///",
			},
		},
	}
}
