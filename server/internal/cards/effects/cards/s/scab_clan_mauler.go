package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scab-Clan Mauler — Creature — Human Berserker {R}{G}, 1/1 (Ravnica
// Remastered, #5948):
//
//	"Bloodthirst 2 (If an opponent was dealt damage this turn, this creature
//	 enters with two +1/+1 counters on it.)
//	 Trample"
//
// The shared Bloodthirst replacement: it reads, as the Mauler enters, whether
// any opponent has been dealt damage this turn (CR 702.54a), and adds the two
// counters to the entry itself, so it is a 3/3 from the moment it is on the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db0b6e35-d7f2-435a-87d7-9e8c36426f4d",
		Name:            "Scab-Clan Mauler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    []game.ReplacementEffect{effects.Bloodthirst(2, "Scab-Clan Mauler")},
	})
}
