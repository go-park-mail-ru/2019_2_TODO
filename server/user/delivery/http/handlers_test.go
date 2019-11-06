package http

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"server/model"
	"server/user"
	"server/user/utils"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestSignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := user.NewMockUsecase(ctrl)

	var userJSON = `{"username": "hello", "password": "123"}`

	userInput := &model.User{
		Username: "hello",
		Password: "123",
		Avatar:   "/images/avatar.png",
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

func TestSignIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := user.NewMockUsecase(ctrl)

	var login = "user"
	var userJSON = `{"username": "user", "password": "123"}`

	userOutput := &model.User{
		ID:       0,
		Username: login,
		Password: "123",
		Avatar:   "/images/avatar.png",
	}

	userOutput.Password = base64.StdEncoding.EncodeToString(
		utils.ConvertPass(userOutput.Password))

	userCRUD.EXPECT().SelectDataByLogin(login).Return(userOutput, nil)

	handler := &Handlers{
		Users: userCRUD,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/signin/", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, handler.handleSignIn(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestSignInGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := user.NewMockUsecase(ctrl)

	inputUser := &model.User{
		Username: "DBTest",
		Password: "123",
		Avatar:   "/images/avatar.png",
	}

	outputJSON := `{"id":0,"username":"DBTest","password":"","image":"http://93.171.139.196:780/images/avatar.png"}`

	handler := &Handlers{
		Users: userCRUD,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/signin/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	utils.SetCookie(c, *inputUser)

	if assert.NoError(t, handler.handleSignInGet(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, outputJSON, strings.Trim(rec.Body.String(), "\n"))
	}
}

func TestOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := user.NewMockUsecase(ctrl)

	handler := &Handlers{
		Users: userCRUD,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, handler.handleSignInGet(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestChangeProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := user.NewMockUsecase(ctrl)

	var login = "user"
	var userJSON = `{"username": "newUser", "password": "newPass"}`

	userDataBefore := &model.User{
		ID:       0,
		Username: login,
		Password: "123",
		Avatar:   "/images/avatar.png",
	}

	newUserData := &model.User{
		ID:       0,
		Username: "newUser",
		Password: "newPass",
		Avatar:   "/images/avatar.png",
	}

	userCRUD.EXPECT().SelectDataByLogin(login).Return(userDataBefore, nil)
	userCRUD.EXPECT().Update(newUserData).Return(int64(1), nil)

	handler := &Handlers{
		Users: userCRUD,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/signin/profile/", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	utils.SetCookie(c, *userDataBefore)

	if assert.NoError(t, handler.handleChangeProfile(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := user.NewMockUsecase(ctrl)

	inputUser := &model.User{
		Username: "DBTest",
		Password: "123",
		Avatar:   "/images/avatar.png",
	}

	outputJSON := `{"id":0,"username":"DBTest","password":"","image":"/images/avatar.png"}`

	handler := &Handlers{
		Users: userCRUD,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/signin/profile/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	utils.SetCookie(c, *inputUser)

	if assert.NoError(t, handler.handleGetProfile(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, outputJSON, strings.Trim(rec.Body.String(), "\n"))
	}
}

func TestLogout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userCRUD := user.NewMockUsecase(ctrl)

	inputUser := &model.User{
		Username: "DBTest",
		Password: "123",
		Avatar:   "/images/avatar.png",
	}

	handler := &Handlers{
		Users: userCRUD,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	utils.SetCookie(c, *inputUser)

	if assert.NoError(t, handler.handleLogout(c)) {
		_, err := req.Cookie("session_token")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Error(t, err)
	}
}
