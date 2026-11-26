package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qal Sisma Behemoth — Creature — Ogre Warrior {2}{R}, 5/5 (Dragons of Tarkir):
//
//	"This creature can't attack or block unless you pay {2}."
//
// Two taxes, paid as the declaration is made (CR 508.1h, 509.1d): {2} for the
// Behemoth when it is declared an attacker and {2} when it is declared a blocker.
//
// No simplification.
func init() {
	const label = "This creature can't attack or block unless you pay {2}."
	effects.Register(effects.Spec{
		OracleID:     "636e9863-29ac-476a-86db-d8f73b4adbf4",
		Name:         "Qal Sisma Behemoth",
		Completeness: effects.CompletenessFull,
		AttackTaxes:  []game.AttackTax{effects.CantAttackUnlessYouPayMana("{2}", label)},
		BlockTaxes:   []game.BlockTax{effects.CantBlockUnlessYouPayMana("{2}", label)},
	})
}
