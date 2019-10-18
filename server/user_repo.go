package main

import (
	"database/sql"
)

type User struct {
	ID       int64
	Username string
	Password string
	Avatar   string
}

type UsersRepository struct {
	DB *sql.DB
}

func (repo *UsersRepository) ListAll() ([]*User, error) {
	users := []*User{}
	rows, err := repo.DB.Query("SELECT id, login, avatar FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		record := &User{}
		err = rows.Scan(&record.ID, &record.Username, &record.Avatar)
		if err != nil {
			return nil, err
		}
		users = append(users, record)
	}
	return users, nil
}

func (repo *UsersRepository) SelectByID(id int64) (*User, error) {
	record := &User{}
	err := repo.DB.
		QueryRow("SELECT id, login, password, avatar FROM users WHERE id = ?", id).
		Scan(&record.ID, &record.Username, &record.Password, &record.Avatar)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Потенциально очень не безопасный запрос
func (repo *UsersRepository) SelectDataByLogin(username string) (*User, error) {
	record := &User{}
	err := repo.DB.
		QueryRow("SELECT id, login, password, avatar FROM users WHERE login = ?", username).
		Scan(&record.ID, &record.Username, &record.Password, &record.Avatar)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (repo *UsersRepository) SelectByLoginAndPassword(elem *User) (*User, error) {
	record := &User{}
	err := repo.DB.
		QueryRow("SELECT id, login, avatar FROM users WHERE login = ? AND password = ?",
			elem.Username, elem.Password).
		Scan(&record.ID, &record.Username, &record.Avatar)
	if err != nil {
		return record, err
	}
	return record, nil
}

func (repo *UsersRepository) Create(elem *User) (int64, error) {
	defaultAvatar := "images/avatar.png"
	result, err := repo.DB.Exec(
		"INSERT INTO users (`login`, `password`, `avatar`) VALUES (?, ?, ?)",
		elem.Username,
		elem.Password,
		defaultAvatar,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *UsersRepository) Update(elem *User) (int64, error) {
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
