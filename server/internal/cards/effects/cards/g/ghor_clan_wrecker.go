package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghor-Clan Wrecker — Creature — Human Warrior {3}{R}, 2/2 (Ravnica
// Allegiance, #5521):
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Menace (This creature can't be blocked except by two or more
//	 creatures.)"
//
// Riot is the engine's keyword (ADR 0109 §10): the entry replacement asks
// for the counter or haste as it enters (CR 702.136a). Menace is the
// printed block rule (CR 702.111b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b113285c-80b2-46dd-874e-201d26979518",
		Name:            "Ghor-Clan Wrecker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot, "menace"},
	})
}
