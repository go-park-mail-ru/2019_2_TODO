package main

import (
	"chat/chatLink/core"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	ADDR    string = ":8080"
	FrontIP        = "localhost"
)

func homeHandler(ctx echo.Context) {
	var homeTempl = template.Must(template.ParseFiles("templates/chat.html"))
	data := struct {
		Host       string
		RoomsCount int
	}{ctx.Request().Host, core.RoomsCount}
	homeTempl.Execute(ctx.Response(), data)
}

func wsHandler(ctx echo.Context) {
	ws, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(ctx.Response(), "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}

	username := "User"
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
		room = core.NewRoom("")
	}

	// Create Player and Conn
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

	e.GET("/", func(ctx echo.Context) error {
		homeHandler(ctx)
		return nil
	})
	e.GET("/chatRoom/", func(ctx echo.Context) error {
		wsHandler(ctx)
		return nil
	})

	e.Logger.Fatal(e.Start(ADDR))
}
