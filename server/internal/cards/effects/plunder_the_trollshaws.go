package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "526f520e-b4bb-4215-9c6b-afc298744f06",
		Name:             "Plunder the Trollshaws",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 1
			if item.CastFromZone == game.ZoneGraveyard {
				n = 2
			}
			return DrawCards{N: n}.Apply(ctx)
		},
	})
}
