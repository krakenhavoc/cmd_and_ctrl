package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reality Shift — Instant {1}{U}:
//
//	"Exile target creature. Its controller manifests the top card of
//	 their library. (That player puts the top card of their library
//	 onto the battlefield face down as a 2/2 creature. If it's a
//	 creature card, it can be turned face up any time for its mana
//	 cost.)"
//
// A removal spell that leaves a body behind, on purpose: the target's
// CONTROLLER — not the caster — manifests, so the exiled creature's
// owner is compensated with a blank 2/2 rather than the caster
// getting to keep the value. The controller is read off the target
// BEFORE the exile (a card that has left the battlefield has no
// controller to ask), and the manifest always happens once the target
// was legal at resolution: nothing in the printed text conditions it
// on "if you do".
//
// No simplification. Manifest is CR 701.34a, carried since ADR 0069;
// this is the first catalog card to call it off a target rather than
// off a fixed number of library cards.
func init() {
	Register(Spec{
		OracleID:     "70dc830e-d05b-4fc7-88dd-879e140b3fbf",
		Name:         "Reality Shift",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			targetID := item.Targets[0].ID
			c, ok := ctx.Game.LookupCardForEffect(targetID)
			if !ok {
				return nil
			}
			controller := c.Controller
			if err := (ExileTarget{Target: targetID}).Apply(ctx); err != nil {
				return err
			}
			return Manifest{Player: controller, N: 1}.Apply(ctx)
		},
	})
}
