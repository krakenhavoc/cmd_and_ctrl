package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Shields — Artifact Creature — Wall {3}, 0/4 (Ice Age, S82 OLD batch 849, #5636):
//
//	"Defender; banding"
//
// Keywords ride PrintedKeywords; banding is the engine's (game/banding.go, ADR 0224).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "530514d9-7274-4835-a4b5-6dccef4da758",
		Name:            "Wall of Shields",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", game.KeywordBanding},
	})
}
