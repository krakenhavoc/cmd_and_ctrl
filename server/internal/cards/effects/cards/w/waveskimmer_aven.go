package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waveskimmer Aven — Creature — Bird Soldier {2}{G}{W}{U}, 2/4 (Shards of
// Alara, #5588):
//
//	"Flying
//	 Exalted"
//
// Both are printed keywords; exalted's trigger is derived by the engine
// (CR 702.83a, game/exalted.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "113d6c3c-3492-47aa-81d1-359d852d8d0a",
		Name:            "Waveskimmer Aven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordExalted},
	})
}
