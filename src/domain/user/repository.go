package user

import (
	"fmt"
	"github.com/aulaga/cloud/src/domain/storage"
	"github.com/google/uuid"
)

type Repository interface {
	One(id uuid.UUID) (User, error)
	OneByName(name string) (User, error)
	Store(user User) error
}

type mockRepository struct {
	users map[string]User
}

func NewMockRepository() Repository {
	st := storage.NewFs("C:\\Users\\raul\\GolandProjects\\cloud\\.attic\\fs")

	fs := storage.NewFilesystem("mockFS", st)
	users := map[string]User{
		"00000000-0000-0000-0000-000000000001": user{uuid.MustParse("00000000-0000-0000-0000-000000000001"), fs},
	}

	return mockRepository{users: users}
}

func (r mockRepository) One(id uuid.UUID) (User, error) {
	return r.OneByName(id.String())
}

func (r mockRepository) OneByName(name string) (User, error) {
	user, ok := r.users[name]
	if !ok {
		return NilUser{}, fmt.Errorf("user not found")
	}

	return user, nil
}

func (r mockRepository) Store(user User) error {
	r.users[user.Id().String()] = user
	return nil
}
