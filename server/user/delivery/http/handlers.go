package http

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/middlewares"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/model"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"

	"github.com/labstack/echo"
	"github.com/microcosm-cc/bluemonday"

	consulapi "github.com/hashicorp/consul/api"
)

var (
	consulAddr = flag.String("addr", "127.0.0.1:8500", "consul addr (8500 in original consul)")
)

var (
	consul       *consulapi.Client
	nameResolver *testNameResolver
)

// Handlers - use UserCRUD
type Handlers struct {
	Users user.Usecase
}

// NewUserHandler - deliver our handlers in http
func NewUserHandler(e *echo.Echo, us user.Usecase) {
	flag.Parse()

	var err error
	config := consulapi.DefaultConfig()
	config.Address = *consulAddr
	consul, err = consulapi.NewClient(config)

	health, _, err := consul.Health().Service("session-api", "", false, nil)
	if err != nil {
		log.Fatalf("cant get alive services")
	}

	servers := []string{}
	for _, item := range health {
		addr := item.Service.Address +
			":" + strconv.Itoa(item.Service.Port)
		servers = append(servers, addr)
	}

	nameResolver = &testNameResolver{
		addr: servers[0],
	}
	log.Println(nameResolver)

	grcpConn, err := grpc.Dial(
		servers[0],
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithBalancer(grpc.RoundRobin(nameResolver)),
	)
	if err != nil {
		log.Fatalf("cant connect to grpc")
	}
	defer grcpConn.Close()

	if len(servers) > 1 {
		var updates []*naming.Update
		for i := 1; i < len(servers); i++ {
			updates = append(updates, &naming.Update{
				Op:   naming.Add,
				Addr: servers[i],
			})
		}
		nameResolver.w.inject(updates)
	}

	utils.SessManager = session.NewAuthCheckerClient(grcpConn)

	go runOnlineServiceDiscovery(servers)

	// ctx := context.Background()
	// step := 1
	// for {
	// 	// проверяем несуществуюущую сессию
	// 	// потому что сейчас между сервисами нет общения
	// 	// получаем загшулку
	// 	sess, err := utils.SessManager.Check(ctx,
	// 		&session.SessionID{
	// 			ID: "not_exist_" + strconv.Itoa(step),
	// 		})
	// 	fmt.Println("get sess", step, sess, err)

	// 	time.Sleep(1500 * time.Millisecond)
	// 	step++
	// }

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

func (h *Handlers) handlePrometheus(ctx echo.Context) error {
	promhttp.Handler()
	return nil
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
	newUserInput.ID = lastID

	if err = utils.SetSession(ctx, *newUserInput); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	// err = utils.SetToken(ctx)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, "Token set error")
	// }

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

	if err = utils.SetSession(ctx, *userRecord); err != nil {
		ctx.JSON(http.StatusInternalServerError, "Cookie set error")
	}

	// err = utils.SetToken(ctx)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, "Token set error")
	// }

	return nil
}

func (h *Handlers) handleSignInGet(ctx echo.Context) error {
	return nil

	session, err := utils.SessManager.Check(
		context.Background(),
		utils.ReadSessionID(ctx),
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Error Checking session")
	}

	cookieUsername := session.Username
	cookieAvatar := session.Avatar

	log.Println(cookieUsername + " " + cookieAvatar)

	if cookieUsername != "" {
		cookieUsernameInput := model.User{}
		if cookieUsername == "Resg" {
			cookieUsernameInput = model.User{
				Username: cookieUsername,
				Avatar:   utils.BackIP + cookieAvatar,
				Admin:    true,
			}
		} else {
			cookieUsernameInput = model.User{
				Username: cookieUsername,
				Avatar:   utils.BackIP + cookieAvatar,
			}
		}
		log.Println(cookieUsernameInput)
		return ctx.JSON(http.StatusCreated, cookieUsernameInput)
	}

	return nil
}

func (h *Handlers) handleOk(ctx echo.Context) error {
	// utils.ClearSession(ctx)
	return nil
}

func (h *Handlers) handleChangeProfile(ctx echo.Context) error {

	changeProfileCredentials := new(model.User)

	if err := ctx.Bind(changeProfileCredentials); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	session, err := utils.SessManager.Check(
		context.Background(),
		utils.ReadSessionID(ctx),
	)

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
	}

	if changeProfileCredentials.Password == "" {
		changeProfileCredentials.Password = oldData.Password
	}

	affected, err := h.Users.Update(changeProfileCredentials)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Users.Update error")
	}
	log.Println("Update affectedRows: ", affected)

	if err = utils.SetSession(ctx, *changeProfileCredentials); err != nil {
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

	if err = utils.SetSession(ctx, *changeData); err != nil {
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

func runOnlineServiceDiscovery(servers []string) {
	currAddrs := make(map[string]struct{}, len(servers))
	for _, addr := range servers {
		currAddrs[addr] = struct{}{}
	}
	ticker := time.Tick(5 * time.Second)
	for _ = range ticker {
		health, _, err := consul.Health().Service("session-api", "", false, nil)
		if err != nil {
			log.Fatalf("cant get alive services")
		}

		newAddrs := make(map[string]struct{}, len(health))
		for _, item := range health {
			addr := item.Service.Address +
				":" + strconv.Itoa(item.Service.Port)
			newAddrs[addr] = struct{}{}
		}

		var updates []*naming.Update
		// проверяем что удалилось
		for addr := range currAddrs {
			if _, exist := newAddrs[addr]; !exist {
				updates = append(updates, &naming.Update{
					Op:   naming.Delete,
					Addr: addr,
				})
				delete(currAddrs, addr)
				fmt.Println("remove", addr)
			}
		}
		// проверяем что добавилось
		for addr := range newAddrs {
			if _, exist := currAddrs[addr]; !exist {
				updates = append(updates, &naming.Update{
					Op:   naming.Add,
					Addr: addr,
				})
				currAddrs[addr] = struct{}{}
				fmt.Println("add", addr)
			}
		}
		if len(updates) > 0 {
			nameResolver.w.inject(updates)
		}
	}
}
