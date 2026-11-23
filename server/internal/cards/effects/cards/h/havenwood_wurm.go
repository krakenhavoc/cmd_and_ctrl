package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Havenwood Wurm — {6}{G} Creature — Wurm 5/6 (Jumpstart 2022, #5948):
//
//	"Flash
//	 Trample"
//
// Two printed keywords, nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a4ddff34-d0ae-481f-870b-b911b249250f",
		Name:            "Havenwood Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "trample"},
	})
}
