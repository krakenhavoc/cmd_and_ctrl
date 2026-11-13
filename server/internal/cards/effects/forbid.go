package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forbid — Instant {1}{U}{U} (Exodus, #5626):
//
//	"Buyback—Discard two cards. (You may discard two cards in addition to
//	 any other costs as you cast this spell. If you do, put this card
//	 into your hand as it resolves.)
//	 Counter target spell."
//
// Cancel behind Demonic Collusion's non-mana buyback (BuybackDiscard).
// The two discards are paid as the spell goes on the stack (CR 601.2h);
// the return to hand is the engine's (CR 702.27a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "6e26a525-2584-490d-b106-621c1a30e1fd",
		Name:          "Forbid",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets:       TargetSpell("target spell"),
		OptionalCosts: []game.AdditionalCost{BuybackDiscard(2)},
		OnResolve:     counterTheTargetSpell,
	})
}
