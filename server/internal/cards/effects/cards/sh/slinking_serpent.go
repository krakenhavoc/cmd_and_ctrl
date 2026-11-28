package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slinking Serpent — Creature — Serpent {2}{U}{B}, 2/3 (Invasion, S82 #5616):
//
//	"Forestwalk"
//
// Forestwalk is a printed keyword: it can't be blocked while the defending player controls a Forest (CR 702.14).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23875679-5c08-44ff-90b3-13a1a758e5f7",
		Name:            "Slinking Serpent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
