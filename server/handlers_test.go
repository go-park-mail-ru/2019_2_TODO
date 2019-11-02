package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

var (
	userJSON = `{"username":"hello","password":"world"}`
)

func TestSignUp(t *testing.T) {
	// mockdb, _, _ := sqlmock.New()
	// defer mockdb.Close()

	// e := echo.New()
	// req := httptest.NewRequest(http.MethodPost, "/signup/", strings.NewReader(userJSON))
	// req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// rec := httptest.NewRecorder()
	// c := e.NewContext(req, rec)

	// usersRepo := &UsersRepository{
	// 	DB: mockdb,
	// }

	// h := &Handlers{
	// 	Users: usersRepo,
	// }

	// if assert.NoError(t, h.handleSignUp(c)) {
	// 	assert.Equal(t, http.StatusCreated, rec.Code)
	// 	assert.Equal(t, userJSON, rec.Body.String())
	// }
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := NewMockUserCRUD(ctrl)

	userInput := &User{
		Username: "hello",
		Password: "",
	}

	userCRUD.EXPECT().Create(userInput).Return(int64(0), nil)

	handler := &Handlers{
		Users: userCRUD,
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/signup/", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, handler.handleSignUp(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

// import (
// 	"bytes"
// 	"net/http"
// 	"net/http/httptest"
// 	"reflect"
// 	"sync"
// 	"testing"
// )

// func TestSignUp(t *testing.T) {
// 	t.Parallel()

// 	h := Handlers{
// 		users: []Credentials{},
// 		mu:    &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{"username": "Sergey", "password": "qwerty"}`))

// 	var expectedUsers = []Credentials{
// 		{
// 			ID:       0,
// 			Username: "Sergey",
// 			Password: "qwerty",
// 		},
// 	}

// 	r := httptest.NewRequest("POST", "/signup/", body)
// 	w := httptest.NewRecorder()

// 	h.handleSignUp(w, r)

// 	if w.Code != http.StatusOK {
// 		t.Error("Failed http Status")
// 	}

// 	reflect.DeepEqual(h.users, expectedUsers)

// }

// func TestSignIn(t *testing.T) {
// 	t.Parallel()

// 	h := Handlers{
// 		users: []Credentials{},
// 		mu:    &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{"username": "Sergey", "password": "qwerty"}`))

// 	r := httptest.NewRequest("POST", "/signup/", body)
// 	w := httptest.NewRecorder()

// 	h.handleSignUp(w, r)

// 	body1 := bytes.NewReader([]byte(`{"username": "Sergey", "password": "qwerty"}`))

// 	r = httptest.NewRequest("POST", "/signin/", body1)
// 	w = httptest.NewRecorder()

// 	h.handleSignIn(w, r)

// 	if w.Code != http.StatusOK {
// 		t.Error("Failed http Status")
// 	}
// }

// func TestSignInGet(t *testing.T) {
// 	t.Parallel()

// 	var expectedResponse = `{"username": "Sergey"}`

// 	h := Handlers{
// 		users: []Credentials{},
// 		mu:    &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{"username": "Sergey", "password": "qwerty"}`))

// 	r := httptest.NewRequest("POST", "/signup/", body)
// 	w := httptest.NewRecorder()

// 	h.handleSignUp(w, r)

// 	body1 := bytes.NewReader([]byte(`{}`))

// 	r = httptest.NewRequest("GET", "/signin/", body1)
// 	w = httptest.NewRecorder()

// 	SetCookie(w, "Sergey")

// 	h.handleSignInGet(w, r)

// 	if w.Code != http.StatusOK {
// 		t.Error("Failed http Status")
// 	}

// 	reflect.DeepEqual(w.Body, expectedResponse)
// }

// func TestChangeProfile(t *testing.T) {
// 	t.Parallel()

// 	h := Handlers{
// 		users: []Credentials{},
// 		mu:    &sync.Mutex{},
// 	}

// 	var expectedUsers = []Credentials{
// 		{
// 			ID:       0,
// 			Username: "Sergey",
// 			Password: "sdhsdh",
// 		},
// 	}

// 	body := bytes.NewReader([]byte(`{"username": "Sergey", "password": "qwerty"}`))

// 	r := httptest.NewRequest("POST", "/signup/", body)
// 	w := httptest.NewRecorder()

// 	h.handleSignUp(w, r)

// 	changeBody := bytes.NewReader([]byte(`{"username": "Sergey", "password": "sdhsdh"}`))

// 	r = httptest.NewRequest("POST", "/profile/", changeBody)
// 	w = httptest.NewRecorder()

// 	h.handleChangeProfile(w, r)

// 	if w.Code != http.StatusOK {
// 		t.Error("Failed http Status")
// 	}

// 	reflect.DeepEqual(h.users, expectedUsers)
// }

// func TestSetAndReadCookie(t *testing.T) {
// 	t.Parallel()

// 	var expectedCookieUsername = "TestNickname"

// 	h := Handlers{
// 		users: []Credentials{},
// 		mu:    &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{}`))

// 	r := httptest.NewRequest("POST", "/", body)
// 	w := httptest.NewRecorder()

// 	SetCookie(w, expectedCookieUsername)

// 	username := h.ReadCookieUsername(w, r)

// 	reflect.DeepEqual(username, expectedCookieUsername)
// }

// func TestReadCookieAvatar(t *testing.T) {
// 	t.Parallel()

// 	var expectedCookieAvatar = "images/avatar.png"

// 	h := Handlers{
// 		users: []Credentials{
// 			{
// 				ID:       0,
// 				Username: "Sergey",
// 				Password: "sdhsdh",
// 				Image:    "images/avatar.png",
// 			},
// 		},
// 		mu: &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{}`))

// 	r := httptest.NewRequest("POST", "/", body)
// 	w := httptest.NewRecorder()

// 	SetCookie(w, "Sergey")

// 	username := h.ReadCookieAvatar(w, r)

// 	reflect.DeepEqual(username, expectedCookieAvatar)
// }

// func TestClearCookie(t *testing.T) {
// 	t.Parallel()

// 	var expectedResponse = ""
// 	var testCookieString = "SomeUsername"

// 	h := Handlers{
// 		users: []Credentials{},
// 		mu:    &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{}`))

// 	r := httptest.NewRequest("POST", "/", body)
// 	w := httptest.NewRecorder()

// 	SetCookie(w, testCookieString)
// 	ClearCookie(w)

// 	response := h.ReadCookieUsername(w, r)

// 	reflect.DeepEqual(response, expectedResponse)
// }

// func TestGetProfile(t *testing.T) {
// 	t.Parallel()

// 	var expectedRequestJSON = `{"username": "Sergey", "image": "images/avatar.png"}`

// 	h := Handlers{
// 		users: []Credentials{
// 			{
// 				ID:       0,
// 				Username: "Sergey",
// 				Password: "sdhsdh",
// 				Image:    "images/avatar.png",
// 			},
// 		},
// 		mu: &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{}`))

// 	r := httptest.NewRequest("GET", "/profile/", body)
// 	w := httptest.NewRecorder()

// 	SetCookie(w, "Sergey")
// 	h.handleGetProfile(w, r)

// 	reflect.DeepEqual(w.Body, expectedRequestJSON)

// 	if w.Code != http.StatusOK {
// 		t.Error("Failed http Status")
// 	}

// }

// func TestHandleLogout(t *testing.T) {
// 	t.Parallel()

// 	var expectedResponse = ""
// 	var testCookieString = "SomeUsername"

// 	h := Handlers{
// 		users: []Credentials{},
// 		mu:    &sync.Mutex{},
// 	}

// 	body := bytes.NewReader([]byte(`{}`))

// 	r := httptest.NewRequest("GET", "/logout/", body)
// 	w := httptest.NewRecorder()

// 	SetCookie(w, testCookieString)
// 	h.handleLogout(w, r)

// 	response := h.ReadCookieUsername(w, r)

// 	reflect.DeepEqual(response, expectedResponse)

// 	if w.Code != http.StatusSeeOther {
// 		t.Error("Failed http Status")
// 	}
// }
