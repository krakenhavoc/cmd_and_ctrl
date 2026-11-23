package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// exileTheFirstLegalTarget is "exile target <thing>" as a trigger or
// ability body: the first chosen target that is still legal at resolution
// is exiled, and nothing happens when none is (CR 608.2b).
func ExileTheFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return ExileTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}
