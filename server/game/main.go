package main

// import (
// 	"fmt"
// 	game "server/game/hand"
// )

// func main() {
// 	deck := game.NewDeck()
// 	player1 := deck.Draw(2)
// 	player2 := deck.Draw(2)
// 	table := deck.Draw(5)
// 	fmt.Println(table)
// 	fmt.Println(player1)
// 	fmt.Println(player2)

// 	var firstHand []game.Card
// 	var secondHand []game.Card
// 	firstHand = append(player1, table...)
// 	secondHand = append(player2, table...)

// 	fmt.Println(firstHand)
// 	fmt.Println(secondHand)

// 	rank1 := game.Evaluate(firstHand)
// 	rank2 := game.Evaluate(secondHand)
// 	if rank1 < rank2 {
// 		fmt.Println(game.RankString(rank1))
// 		fmt.Println("First player wins!")
// 	} else {
// 		fmt.Println(game.RankString(rank2))
// 		fmt.Println("Second player wins!")
// 	}
// }

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"server/game/core"

	"github.com/gorilla/websocket"
)

const (
	ADDR string = ":8080"
)

func homeHandler(c http.ResponseWriter, r *http.Request) {
	var homeTempl = template.Must(template.ParseFiles("templates/home.html"))
	data := struct {
		Host       string
		RoomsCount int
	}{r.Host, core.RoomsCount}
	homeTempl.Execute(c, data)
}

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
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)

	if err := http.ListenAndServe(ADDR, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
