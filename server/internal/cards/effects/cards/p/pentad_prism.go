package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pentad Prism — Artifact {2}:
//
//	"Sunburst (This artifact enters with a charge counter on it for
//	 each color of mana spent to cast it.)
//	 Remove a charge counter from this artifact: Add one mana of any
//	 color."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552): a noncreature gets charge counters (CR 702.44a).
// The mana ability's only cost is the counter, the Vivid lands' shape
// without the {T}, so the Prism can be cashed out all at once.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "72e8d67a-41e6-4074-bcbd-b54ead995237",
		Name:                "Pentad Prism",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{RemoveCounters: effects.RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "Remove a charge counter from this artifact: Add one mana of any color",
		}},
	})
}
