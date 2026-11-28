package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shield Bearer — Creature — Human Soldier {1}{W}, 0/3 (Ice Age, S82 OLD batch 849, #5636):
//
//	"Banding"
//
// Keywords ride PrintedKeywords; banding is the engine's (game/banding.go, ADR 0224).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "75decba6-d188-4144-8e74-58285be0fe9f",
		Name:            "Shield Bearer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
	})
}
