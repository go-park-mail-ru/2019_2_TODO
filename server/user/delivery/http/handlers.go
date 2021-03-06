package http

import (
	"context"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/middlewares"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/model"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/labstack/echo"
	"github.com/microcosm-cc/bluemonday"
)

// Handlers - use UserCRUD
type Handlers struct {
	Users user.Usecase
}

// NewUserHandler - deliver our handlers in http
func NewUserHandler(e *echo.Echo, us user.Usecase) {
	handlers := Handlers{Users: us}

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/", handlers.handleOk)
	e.GET("/signin/", handlers.handleSignInGet)
	e.GET("/signin/profile/", handlers.handleGetProfile)
	e.GET("/logout/", handlers.handleLogout)

	e.POST("/signup/", handlers.handleSignUp)
	e.POST("/signin/", handlers.handleSignIn)
	e.POST("/signin/profile/", handlers.handleChangeProfile, middlewares.JWTMiddlewareCustom)
	e.POST("/signin/profileImage/", handlers.handleChangeImage, middlewares.JWTMiddlewareCustom)
}

func (h *Handlers) handleSignUp(ctx echo.Context) error {
	log.Println(utils.SessManager)
	newUserInput := new(model.User)

	if err := ctx.Bind(newUserInput); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	sanitizer := bluemonday.UGCPolicy()
	newUserInput.Username = sanitizer.Sanitize(newUserInput.Username)

	newUserInput.Avatar = "/images/avatar.png"

	lastID, err := h.Users.Create(newUserInput)
	if err != nil {
		log.Println("Items.Create err:", err)
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	userLeader := &leaderBoardModel.UserLeaderBoard{
		ID:       lastID,
		Username: newUserInput.Username,
		Points:   "1000",
	}

	_, err = h.Users.CreateLeader(userLeader)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "LeaderBoard Error")
	}

	log.Println("Last id: ", lastID)
	newUserInput.ID = lastID

	var cookie *http.Cookie
	if cookie, err = utils.SetSession(ctx, newUserInput); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	err = utils.SetToken(ctx, cookie)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Token set error")
	}

	log.Println(newUserInput.Username)

	return nil
}

func (h *Handlers) handleSignIn(ctx echo.Context) error {
	authCredentials := new(model.User)

	if err := ctx.Bind(authCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	authCredentials.Username = utils.Sanitizer.Sanitize(authCredentials.Username)

	userRecord, err := h.Users.SelectDataByLogin(authCredentials.Username)

	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, "No such user!")
	}

	passToCheck, err := base64.StdEncoding.DecodeString(userRecord.Password)

	if err != nil || !utils.CheckPass(passToCheck, authCredentials.Password) {
		return ctx.JSON(http.StatusUnauthorized, "Incorrect password!")
	}

	log.Println("UserData: ID - ", userRecord.ID, " Login - ", userRecord.Username,
		" Avatar - ", userRecord.Avatar)

	var cookie *http.Cookie
	if cookie, err = utils.SetSession(ctx, userRecord); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	err = utils.SetToken(ctx, cookie)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Token set error")
	}

	return nil
}

func (h *Handlers) handleSignInGet(ctx echo.Context) error {
	session, err := utils.SessManager.Check(
		context.Background(),
		utils.ReadSessionID(ctx),
	)

	if err != nil {
		return nil
	}

	sessIDAndUserID := utils.ReadSessionIDAndUserID(ctx)
	if sessIDAndUserID == nil {
		return ctx.JSON(http.StatusInternalServerError, "Not ok")
	}

	userID, err := strconv.Atoi(sessIDAndUserID[1])
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Not ok")
	}
	cookieUsername := session.Username
	cookieAvatar := session.Avatar

	log.Println(cookieUsername + " " + cookieAvatar)

	if cookieUsername != "" {
		cookieUsernameInput := model.User{}
		if cookieUsername == "Resg" {
			cookieUsernameInput = model.User{
				ID:       int64(userID),
				Username: cookieUsername,
				Avatar:   cookieAvatar,
				Admin:    true,
			}
		} else {
			cookieUsernameInput = model.User{
				ID:       int64(userID),
				Username: cookieUsername,
				Avatar:   cookieAvatar,
			}
		}
		log.Println(cookieUsernameInput)
		return ctx.JSON(http.StatusCreated, cookieUsernameInput)
	}

	return nil
}

func (h *Handlers) handleOk(ctx echo.Context) error {
	return nil
}

func (h *Handlers) handleChangeProfile(ctx echo.Context) error {
	log.Println("Im here")

	changeProfileCredentials := new(model.User)

	if err := ctx.Bind(changeProfileCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	log.Println(changeProfileCredentials)

	session, err := utils.СheckSession(ctx)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Error Checking session")
	}

	oldUsername := session.Username

	oldData, err := h.Users.SelectDataByLogin(oldUsername)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Users.SelectDataByLogin error")
	}

	changeProfileCredentials.ID = oldData.ID
	changeProfileCredentials.Avatar = oldData.Avatar
	if changeProfileCredentials.Username == "" {
		changeProfileCredentials.Username = oldData.Username
	} else {
		elem, err := h.Users.SelectLeaderByID(changeProfileCredentials.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "LeaderBoard Error")
		}
		elem.Username = changeProfileCredentials.Username
		_, err = h.Users.UpdateLeader(elem)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "LeaderBoard Error")
		}
	}

	if changeProfileCredentials.Password == "" {
		changeProfileCredentials.Password = oldData.Password
	}

	affected, err := h.Users.Update(changeProfileCredentials)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Users.Update error")
	}
	log.Println("Update affectedRows: ", affected)

	if _, err = utils.SetSession(ctx, changeProfileCredentials); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	return nil
}

func (h *Handlers) handleChangeImage(ctx echo.Context) error {
	session, err := utils.SessManager.Check(
		context.Background(),
		utils.ReadSessionID(ctx),
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Error Checking session")
	}

	username := session.Username

	fileName, err := loadAvatar(ctx, username)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "")
	}

	changeData := new(model.User)

	changeData.Avatar = "/images/" + fileName

	oldData, err := h.Users.SelectDataByLogin(username)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Users.Update error:")
	}

	changeData.ID = oldData.ID
	changeData.Username = oldData.Username
	changeData.Password = oldData.Password

	affected, err := h.Users.Update(changeData)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Users.Update error")
	}
	log.Println("Update affectedRows: ", affected)

	if _, err = utils.SetSession(ctx, changeData); err != nil {
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

	dst, err := os.Create(utils.PathToImages + `images/` + file.Filename)
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

	session, err := utils.SessManager.Check(
		context.Background(),
		utils.ReadSessionID(ctx),
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Error Checking session")
	}

	cookiesData := model.User{
		Username: session.Username,
		Avatar:   session.Avatar,
	}

	if cookiesData.Username == "" {
		return nil
	}

	return ctx.JSON(http.StatusOK, cookiesData)
}

func (h *Handlers) handleLogout(ctx echo.Context) error {
	utils.ClearSession(ctx)
	return ctx.JSON(http.StatusOK, "")
}
