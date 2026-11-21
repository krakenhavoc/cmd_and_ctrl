package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blbMPFillGraveyard puts n plain cards into p's graveyard, then nudges
// the battlefield so the layer cache is refreshed (a direct graveyard
// push is no game action).
func BlbMPFillGraveyard(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		PushGraveyardCardTyped(p, "Old Card", "Sorcery")
	}
	BlbB1Push(g, p.ID, "Nudge", "Creature — Bear", "", 1, 1)
}
