package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moon-Vigil Adherents — Creature — Elf Druid {2}{G}{G} (Lorwyn Eclipsed, #4458):
//
//	"Trample
//	 This creature gets +1/+1 for each creature you control and each
//	 creature card in your graveyard."
//
// Counted on every layer pass (7c), the Adherents themselves included.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d0fa43f6-d8f6-4674-b538-6d546b7dcf40",
		Name:            "Moon-Vigil Adherents",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each creature you control and each creature card in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() })
				n += len(effects.B42CardsInGraveyardOf(g, source.Controller, func(c game.Card) bool { return c.IsCreature() }))
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
