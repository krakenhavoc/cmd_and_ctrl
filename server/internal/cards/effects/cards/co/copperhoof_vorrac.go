package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Copperhoof Vorrac — Creature — Boar Beast {3}{G}{G}, 2/2 (Mirrodin,
// #5606):
//
//	"This creature gets +1/+1 for each untapped permanent your opponents
//	 control."
//
// Faeburrow Elder's self-scaling layer 7c modification, counting every
// untapped permanent a player other than the Vorrac's controller
// controls, lands and tokens included. Tapping and untapping bump the
// layers, so the size follows each tap as it happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5fed574-6ac3-4318-b0a6-27082e07f3bd",
		Name:         "Copperhoof Vorrac",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each untapped permanent your opponents control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller != source.Controller && !p.Tapped {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
