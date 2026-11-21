package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// It'll Quench Ya! — Instant — Lesson {1}{U}:
//
//	"Counter target spell unless its controller pays {2}."
//
// Spell Pierce's pay-unless with no restriction on the spell: the
// spell's controller is asked, and the spell is countered unless they
// pay (CR 118.12a).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb184995-700f-4900-994a-4ce7fd12f942",
		Name:         "It'll Quench Ya!",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{2}",
					Question: "It'll Quench Ya! — pay {2} or your spell is countered",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
