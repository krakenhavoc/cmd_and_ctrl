package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b145_helpers.go — shared bodies for S82 OLD batch 145 (#5516).

// OldB145BounceTheTargetPermanent is "Return target nonland permanent to its
// owner's hand" for a spell's first legal target that is still on the
// battlefield; a target that has gone is left alone (CR 608.2b).
func OldB145BounceTheTargetPermanent(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !OnBattlefield(ctx.Game, t.ID) {
			continue
		}
		return BounceToHand{Target: t.ID}.Apply(ctx)
	}
	return nil
}
