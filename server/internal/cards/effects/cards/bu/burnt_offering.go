package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burnt Offering — Instant {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Add X mana in any combination of {B} and/or {R}, where X is the
//	 sacrificed creature's mana value."
//
// X is the sacrificed creature's mana value as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1).
// Each of the X mana is its own {B}-or-{R} pick, so any split is
// possible; a token that is no copy adds nothing.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "86eb30a0-0beb-42db-9ddf-cf8be6c99dd3",
		Name:           "Burnt Offering",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.SacrificedManaValue(ctx)
			if x <= 0 {
				return nil
			}
			return effects.AddMana{Amount: effects.ManaN(x, "B", "R")}.Apply(ctx)
		},
	})
}
