package main

import (
	"chat/chatLink/core"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	ADDR    string = ":8080"
	FrontIP        = "localhost"
)

type JSONRooms struct {
	Rooms []string `json:"rooms"`
}

func getRooms(ctx echo.Context) error {
	var rooms = []string{}
	for r := range core.AllRooms {
		rooms = append(rooms, r)
	}
	var jsonRooms = &JSONRooms{
		Rooms: rooms,
	}
	return ctx.JSON(http.StatusOK, jsonRooms)
}

func wsHandler(ctx echo.Context) {
	ws, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(ctx.Response(), "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}

	username := "User" + strconv.Itoa(int(core.IDUser))
	params, _ := url.ParseQuery(ctx.Request().URL.RawQuery)
	if len(params["name"]) > 0 {
		username = params["name"][0]
	}

	// Get or create a room
	var room *core.Room
	if len(core.FreeRooms) > 0 {
		for _, r := range core.FreeRooms {
			room = r
			break
		}
	} else {
		room = core.NewRoom(username)
	}

	// Create User and Conn
	user := core.NewUser(username, false)
	uConn := core.NewUserConn(ws, user, room)
	// Join Player to room
	room.Join <- uConn

	log.Printf("User: %s has joined to room: %s", uConn.Msg.Autor, room.Name)
}

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${method}] ${remote_ip}, ${uri} ${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{FrontIP},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	e.GET("getRooms", func(ctx echo.Context) error {
		getRooms(ctx)
		return nil
	})

	e.GET("/chatRoom/", func(ctx echo.Context) error {
		wsHandler(ctx)
		return nil
	})

	e.Logger.Fatal(e.Start(ADDR))
}
