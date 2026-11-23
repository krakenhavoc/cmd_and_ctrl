package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strangleroot Geist — Creature — Spirit {G}{G}, 2/1 (Dark Ascension, #5571):
//
//	"Haste
//	 Undying (When this creature dies, if it had no +1/+1 counters on
//	 it, return it to the battlefield under its owner's control with
//	 a +1/+1 counter on it.)"
//
// Haste and undying are PrintedKeywords (#2075). The undying return is a
// new object, so it has haste again and may attack the turn it returns.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "af12758f-4a7b-4156-8942-de4716aa0623",
		Name:            "Strangleroot Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", game.KeywordUndying},
	})
}
