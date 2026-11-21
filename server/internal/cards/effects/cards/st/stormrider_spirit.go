package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormrider Spirit — Creature — Spirit {4}{U}, 3/3 (Innistrad: Midnight
// Hunt Commander reprint, S82):
//
//	"Flash
//	 Flying"
//
// Both words are printed keywords: flash lets it be cast at instant speed,
// flying is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c378f72-917c-49d4-ba68-51ebaf18d976",
		Name:            "Stormrider Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
