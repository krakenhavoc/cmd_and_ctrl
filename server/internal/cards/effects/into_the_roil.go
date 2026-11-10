package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c2898bbd-82a4-4d26-b6ee-169b0ebe71b4",
		Name:         "Into the Roil",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OptionalCosts: []game.AdditionalCost{
			Kicker("{1}{U}"),
		},
		OnResolve: bounceTheTargetThenDrawIfKicked,
	})
}
