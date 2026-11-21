package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre Charger — Creature — Elemental Warrior {R}{R}, 1/1 (Shadowmoor,
// #5591):
//
//	"Haste
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Haste is a printed keyword. thisGetsRow, Kranioceros' self pump: an
// activation with no {T}, so it can be activated any number of times. A
// hybrid symbol is paid with either colour (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "50bc464e-1c8d-455c-a870-ddbbc4618343",
		Name:            "Pyre Charger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{R}", 1, 0, "Pyre Charger")},
	})
}
