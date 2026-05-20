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
	"math/rand"
	"sync"
	"time"
)

type Carta int

const (
	NumJogadores = 4
	CartasPorMao = 3
)

// =============================================================================
// LOG SEGURO (thread-safe)
// =============================================================================

var mu sync.Mutex

func logf(formato string, args ...any) {
	mu.Lock()
	fmt.Printf(formato, args...)
	mu.Unlock()
}

func novoBaralho() []Carta {
	var b []Carta
	for v := 1; v <= 13; v++ {
		for k := 0; k < 4; k++ {
			b = append(b, Carta(v))
		}
	}
	return b
}

func embaralhar(b []Carta) {
	rand.Shuffle(len(b), func(i, j int) { b[i], b[j] = b[j], b[i] })
}

func temTrinca(mao []Carta) bool {
	return mao[0] == mao[1] && mao[1] == mao[2]
}

func escolherDescarte(mao []Carta) Carta {
	freq := map[Carta]int{}
	for _, c := range mao {
		freq[c]++
	}
	var pior Carta
	min := 999
	for _, c := range mao {
		if freq[c] < min || (freq[c] == min && c < pior) {
			min, pior = freq[c], c
		}
	}
	return pior
}

func removerCarta(mao []Carta, alvo Carta) []Carta {
	for i, c := range mao {
		if c == alvo {
			mao[i] = mao[len(mao)-1]
			return mao[:len(mao)-1]
		}
	}
	return mao
}

func maoStr(mao []Carta) string {
	s := "["
	for i, c := range mao {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%2d", int(c))
	}
	return s + "]"
}

var batidaCh = make(chan int) // so 1 player consegue enviar

// =============================================================================
// JOGADOR  (cada chamada vira uma goroutine independente)
// =============================================================================

func jogador(id int, mao []Carta, recebe <-chan Carta, envia chan<- Carta, reagir <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// --- checagem inicial (raramente ocorre, mas possível) -----------
	if temTrinca(mao) {
		logf("Jogador %d BATEU de inicio com trinca de %d!\n", id+1, int(mao[0]))
		select {
			case batidaCh <- id: // avisa o árbitro
			case <-reagir:       // outro já bateu antes
		}
		return
	}

	for {
		descarte := escolherDescarte(mao)
		mao = removerCarta(mao, descarte)
		logf("Jogador %d descarta %2d\n", id+1, int(descarte))

		select {
			case envia <- descarte:
			case <-reagir:
				return
		}

		var nova Carta
		select {
			case nova = <-recebe:
			case <-reagir:
				return
		}

		mao = append(mao, nova)
		logf("Jogador %d recebeu %2d | mao: %v\n", id+1, int(nova), maoStr(mao))

		if temTrinca(mao) {
			logf("Jogador %d BATEU com trinca de %d!\n", id+1, int(mao[0]))
			select {
				case batidaCh <- id: // tenta ser o primeiro a bater
				case <-reagir:       // outro já bateu, encerra limpo
			}
			return
		}

	} // fim do loop principal
}

// =============================================================================
// ÁRBITRO  (goroutine separada)
// =============================================================================

func arbitro(reagirChs []chan struct{}) {
	vencedor := <-batidaCh
	logf("\nJogador %d bateu! Aguardando reacoes...\n", vencedor+1)

	responderam := make(chan int, NumJogadores) // em ordem

	// envia sinal de fim para TODOS em goroutines paralelas
	for i, ch := range reagirChs {
		go func(jid int, c chan struct{}) {
			c <- struct{}{}    // sinal de encerramento
			responderam <- jid // registra quem reagiu
		}(i, ch)
	}

	var ultimo int
	for k := 0; k < NumJogadores; k++ {
		ultimo = <-responderam
		logf("   Jogador %d reagiu (posicao %d)\n", ultimo+1, k+1)
	}

	logf("\nJogador %d foi o ULTIMO a reagir — perdeu!\n", ultimo+1)
	logf("Jogador %d venceu a rodada!\n\n", vencedor+1)
}

func main() {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck

	baralho := novoBaralho()
	embaralhar(baralho)

	maos := make([][]Carta, NumJogadores)
	for i := range maos {
		ini := i * CartasPorMao
		maos[i] = append([]Carta{}, baralho[ini:ini+CartasPorMao]...)
	}

	fmt.Println("=== JOGO DE TRINCA ===")
	for i, m := range maos {
		logf("Jogador %d comeca com: %v\n", i+1, maoStr(m))
	}
	fmt.Println("----------------------")

	// channel de cartas (o buffer=1 quebra o deadlock circular do anel)
	channels := make([]chan Carta, NumJogadores)
	for i := range channels {
		channels[i] = make(chan Carta, 1)
	}

	// channels de reagir (unbuffered - árbitro envia via goroutine)
	reagirChs := make([]chan struct{}, NumJogadores)
	for i := range reagirChs {
		reagirChs[i] = make(chan struct{})
	}
	
	var wg sync.WaitGroup

	// =================================================================
	// LANÇA GOROUTINES DOS JOGADORES
	// =================================================================
	for i := 0; i < NumJogadores; i++ {
		wg.Add(1)

		// anel: recebe do vizinho da direita, envia para a esquerda
		recebe := channels[(i+NumJogadores-1)%NumJogadores]
		envia := channels[i]

		go jogador(i, maos[i], recebe, envia, reagirChs[i], &wg)
	}

	go arbitro(reagirChs)
	wg.Wait()

	fmt.Println("=== FIM DO JOGO ===")
}
