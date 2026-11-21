package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tend the Pests — Instant {B}{G}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Create X 1/1 black and green Pest creature tokens with "When this
//	 token dies, you gain 1 life," where X is the sacrificed creature's
//	 power."
//
// X is the sacrificed creature's power as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1).
// A copy makes the ORIGINAL's number of Pests: it uses the creature
// sacrificed for the original spell, and sacrifices nothing of its own
// (the 2021-04-16 ruling, CR 707.10). A negative power makes none
// (CR 107.1b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "9ffce1d5-e29c-4dab-8b97-6b32655785d4",
		Name:           "Tend the Pests",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.PestToken(), N: ctx.SacrificedPower()}.Apply(ctx)
		},
	})
}
