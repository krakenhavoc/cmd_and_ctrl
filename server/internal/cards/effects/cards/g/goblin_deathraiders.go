package cardsg

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Goblin Deathraiders — Creature — Goblin Warrior {B}{R}, 3/1 (Duel Decks:
// Speed vs. Cunning, #5555):
//
//	"Trample"
//
// A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d5142292-446a-413b-85a0-be8a59cec1f8",
		Name:            "Goblin Deathraiders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
