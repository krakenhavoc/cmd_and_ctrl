package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rustwing Falcon — Creature — Bird {W}, 1/2 (Core Set 2019, #5527):
//
//	"Flying"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "438ee544-78a2-4e32-9343-70838b60ac16",
		Name:            "Rustwing Falcon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
