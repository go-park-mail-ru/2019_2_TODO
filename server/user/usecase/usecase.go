package usecase

import (
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/model"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user"
)

// NewUserUsecase - create new userCase
func NewUserUsecase(userRepo user.Repository) user.Usecase {
	return userUsecase{repo: userRepo}
}

type userUsecase struct {
	repo user.Repository
}

func (u userUsecase) ListAll() ([]*model.User, error) {
	return u.repo.ListAll()
}

func (u userUsecase) SelectByID(id int64) (*model.User, error) {
	return u.repo.SelectByID(id)
}

func (u userUsecase) SelectDataByLogin(login string) (*model.User, error) {
	return u.repo.SelectDataByLogin(login)
}

func (u userUsecase) Create(user *model.User) (int64, error) {
	return u.repo.Create(user)
}

func (u userUsecase) Update(user *model.User) (int64, error) {
	return u.repo.Update(user)
}

func (u userUsecase) Delete(id int64) (int64, error) {
	return u.repo.Delete(id)
}
