package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plunder the Trollshaws — Instant {1}{U} (The Hobbit, HOB):
//
//	"Draw a card. If this spell was cast from a graveyard, draw two cards
//	 instead.
//	 Flashback {3}{U}"
//
// "Cast from a graveyard" reads CastFromZone, as Increasing Vengeance does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "526f520e-b4bb-4215-9c6b-afc298744f06",
		Name:             "Plunder the Trollshaws",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if item.CastFromZone == game.ZoneGraveyard {
				n = 2
			}
			return effects.DrawCards{N: n}.Apply(ctx)
		},
	})
}
