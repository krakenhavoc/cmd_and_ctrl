package effects

import (
	"strings"

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
	Register(Spec{
		OracleID:       "86eb30a0-0beb-42db-9ddf-cf8be6c99dd3",
		Name:           "Burnt Offering",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := sacrificedManaValue(ctx)
			if x <= 0 {
				return nil
			}
			return AddMana{Produced: strings.Repeat("{B|R}", x)}.Apply(ctx)
		},
	})
}
