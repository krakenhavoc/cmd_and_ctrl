package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icy Reception — Instant {1}{U}:
//
//	"Choose one —
//	 • Counter target creature or legendary spell unless its controller
//	   pays {3}.
//	 • Target creature gets -5/-0 until end of turn."
//
// Mode one asks its spell's controller the CR 118.12 question through
// CounterUnlessPaid, which stops the table while it is open. A
// legendary noncreature spell (a legendary artifact or enchantment) is
// a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "612c7807-be70-4e2f-afd2-3d3c03eddbbe",
		Name:         "Icy Reception",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Counter target creature or legendary spell unless its controller pays {3}.",
				effects.TargetSpell("target creature or legendary spell", effects.Or(effects.Creature(), effects.Legendary()))),
			effects.ModeWithPurpose(effects.Mode("Target creature gets -5/-0 until end of turn.",
				effects.TargetCreature("target creature")), effects.ForTargets(effects.HarmUntilEOT(0, -5, 0))),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				t, ok := effects.ModeTarget(ctx, 0)
				if !ok {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{3}",
					Question: "Icy Reception — pay {3} or your spell is countered?",
				}.Apply(ctx)
			case ctx.HasMode(1):
				t, ok := effects.ModeTarget(ctx, 0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: -5, Label: "Icy Reception — -5/-0"}.Apply(ctx)
			}
			return nil
		},
	})
}
