package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benevolent Unicorn — Creature — Unicorn {1}{W}, 1/2 (Mirage, #5633):
//
//	"If a spell would deal damage to a permanent or player, it deals that much
//	 damage minus 1 to that permanent or player instead."
//
// A replacement effect (CR 614.1a) on every spell's damage, anyone's and to
// anything; damage from an ability or a creature is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b437c963-def5-40d6-b567-ae9f1d2a0fa5",
		Name:         "Benevolent Unicorn",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB794DamageMinusOneFromSpells("Benevolent Unicorn")},
	})
}
