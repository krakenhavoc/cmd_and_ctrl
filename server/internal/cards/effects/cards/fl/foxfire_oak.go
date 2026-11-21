package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foxfire Oak — Creature — Treefolk Shaman {5}{G}, 3/6 (Shadowmoor,
// #5591):
//
//	"{R/G}{R/G}{R/G}: This creature gets +3/+0 until end of turn."
//
// thisGetsRow, Kranioceros' self pump: an activation with no {T}, so it
// can be activated any number of times. A hybrid symbol is paid with
// either colour (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "925775c6-b6d2-4793-a363-bd5aca871aa8",
		Name:         "Foxfire Oak",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{R/G}{R/G}{R/G}", 3, 0, "Foxfire Oak")},
	})
}
