package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:            "72e8d67a-41e6-4074-bcbd-b54ead995237",
		Name:                "Pentad Prism",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "Remove a charge counter from this artifact: Add one mana of any color",
		}},
	})
}
