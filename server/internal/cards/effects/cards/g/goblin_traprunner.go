package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Traprunner — Creature — Goblin {3}{R}, 4/2 (Modern Horizons 2):
//
//	"Whenever this creature attacks, flip three coins. For each flip you win,
//	 create a 1/1 red Goblin creature token that's tapped and attacking."
//
// You call each of the three flips (CR 705); each win is a Goblin that enters
// tapped and attacking (CR 508.4), the attacked player or planeswalker your
// choice, never declared as an attacker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbdffef0-9119-4650-9f92-58e8a664db6f",
		Name:         "Goblin Traprunner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Goblin Traprunner — flip three coins; for each flip you win, create a 1/1 red Goblin creature token that's tapped and attacking",
				effects.OldB033TraprunnerFlip),
		},
	})
}
