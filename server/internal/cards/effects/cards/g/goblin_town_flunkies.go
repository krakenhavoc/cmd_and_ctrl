package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin-town Flunkies — Creature — Goblin Soldier {1}{R}, 1/1 (The Hobbit):
//
//	"Haste
//	 When this creature enters, amass Goblins 1."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "85ceb4ac-7a1c-4c7e-95a3-515a5b5d6116",
		Name:            "Goblin-town Flunkies",
		Purpose:         game.Purpose{Amass: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Goblin-town Flunkies — amass Goblins 1", effects.Do(effects.Amass{Subtype: "Goblin", N: 1})),
		},
	})
}
