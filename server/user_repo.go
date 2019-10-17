package main

import (
	"database/sql"
)

type User struct {
	ID       int64
	Login    string
	Password string
	Avatar   string
}

type UsersRepository struct {
	DB *sql.DB
}

func (repo *UsersRepository) ListAll() ([]*User, error) {
	items := []*User{}
	rows, err := repo.DB.Query("SELECT id, title, updated FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		post := &User{}
		err = rows.Scan(&post.ID, &post.Login, &post.Avatar)
		if err != nil {
			return nil, err
		}
		items = append(items, post)
	}
	return items, nil
}

func (repo *UsersRepository) SelectByID(id int64) (*User, error) {
	post := &User{}
	err := repo.DB.
		QueryRow("SELECT id, title, updated, description FROM items WHERE id = ?", id).
		Scan(&post.ID, &post.Login, &post.Password, &post.Avatar)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (repo *UsersRepository) Create(elem *User) (int64, error) {
	defaultAvatar := "/images/avatar.png"
	result, err := repo.DB.Exec(
		"INSERT INTO items (`login`, `password`, `avatar`) VALUES (?, ?, ?)",
		elem.Login,
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
		"UPDATE items SET"+
			"`login` = ?"+
			",`password` = ?"+
			",`avatar` = ?"+
			"WHERE id = ?",
		elem.Login,
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
		"DELETE FROM items WHERE id = ?",
		id,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
