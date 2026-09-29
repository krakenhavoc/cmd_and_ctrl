package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crashing Drawbridge — Artifact Creature — Wall {2}, 0/4:
//
//	"Defender
//	 {T}: Creatures you control gain haste until end of turn."
//
// Defender rides PrintedKeywords like any other combat keyword. The
// tap ability is an ordinary S32 turn-scoped keyword grant
// (`GrantKeywordUntilEOT`) over "creatures you control" — the same
// `And(Creature(), YouControl())` match Overrun uses — snapshotted
// once at activation per CR 611.2c, so a creature cast afterward this
// turn doesn't get haste from it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "328d5ef0-b25e-4f2a-80fe-35c6ae419e8a",
		Name:            "Crashing Drawbridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label: "{T}: Creatures you control gain haste until end of turn.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"haste"},
					Label:    "Crashing Drawbridge — haste",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
