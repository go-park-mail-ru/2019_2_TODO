package http

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"server/middlewares"
	"server/model"
	"server/user"
	"server/user/utils"

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

	log.Println("Last id: ", lastID)

	if err = utils.SetCookie(ctx, *newUserInput); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	err = utils.SetToken(ctx)
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

	if err = utils.SetCookie(ctx, *userRecord); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	err = utils.SetToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Token set error")
	}

	return nil
}

func (h *Handlers) handleSignInGet(ctx echo.Context) error {
	cookieUsername := utils.ReadCookieUsername(ctx)
	cookieAvatar := utils.ReadCookieAvatar(ctx)

	log.Println(cookieUsername + " " + cookieAvatar)

	if cookieUsername != "" {
		cookieUsernameInput := model.User{
			Username: cookieUsername,
			Avatar:   utils.BackIP + cookieAvatar,
		}
		log.Println(cookieUsernameInput)
		return ctx.JSON(http.StatusCreated, cookieUsernameInput)
	}

	return nil
}

func (h *Handlers) handleOk(ctx echo.Context) error {
	utils.ClearCookie(ctx)
	return nil
}

func (h *Handlers) handleChangeProfile(ctx echo.Context) error {

	changeProfileCredentials := new(model.User)

	if err := ctx.Bind(changeProfileCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	oldUsername := utils.ReadCookieUsername(ctx)

	oldData, err := h.Users.SelectDataByLogin(oldUsername)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Users.SelectDataByLogin error")
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
		return ctx.JSON(http.StatusInternalServerError, "Users.Update error")
	}
	log.Println("Update affectedRows: ", affected)

	if err = utils.SetCookie(ctx, *changeProfileCredentials); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	return nil
}

func (h *Handlers) handleChangeImage(ctx echo.Context) error {

	username := utils.ReadCookieUsername(ctx)

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

	if err = utils.SetCookie(ctx, *changeData); err != nil {
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

	cookiesData := model.User{
		Username: utils.ReadCookieUsername(ctx),
		Avatar:   utils.ReadCookieAvatar(ctx),
	}

	if cookiesData.Username == "" {
		return nil
	}

	return ctx.JSON(http.StatusOK, cookiesData)
}

func (h *Handlers) handleLogout(ctx echo.Context) error {
	utils.ClearCookie(ctx)
	return ctx.JSON(http.StatusOK, "")
}
