package main

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/microcosm-cc/bluemonday"
)

// UserCRUD - User DataBase interface
type UserCRUD interface {
	ListAll() ([]*User, error)
	SelectByID(int64) (*User, error)
	SelectDataByLogin(string) (*User, error)
	Create(*User) (int64, error)
	Update(*User) (int64, error)
	Delete(int64) (int64, error)
}

// Handlers - use UserCRUD
type Handlers struct {
	Users UserCRUD
}

func (h *Handlers) handleSignUp(ctx echo.Context) error {
	newUserInput := new(User)

	if err := ctx.Bind(newUserInput); err != nil {
		log.Println(ctx.Request().Header.Get("Content-Type"))
		log.Println(err)
		log.Println(newUserInput)
		log.Println(ctx.Request().Body)
		return ctx.JSON(http.StatusBadRequest, "")
	}

	sanitizer := bluemonday.UGCPolicy()
	newUserInput.Username = sanitizer.Sanitize(newUserInput.Username)

	newUserInput.Password = base64.StdEncoding.EncodeToString(
		convertPass(newUserInput.Password))

	newUserInput.Avatar = "/images/avatar.png"

	log.Println(newUserInput.Password)

	lastID, err := h.Users.Create(newUserInput)
	if err != nil {
		log.Println("Items.Create err:", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	log.Println("Last id: ", lastID)

	if err = SetCookie(ctx, *newUserInput); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	log.Println(newUserInput.Username)

	return nil
}

func (h *Handlers) handleSignIn(ctx echo.Context) error {

	authCredentials := new(User)

	if err := ctx.Bind(authCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	sanitizer := bluemonday.UGCPolicy()
	authCredentials.Username = sanitizer.Sanitize(authCredentials.Username)

	userRecord, err := h.Users.SelectDataByLogin(authCredentials.Username)

	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, "No such user!")
	}

	passToCheck, err := base64.StdEncoding.DecodeString(userRecord.Password)

	if err != nil || !checkPass(passToCheck, authCredentials.Password) {
		return ctx.JSON(http.StatusUnauthorized, "Incorrect password!")
	}

	log.Println("UserData: ID - ", userRecord.ID, " Login - ", userRecord.Username,
		" Avatar - ", userRecord.Avatar)

	if err = SetCookie(ctx, *userRecord); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	return nil
}

func (h *Handlers) handleSignInGet(ctx echo.Context) error {
	cookieUsername := ReadCookieUsername(ctx)
	cookieAvatar := ReadCookieAvatar(ctx)

	log.Println(cookieUsername + " " + cookieAvatar)

	if cookieUsername != "" {
		cookieUsernameInput := User{
			Username: cookieUsername,
			Avatar:   backIP + cookieAvatar,
		}
		return ctx.JSON(http.StatusCreated, cookieUsernameInput)
	}

	return nil
}

func (h *Handlers) handleOk(ctx echo.Context) error {
	return nil
}

func (h *Handlers) handleChangeProfile(ctx echo.Context) error {

	changeProfileCredentials := new(User)

	if err := ctx.Bind(changeProfileCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	oldUsername := ReadCookieUsername(ctx)

	oldData, err := h.Users.SelectDataByLogin(oldUsername)

	if err != nil {
		log.Println("Users.Update error: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	changeProfileCredentials.ID = oldData.ID
	changeProfileCredentials.Avatar = oldData.Avatar
	if changeProfileCredentials.Username == "" {
		changeProfileCredentials.Username = oldData.Username
	}

	if changeProfileCredentials.Password == "" {
		changeProfileCredentials.Password = oldData.Password
	} else {
		changeProfileCredentials.Password = base64.StdEncoding.EncodeToString(
			convertPass(changeProfileCredentials.Password))
	}

	affected, err := h.Users.Update(changeProfileCredentials)
	if err != nil {
		log.Println("Users.Update error: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}
	log.Println("Update affectedRows: ", affected)

	if err = SetCookie(ctx, *changeProfileCredentials); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	return nil
}

func (h *Handlers) handleChangeImage(ctx echo.Context) error {

	username := ReadCookieUsername(ctx)

	fileName, err := loadAvatar(ctx, username)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	changeData := new(User)

	changeData.Avatar = "/images/" + fileName

	oldData, err := h.Users.SelectDataByLogin(username)
	if err != nil {
		log.Println("Users.Update error: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	changeData.ID = oldData.ID
	changeData.Username = oldData.Username
	changeData.Password = oldData.Password

	affected, err := h.Users.Update(changeData)
	if err != nil {
		log.Println("Users.Update error: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}
	log.Println("Update affectedRows: ", affected)

	if err = SetCookie(ctx, *changeData); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	return nil
}

func loadAvatar(ctx echo.Context, username string) (string, error) {
	file, err := ctx.FormFile("image")
	if err != nil {
		log.Println("Error formFile")
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		log.Println("Error file while opening")
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(pathToImages + `images/` + file.Filename)
	if err != nil {
		log.Println("Error creating file")
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		log.Println("Error copy file")
		return "", err
	}

	return file.Filename, nil
}

func (h *Handlers) handleGetProfile(ctx echo.Context) error {

	cookiesData := User{
		Username: ReadCookieUsername(ctx),
		Avatar:   ReadCookieAvatar(ctx),
	}

	if cookiesData.Username == "" {
		return nil
	}

	return ctx.JSON(http.StatusOK, cookiesData)
}

func (h *Handlers) handleLogout(ctx echo.Context) error {
	ClearCookie(ctx)
	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) checkUsersForTesting(ctx echo.Context) error {
	if ReadCookieUsername(ctx) != "" {
		log.Println("Success checking cook")
	}

	users, err := h.Users.ListAll()
	if err != nil {
		log.Println("Error while getting all users: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	ctx.JSON(http.StatusOK, users)

	log.Println(users)

	return nil
}
