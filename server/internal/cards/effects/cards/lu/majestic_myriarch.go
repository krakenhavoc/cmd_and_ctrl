package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Majestic Myriarch — {4}{G} Creature — Chimera */* (Commander 2020):
//
//	"Majestic Myriarch's power and toughness are each equal to twice the
//	 number of creatures you control.
//	 At the beginning of each combat, this creature gains flying until end of
//	 turn if you control a creature with flying. The same is true for first
//	 strike, double strike, deathtouch, haste, hexproof, indestructible,
//	 lifelink, menace, reach, trample, and vigilance."
//
// The first line is Battle Squadron's characteristic-defining ability at
// twice the count. The second is one trigger at the start of every player's
// combat: as it resolves, each keyword of the list that a creature the
// controller controls has (read through the layers, the Myriarch itself
// included) is granted to the Myriarch until end of turn. One grant of the
// whole list, since the keywords are read together at resolution (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c39dca94-3849-4edb-bba9-e7a2eefa5021",
		Name:         "Majestic Myriarch",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Majestic Myriarch's power and toughness are each equal to twice the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 2 * effects.B04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat("Majestic Myriarch — gains each keyword a creature you control has until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					var gained []string
					for _, kw := range majesticMyriarchKeywords {
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == item.Controller && c.IsCreature() && game.HasKeyword(&c, kw) {
								gained = append(gained, kw)
								break
							}
						}
					}
					if len(gained) == 0 {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: gained, Label: "Majestic Myriarch — keywords from creatures you control"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var majesticMyriarchKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "haste", "hexproof",
	"indestructible", "lifelink", "menace", "reach", "trample", "vigilance",
}
