package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siege Wurm — Creature — Wurm {5}{G}{G}, 5/5 (Ravnica: City of Guilds,
// #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Trample"
//
// Autochthon Wurm's shape: convoke is the engine's tap cost (CR 702.51)
// and trample is printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a802eae-abfb-47fd-8458-1689055531ef",
		Name:            "Siege Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		TapCost:         effects.Convoke(),
	})
}
