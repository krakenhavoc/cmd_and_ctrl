package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loch Korrigan — Creature — Spirit {3}{B}, 1/1 (Shadowmoor, #5591):
//
//	"{U/B}: This creature gets +1/+1 until end of turn."
//
// thisGetsRow, Kranioceros' self pump: an activation with no {T}, so it
// can be activated any number of times. A hybrid symbol is paid with
// either colour (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f281c47f-3587-4a3f-8726-ec59568fef94",
		Name:         "Loch Korrigan",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{U/B}", 1, 1, "Loch Korrigan")},
	})
}
