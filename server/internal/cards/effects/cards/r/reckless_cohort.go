package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Cohort — Creature — Human Warrior Ally {1}{R}, 2/2 (Battle for
// Zendikar, #5547):
//
//	"This creature attacks each combat if able unless you control another
//	 Ally."
//
// The requirement is a static that switches off while another Ally is
// under your control (CR 508.1d); "Ally" is a word of the card's text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "015b1006-decb-4d94-940e-bca550e0ad5b",
		Name:         "Reckless Cohort",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB314RecklessCohortMustAttack()},
	})
}
