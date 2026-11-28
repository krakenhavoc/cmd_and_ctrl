package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kjeldoran Skycaptain — Creature — Human Soldier {4}{W}, 2/2 (Ice Age, S82 OLD batch 849, #5636):
//
//	"Flying; first strike; banding"
//
// Keywords ride PrintedKeywords; banding is the engine's (game/banding.go, ADR 0224).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7160ebce-ee3f-42ac-a3f7-b55f17f2c49f",
		Name:            "Kjeldoran Skycaptain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", game.KeywordBanding},
	})
}
