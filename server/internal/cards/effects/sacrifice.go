package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sacrifice — Instant {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Add an amount of {B} equal to the sacrificed creature's mana value."
//
// The sacrificed creature's mana value as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1). An
// animated land or a token that is no copy has mana value 0 and adds
// nothing (the 2004-10-04 ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "068b3692-411b-44d4-a7e9-005262760cfc",
		Name:           "Sacrifice",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := sacrificedManaValue(ctx)
			if x <= 0 {
				return nil
			}
			return AddMana{Produced: strings.Repeat("{B}", x)}.Apply(ctx)
		},
	})
}
