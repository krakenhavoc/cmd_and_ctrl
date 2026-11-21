package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noxious Groodion — Creature — Beast {2}{B}, 2/2 (Ravnica Allegiance,
// #5521):
//
//	"Deathtouch"
//
// Deathtouch is an engine keyword carried as printed data (CR 702.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aed587bf-c6e5-4850-baeb-2715cb27bf4c",
		Name:            "Noxious Groodion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
