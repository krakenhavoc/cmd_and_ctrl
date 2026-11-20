package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caustic Hound — Creature — Phyrexian Dog {5}{B}, 4/4 (Mirrodin Besieged,
// #5576):
//
//	"When this creature dies, each player loses 4 life."
//
// Deathcurse Ogre's dies trigger at 4 (nemEachPlayerLoses): every player
// still in the game, its controller included, loses the life. Dying looks
// back in time (CR 603.10a), so it triggers from the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "032cc574-f7e7-4753-9a30-76cbd45d9563",
		Name:         "Caustic Hound",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Caustic Hound — each player loses 4 life", nemEachPlayerLoses(4)),
		},
	})
}
