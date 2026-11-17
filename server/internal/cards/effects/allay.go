package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Allay — Instant {1}{W} (Exodus, #5626):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Destroy target enchantment."
//
// Capsize's buyback (the shared Buyback constructor, CR 702.27a) on the
// shared "Destroy target [permanent]." body. Enchantment() reads the
// post-layer type set, so a permanent that stops being an enchantment in
// response is no longer a legal target (CR 608.2b). The return to hand
// is the engine's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "664d982a-5cd8-4289-afb8-fa2e083042c9",
		Name:          "Allay",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:       TargetPermanent("target enchantment", Enchantment()),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve:     destroyTheTargetPermanent,
	})
}
