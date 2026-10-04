package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Worthy Cause — Instant {W}:
//
//	"Buyback {2}
//	 As an additional cost to cast this spell, sacrifice a creature.
//	 You gain life equal to the sacrificed creature's toughness."
//
// The life is the sacrificed creature's toughness as it last existed on
// the battlefield (CR 608.2h), read off the payment record (ADR 0113
// §1); zero or negative toughness gains nothing (CR 107.1b). Buyback is
// the optional cost; the sacrifice is owed either way.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "b24063f7-157f-47c0-919a-33856d53b902",
		Name:           "Worthy Cause",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OptionalCosts:  []game.AdditionalCost{Buyback("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GainLife{Amount: sacrificedToughness(ctx)}.Apply(ctx)
		},
	})
}
