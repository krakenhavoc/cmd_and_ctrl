package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repeal — Instant {X}{U} (Ravnica: Clue Edition, #5929):
//
//	"Return target nonland permanent with mana value X to its owner's
//	 hand.
//	 Draw a card."
//
// Disembowel's target (mana value equal to the announced X, judged at
// announce and again at resolution) with Into the Roil's bounce, then the keyed
// draw (bounceTheTargetThenDraw).
// A target that became illegal counters the whole spell, draw included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1f2b26c-91b2-4f3c-9f48-6429ec3fde48",
		Name:         "Repeal",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent with mana value X", Nonland()).WithManaValueEqualsX(),
		OnResolve:    bounceTheTargetThenDraw,
	})
}
