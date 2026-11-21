package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleric of Chill Depths — Creature — Merfolk Cleric {1}{U}, 1/3
// (Zendikar Rising, #5508):
//
//	"Whenever this creature blocks a creature, that creature doesn't
//	 untap during its controller's next untap step."
//
// Labyrinth Minotaur's text, through the same constructor
// (WheneverThisBlocksItDoesntUntap): once for each attacker it blocks
// (CR 509.3b), and "that creature" is the attacker as it was when
// blocked, so one that left and came back before the trigger resolved is
// a new object and is not marked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99c13f8d-7f18-48ce-86da-c74ddb6c1205",
		Name:         "Cleric of Chill Depths",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.WheneverThisBlocksItDoesntUntap(
			"Cleric of Chill Depths — the creature it blocks doesn't untap during its controller's next untap step")},
	})
}
