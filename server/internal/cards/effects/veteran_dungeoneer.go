package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran Dungeoneer — Creature — Human Warrior {3}{W}, 3/4:
//
//	"When this creature enters, venture into the dungeon."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ce7b966-2a10-4250-8c0d-525e73efe2f4",
		Name:         "Veteran Dungeoneer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{WhenThisEntersVenture("Veteran Dungeoneer")},
	})
}
