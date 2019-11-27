package main

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/chat/chatLink/core"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/session"
	"google.golang.org/grpc"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	ListenAddr = "172.26.112.3:81"
	FrontIP    = "http://93.171.139.195:781"
)

var (
	sessManager session.AuthCheckerClient
)

type tokenAuth struct {
	Token string
}

func (t *tokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"session_token": t.Token,
	}, nil
}

func (c *tokenAuth) RequireTransportSecurity() bool {
	return false
}

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

	username := "User"
	params, _ := url.ParseQuery(ctx.Request().URL.RawQuery)
	if len(params["name"]) > 0 {
		username = params["name"][0]
	}

	ctxSes := context.Background()
	nothing := &session.Nothing{}
	sessionData, err := sessManager.HandleSignInGet(ctxSes, nothing)

	if err != nil {
		log.Println("Error while grcp session check")
	}

	emptySession := &session.Session{}
	if sessionData != emptySession {
		username = sessionData.GetUsername()
	}

	roomName := username
	// Get or create a room
	var room *core.Room

	if username == "Resg" {
		if len(params["room"]) > 0 {
			roomName = params["room"][0]
		}
		room = core.AllRooms[roomName]
	} else {
		room = core.NewRoom(username)
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

	grcpConn, err := grpc.Dial(
		utils.FrontIP,
		grpc.WithPerRPCCredentials(&tokenAuth{"session_token"}),
		grpc.WithInsecure(),
	)

	if err != nil {
		log.Fatalf("cant connect to grpc")
	}

	defer grcpConn.Close()

	sessManager = session.NewAuthCheckerClient(grcpConn)

	e.GET("/getRooms/", func(ctx echo.Context) error {
		getRooms(ctx)
		return nil
	})

	e.GET("/chatRoom/", func(ctx echo.Context) error {
		wsHandler(ctx)
		return nil
	})

	e.Logger.Fatal(e.Start(ListenAddr))
}
