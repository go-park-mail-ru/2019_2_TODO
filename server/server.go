package main

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	userhttp "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/delivery/http"
	"google.golang.org/grpc"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/repository"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/usecase"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${method}] ${remote_ip}, ${uri} ${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{utils.FrontIP, utils.FrontIPChat},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	e.Static("/images", "images")

	userhttp.NewUserHandler(e, usecase.NewUserUsecase(repository.NewUserMemoryRepository()))

	grcpConn, err := grpc.Dial(
		"127.0.0.1:8080",
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("cant connect to grpc")
	}
	defer grcpConn.Close()

	utils.SessManager = session.NewAuthCheckerClient(grcpConn)

	e.Logger.Fatal(e.Start(utils.ListenAddr), "cert.pem", "key.pem", nil)
}
