package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "bcbef58b-0c16-40f2-9b5c-14dfcc00f5e5",
		Name:         "Furnace Gremlin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Furnace Gremlin — +1/+0"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Furnace Gremlin — incubate X, where X is its power", effects.IncubateItsPower),
		},
	})
}
