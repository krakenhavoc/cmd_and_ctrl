package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Shade — Creature — Shade {B}{B}{B}, 3/3 (Dominaria, #5530):
//
//	"{B}: This creature gets +1/+1 until end of turn."
//
// Nantuko Shade's pump, the shared selfPumpRow.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b90c7d1-38c6-4210-bc76-68c5fcb91202",
		Name:         "Dread Shade",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Dread Shade", "{B}", 1, 1)},
	})
}
