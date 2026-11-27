package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guma — Creature — Cat {2}{R} 2/2 (Urza's Saga, #5624):
//
//	"Protection from blue"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a9cbf2ed-9896-41b5-b17f-e2fcb3cb82fa",
		Name:            "Guma",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from blue"},
	})
}
