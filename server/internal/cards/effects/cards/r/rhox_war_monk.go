package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhox War Monk — {G}{W}{U} Creature — Rhino Monk 3/4 (Shards of Alara, S82 OLD b515):
//
//	"Lifelink"
//
// Keywords only: each is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5c05c350-1b11-4dda-9de6-01c7e0f07ffc",
		Name:            "Rhox War Monk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
