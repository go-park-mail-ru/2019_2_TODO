package user

import "server/model"

// Usecase - some funcs work with our Usecase
type Usecase interface {
	ListAll() ([]*model.User, error)
	SelectByID(int64) (*model.User, error)
	SelectDataByLogin(string) (*model.User, error)
	Create(*model.User) (int64, error)
	Update(*model.User) (int64, error)
	Delete(int64) (int64, error)
}
