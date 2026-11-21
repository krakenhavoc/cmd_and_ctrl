package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timber Wolves — Creature — Wolf {G}, 1/1 (Limited Edition Alpha):
//
//	"Banding"
//
// Banding is the engine's (game/banding.go, ADR 0224).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35d07ac9-b184-4b5f-8192-34b1db042f69",
		Name:            "Timber Wolves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
	})
}
