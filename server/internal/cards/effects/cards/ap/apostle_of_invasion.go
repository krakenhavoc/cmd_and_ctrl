package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apostle of Invasion — Creature — Phyrexian Angel {4}{W}{W}, 4/4:
//
//	"Flying
//	 Corrupted — As long as an opponent has three or more poison counters,
//	 this creature has double strike."
//
// A layer 6 grant on the source alone, re-answered on every recompute
// (SelfKeywordWhileCorrupted).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d56f1543-2b2d-4ce3-bcfe-41224a0cece4",
		Name:            "Apostle of Invasion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.SelfKeywordWhileCorrupted("double strike")},
	})
}
