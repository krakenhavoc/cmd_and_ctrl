package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kjeldoran Phalanx — Creature — Human Soldier {5}{W}, 2/5 (Ice Age, S82 OLD batch 849, #5636):
//
//	"First strike; banding"
//
// Keywords ride PrintedKeywords; banding is the engine's (game/banding.go, ADR 0224).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a8563a4f-a23b-4cb5-b6e5-795bf8d74ee6",
		Name:            "Kjeldoran Phalanx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", game.KeywordBanding},
	})
}
