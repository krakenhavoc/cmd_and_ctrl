package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agate-Blade Assassin — Creature — Lizard Assassin {1}{B}, 1/3
// (Bloomburrow, #4467):
//
//	"Whenever this creature attacks, defending player loses 1 life and
//	 you gain 1 life."
//
// Brutal Hordechief's drain (attackDrainTriggerWhen) on its own attack
// only: the defending player is read off the attack as it is declared,
// so attacking a planeswalker or a battle drains the player behind it,
// and a defending player who has left the game loses nothing while you
// still gain 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "381a3e8e-71dd-48e4-ab62-53478bde4a14",
		Name:         "Agate-Blade Assassin",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attackDrainTriggerWhen("Agate-Blade Assassin", ThisAttacked)},
	})
}
