package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jirina Kudro — Legendary Creature — Human Soldier {1}{R}{W}{B}, 3/3
// (Commander 2020, #5513):
//
//	"When Jirina Kudro enters, create a 1/1 white Human Soldier creature
//	 token for each time you've cast a commander from the command zone this
//	 game.
//	 Other Humans you control get +2/+0."
//
// The count is Commander's Insight's: every cast of a commander you made
// from the command zone this game (Player.CommanderCasts, the same tally the
// commander tax reads, so a second commander's casts count too), read as the
// trigger resolves. The anthem is Magda, Brazen Outlaw's layer 7c modify over
// the other Humans you control, the Human Soldier tokens among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b122bf99-b003-452a-a854-b25597ff7ca2",
		Name:         "Jirina Kudro",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jirina Kudro — create a 1/1 white Human Soldier for each time you've cast a commander from the command zone this game",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := effects.SocB3CommanderCastsFromCommandZone(ctx.Game, ctx.Controller())
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: n}.Apply(ctx)
				}))),
		},
		Static: []game.StaticAbility{{
			Label:     "Other Humans you control get +2/+0.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.B16OtherCreaturesYouControlOfSubtype("Human"),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
			},
		}},
	})
}
