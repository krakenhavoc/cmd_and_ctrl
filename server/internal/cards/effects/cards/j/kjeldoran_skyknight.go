package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kjeldoran Skyknight — Creature — Human Knight {2}{W}, 1/1 (Ice Age, S82 OLD batch 849, #5636):
//
//	"Flying; first strike; banding"
//
// Keywords ride PrintedKeywords; banding is the engine's (game/banding.go, ADR 0224).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b06db20f-84c4-4e2e-a184-3cd70ea0d10d",
		Name:            "Kjeldoran Skyknight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", game.KeywordBanding},
	})
}
