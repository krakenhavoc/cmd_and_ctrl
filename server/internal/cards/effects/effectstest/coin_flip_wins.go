package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emitCoinFlip fakes one coin's EventFlipCoin, which is all a "whenever
// you win a coin flip" trigger watches, and runs the priority boundary
// that puts a trigger it caused on the stack, as a real flip's
// resolution does.
func EmitCoinFlip(g *game.Game, flipper uuid.UUID, won bool) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventFlipCoin, Actor: flipper, Label: "heads", Call: "heads", Won: won})
	})
	g.RunStateChecksForTest()
}

// battlefieldTokenNamed returns the first token with this name on the
// battlefield.
func BattlefieldTokenNamed(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if effects.IsToken(c) && c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s token on the battlefield", name)
	return game.Card{}
}
