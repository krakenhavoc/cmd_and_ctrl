package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paperfin Rascal — Creature — Merfolk Rogue {2}{U}, 2/2 (Lorwyn, #5593):
//
//	"When this creature enters, clash with an opponent. If you win, put a
//	 +1/+1 counter on this creature. (Each clashing player reveals the top
//	 card of their library, then puts that card on their choice of the top
//	 or bottom. A player wins if their card had a greater mana value.)"
//
// The clash is the engine's CR 701.30 action; the counter is the clash's
// tail, put on the creature only while it is still the same permanent
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdec1d22-2f89-424b-b48f-46c9158943af",
		Name:         "Paperfin Rascal",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.OldB544WhenThisEntersClashForCounter("Paperfin Rascal")},
	})
}
