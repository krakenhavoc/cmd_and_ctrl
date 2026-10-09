package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burnout Bashtronaut — Creature — Goblin Warrior {R}, 1/1:
//
//	"Menace
//	 Start your engines!
//	 {2}: This creature gets +1/+0 until end of turn.
//	 Max speed — This creature has double strike."
//
// ADR 0138 (#2122). Double strike is a layer 6 grant switched on by its
// controller's speed (CR 702.178a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1b122fbb-c2e1-42d1-bfc5-fcacacdfa0bc",
		Name:            "Burnout Bashtronaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", StartYourEngines},
		Activated: []ActivatedAbility{{
			Label: "{2}: This creature gets +1/+0 until end of turn.",
			Cost:  ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Burnout Bashtronaut — +1/+0"}.Apply(ctx)
			},
		}},
		Static: MaxSpeedSelfKeywords("double strike"),
	})
}
