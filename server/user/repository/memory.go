package repository

import (
	"database/sql"
	"encoding/base64"
	"log"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/model"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"
)

// NewUserMemoryRepository - create connection and return new repository
func NewUserMemoryRepository() user.Repository {
	dsn := utils.DataBaseConfig
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		log.Println("Error while Ping")
	}

	return &UsersRepository{
		DB: db,
	}
}

// UsersRepository - uses pointer of any DataBase
type UsersRepository struct {
	DB *sql.DB
}

// ListAll - show all public information
func (repo *UsersRepository) ListAll() ([]*model.User, error) {
	users := []*model.User{}
	rows, err := repo.DB.Query("SELECT id, login, avatar FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		record := &model.User{}
		err = rows.Scan(&record.ID, &record.Username, &record.Avatar)
		if err != nil {
			return nil, err
		}
		users = append(users, record)
	}
	return users, nil
}

// CreateLeader - create new user in leaderboard
func (repo *UsersRepository) CreateLeader(elem *leaderBoardModel.UserLeaderBoard) (int64, error) {
	result, err := repo.DB.Exec(
		"INSERT INTO leaderboard (`id`, `username`, `points`) VALUES (?, ?)",
		elem.ID,
		elem.Username,
		elem.Points,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// SelectLeaderByID - select all user`s data by ID
func (repo *UsersRepository) SelectLeaderByID(id int64) (*leaderBoardModel.UserLeaderBoard, error) {
	record := &leaderBoardModel.UserLeaderBoard{}
	err := repo.DB.
		QueryRow("SELECT id, username, points FROM leaderboard WHERE id = ?", id).
		Scan(&record.ID, &record.Username, &record.Points)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// UpdateLeader - update user`s data in DataBase
func (repo *UsersRepository) UpdateLeader(elem *leaderBoardModel.UserLeaderBoard) (int64, error) {
	result, err := repo.DB.Exec(
		"UPDATE leaderboard SET"+
			"`username` = ?"+
			",`points` = ?"+
			"WHERE id = ?",
		elem.Username,
		elem.Points,
		elem.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// SelectByID - select all user`s data by ID
func (repo *UsersRepository) SelectByID(id int64) (*model.User, error) {
	record := &model.User{}
	err := repo.DB.
		QueryRow("SELECT id, login, password, avatar FROM users WHERE id = ?", id).
		Scan(&record.ID, &record.Username, &record.Password, &record.Avatar)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// SelectDataByLogin - select all user`s data by Login
func (repo *UsersRepository) SelectDataByLogin(username string) (*model.User, error) {
	record := &model.User{}
	err := repo.DB.
		QueryRow("SELECT id, login, password, avatar FROM users WHERE login = ?", username).
		Scan(&record.ID, &record.Username, &record.Password, &record.Avatar)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Create - create new user in dataBase with default avatar
func (repo *UsersRepository) Create(elem *model.User) (int64, error) {
	elem.Password = base64.StdEncoding.EncodeToString(
		utils.ConvertPass(elem.Password))
	result, err := repo.DB.Exec(
		"INSERT INTO users (`login`, `password`, `avatar`) VALUES (?, ?, ?)",
		elem.Username,
		elem.Password,
		elem.Avatar,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update - update user`s data in DataBase
func (repo *UsersRepository) Update(elem *model.User) (int64, error) {
	elem.Password = base64.StdEncoding.EncodeToString(
		utils.ConvertPass(elem.Password))
	result, err := repo.DB.Exec(
		"UPDATE users SET"+
			"`login` = ?"+
			",`password` = ?"+
			",`avatar` = ?"+
			"WHERE id = ?",
		elem.Username,
		elem.Password,
		elem.Avatar,
		elem.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete - delete user`s record in DataBase
func (repo *UsersRepository) Delete(id int64) (int64, error) {
	result, err := repo.DB.Exec(
		"DELETE FROM users WHERE id = ?",
		id,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
