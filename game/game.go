package game

import (
	"game/core"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	ListenAddr = "172.26.112.3:82"
	FrontIP    = "http://93.171.139.195:782"
)

type JSONRooms struct {
	Rooms map[string]int `json:"rooms"`
}

func (h *Handlers) getRooms(ctx echo.Context) error {
	if len(core.FreeRooms) == 2 {
		for i := 0; i < 4; i++ {
			core.NewRoom("")
		}
	}
	var rooms = map[string]int{}
	for r, room := range core.FreeRooms {
		rooms[r] = len(room.PlayerConns)
	}
	var jsonRooms = &JSONRooms{
		Rooms: rooms,
	}
	return ctx.JSON(http.StatusOK, jsonRooms)
}

func (h *Handlers) wsHandler(ctx echo.Context) error {
	ws, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(ctx.Response(), "Not a websocket handshake", 400)
		return err
	} else if err != nil {
		return err
	}

	playerName := "Player"
	var playerStartChips int = 1000
	params, _ := url.ParseQuery(ctx.Request().URL.RawQuery)
	if len(params["name"]) > 0 {
		playerName = params["name"][0]
	}

	var roomName string = "newRoom"
	if len(params["roomName"]) > 0 {
		roomName = params["roomName"][0]
	}

	// Get or create a room
	var room *core.Room
	if roomName != "newRoom" {
		room = core.AllRooms[roomName]
	} else {
		room = core.NewRoom("")
	}

	// Create Player and Conn
	player := core.NewPlayer(playerName, playerStartChips)
	pConn := core.NewPlayerConn(ws, player, room)
	// Join Player to room
	room.Join <- pConn

	log.Printf("Player: %s has joined to room: %s", pConn.Name, room.Name)

	return nil
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

	e.GET("/rooms/", func(ctx echo.Context) error {
		getRooms(ctx)
		return nil
	})

	e.GET("/multiplayer/", func(ctx echo.Context) error {
		wsHandler(ctx)
		return nil
	})

	e.Logger.Fatal(e.Start(ListenAddr))
}
