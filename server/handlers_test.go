package main

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/golang/mock/gomock"
// 	"github.com/labstack/echo"
// 	"github.com/stretchr/testify/assert"
// )

// var (
// 	userJSON = `{"username":"hello","password":"world"}`
// )

// func TestSignUp(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	userCRUD := NewMockUserCRUD(ctrl)

// 	userInput := &User{
// 		Username: "hello",
// 		Password: "",
// 	}

// 	userCRUD.EXPECT().Create(userInput).Return(int64(0), nil)

// 	handler := &Handlers{
// 		Users: userCRUD,
// 	}
// 	e := echo.New()
// 	req := httptest.NewRequest(http.MethodPost, "/signup/", strings.NewReader(userJSON))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)

// 	if assert.NoError(t, handler.handleSignUp(c)) {
// 		assert.Equal(t, http.StatusOK, rec.Code)
// 	}
// }
