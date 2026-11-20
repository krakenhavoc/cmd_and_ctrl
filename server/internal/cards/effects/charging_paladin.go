package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charging Paladin — Creature — Human Knight {2}{W}, 2/2 (Portal, #5630):
//
//	"Whenever this creature attacks, it gets +0/+3 until end of turn."
//
// Vicious Kavu's attack trigger with +0/+3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cccf44e-c05e-4239-b643-54b559c98552",
		Name:         "Charging Paladin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Charging Paladin — it gets +0/+3 until end of turn",
				thisGetsUntilEndOfTurn(0, 3, "Charging Paladin — +0/+3")),
		},
	})
}
