package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "5753f4d6-08d0-4e12-85c5-875ee6441626",
		Name:             "Tidings of War",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 1
			if item.CastFromZone == game.ZoneGraveyard {
				n = 3
			}
			return Amass{Subtype: "Goblin", N: n}.Apply(ctx)
		},
	})
}
