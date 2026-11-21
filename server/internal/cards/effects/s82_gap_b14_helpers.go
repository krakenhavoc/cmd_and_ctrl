package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapTwoDamageToChosenTarget is "this <permanent> deals 2 damage to any
// target" as a trigger body: the first legal target the pick stamped into the
// item (CR 608.2b), damaged by the source (last-known if it has left). Shared
// by Burning Vengeance and Ashmouth Dragon.
func GapTwoDamageToChosenTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
	}
	return nil
}
