// -----------------------------------------------------------------------------
// File: dorminhoco.go
//
// Desc: Implementacao do exercicio 2, o jogo do dorminhoco.
//
// Authors: Grupo F.
// -----------------------------------------------------------------------------
package main

import (
	"fmt"
	"sync"
)

const (
	kPlayers int 	= 5
	kHand	 int	= 3
)

type player_t struct {
	uid 		int
	hand		[]int
	sendCard 	chan<- int
	receiveCard	<-chan int
}

type game_t struct {
	knock		chan int
	react		chan int
	stop		chan struct{}
}

func NewGame() *game_t {
	return &game_t {
		knock: make( chan int, 1 ),
		react: make( chan int, kPlayers ),
		stop: make( chan struct{} ),
	}
}

func Play( players *player_t, game *game_t, wg *sync.WaitGroup ) {
	defer wg.Done()

	// TODO: toda logica ferrada
}

func SysPrintf( players... *player_t ) {
	for player := range players {
		fmt.Printf( "Jogador %d ", players[player].uid )
	}
}

func InitPlayer( player **player_t, idx int, aHand []int, send chan<- int, receive <-chan int ) {
	*player = &player_t {
		uid: idx,
		hand: aHand,
		sendCard: send,
		receiveCard: receive,
	}
}

func main() {
	var wg			  sync.WaitGroup
	var deck		= make( []int, 0, kPlayers * kHand )
	var channels 	= make( []chan int, kPlayers )
	var players 	= make( []*player_t, kPlayers )
	var game		= NewGame()

	// TODO: montar o deck

	for idx := 0; idx < kPlayers; idx++ {
		hand := deck[ :kHand ]
		deck =	deck[ kHand:]

		InitPlayer( &players[idx], idx, hand, channels[(idx + 1) % kPlayers], channels[idx] )
		
		wg.Add( 1 )
		go Play( players[idx], game, &wg )
	}

	SysPrintf( players... )

	wg.Wait()

	// TODO: logica de resultado
}
