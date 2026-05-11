// -----------------------------------------------------------------------------
// File: deadlock.go
//
// Desc: Implementacao do deadlock do exercicio 1.
//
// Authors: Grupo F.
// -----------------------------------------------------------------------------
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type state_t uint8

const (
	kThinking 	state_t	= 1
	kEating		state_t	= ( 1 << 1 )
	kHungry		state_t	= ( 1 << 2 )
	kSize		int		= 5
)

type philosopher_t struct {
	uid			uint8
	state 		state_t
	meals		uint32
}

//-----------------------------------------------------------------------------
// Name: Eat()
// Desc: Representa a secao critica. O estado do filosofo eh atualizado para
// 		 kEating ( comendo ). Usamos atomic.AddUint32 para incrementar meals
//		 de forma segura.
//-----------------------------------------------------------------------------
func Eat( philosopher *philosopher_t ) {
	philosopher.state &= ^( kThinking | kHungry )
	philosopher.state |= kEating

	// philosopher.meals++ nao e atomico ( 3 instrucoes por baixo dos panos ) -> pode dar ruim
	fmt.Printf( "Filosofo %d esta comendo\n", philosopher.uid )
	atomic.AddUint32( &philosopher.meals, 1 )
}

//-----------------------------------------------------------------------------
// Name: Think()
// Desc: Simula o estado de pensamento do filosofo.
//-----------------------------------------------------------------------------
func Think( philosopher *philosopher_t ) {
	philosopher.state &= ^( kEating | kHungry )
	philosopher.state |= kThinking
}

//-----------------------------------------------------------------------------
// Name: Hungry()
// Desc: Simula o estado de fome para o filosofo.
//-----------------------------------------------------------------------------
func Hungry( philosopher *philosopher_t ) {
	philosopher.state &= ^( kThinking | kEating )
	philosopher.state |= kHungry
}

//-----------------------------------------------------------------------------
// Name: PickFork()
// Desc: Todos os filosofos pegam o garfo direito primeiro, depois o esquerdo.
//       Isso causa deadlock pois todos ficam esperando o garfo do vizinho,
//       satisfazendo a condicao de espera circular de Coffman.
//-----------------------------------------------------------------------------
func PickFork( idx int, philosophers []*philosopher_t, fork []chan struct{} ) {
	left  := idx
	right := ( idx + 1 ) % kSize

	fmt.Printf( "Filosofo %d tentando pegar garfos\n", philosophers[idx].uid )

	<-fork[ right ]
	<-fork[ left ]

	fmt.Printf( "Filosofo %d pegou os garfos\n", philosophers[idx].uid )
}

//-----------------------------------------------------------------------------
// Name: ReleaseFork()
// Desc: Faz o filosofo soltar o garfo.
//-----------------------------------------------------------------------------
func ReleaseFork( idx int, fork []chan struct{} ) {
	left  := idx
	right := ( idx + 1 ) % kSize

	fork[ left ]  <- struct{}{}
	fork[ right ] <- struct{}{}
}

//-----------------------------------------------------------------------------
// Name: Dine()
// Desc: Logica principal. Se o filosofo estiver com fome, ele pega o garfo.
//		 Depois, ele come. Apos determinado tempo (time.sleep para verificar
//		 data race), ele solta o garfo e volta a pensar por determinado tempo.
// 		 Repete N vezes
//-----------------------------------------------------------------------------
func Dine( idx int, philosophers []*philosopher_t, fork []chan struct{}, wg *sync.WaitGroup ) {
	defer wg.Done()

	for n := 0; n < 100; n++ {
		Hungry( philosophers[idx] )

		if philosophers[ idx ].state & kHungry != 0 {
			PickFork( idx, philosophers, fork )
		}

		Eat( philosophers[idx] )

		if philosophers[ idx ].state & kEating != 0 {
			time.Sleep( time.Microsecond * 10 )
			ReleaseFork( idx, fork )
		}

		Think( philosophers[idx] )
	}
}

//-----------------------------------------------------------------------------
// Name: main()
// Desc: Ponto de inicio. Inicializa os filosofos (ponteiro que aponta pra uma
//		 lista de ponteiros de filosofos), depois inicializa os garfos (aqui eh
//		 um canal com tamanho um, simulando um mutex). A gente usa waitGroup
//		 para garantir que o main nao acabe antes de todo programa rodar.
//-----------------------------------------------------------------------------
func main() {
	var philosophers = 	make( []*philosopher_t, kSize )
	var forks		 = 	make( []chan struct{} , kSize )
	var wg			 	sync.WaitGroup

	for idx := range kSize {
		philosophers[ idx ] = &philosopher_t{ uid: uint8( idx ) }
		forks[ idx ] = make( chan struct{}, 1 )
		forks[ idx ] <- struct{}{}
	}

	for idx := range kSize {
		wg.Add( 1 )
		go Dine( idx, philosophers, forks, &wg )
	}
	
	wg.Wait()

	for idx := range kSize {
		fmt.Printf( "Filosofo %d comeu %d vezes\n", philosophers[idx].uid, philosophers[idx].meals )
	}
}