package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// target_player_drain.go — "Target player loses N life and you gain N
// life", the body of Tendrils of Agony and Syphon Life. Named once so the
// next drain spell reuses it instead of copying nine lines (the clone
// gate, TestNoNewExactClonesInTheCatalog).

// targetPlayerLosesLifeYouGain is the OnResolve of a "target player loses
// N life and you gain N life" spell. The target is the spell's only
// target; one that is no longer a player (it left the game in response)
// does nothing, and the caster gains nothing either — the sentence is one
// instruction about that player (CR 608.2b).
func targetPlayerLosesLifeYouGain(n int) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
			return nil
		}
		if err := (GainLife{Player: item.Targets[0].ID, Amount: -n}).Apply(ctx); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: n}.Apply(ctx)
	}
}
