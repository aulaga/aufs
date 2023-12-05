package storage

import (
	webdav "github.com/aulaga/webdav"
)

type Node interface {
	Path() string
	Storage() Storage
}

type File interface {
	Node
	webdav.File
	webdav.CustomPropsHolder
}
