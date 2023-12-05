package user

import (
	"github.com/aulaga/cloud/src/domain/storage"
	"github.com/google/uuid"
)

type User interface {
	Id() uuid.UUID
	Filesystem() *storage.Filesystem
}

type NilUser struct {
}

func (u NilUser) Id() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000000")
}

func (u NilUser) Filesystem() *storage.Filesystem {
	return nil
}

type user struct {
	id uuid.UUID
	fs *storage.Filesystem
}

func (u user) Id() uuid.UUID {
	return u.id
}

func (u user) Filesystem() *storage.Filesystem {
	return u.fs
}
