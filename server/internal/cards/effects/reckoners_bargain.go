package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckoner's Bargain — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature. You gain life equal to the sacrificed permanent's mana
//	 value. Draw two cards."
//
// The life is the sacrificed permanent's mana value as it last existed
// on the battlefield (CR 608.2h), read off the payment record (ADR 0113
// §1); a token that is no copy is 0. Then the two cards, as printed.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "044ea111-ce87-49ab-98f7-ae8447f48ac5",
		Name:           "Reckoner's Bargain",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GainLife{Amount: sacrificedManaValue(ctx)}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 2}.Apply(ctx)
		},
	})
}
