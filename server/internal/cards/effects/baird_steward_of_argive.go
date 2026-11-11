package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baird, Steward of Argive — Legendary Creature {2}{W}{W}, 2/4:
//
//	"Vigilance
//	 Creatures can't attack you or planeswalkers you control unless their
//	 controller pays {1} for each of those creatures."
func init() {
	Register(Spec{
		OracleID:        "69aa6809-5468-413a-94c9-562e6de93957",
		Name:            "Baird, Steward of Argive",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		AttackTaxes: []game.AttackTax{
			ProtectingPlaneswalkers(AttackTax("{1}", "Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.")),
		},
	})
}
