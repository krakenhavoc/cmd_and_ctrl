package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Balloon Brigade — Creature — Goblin Warrior {R}, 1/1 (Conspiracy:
// Take the Crown, S82 older sets batch 273, #5540):
//
//	"{R}: This creature gains flying until end of turn."
//
// A layer 6 grant pinned to the source and ended at cleanup (CR 514.2);
// each activation is its own, and a Brigade that has left the battlefield
// grants nothing (the shared ThisGainsKeywordUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10bc98b0-3fdc-46d1-8d3b-6d160e9dd62f",
		Name:         "Goblin Balloon Brigade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Goblin Balloon Brigade — flying until end of turn"),
		}},
	})
}
