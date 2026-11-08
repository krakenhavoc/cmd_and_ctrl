package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whoosh! — Instant {1}{U} (Marvel Super Heroes Commander, #4455):
//
//	"Kicker {1}{U}
//	 Return target nonland permanent to its owner's hand. If this spell was
//	 kicked, draw a card."
//
// Boomerang Basics' bounce with the draw waiting behind it as the bounce's
// continuation (#993), gated on the kicker record (Burst Lightning's
// ctx.WasKicked) instead of on who controlled the permanent:
// bounceTheTargetThenDrawIfKicked, which Into the Roil shares.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "855353e2-b686-482a-b987-c6ae456f088c",
		Name:         "Whoosh!",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OptionalCosts: []game.AdditionalCost{
			Kicker("{1}{U}"),
		},
		OnResolve: bounceTheTargetThenDrawIfKicked,
	})
}
