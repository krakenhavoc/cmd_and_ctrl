package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of Vitu-Ghazi — Creature — Elemental {6}{G}{W}, 4/7 (Ravnica:
// City of Guilds, #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Vigilance"
//
// Convoke is the engine's tap cost (CR 702.51); vigilance is printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7d5a9ca-27ec-4843-a9fd-0de0d67def39",
		Name:            "Guardian of Vitu-Ghazi",
		Completeness:    effects.CompletenessFull,
		TapCost:         effects.Convoke(),
		PrintedKeywords: []string{"vigilance"},
	})
}
