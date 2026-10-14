package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furnace Gremlin — Creature — Phyrexian Gremlin {1}{R}, 1/2 (March of
// the Machine):
//
//	"{1}{R}: This creature gets +1/+0 until end of turn.
//	 When this creature dies, incubate X, where X is its power."
//
// The dies trigger reads last-known power (incubateItsPower), so pumps
// that were on it when it died count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bcbef58b-0c16-40f2-9b5c-14dfcc00f5e5",
		Name:         "Furnace Gremlin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Furnace Gremlin — +1/+0"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Furnace Gremlin — incubate X, where X is its power", incubateItsPower),
		},
	})
}
