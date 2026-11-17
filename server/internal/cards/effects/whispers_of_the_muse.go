package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whispers of the Muse — Instant {U} (Tempest, #5628):
//
//	"Buyback {5} (You may pay an additional {5} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Draw a card."
//
// Capsize's buyback (the shared Buyback constructor, CR 702.27a) on a
// one-card draw. The spell itself only draws; the engine puts it back
// into its owner's hand as it resolves when the buyback was paid, and
// into the graveyard otherwise.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a4a8af4e-8686-45ba-94fd-ccf6cc8adaae",
		Name:          "Whispers of the Muse",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Draws: 1},
		OptionalCosts: []game.AdditionalCost{Buyback("{5}")},
		OnResolve:     ResolveSteps(DrawCards{N: 1}),
	})
}
