package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quench — Instant {1}{U} (Ravnica Allegiance, #5521):
//
//	"Counter target spell unless its controller pays {2}."
//
// It'll Quench Ya!'s counter: the spell's controller is asked to pay {2}
// as it resolves, and the spell is countered if they don't (CR 118.12).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa1b722a-109e-4a43-bd7b-818292728cb3",
		Name:         "Quench",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{2}",
					Question: "Quench — pay {2} or your spell is countered",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
