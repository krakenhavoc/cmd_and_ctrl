package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Dragon — Creature — Dragon {4}{R}{R}, 5/5 (Foundations, #4465):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Furnace Gremlin's pump at {R}, on a flier. The ability has no tap in
// its cost, so it can be activated any number of times, summoning
// sickness or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "711eea87-0fa3-46e0-a42b-fa5a86455f04",
		Name:            "Shivan Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Shivan Dragon — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
