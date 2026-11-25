package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b283_helpers.go — shared pieces of S82 older-set batch 283
// (Shadows over Innistrad: delirium, #5543).

// S82OldB283MinusOneOnFirstTarget is "target creature gets -1/-1 until end
// of turn" as a trigger body: the first chosen target still legal as the
// trigger resolves gets the boost (CR 608.2b), and nothing happens when
// none is. `name` is the source card, for the effect's label.
func S82OldB283MinusOneOnFirstTarget(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return BoostUntilEOT{Target: t.ID, Power: -1, Toughness: -1, Label: name + " — -1/-1"}.Apply(ctx)
		}
		return nil
	}
}
