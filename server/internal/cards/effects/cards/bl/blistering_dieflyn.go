package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blistering Dieflyn — Creature — Imp {3}{R}, 0/1 (Shadowmoor, #5591):
//
//	"Flying
//	 {B/R}: This creature gets +1/+0 until end of turn."
//
// Flying is a printed keyword. thisGetsRow, Kranioceros' self pump: an
// activation with no {T}, so it can be activated any number of times. A
// hybrid symbol is paid with either colour (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4cd6ad78-7441-4dd7-b486-d9575e8e0947",
		Name:            "Blistering Dieflyn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{B/R}", 1, 0, "Blistering Dieflyn")},
	})
}
