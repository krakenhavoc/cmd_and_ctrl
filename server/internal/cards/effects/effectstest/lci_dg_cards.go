package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lciDGMillOne mills one card of the given type line from p's library, a
// descent when it is a permanent card (CR 700.11).
func LciDGMillOne(t *testing.T, g *game.Game, p *game.Player, typeLine string) {
	t.Helper()
	SsCard(p.Library, p.ID, "Milled Card", typeLine, 2, 2)
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(p.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
}
