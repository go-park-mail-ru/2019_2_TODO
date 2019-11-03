package main

import (
	"fmt"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

func TestListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	var elemID int64 = 1

	rows := sqlmock.
		NewRows([]string{"id", "login", "avatar"})
	expectResult := []*User{
		{elemID, "toringol", "", "default"},
		{elemID + 1, "sergey", "", "default"},
	}

	for _, item := range expectResult {
		rows = rows.AddRow(item.ID, item.Username, item.Avatar)
	}

	mock.
		ExpectQuery("SELECT id, login, avatar FROM users").
		WillReturnRows(rows)

	repo := &UsersRepository{
		DB: db,
	}

	items, err := repo.ListAll()

	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(items[0], expectResult[0]) && !reflect.DeepEqual(items[1], expectResult[1]) {
		t.Errorf("results not match, want %v, have %v or want %v, have %v", expectResult[0], items[0],
			expectResult[1], items[1])
		return
	}

	// query error
	mock.
		ExpectQuery("SELECT id, login, avatar FROM users").
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.ListAll()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// row scan error
	rows = sqlmock.NewRows([]string{"id", "login"}).
		AddRow(1, "username")

	mock.
		ExpectQuery("SELECT id, login, avatar FROM users").
		WillReturnRows(rows)

	_, err = repo.ListAll()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestSelectByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	var elemID int64 = 1

	rows := sqlmock.
		NewRows([]string{"id", "login", "password", "avatar"})
	expect := []*User{
		{elemID, "toringol", "12345", "default"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.ID, item.Username, item.Password, item.Avatar)
	}

	mock.
		ExpectQuery("SELECT id, login, password, avatar FROM users WHERE").
		WithArgs(elemID).
		WillReturnRows(rows)

	repo := &UsersRepository{
		DB: db,
	}

	item, err := repo.SelectByID(elemID)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(item, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], item)
		return
	}

	// query error
	mock.
		ExpectQuery("SELECT id, login, password, avatar FROM users WHERE").
		WithArgs(elemID).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.SelectByID(elemID)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// row scan error
	rows = sqlmock.NewRows([]string{"id", "login"}).
		AddRow(1, "username")

	mock.
		ExpectQuery("SELECT id, login, password, avatar FROM users WHERE").
		WithArgs(elemID).
		WillReturnRows(rows)

	_, err = repo.SelectByID(elemID)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

}

func TestSelectDataByLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	var elemID int64 = 1
	login := "toringol"

	rows := sqlmock.
		NewRows([]string{"id", "login", "password", "avatar"})
	expect := []*User{
		{elemID, "toringol", "12345", "default"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.ID, item.Username, item.Password, item.Avatar)
	}

	mock.
		ExpectQuery("SELECT id, login, password, avatar FROM users WHERE").
		WithArgs(login).
		WillReturnRows(rows)

	repo := &UsersRepository{
		DB: db,
	}

	item, err := repo.SelectDataByLogin(login)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(item, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], item)
		return
	}

	// query error
	mock.
		ExpectQuery("SELECT id, login, password, avatar FROM users WHERE").
		WithArgs(login).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.SelectDataByLogin(login)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// row scan error
	rows = sqlmock.NewRows([]string{"id", "login"}).
		AddRow(1, "toringol")

	mock.
		ExpectQuery("SELECT id, login, password, avatar FROM users WHERE").
		WithArgs(login).
		WillReturnRows(rows)

	_, err = repo.SelectDataByLogin(login)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := &UsersRepository{
		DB: db,
	}

	username := "login"
	password := "password"
	defaultAvatar := "/images/avatar.png"
	testItem := &User{
		Username: username,
		Password: password,
		Avatar:   defaultAvatar,
	}

	//ok query
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(username, password, defaultAvatar).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := repo.Create(testItem)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if id != 1 {
		t.Errorf("bad id: want %v, have %v", id, 1)
		return
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// query error
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(username, password, defaultAvatar).
		WillReturnError(fmt.Errorf("bad query"))

	_, err = repo.Create(testItem)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// result error
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(username, password, defaultAvatar).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("bad_result")))

	_, err = repo.Create(testItem)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	var elemID int64 = 1

	rows := sqlmock.
		NewRows([]string{"id", "login", "password", "avatar"})
	testInput := []*User{
		{elemID, "toringol", "12345", "default"},
		{elemID + 1, "user", "pass", "default"},
	}

	expect := &User{elemID, "sergey", "23623", "default"}

	for _, item := range testInput {
		rows = rows.AddRow(item.ID, item.Username, item.Password, item.Avatar)
	}

	mock.
		ExpectExec(`UPDATE users SET`).
		WithArgs(expect.Username, expect.Password, expect.Avatar, expect.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := &UsersRepository{
		DB: db,
	}

	rowsAffected, err := repo.Update(expect)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if rowsAffected != 1 {
		t.Errorf("bad rowsAffected: want %v, have %v", rowsAffected, 1)
		return
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// query error
	mock.
		ExpectExec(`UPDATE users SET`).
		WithArgs(expect.Username, expect.Password, expect.Avatar, expect.ID).
		WillReturnError(fmt.Errorf("bad query"))

	_, err = repo.Update(expect)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// result error
	mock.
		ExpectExec(`UPDATE users SET`).
		WithArgs(expect.Username, expect.Password, expect.Avatar, expect.ID).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("bad_result")))

	_, err = repo.Update(expect)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	var elemID int64 = 1

	rows := sqlmock.
		NewRows([]string{"id", "login", "password", "avatar"})
	testInput := []*User{
		{elemID, "toringol", "12345", "default"},
		{elemID + 1, "user", "pass", "default"},
	}

	for _, item := range testInput {
		rows = rows.AddRow(item.ID, item.Username, item.Password, item.Avatar)
	}

	mock.
		ExpectExec(`DELETE FROM users WHERE`).
		WithArgs(elemID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := &UsersRepository{
		DB: db,
	}

	rowsAffected, err := repo.Delete(elemID)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if rowsAffected != 1 {
		t.Errorf("bad rowsAffected: want %v, have %v", rowsAffected, 1)
		return
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// query error
	mock.
		ExpectExec(`DELETE FROM users WHERE`).
		WithArgs(elemID).
		WillReturnError(fmt.Errorf("bad query"))

	_, err = repo.Delete(elemID)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// result error
	mock.
		ExpectExec(`DELETE FROM users WHERE`).
		WithArgs(elemID).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("bad_result")))

	_, err = repo.Delete(elemID)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
