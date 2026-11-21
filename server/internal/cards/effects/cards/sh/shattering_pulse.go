package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "77953b7a-813e-45f5-ad7f-ab72a79a0112",
		Name:          "Shattering Pulse",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:       effects.TargetPermanent("target artifact", effects.Artifact()),
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{3}")},
		OnResolve:     effects.DestroyTheTargetPermanent,
	})
}
