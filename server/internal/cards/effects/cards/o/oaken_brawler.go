package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oaken Brawler — Creature — Treefolk Warrior {3}{W}, 2/4 (Lorwyn, #5593):
//
//	"When this creature enters, clash with an opponent. If you win, put a
//	 +1/+1 counter on this creature. (Each clashing player reveals the top
//	 card of their library, then puts that card on their choice of the top
//	 or bottom. A player wins if their card had a greater mana value.)"
//
// The clash is the engine's CR 701.30 action; the counter is the clash's
// tail, put on the Brawler only while it is still the same permanent
// (CR 400.7). Nath's Elite shares the trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "819351c6-b0d6-42dc-9bd9-d1a3851aad7c",
		Name:         "Oaken Brawler",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.OldB543WhenThisEntersClashForCounter("Oaken Brawler")},
	})
}
