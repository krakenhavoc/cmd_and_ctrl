package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// domainLands gives `seat` one land per named basic type, plus a second
// Plains (counted once) and a nonbasic Cave (no type), and an opponent's
// Swamp that must not count.
func DomainLands(g *game.Game, seat, other *game.Player, types ...string) {
	for _, ty := range types {
		PushPermanent(g, seat.ID, game.Card{Name: ty, TypeLine: "Basic Land — " + ty})
	}
	PushPermanent(g, seat.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	PushPermanent(g, seat.ID, game.Card{Name: "Cave", TypeLine: "Land — Cave"})
	PushPermanent(g, other.ID, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
}
