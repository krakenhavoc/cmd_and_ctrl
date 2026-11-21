package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// waryFarmerTokenFor creates a 1/1 Shapeshifter creature token under
// `who`, an entry the turn's tally records.
func WaryFarmerTokenFor(g *game.Game, who uuid.UUID) {
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(who, effects.TokenCard("1/1 colorless Shapeshifter with changeling"), 1)
	})
}
