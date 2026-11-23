package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baird, Steward of Argive — Legendary Creature {2}{W}{W}, 2/4:
//
//	"Vigilance
//	 Creatures can't attack you or planeswalkers you control unless their
//	 controller pays {1} for each of those creatures."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69aa6809-5468-413a-94c9-562e6de93957",
		Name:            "Baird, Steward of Argive",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		AttackTaxes: []game.AttackTax{
			effects.ProtectingPlaneswalkers(effects.AttackTax("{1}", "Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.")),
		},
	})
}
