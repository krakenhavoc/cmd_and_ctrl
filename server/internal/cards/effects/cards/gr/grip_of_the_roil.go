package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grip of the Roil — Instant {2}{U} (Oath of the Gatewatch, #5545):
//
//	"Surge {1}{U} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Tap target creature. It doesn't untap during its controller's next
//	 untap step.
//	 Draw a card."
//
// Rooftop Nuisance's tap-and-freeze, then the caster draws. With its only
// target illegal the spell is countered on resolution and draws nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c3c87292-2fc8-4759-b7b6-18a8b539b5d1",
		Name:             "Grip of the Roil",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{1}{U}")},
		Targets:          effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Grip of the Roil — doesn't untap during its controller's next untap step"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
