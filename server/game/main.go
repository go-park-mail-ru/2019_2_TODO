package main

import (
	"fmt"
	game "server/game/hand"
)

func main() {
	deck := game.NewDeck()
	player1 := deck.Draw(2)
	player2 := deck.Draw(2)
	table := deck.Draw(5)
	fmt.Println(table)
	fmt.Println(player1)
	fmt.Println(player2)

	var firstHand []game.Card
	var secondHand []game.Card
	firstHand = append(player1, table...)
	secondHand = append(player2, table...)

	fmt.Println(firstHand)
	fmt.Println(secondHand)

	rank1 := game.Evaluate(firstHand)
	rank2 := game.Evaluate(secondHand)
	if rank1 < rank2 {
		fmt.Println(game.RankString(rank1))
		fmt.Println("First player wins!")
	} else {
		fmt.Println(game.RankString(rank2))
		fmt.Println("Second player wins!")
	}
}
