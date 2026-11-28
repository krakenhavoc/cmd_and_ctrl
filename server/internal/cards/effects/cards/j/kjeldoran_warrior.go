package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kjeldoran Warrior — Creature — Human Warrior {W}, 1/1 (Ice Age, S82 OLD batch 849, #5636):
//
//	"Banding"
//
// Keywords ride PrintedKeywords; banding is the engine's (game/banding.go, ADR 0224).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73b70a48-b62c-4fdf-9c2d-c6d6688f9ac8",
		Name:            "Kjeldoran Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
	})
}
