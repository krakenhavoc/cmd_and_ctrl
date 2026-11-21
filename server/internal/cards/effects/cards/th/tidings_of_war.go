package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidings of War — Sorcery {R} (The Hobbit, HOB):
//
//	"Amass Goblins 1. If this spell was cast from a graveyard, amass Goblins
//	 3 instead.
//	 Flashback {3}{R}"
//
// "Cast from a graveyard" reads CastFromZone, as Increasing Vengeance does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5753f4d6-08d0-4e12-85c5-875ee6441626",
		Name:             "Tidings of War",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if item.CastFromZone == game.ZoneGraveyard {
				n = 3
			}
			return effects.Amass{Subtype: "Goblin", N: n}.Apply(ctx)
		},
	})
}
