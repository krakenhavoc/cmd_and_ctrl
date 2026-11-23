package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the Roil — Instant {1}{U} (Foundations, #4465):
//
//	"Kicker {1}{U} (You may pay an additional {1}{U} as you cast this
//	 spell.)
//	 Return target nonland permanent to its owner's hand. If this spell
//	 was kicked, draw a card."
//
// Whoosh!'s text, so it shares Whoosh!'s resolution
// (bounceTheTargetThenDrawIfKicked): the bounce, with the draw waiting
// behind it as a keyed continuation, gated on the kicker record.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2898bbd-82a4-4d26-b6ee-169b0ebe71b4",
		Name:         "Into the Roil",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OptionalCosts: []game.AdditionalCost{
			effects.Kicker("{1}{U}"),
		},
		OnResolve: effects.BounceTheTargetThenDrawIfKicked,
	})
}
