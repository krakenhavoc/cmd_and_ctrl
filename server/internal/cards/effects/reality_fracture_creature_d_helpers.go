package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_d_helpers.go — helpers for the Reality
// Fracture creatures of slice fra-creature-d.

// rfCreatureDTapped is the predicate "tapped" (Ob Nixilis, the
// Ascended's "all tapped creatures your opponents control").
func rfCreatureDTapped() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped }
}
