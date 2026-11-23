package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valeron Outlander — Creature — Human Scout {G}{W}, 2/2 (Conflux,
// #5587):
//
//	"Protection from black"
//
// Protection from black is a printed keyword, read by targeting,
// blocking, damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01d70534-7c26-4a54-9eca-d0ebbefc2159",
		Name:            "Valeron Outlander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
	})
}
