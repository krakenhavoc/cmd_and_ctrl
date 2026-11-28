package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Willow Faerie — Creature — Faerie {1}{G}, 1/2 (Homelands, S82 OLD b826,
// #5635):
//
//	"Flying"
//
// A vanilla flier: the printed keyword is all there is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5390bb44-985e-473b-883e-c9562c06f210",
		Name:            "Willow Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
