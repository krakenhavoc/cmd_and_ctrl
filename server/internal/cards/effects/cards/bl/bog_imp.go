package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Imp — Creature — Imp {1}{B}, 1/1
// (The Dark, #5638): "Flying". A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "45b94e3c-a905-435b-aee5-bec9239fd24c",
		Name:            "Bog Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
