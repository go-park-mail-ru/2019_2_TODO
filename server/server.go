package main

import (
	"net/http"

	userhttp "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/delivery/http"

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

	e.Logger.Fatal(e.Start(utils.ListenAddr))
}
