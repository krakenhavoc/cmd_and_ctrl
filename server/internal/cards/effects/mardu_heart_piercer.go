package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mardu Heart-Piercer — Creature — Human Archer {3}{R}, 2/3 (Jumpstart 2022,
// #5948):
//
//	"Raid — When this creature enters, if you attacked this turn, this
//	 creature deals 2 damage to any target."
//
// Storm Fleet Pyromancer's raid trigger: an intervening "if" (CR 603.4),
// the Archer deals the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4b15cba-9cd7-422d-9ca6-39243bf67dca",
		Name:         "Mardu Heart-Piercer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{gapB157RaidDamage("Mardu Heart-Piercer", 2)},
	})
}
