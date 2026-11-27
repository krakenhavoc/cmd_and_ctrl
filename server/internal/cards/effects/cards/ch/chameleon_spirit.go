package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chameleon Spirit — Creature — Illusion Spirit {3}{U}, */* (Mercadian
// Masques, #5619):
//
//	"As this creature enters, choose a color.
//	 Chameleon Spirit's power and toughness are each equal to the number of
//	 permanents of the chosen color your opponents control."
//
// Prism Ring's as-enters colour (the colour selects among the opposition's
// permanents, so it is declared a filter) and a layer 7a characteristic-
// defining ability (CR 604.3) in the shape of Avalanche of Sector 7's count
// over permanents controlled by other players. Before a colour is chosen, or
// with no permanents of it, it is 0/0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ae547df-d280-41fb-81b4-24e52d60c25b",
		Name:         "Chameleon Spirit",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForFilter, "Chameleon Spirit"),
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the number of permanents of the chosen color your opponents control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if source.ChosenColor != "" {
					for _, p := range g.Battlefield.Cards {
						if p.Controller != source.Controller && p.HasColor(source.ChosenColor) {
							n++
						}
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
