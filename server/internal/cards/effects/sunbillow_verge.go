package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunbillow Verge — Land:
//
//	"{T}: Add {W}.
//	 {T}: Add {R}. Activate only if you control a Mountain or a Plains."
//
// The Foundations verge shape verges.go already builds for seven of the
// cycle, written out here with the Plains/Mountain pair: an unconditional
// first colour and a second behind a `ManaAbility.Condition` that reads
// post-layer land types. The verge's own type line has no land types, so
// it never satisfies its own gate.
//
// No simplification.
func init() {
	mountain, plains := MatchLandSubtype("Mountain"), MatchLandSubtype("Plains")
	Register(Spec{
		OracleID:     "a202276b-1f1b-4277-95ee-26877a204f5e",
		Name:         "Sunbillow Verge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W}",
				Label:    "Add {W}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Label:    "Add {R} (only if you control a Mountain or a Plains)",
				Condition: ControlsAtLeast(1, func(c game.Card) bool {
					return mountain(c) || plains(c)
				}),
			},
		},
	})
}
