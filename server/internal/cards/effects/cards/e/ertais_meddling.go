package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ertai's Meddling — Instant {X}{U} (Tempest, S82 OLD batch 759, #5628):
//
//	"X can't be 0.
//	 Target spell's controller exiles it with X delay counters on it.
//	 At the beginning of each of that player's upkeeps, if that card is
//	 exiled, remove a delay counter from it. If the card has no delay
//	 counters on it, the player puts it onto the stack as a copy of the
//	 original spell."
//
// The resolution is the #6221 seam, one call on the target spell
// (game.ExileSpellWithDelayCountersForEffect; CR 707.10); the spell is
// exiled, not countered. "X can't be 0" is Spec.XFloor (game/x_floor.go),
// which refuses a cast announcing X=0. A target that has left the stack
// (illegal, CR 608.2b) leaves the Meddling to do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "609c529a-b3b3-46d5-b9d0-ae0117507d6e",
		Name:         "Ertai's Meddling",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		XFloor:       1,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				_, err := ctx.Game.ExileSpellWithDelayCountersForEffect(t.ID, ctx.X(), ctx.Source(), "")
				return err
			}
			return nil
		},
	})
}
