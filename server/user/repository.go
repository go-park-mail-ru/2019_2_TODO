package user

import (
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/model"
)

// Repository - some funcs work with out repository
type Repository interface {
	ListAll() ([]*model.User, error)
	SelectByID(int64) (*model.User, error)
	SelectDataByLogin(string) (*model.User, error)
	Create(*model.User) (int64, error)
	CreateLeader(*leaderBoardModel.UserLeaderBoard) (int64, error)
	SelectLeaderByID(id int64) (*leaderBoardModel.UserLeaderBoard, error)
	UpdateLeader(elem *leaderBoardModel.UserLeaderBoard) (int64, error)
	Update(*model.User) (int64, error)
	Delete(int64) (int64, error)
}
