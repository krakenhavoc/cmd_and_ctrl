package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fear_of_abduction_exile.go — Fear of Abduction's linked exile, shared with
// the Helvault.

// FearOfAbductionExile exiles the target creature, linked to the Fear.
func FearOfAbductionExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileTarget{Target: t.ID, Linked: true}.Apply(ctx)
	}
	return nil
}
