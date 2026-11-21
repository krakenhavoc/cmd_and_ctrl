package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "664d982a-5cd8-4289-afb8-fa2e083042c9",
		Name:          "Allay",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:       effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{3}")},
		OnResolve:     effects.DestroyTheTargetPermanent,
	})
}
