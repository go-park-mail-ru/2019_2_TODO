package core

import "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/hand"

type Game struct {
	Players             []*playerConn
	TableCards          []hand.Card
	Bank                int
	Dealer              int
	MinBet              int
	PlayerCounter       int
	MaxBet              int
	StageCounter        int
	PositionToNextStage int
}

func (game *Game) StartGame() {
	game.DealerChange()
	game.PlayerCounter = game.Dealer
	deck := hand.NewDeck()

	for _, player := range game.Players {
		player.Hand = deck.Draw(2)
		player.sendState("showPlayerCards")
	}

	game.TableCards = deck.Draw(5)
	game.SetBlind()
	game.PositionToNextStage = game.PlayerCounter
}

func (game *Game) DealerChange() {
	game.Dealer = (game.Dealer + 1) % len(game.Players)
}

func (game *Game) StageCounterChange() {
	game.MaxBet = 0
	game.PlayerCounter = game.Dealer
	game.StageCounter = (game.StageCounter + 1) % 5
}

func (game *Game) PlayerCounterChange() {
	game.PlayerCounter = (game.PlayerCounter + 1) % len(game.Players)
}

func (game *Game) SetBlind() {
	game.Players[game.PlayerCounter].Bet = game.MinBet
	game.Players[game.PlayerCounter].Chips -= game.MinBet
	for _, player := range game.Players {
		player.sendNewPlayer(game.Players[game.PlayerCounter], "updatePlayerScore")
	}
	game.PlayerCounterChange()
	game.Players[game.PlayerCounter].Bet = game.MinBet * 2
	game.Players[game.PlayerCounter].Chips -= game.MinBet * 2
	for _, player := range game.Players {
		player.sendNewPlayer(game.Players[game.PlayerCounter], "updatePlayerScore")
	}
	game.PlayerCounterChange()
}
