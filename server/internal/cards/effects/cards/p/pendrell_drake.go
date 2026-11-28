package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pendrell Drake — Creature — Drake 2/3 {3}{U} (Urza's Saga, S82 OLD b722):
//
//	Flying
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Flying is a printed keyword; cycling is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d8981df-4cdc-4fd4-acd8-095fa3281281",
		Name:            "Pendrell Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
