package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rune-Cervin Rider — Creature — Elf Knight {3}{W}, 2/2 (Shadowmoor,
// #5591):
//
//	"Flying
//	 {G/W}{G/W}: This creature gets +1/+1 until end of turn."
//
// Flying is a printed keyword. thisGetsRow, Kranioceros' self pump: an
// activation with no {T}, so it can be activated any number of times. A
// hybrid symbol is paid with either colour (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5c6956ad-8031-4708-82ed-75491755d2a4",
		Name:            "Rune-Cervin Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{G/W}{G/W}", 1, 1, "Rune-Cervin Rider")},
	})
}
