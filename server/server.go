package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()

	dsn := dataBaseConfig
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)

	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		log.Println("Error while Ping")
	}

	usersRepo := &UsersRepository{
		DB: db,
	}

	handlers := Handlers{
		Users: usersRepo,
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${method}] ${remote_ip}, ${uri} ${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{frontIP},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	e.Static("/images", "images")

	e.GET("/", handlers.handleOk)
	e.GET("/checkUsers/", handlers.checkUsersForTesting)
	e.GET("/signin/", handlers.handleSignInGet)
	e.GET("/signin/profile/", handlers.handleGetProfile)
	e.GET("/logout/", handlers.handleLogout)

	e.POST("/signup/", handlers.handleSignUp)
	e.POST("/signin/", handlers.handleSignIn)
	e.POST("/signin/profile/", handlers.handleChangeProfile)
	e.POST("/signin/profileImage/", handlers.handleChangeImage)

	e.Logger.Fatal(e.Start(listenAddr))
}
