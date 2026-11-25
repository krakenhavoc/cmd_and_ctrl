package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rowan's Battleguard — Creature — Human Knight {3}{R}, 3/3 (Throne of
// Eldraine, #5516):
//
//	"First strike
//	 As long as you control a Rowan planeswalker, this creature gets +3/+0."
//
// A layer 7c static re-read on every recompute; "Rowan" is a planeswalker
// type, matched on the permanent's subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce07104c-8573-4dc0-a492-f9548d4317d2",
		Name:            "Rowan's Battleguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{{
			Label:    "As long as you control a Rowan planeswalker, this creature gets +3/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfCreature(target, g, source) &&
					effects.CountPermanentsYouControl(g, source, func(o *game.Card) bool {
						return o.IsPlaneswalker() && o.HasSubtype("Rowan")
					}) > 0
			},
			Apply: effects.PumpBy(3, 0),
		}},
	})
}
