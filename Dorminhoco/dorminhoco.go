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

type player_t struct {
	uid int
}

func Play( wg *sync.WaitGroup) {
	defer wg.Done()
}

func main() {
	var wg		sync.WaitGroup

	Play( &wg )

	fmt.Printf( "Fernando A" )
}
