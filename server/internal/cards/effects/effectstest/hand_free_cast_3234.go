package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func OfferFrom(t *testing.T, g *game.Game, me *game.Player, f func(*effects.Context) error) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := f(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("offer: %v", err)
		}
	})
}
