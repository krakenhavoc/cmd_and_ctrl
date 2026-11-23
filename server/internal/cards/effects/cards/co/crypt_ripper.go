package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crypt Ripper — Creature — Shade {2}{B}{B}, 2/2 (Zendikar, #5583):
//
//	"Haste
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Haste is the printed keyword; the pump is the shared firebreathing
// row (self_pump_row.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "18bcf4c5-7abc-4ca5-b805-4bbebf4e433c",
		Name:            "Crypt Ripper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []effects.ActivatedAbility{effects.SelfPumpRow("Crypt Ripper", "{B}", 1, 1)},
	})
}
