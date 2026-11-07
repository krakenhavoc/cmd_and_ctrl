package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greasewrench Goblin — Creature — Goblin Artificer {R}, 2/1 (Aetherdrift,
// #4464):
//
//	"Exhaust — {2}{R}: Discard up to two cards, then draw that many
//	 cards. Put a +1/+1 counter on this creature. (Activate each exhaust
//	 ability only once.)"
//
// Cathartic Pyre's second bullet (b39MayDiscardThenDraw with
// drawAsManyAsDiscarded) as an exhaust ability (CR 702.177a): the
// discard is chosen as the ability resolves, zero to two cards, and the
// draw is as many as were really discarded. The counter is the next
// step, so it waits for the discard prompt (CR 608.2c), and it goes only
// on the Goblin that activated it (plusOneCounterOnThisIfStillHere): one
// that left and came back gets nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "952300de-766e-4ceb-bd36-e0b721bebc17",
		Name:         "Greasewrench Goblin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {2}{R}: Discard up to two cards, then draw that many cards. Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Exhaust: true,
			Cost:    ManaCost("{2}{R}"),
			Effect: Do(
				Instruction(b39MayDiscardThenDraw(2, false,
					"Greasewrench Goblin — discard up to two cards, then draw that many", drawAsManyAsDiscarded)),
				Instruction(plusOneCounterOnThisIfStillHere),
			),
		}},
	})
}
