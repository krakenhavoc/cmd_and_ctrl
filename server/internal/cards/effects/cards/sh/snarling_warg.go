package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snarling Warg — Creature — Wolf {3}{B}, 3/4 (The Lord of the Rings: Tales
// of Middle-earth):
//
//	"Menace
//	 As long as you control a Goblin or Orc, this creature gets +1/+0."
//
// Battle Brawler's conditional pump: a layer-7c +1/+0 on the source alone,
// read live on every layer pass. The Goblin or Orc is any permanent the
// controller controls with either creature type, the Warg itself included
// only if it were one (it is a Wolf). The type words are read through the
// source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "978a7a96-ca79-4f73-a4c6-99d662605969",
		Name:            "Snarling Warg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:    "As long as you control a Goblin or Orc, this creature gets +1/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if target.InstanceID != source.InstanceID {
					return false
				}
				goblin, orc := source.TextSubtype("Goblin"), source.TextSubtype("Orc")
				for i := range g.Battlefield.Cards {
					c := &g.Battlefield.Cards[i]
					if c.Controller == source.Controller && (c.HasSubtype(goblin) || c.HasSubtype(orc)) {
						return true
					}
				}
				return false
			},
			Apply: effects.PumpBy(1, 0),
		}},
	})
}
