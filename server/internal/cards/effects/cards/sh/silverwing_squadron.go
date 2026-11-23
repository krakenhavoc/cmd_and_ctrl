package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverwing Squadron — Creature — Human Knight {5}{W}, */*:
//
//	"Flying, vigilance
//	 Silverwing Squadron's power and toughness are each equal to the
//	 number of creatures you control.
//	 Whenever this creature attacks, create a number of 2/2 white Knight
//	 creature tokens with vigilance equal to the number of opponents you
//	 have."
//
// The size is a characteristic-defining ability (CR 604.3, layer 7a) that
// counts the Squadron itself. The tokens enter untapped and not attacking
// (the clause says only "create"); the count is the opponents still in
// the game when the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba968a1e-c946-4cbe-9727-160c73cde3f4",
		Name:            "Silverwing Squadron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Label:    "Power and toughness are each equal to the number of creatures you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B04CreaturesControlled(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Silverwing Squadron — create a 2/2 white Knight with vigilance for each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: len(ctx.Opponents())}.Apply(ctx)
				}),
		},
	})
}
