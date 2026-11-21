package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lhurgoyf — Creature — Lhurgoyf {2}{G}{G}, */1+*:
//
//	"Lhurgoyf's power is equal to the number of creature cards in all
//	 graveyards and its toughness is equal to that number plus 1."
//
// A characteristic-defining ability (CR 604.3, CR 208.2a) in layer 7a,
// Mortivore's count (b41CreatureCardsInAllGraveyards) with the toughness
// one more, so it works on the battlefield and counts every graveyard's
// creature cards as they are.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fe76004-b2f4-4b11-abce-75fd76f68b4d",
		Name:         "Lhurgoyf",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the number of creature cards in all graveyards and toughness is equal to that number plus 1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := effects.B41CreatureCardsInAllGraveyards(g)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
	})
}
