package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b52_helpers.go — shared bodies for Foundations Jumpstart batch 52
// (#5929).

// mayReturnALandYouControl is "you may return a land you control to its owner's
// hand" as a ChoosePermanents pick of none or one land from the chooser's own
// board, made as the effect resolves (Kefnet the Mindful, Sutina, Speaker of
// the Tajuru). `keyed` is what happens to the pick.
func mayReturnALandYouControl(question string, keyed PermanentsThen) ChoosePermanents {
	return ChoosePermanents{
		Question: question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, id := range PermanentsControlledBy(g, of) {
				if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
					out = append(out, id)
				}
			}
			return out, 0, 1
		},
		Keyed: keyed,
	}
}
