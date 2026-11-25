package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earthshaking Si — Creature — Beast {5}{G}, 5/5 (Global Series Jiang Yanggu & Mu Yanling, S82 OLD batch 201, #5528):
//
//	"Trample"
//
// A keyword body: trample is the printed keyword and nothing else is declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "62c4d349-a18e-40a2-8cc3-244c219dec7f",
		Name:            "Earthshaking Si",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
