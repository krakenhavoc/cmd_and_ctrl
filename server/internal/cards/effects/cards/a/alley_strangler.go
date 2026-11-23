package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alley Strangler — Creature — Aetherborn Rogue {2}{B}, 2/3 (Aether Revolt, #5536):
// "Menace". A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb9af562-7384-4575-93bb-34f1ada23735",
		Name:            "Alley Strangler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
