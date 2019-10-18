package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
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

type UserCRUD interface {
	ListAll() ([]*User, error)
	SelectByID(int64) (*User, error)
	SelectDataByLogin(string) (*User, error)
	SelectByLoginAndPassword(*User) (*User, error)
	Create(*User) (int64, error)
	Update(*User) (int64, error)
	Delete(int64) (int64, error)
}

type Handlers struct {
	Users UserCRUD
}

func (h *Handlers) handleSignUp(ctx echo.Context) error {

	newUserInput := new(User)

	if err := ctx.Bind(newUserInput); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	lastID, err := h.Users.Create(newUserInput)
	if err != nil {
		log.Println("Items.Create err:", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	log.Println("Last id: ", lastID)

	SetCookie(ctx, newUserInput.Username)

	log.Println(newUserInput.Username)

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleSignIn(ctx echo.Context) error {

	authCredentials := new(User)

	if err := ctx.Bind(authCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	userRecord, err := h.Users.SelectByLoginAndPassword(authCredentials)

	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, "")
	}

	log.Println("UserData: ID - ", userRecord.ID, " Login - ", userRecord.Username,
		" Avatar - ", userRecord.Avatar)

	SetCookie(ctx, authCredentials.Username)

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

	changeProfileCredentials := new(User)

	if err := ctx.Bind(changeProfileCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	oldUsername := h.ReadCookieUsername(ctx)

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
	}

	affected, err := h.Users.Update(changeProfileCredentials)
	if err != nil {
		log.Println("Users.Update error: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}
	log.Println("Update affectedRows: ", affected)

	ClearCookie(ctx)
	SetCookie(ctx, changeProfileCredentials.Username)

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) handleChangeImage(ctx echo.Context) error {

	username := h.ReadCookieUsername(ctx)

	err := loadAvatar(ctx, username)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	changeData := new(User)

	changeData.Avatar = "images/" + username + ".png"

	oldData, err := h.Users.SelectDataByLogin(username)
	if err != nil {
		log.Println("Users.Update error: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	changeData.ID = oldData.ID

	affected, err := h.Users.Update(changeData)
	if err != nil {
		log.Println("Users.Update error: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}
	log.Println("Update affectedRows: ", affected)

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

	cookiesData := CredentialsInput{
		Username: h.ReadCookieUsername(ctx),
		Image:    h.ReadCookieAvatar(ctx),
	}

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

	users, err := h.Users.ListAll()
	if err != nil {
		log.Println("Error while getting all users: ", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	ctx.JSON(http.StatusOK, users)

	log.Println(users)

	return ctx.JSON(http.StatusOK, "")
}

// func (h *Handlers) checkUsername(accounts []Credentials, authCredentials *CredentialsInput) error {
// 	if len(accounts) == 0 {
// 		return errors.New("No users")
// 	}

// 	sort.Slice(accounts[:], func(i, j int) bool {
// 		return accounts[i].Username < accounts[j].Username
// 	})

// 	iter := sort.Search(len(accounts), func(i int) bool {
// 		return accounts[i].Username == authCredentials.Username
// 	})

// 	if iter < len(accounts) && accounts[iter].Username == authCredentials.Username {
// 		return nil
// 	} else {
// 		return errors.New("No such user")
// 	}
// }

// func (h *Handlers) checkPassword(accounts []Credentials, authCredentials *CredentialsInput) error {

// 	if len(accounts) == 0 {
// 		return errors.New("No users")
// 	}

// 	sort.Slice(accounts[:], func(i, j int) bool {
// 		return accounts[i].Password < accounts[j].Password
// 	})

// 	iter := sort.Search(len(accounts), func(i int) bool {
// 		return accounts[i].Password == authCredentials.Password
// 	})

// 	if iter < len(accounts) && accounts[iter].Password == authCredentials.Password {
// 		return nil
// 	} else {
// 		return errors.New("Wrong password")
// 	}
// }

// func (h *Handlers) changeProfile(accounts []Credentials, changeProfileCredentials *CredentialsInput, oldUsername string) {
// 	sort.Slice(accounts[:], func(i, j int) bool {
// 		return accounts[i].Username < accounts[j].Username
// 	})

// 	iter := sort.Search(len(accounts), func(i int) bool {
// 		return accounts[i].Username == oldUsername
// 	})

// 	if changeProfileCredentials.Username != "" {
// 		accounts[iter].Username = changeProfileCredentials.Username
// 	}
// 	if changeProfileCredentials.Password != "" {
// 		accounts[iter].Password = changeProfileCredentials.Password
// 	}
// 	if changeProfileCredentials.Image != "" {
// 		accounts[iter].Image = changeProfileCredentials.Image
// 	}
// }

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
			accounts, err := h.Users.ListAll()
			if err != nil {
				return ""
			}
			sort.Slice(accounts[:], func(i, j int) bool {
				return accounts[i].Username < accounts[j].Username
			})

			iter := sort.Search(len(accounts), func(i int) bool {
				return accounts[i].Username == value["username"]
			})
			if iter >= len(accounts) {
				return "images/avatar.png"
			}
			return accounts[iter].Avatar
		}
	}
	return ""
}
