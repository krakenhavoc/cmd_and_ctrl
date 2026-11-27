package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b557_helpers.go — S82 set OLD, batch 557 (Lorwyn).

// CounterFirstTargetAndExile is "counter target spell. If that spell is
// countered this way, exile it instead of putting it into its owner's
// graveyard" (Horribly Awry, Faerie Trickery): the first legal target is
// countered and goes to exile, and nothing happens when it is no longer legal.
func CounterFirstTargetAndExile(_ *game.StackItem, ctx *Context) error {
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return CounterTarget{StackID: targets[0].ID, Dest: game.ZoneRef{Kind: game.ZoneExile}}.Apply(ctx)
}
