package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roc of Kher Ridges — Creature — Bird {3}{R}, 3/3
// (Limited Edition Alpha, #5642): "Flying". A printed keyword and
// nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01f07ab9-3c3e-4fea-9cd6-543a5b7056b8",
		Name:            "Roc of Kher Ridges",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
