package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morbid Curiosity — Sorcery {1}{B}{B}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature. Draw cards equal to the mana value of the sacrificed
//	 permanent."
//
// "The mana value of the permanent immediately before you sacrificed
// it" (the 2016-09-20 ruling): its last-known mana value (CR 608.2h),
// read off the payment record (ADR 0113 §1). An {X} in its cost is 0, and
// a token that is no copy has none. A copy draws the original's number
// and sacrifices nothing (the same ruling, CR 707.10).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "59c620fb-a58d-4038-a7e2-0185324b5123",
		Name:           "Morbid Curiosity",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{N: effects.SacrificedManaValue(ctx)}.Apply(ctx)
		},
	})
}
