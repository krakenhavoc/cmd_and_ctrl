package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gluttonous Troll — Creature — Troll {2}{B}{G}, 3/3 (Commander 2021,
// #5504):
//
//	"Trample
//	 When this creature enters, create a number of Food tokens equal to the
//	 number of opponents you have.
//	 {1}{G}, Sacrifice another nonland permanent: This creature gets +2/+2
//	 until end of turn."
//
// The Food count is the opponents still in the game as the trigger
// resolves. The pump's cost sacrifices another nonland permanent (object
// identity, so the Troll can't pay with itself, and a land can't pay),
// and it can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a3712f9-6bed-4068-9064-2b5240fc7b28",
		Name:            "Gluttonous Troll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gluttonous Troll — create a Food for each opponent", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := len(ctx.Opponents())
				if n == 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.FoodToken(), N: n}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice another nonland permanent: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeAnotherN(1, "another nonland permanent", effects.Nonland())),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Gluttonous Troll — +2/+2 until end of turn"),
		}},
	})
}
