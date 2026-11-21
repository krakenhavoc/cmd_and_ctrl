package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Four Knocks — Enchantment {2}{W} (Doctor Who):
//
//	"Vanishing 4 (This enchantment enters with four time counters on it.
//	 At the beginning of your upkeep, remove a time counter from it. When
//	 the last is removed, sacrifice it.)
//	 At the beginning of your first main phase, draw a card."
//
// Vanishing is cmdr2B17Vanishing* (CR 702.63a): the entry counters are a
// replacement, the upkeep removal and the last-counter sacrifice are two
// triggers. The draw is a plain beginning-of-first-main trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64fe83b1-5aae-4433-82c3-a11c8c83137e",
		Name:         "Four Knocks",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(4)},
		Triggered: append(effects.Cmdr2B17VanishingTriggers(),
			effects.AtYourPrecombatMain("Four Knocks — draw a card", effects.Do(effects.DrawCards{N: 1}))),
	})
}
