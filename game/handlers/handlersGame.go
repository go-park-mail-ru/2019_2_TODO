package handlers

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/core"
	repository "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/repositoryLeaders"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

const (
	LEADERSSIZE int = 10
)

type HandlersGame struct {
	Usecase *repository.LeadersRepository
}

type JSONRooms struct {
	Rooms map[string]int `json:"rooms"`
}

type JSONLeaders struct {
	Leaders []*leaderBoardModel.UserLeaderBoard `json:"leaders"`
}

var (
	SessManager session.AuthCheckerClient
)

func (h *HandlersGame) GetRooms(ctx echo.Context) error {
	ws, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(ctx.Response(), "Not a websocket handshake", 400)
		return err
	} else if err != nil {
		return err
	}

	go func() {
		for {
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
			err := ws.WriteJSON(jsonRooms)
			if err != nil {
				ws.Close()
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	return nil
}

func (h *HandlersGame) WsHandler(ctx echo.Context) error {
	ws, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(ctx.Response(), "Not a websocket handshake", 400)
		return err
	} else if err != nil {
		return err
	}

	params, err := url.ParseQuery(ctx.Request().URL.RawQuery)
	if err != nil || !(len(params["session_token"]) > 0) {
		return ctx.JSON(http.StatusInternalServerError, "Smth wrong with parseQuery")
	}
	log.Println(params["session_token"][0])

	cookieSessionID := ReadSessionIDAndUserID(params["session_token"][0])
	if cookieSessionID == nil {
		return ctx.JSON(http.StatusUnauthorized, "Firstly log in")
	}

	userID, err := strconv.Atoi(cookieSessionID[1])
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Smth wrong with cookie")
	}

	user, err := h.Usecase.SelectLeaderByID(int64(userID))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Smth wrong with database")
	}

	playerName := user.Username
	playerStartChips, err := strconv.Atoi(user.Points)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Atoi error")
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

func (h *HandlersGame) LeaderBoardTopHandler(ctx echo.Context) error {
	leaders, err := h.Usecase.ListAllLeaders()
	if err != nil {
		return err
	}

	result := &JSONLeaders{
		Leaders: partitionSort(leaders),
	}

	return ctx.JSON(http.StatusOK, result)
}

func partitionSort(leaders []*leaderBoardModel.UserLeaderBoard) []*leaderBoardModel.UserLeaderBoard {
	var result []*leaderBoardModel.UserLeaderBoard

	var tmp *leaderBoardModel.UserLeaderBoard = &leaderBoardModel.UserLeaderBoard{}
	var maxIndex int

	for i := 0; i < LEADERSSIZE; i++ {
		tmp = leaders[0]
		maxIndex = 0

		for j := 1; j < len(leaders); j++ {
			if leaders[j].Points > tmp.Points {
				tmp = leaders[j]
				maxIndex = j
			}
		}

		result = append(result, tmp)

		leaders[maxIndex] = leaders[len(leaders)-1]
		leaders[len(leaders)-1] = nil
		leaders = leaders[:len(leaders)-1]
	}
	return result
}

func ReadSessionIDAndUserID(cookie string) []string {
	cookieStr := "session_token" + cookie
	value := make(map[string]string)
	if err := utils.CookieHandler.Decode("session_token", cookieStr, &value); err == nil {
		log.Println("Im here")
		var result = []string{}
		result = append(result, value["session_id"])
		result = append(result, value["user_id"])
		return result
	}
	return nil
}
