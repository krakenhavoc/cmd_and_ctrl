package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kederekt Creeper — {U}{B}{R} Creature — Horror 2/3 (Shards of Alara, S82 OLD b515):
//
//	"Menace
//	 Deathtouch"
//
// Keywords only: each is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f8e0b2c-f1a5-40ad-b8d8-7c06a012e909",
		Name:            "Kederekt Creeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "deathtouch"},
	})
}
