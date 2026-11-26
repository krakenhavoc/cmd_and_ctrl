package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hanweir Militia Captain // Westvale Cult Leader — {1}{W} Creature —
// Human Soldier 2/2 // Creature — Human Cleric */* (Shadows over Innistrad,
// #5543):
//
//	Front: "At the beginning of your upkeep, if you control four or more
//	        creatures, transform this creature."
//	Back:  "Westvale Cult Leader's power and toughness are each equal to the
//	        number of creatures you control.
//	        At the beginning of your end step, create a 1/1 white and black
//	        Human Cleric creature token."
//
// The front's "if" is an intervening if (CR 603.4): checked as the upkeep
// begins and again as the ability resolves. The back face's power and
// toughness are a characteristic-defining ability (CR 604.3, layer 7a), read
// through the layers.
//
// No simplification.
func init() {
	const oracle = "2d8eba16-cfca-4a38-bde0-6ee36c67231a"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Hanweir Militia Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control four or more creatures", func(g *game.Game, you uuid.UUID) bool {
				return effects.B04CreaturesControlled(g, you) >= 4
			}, effects.AtYourUpkeep("Hanweir Militia Captain — transform this creature", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Westvale Cult Leader",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Westvale Cult Leader's power and toughness are each equal to the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Westvale Cult Leader — create a 1/1 white and black Human Cleric creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white and black Human Cleric"), N: 1})),
		},
	})
}
