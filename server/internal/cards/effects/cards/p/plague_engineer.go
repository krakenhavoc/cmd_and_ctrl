package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plague Engineer — Creature — Phyrexian Carrier {2}{B}, 2/2:
//
//	"Deathtouch
//	 As this creature enters, choose a creature type.
//	 Creatures of the chosen type your opponents control get -1/-1."
//
// The type is the permanent's NamedTribe (Cover of Darkness' as-enters
// prompt). The -1/-1 is a layer 7c modification over every creature of that
// type, changelings included, that an opponent of the Engineer's controller
// controls; your own creatures of the type are untouched. Until the type is
// chosen it shrinks nothing. It follows the Engineer's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df130616-8efe-42ed-a7f0-b58bef3f11c4",
		Name:            "Plague Engineer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		AsEnters:        effects.ChooseCreatureTypeAsEnters("Plague Engineer — choose a creature type"),
		Static: []game.StaticAbility{{
			Label:    "Creatures of the chosen type your opponents control get -1/-1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller != source.Controller &&
					effects.TribeFilter{Chosen: true}.Matches(target, g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
				c.Toughness--
			},
		}},
	})
}
