package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Final Frontier — Sorcery {2}{G} (Star Trek, #4973):
//
//	"Search your library for a basic land card or Planet card, put it
//	 onto the battlefield tapped, then shuffle.
//	 Beam me up {1}{G} (You may cast this card from your graveyard for
//	 {1}{G} if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// The search takes a basic land or any land card with the Planet land
// type (CR 205.3i), and may find nothing. Beam me up is trekB2BeamMeUp,
// Federation Resupply's: a graveyard cast plus the return of a creature
// you control, and the spell is exiled however it leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "26f7e903-2a00-4878-b8b7-d5abe35c9526",
		Name:             "Final Frontier",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Lands: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{trekB2BeamMeUp("{1}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return trekB20SearchBasicOrPlanetTapped(ctx, "Final Frontier — a basic land card or Planet card to put onto the battlefield tapped")
		},
	})
}
