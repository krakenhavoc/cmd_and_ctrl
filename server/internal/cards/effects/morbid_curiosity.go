package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "59c620fb-a58d-4038-a7e2-0185324b5123",
		Name:           "Morbid Curiosity",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{N: sacrificedManaValue(ctx)}.Apply(ctx)
		},
	})
}
