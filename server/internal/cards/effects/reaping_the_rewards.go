package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reaping the Rewards — Instant {W} (Exodus, #5626):
//
//	"Buyback—Sacrifice a land. (You may sacrifice a land in addition to
//	 any other costs as you cast this spell. If you do, put this card
//	 into your hand as it resolves.)
//	 You gain 2 life."
//
// Constant Mists' non-mana buyback (BuybackSacrifice, sacrificed with
// the spell already on the stack, CR 601.2h) on a two-life gain. The
// return to hand is the engine's (CR 702.27a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "b4177445-46d8-4427-a345-d786c820c3af",
		Name:          "Reaping the Rewards",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{LifeGain: 2, Answers: game.AnswerValue},
		OptionalCosts: []game.AdditionalCost{BuybackSacrifice("a land", Land())},
		OnResolve:     ResolveSteps(GainLife{Amount: 2}),
	})
}
