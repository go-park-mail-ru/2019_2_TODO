package repository

import (
	"database/sql"
	"log"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"
)

// NewUserMemoryRepository - create connection and return new repository
func NewUserMemoryRepository() *LeadersRepository {
	dsn := utils.DataBaseConfig
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		log.Println("Error while Ping")
	}

	return &LeadersRepository{
		DB: db,
	}
}

// UsersRepository - uses pointer of any DataBase
type LeadersRepository struct {
	DB *sql.DB
}

// ListAllLeaders - show all public information about leaders
func (repo *LeadersRepository) ListAllLeaders() ([]*leaderBoardModel.UserLeaderBoard, error) {
	leaders := []*leaderBoardModel.UserLeaderBoard{}
	rows, err := repo.DB.Query("SELECT id, username, points FROM leaderboard")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		record := &leaderBoardModel.UserLeaderBoard{}
		err = rows.Scan(&record.ID, &record.Username, &record.Points)
		if err != nil {
			return nil, err
		}
		leaders = append(leaders, record)
	}
	return leaders, nil
}

// SelectByID - select all user`s data by ID
func (repo *LeadersRepository) SelectLeaderByID(id int64) (*leaderBoardModel.UserLeaderBoard, error) {
	record := &leaderBoardModel.UserLeaderBoard{}
	err := repo.DB.
		QueryRow("SELECT id, username, points FROM leaderboard WHERE id = ?", id).
		Scan(&record.ID, &record.Username, &record.Points)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Create - create new user in dataBase with default avatar
func (repo *LeadersRepository) CreateLeader(elem *leaderBoardModel.UserLeaderBoard) (int64, error) {
	result, err := repo.DB.Exec(
		"INSERT INTO leaderboard (`username`, `points`) VALUES (?, ?)",
		elem.Username,
		elem.Points,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update - update user`s data in DataBase
func (repo *LeadersRepository) UpdateLeader(elem *leaderBoardModel.UserLeaderBoard) (int64, error) {
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

// Delete - delete user`s record in DataBase
func (repo *LeadersRepository) DeleteLeader(id int64) (int64, error) {
	result, err := repo.DB.Exec(
		"DELETE FROM users WHERE id = ?",
		id,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
