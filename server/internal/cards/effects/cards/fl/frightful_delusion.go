package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frightful Delusion — Instant {2}{U} (Innistrad, S82 OLD batch 450,
// #5572):
//
//	"Counter target spell unless its controller pays {1}. That player
//	 discards a card."
//
// Mindswipe's two steps: the pay-or-counter prompt settles first, then
// the spell's controller (read as last known, CR 608.2h) discards a card
// of their choice whether or not they paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13ca3ede-a4c8-463d-b160-99b4071034b0",
		Name:         "Frightful Delusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  targets[0].ID,
					Cost:     "{1}",
					Question: "Frightful Delusion — pay {1} or your spell is countered",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if len(ctx.Item.Targets) == 0 {
					return nil
				}
				controller, ok := ctx.Game.LastKnownSpellControllerForEffect(ctx.Item.Targets[0].ID)
				if !ok || controller == uuid.Nil {
					return nil
				}
				ctx.Game.DiscardChoiceForEffect(controller, 1)
				return nil
			}),
		),
	})
}
