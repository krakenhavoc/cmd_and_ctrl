package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghor-Clan Savage — Creature — Centaur Berserker {3}{G}{G}, 2/3 (Duel Decks:
// Heroes vs. Monsters, S82 old sets):
//
//	"Bloodthirst 3 (If an opponent was dealt damage this turn, this creature
//	 enters with three +1/+1 counters on it.)"
//
// Battering Wurm's Bloodthirst, at three: its own entry replacement, read as
// the creature enters (CR 702.54a), counting any damage dealt to an opponent
// this turn from any source, and never a drain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28f35fa1-befa-40b4-95ab-30bb53d339ab",
		Name:         "Ghor-Clan Savage",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(3, "Ghor-Clan Savage")},
	})
}
