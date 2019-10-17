package main

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	frontIp      = "http://93.171.139.195:780"
	backIp       = "http://93.171.139.196:780"
	pathToImages = `/root/golang/test/2019_2_TODO/server/`
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

type Handlers struct {
	users []Credentials
	mu    *sync.Mutex
}

func (h *Handlers) handleSignUp(ctx echo.Context) error {

	newUserInput := new(CredentialsInput)

	if err := ctx.Bind(newUserInput); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	h.mu.Lock()

	var idUser uint64 = 0
	var defaultImage = "images/avatar.png"

	if len(h.users) > 0 {
		idUser = h.users[len(h.users)-1].ID + 1
	}

	h.users = append(h.users, Credentials{
		ID:       idUser,
		Username: newUserInput.Username,
		Password: newUserInput.Password,
		Image:    defaultImage,
	})
	h.mu.Unlock()

	SetCookie(ctx, newUserInput.Username)

	log.Println(newUserInput.Username)

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleSignIn(ctx echo.Context) error {

	authCredentials := new(CredentialsInput)

	if err := ctx.Bind(authCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	h.mu.Lock()

	accounts := h.users

	err := h.checkUsername(accounts, authCredentials)

	if err != nil {
		return ctx.JSON(http.StatusOK, "")
	}

	err = h.checkPassword(accounts, authCredentials)

	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, "")
	}

	SetCookie(ctx, authCredentials.Username)

	h.mu.Unlock()

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleSignInGet(ctx echo.Context) error {
	cookieUsername := h.ReadCookieUsername(ctx)
	cookieAvatar := h.ReadCookieAvatar(ctx)

	log.Println(cookieUsername)
	log.Println(cookieAvatar)

	if cookieUsername != "" {
		cookieUsernameInput := CredentialsInput{
			Username: cookieUsername,
			Image:    backIp + cookieAvatar,
		}

		return ctx.JSON(http.StatusCreated, cookieUsernameInput)
	}

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleOk(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleChangeProfile(ctx echo.Context) error {

	changeProfileCredentials := new(CredentialsInput)

	if err := ctx.Bind(changeProfileCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	h.mu.Lock()

	oldUsername := h.ReadCookieUsername(ctx)

	h.changeProfile(h.users, changeProfileCredentials, oldUsername)

	ClearCookie(ctx)
	SetCookie(ctx, changeProfileCredentials.Username)

	h.mu.Unlock()

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleChangeImage(ctx echo.Context) error {

	username := h.ReadCookieUsername(ctx)

	err := loadAvatar(ctx, username)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	changeData := new(CredentialsInput)

	changeData.Image = "images/" + username + ".png"

	h.changeProfile(h.users, changeData, username)

	return ctx.JSON(http.StatusOK, "")
}

func loadAvatar(ctx echo.Context, username string) error {
	file, err := ctx.FormFile("image")
	if err != nil {
		log.Println("Error formFile")
		return err
	}

	src, err := file.Open()
	if err != nil {
		log.Println("Error file while opening")
		return err
	}
	defer src.Close()

	dst, err := os.Create(pathToImages + `images/` + file.Filename)
	os.Rename(pathToImages+"images/"+file.Filename,
		pathToImages+"images/"+username+".png")
	if err != nil {
		log.Println("Error creating file")
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		log.Println("Error copy file")
		return err
	}

	return nil
}

func (h *Handlers) handleGetProfile(ctx echo.Context) error {

	h.mu.Lock()

	cookiesData := CredentialsInput{
		Username: h.ReadCookieUsername(ctx),
		Image:    h.ReadCookieAvatar(ctx),
	}

	h.mu.Unlock()

	return ctx.JSON(http.StatusOK, cookiesData)
}

func (h *Handlers) handleGetImage(ctx echo.Context) error {
	avatar := h.ReadCookieAvatar(ctx)

	log.Println(avatar)

	http.ServeFile(ctx.Response(), ctx.Request(), pathToImages+avatar)

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleLogout(ctx echo.Context) error {
	ClearCookie(ctx)
	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) checkUsersForTesting(ctx echo.Context) error {
	if h.ReadCookieUsername(ctx) != "" {
		log.Println("Success checking cook")
	}

	h.mu.Lock()
	ctx.JSON(http.StatusOK, h.users)
	h.mu.Unlock()

	log.Println(h.users)

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) checkUsername(accounts []Credentials, authCredentials *CredentialsInput) error {
	if len(accounts) == 0 {
		return errors.New("No users")
	}

	sort.Slice(accounts[:], func(i, j int) bool {
		return accounts[i].Username < accounts[j].Username
	})

	iter := sort.Search(len(accounts), func(i int) bool {
		return accounts[i].Username == authCredentials.Username
	})

	if iter < len(accounts) && accounts[iter].Username == authCredentials.Username {
		return nil
	} else {
		return errors.New("No such user")
	}
}

func (h *Handlers) checkPassword(accounts []Credentials, authCredentials *CredentialsInput) error {

	if len(accounts) == 0 {
		return errors.New("No users")
	}

	sort.Slice(accounts[:], func(i, j int) bool {
		return accounts[i].Password < accounts[j].Password
	})

	iter := sort.Search(len(accounts), func(i int) bool {
		return accounts[i].Password == authCredentials.Password
	})

	if iter < len(accounts) && accounts[iter].Password == authCredentials.Password {
		return nil
	} else {
		return errors.New("Wrong password")
	}
}

func (h *Handlers) changeProfile(accounts []Credentials, changeProfileCredentials *CredentialsInput, oldUsername string) {
	sort.Slice(accounts[:], func(i, j int) bool {
		return accounts[i].Username < accounts[j].Username
	})

	iter := sort.Search(len(accounts), func(i int) bool {
		return accounts[i].Username == oldUsername
	})

	if changeProfileCredentials.Username != "" {
		accounts[iter].Username = changeProfileCredentials.Username
	}
	if changeProfileCredentials.Password != "" {
		accounts[iter].Password = changeProfileCredentials.Password
	}
	if changeProfileCredentials.Image != "" {
		accounts[iter].Image = changeProfileCredentials.Image
	}
}

func SetCookie(ctx echo.Context, userName string) {
	value := map[string]string{
		"username": userName,
	}

	if encoded, err := cookieHandler.Encode("session_token", value); err == nil {
		expiration := time.Now().Add(24 * time.Hour)
		cookie := http.Cookie{
			Name:    "session_token",
			Value:   encoded,
			Path:    "/",
			Expires: expiration,
		}

		http.SetCookie(ctx.Response(), &cookie)
	}
}

func ClearCookie(ctx echo.Context) {
	cookie := http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(ctx.Response(), &cookie)
}

func (h *Handlers) ReadCookieUsername(ctx echo.Context) string {
	if cookie, err := ctx.Request().Cookie("session_token"); err == nil {
		value := make(map[string]string)
		if err = cookieHandler.Decode("session_token", cookie.Value, &value); err == nil {
			return value["username"]
		}
	}
	return ""
}

func (h *Handlers) ReadCookieAvatar(ctx echo.Context) string {
	if cookie, err := ctx.Request().Cookie("session_token"); err == nil {
		value := make(map[string]string)
		if err = cookieHandler.Decode("session_token", cookie.Value, &value); err == nil {
			accounts := h.users
			sort.Slice(accounts[:], func(i, j int) bool {
				return accounts[i].Username < accounts[j].Username
			})

			iter := sort.Search(len(accounts), func(i int) bool {
				return accounts[i].Username == value["username"]
			})
			if iter >= len(accounts) {
				return "images/avatar.png"
			}
			return accounts[iter].Image
		}
	}
	return ""
}
