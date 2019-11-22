package main

import (
	"log"
	"net/http"
	"net/url"
	"server/game/core"
	"server/user/repository"
	"server/user/usecase"
	"server/user/utils"

	userhttp "server/user/delivery/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/go-sql-driver/mysql"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}

	playerName := "Player"
	var playerStartChips int32 = 1000
	params, _ := url.ParseQuery(r.URL.RawQuery)
	if len(params["name"]) > 0 {
		playerName = params["name"][0]
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
	player := core.NewPlayer(playerName, playerStartChips)
	pConn := core.NewPlayerConn(ws, player, room)
	// Join Player to room
	room.Join <- pConn

	log.Printf("Player: %s has joined to room: %s", pConn.Name, room.Name)
}

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${method}] ${remote_ip}, ${uri} ${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{utils.FrontIP},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	e.Static("/images", "images")

	userhttp.NewUserHandler(e, usecase.NewUserUsecase(repository.NewUserMemoryRepository()))
	http.HandleFunc("/multiplayer", wsHandler)

	e.Logger.Fatal(e.Start(utils.ListenAddr))
}
