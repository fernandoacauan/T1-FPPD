// -----------------------------------------------------------------------------
// File: solution2.go
//
// Desc: Implementacao da segunda solucao do exercicio 1, usando monitores.
//
// Authors: Grupo F.
// -----------------------------------------------------------------------------
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
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

type monitor_t struct {
	forks		[]chan struct{}
	mu			sync.Mutex
	cond		*sync.Cond
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
// Desc: Tenta adquirir os dois garfos via select. Se o garfo direito nao esti
//		 ver disponivel, devolve o esquerdo ja adquirido e dorme via cond.Wait()
//		 ate ser acordado pelo Broadcast do ReleaseFork. A condicao de posse e
//		 espera eh quebrada.
//-----------------------------------------------------------------------------
func PickFork( idx int, philosopher *philosopher_t, monitor *monitor_t ) {
	left  := idx
	right := ( idx + 1 ) % kSize

	monitor.mu.Lock()

	fmt.Printf( "Filosofo %d tentando pegar os garfos\n", philosopher.uid )

	for {
		select {
			case <-monitor.forks[ left ]: {
				select {
					case <-monitor.forks[ right ]: {
						monitor.mu.Unlock()
						fmt.Printf( "Filosofo %d pegou os garfos\n", philosopher.uid )
						return
					}
					default: {
						monitor.forks[ left ] <- struct{}{} // devolve o esquerdo
						monitor.cond.Wait()
					}
				}
			}
			default: {
				monitor.cond.Wait()
			}
		}
	}
}

//-----------------------------------------------------------------------------
// Name: ReleaseFork()
// Desc: Devolve os dois garfos aos channels e acorda todos os filosofos que
//		 estao dormindo via Broadcast, permitindo que tentem de novo.
//-----------------------------------------------------------------------------
func ReleaseFork( idx int, monitor *monitor_t ) {
	left  := idx
	right := ( idx + 1 ) % kSize

	monitor.mu.Lock()

	monitor.forks[ left ]  <- struct{}{}
	monitor.forks[ right ] <- struct{}{}

	monitor.cond.Broadcast()
	monitor.mu.Unlock()
}

//-----------------------------------------------------------------------------
// Name: Dine()
// Desc: Logica principal. O ciclo de vida do filosofo: fica com fome, tenta pe
//		 gar os garfos via monitor, come, solta os garfos e volta a pensar.
//-----------------------------------------------------------------------------
func Dine( idx int, philosophers []*philosopher_t, monitor *monitor_t, wg *sync.WaitGroup ) {
	defer wg.Done()

	for n := 0; n < 5; n++ {
		Hungry( philosophers[idx] )

		if philosophers[ idx ].state & kHungry != 0 {
			PickFork( idx, philosophers[idx], monitor )
		}

		Eat( philosophers[idx] )

		if philosophers[ idx ].state & kEating != 0 {
			ReleaseFork( idx, monitor )
		}

		Think( philosophers[idx] )
	}
}

//-----------------------------------------------------------------------------
// Name: main()
// Desc: Ponto de inicio. Inicializa os filosofos e o monitor. O monitor encap
//		 sula os garfos como channels de buffer 1, um mutex e uma variavel de
//		 condicao. Cada filosofo roda como uma goroutine independente.
//-----------------------------------------------------------------------------
func main() {
	var philosophers 	= 	make( []*philosopher_t, kSize )
	var monitor			= 	&monitor_t{ forks: make( []chan struct{}, kSize ) }
	var wg					sync.WaitGroup
	monitor.cond 		= 	sync.NewCond( &monitor.mu )

	for idx := range kSize {
		philosophers[ idx ]		= &philosopher_t{ uid: uint8( idx ) }
		monitor.forks[ idx ]	= make( chan struct{}, 1 )
		monitor.forks[ idx ] 	<- struct{}{}
	}

	for idx := range kSize {
		wg.Add( 1 )
		go Dine( idx, philosophers, monitor, &wg )
	}

	wg.Wait()

	for idx := range kSize {
		fmt.Printf( "Filosofo %d comeu %d vezes\n", philosophers[idx].uid, philosophers[idx].meals )
	}
}
