package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitsune's Technique — Instant {4}{U}{U} (TMT, #4457):
//
//	"Sneak {1}{U} (You may cast this spell for {1}{U} if you also return
//	 an unblocked attacker you control to hand during the declare
//	 blockers step.)
//	 Target opponent mills half their library, rounded up."
//
// Sneak is Sneak(cost) (CR 702.190). Traumatize's body with the other
// rounding (CR 107.1a): the half is of the library as the spell
// resolves, and an opponent who left the game in response makes the
// spell do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3e0f8ab3-75a6-49a2-a228-fa720cbde3d8",
		Name:             "Kitsune's Technique",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{1}{U}")},
		Targets:          effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if p := ctx.PlayerByID(t.ID); t.Kind == game.TargetPlayer && p != nil && p.Library != nil {
					return effects.MillCards{Player: t.ID, N: (p.Library.Size() + 1) / 2}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
