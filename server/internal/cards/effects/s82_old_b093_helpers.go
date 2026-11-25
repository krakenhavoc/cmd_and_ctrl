package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB093OptionalLandYouControl is the candidate list of "you may return a land
// you control to its owner's hand" (Tazeem Raptor): every land `of` controls,
// with a floor of zero and a ceiling of one.
func OldB093OptionalLandYouControl(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out, 0, 1
}
