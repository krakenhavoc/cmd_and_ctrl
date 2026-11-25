package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pack Mastiff — Creature — Dog {1}{R}, 2/2 (Core Set 2020):
//
//	"{1}{R}: Each creature you control named Pack Mastiff gets +1/+0 until end
//	 of turn."
//
// A mass pump until end of turn whose set is fixed as the ability resolves
// (CR 611.2c): every creature you control with the name, itself included. A
// creature that enters later, or a copy that gains the name later, is not
// pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f62d4437-bac6-49c8-9aeb-07de78551df7",
		Name:         "Pack Mastiff",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Each creature you control named Pack Mastiff gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.HasName(c, "Pack Mastiff") }),
					Power: 1, Label: "Pack Mastiff — +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
