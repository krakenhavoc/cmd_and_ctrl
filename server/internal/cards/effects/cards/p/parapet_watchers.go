package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parapet Watchers — Creature — Kithkin Soldier {2}{U}, 2/2 (Shadowmoor,
// #5591):
//
//	"{W/U}: This creature gets +0/+1 until end of turn."
//
// thisGetsRow, Kranioceros' self pump: an activation with no {T}, so it
// can be activated any number of times. A hybrid symbol is paid with
// either colour (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2ffce33-ab1a-40f1-959b-31583219ed1a",
		Name:         "Parapet Watchers",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{W/U}", 0, 1, "Parapet Watchers")},
	})
}
