package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firestorm Hellkite — Creature — Dragon {4}{U}{R}, 6/6 (Visions, S82 older
// sets batch 789, #5631):
//
//	"Flying, trample
//	 Cumulative upkeep {U}{R}"
//
// The age counter goes on first, then the cost is charged once per counter
// (CR 702.24a): two counters cost {U}{R}{U}{R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "533e41cc-7cae-47f2-886b-8a3ffb82bf81",
		Name:            "Firestorm Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Firestorm Hellkite — cumulative upkeep {U}{R}", "{U}{R}"),
		},
	})
}
