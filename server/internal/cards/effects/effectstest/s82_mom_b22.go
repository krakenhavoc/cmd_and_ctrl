package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushLibraryTyped puts a card on top of a library.
func PushLibraryTyped(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func B22SetLore(g *game.Game, saga *game.Card, n int) {
	g.WithWriteLock(func() {
		if saga.Counters == nil {
			saga.Counters = map[string]int{}
		}
		saga.Counters[game.CounterLore] = n
	})
}
