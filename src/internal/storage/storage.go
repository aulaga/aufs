package storage

import (
	"fmt"
	aufs "github.com/aulaga/cloud/src/filesystem"
	"io"
	"path/filepath"
	"strings"
)

func CreateFile(storage aufs.Storage, path string, reader io.Reader) error {
	file, err := storage.Open(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, reader)
	return err
}

func copyDir(srcStorage aufs.Storage, dstStorage aufs.Storage, srcPath string, dstPath string) error {
	infos, err := srcStorage.ListDir(srcPath, false)
	if err != nil {
		return err
	}

	dstPath = strings.TrimRight(dstPath, "/\\")
	for _, srcFileInfo := range infos {
		relPathToFile, err := filepath.Rel(srcPath, srcFileInfo.Path())
		if err != nil {
			return err
		}

		dstFilePath := fmt.Sprintf("%s/%s", dstPath, relPathToFile)
		err = ManualCopy(srcStorage, dstStorage, srcFileInfo.Path(), dstFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// ManualCopy manually copies a path from srcStorage to dstStorage. This function is storage-agnostic.
func ManualCopy(srcStorage aufs.Storage, dstStorage aufs.Storage, srcPath string, dstPath string) error {
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

func ManualDelete(storage aufs.Storage, path string) error {
	if path == "" || path == "." {
		return fmt.Errorf("cannot delete root of path")
	}

	info, err := storage.Stat(path)
	if err != nil {
		return err // TODO file doesnt exist?
	}

	deleteFn := func(storage aufs.Storage, info aufs.NodeInfo) error {
		return storage.Delete(info.Path())
	}

	return walkFs(storage, info, deleteFn)
}

func walkFs(storage aufs.Storage, info aufs.NodeInfo, operationFunc func(aufs.Storage, aufs.NodeInfo) error) error {
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
