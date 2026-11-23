package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frontline Rebel — Creature — Human Warrior {2}{R}, 3/3 (Aether Revolt,
// #5536):
//
//	"This creature attacks each combat if able."
//
// Juggernaut's attack requirement (CR 508.1d): passing the declaration
// with it home while it can attack is refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7471e445-4655-4336-9769-3debd197b77c",
		Name:         "Frontline Rebel",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
