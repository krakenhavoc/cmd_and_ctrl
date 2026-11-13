package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shattering Pulse — Instant {1}{R} (Exodus, #5626):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Destroy target artifact."
//
// Shatter behind Capsize's buyback (the shared Buyback constructor,
// CR 702.27a). The return to hand is the engine's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "77953b7a-813e-45f5-ad7f-ab72a79a0112",
		Name:          "Shattering Pulse",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:       TargetPermanent("target artifact", Artifact()),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve:     destroyTheTargetPermanent,
	})
}
