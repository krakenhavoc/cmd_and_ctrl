package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Retreat to Valakut — Enchantment {2}{R} (Battle for Zendikar):
//
//	"Landfall — Whenever a land you control enters, choose one —
//	 • Target creature gets +2/+0 until end of turn.
//	 • Target creature can't block this turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91108ed7-5b24-4b51-b01d-85005998c69d",
		Name:         "Retreat to Valakut",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.LandEnteredUnderYourControl,
			Key:       "Retreat to Valakut — target creature gets +2/+0, or target creature can't block (landfall)",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseOne(
				effects.ModeDoing("Target creature gets +2/+0 until end of turn.", effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Retreat to Valakut — +2/+0"}.Apply(ctx)
					}),
				effects.ModeDoing("Target creature can't block this turn.", effects.Harmful(effects.TargetCreature("target creature")),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Retreat to Valakut — can't block"}.Apply(ctx)
					}),
			),
		}},
	})
}
