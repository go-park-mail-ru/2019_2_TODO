package handlers

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"

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
	for i := 0; i < LEADERSSIZE; i++ {
		tmp = leaders[0]
		var j int = 1
		for j = 1; j < len(leaders); j++ {
			if leaders[j].Points > tmp.Points {
				tmp = leaders[j]
			}
		}
		result = append(result, tmp)
		leaders[j] = leaders[len(leaders)-1]
		leaders[len(leaders)-1] = nil
		leaders = leaders[:len(leaders)-1]
	}
	return result
}
