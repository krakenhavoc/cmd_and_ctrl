package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Open Communications — Sorcery {U} (Star Trek, #4973):
//
//	"Draw a card.
//	 Beam me up {2}{U} (You may cast this card from your graveyard for
//	 {2}{U} if you also return a creature you control to its owner's hand.
//	 Then exile this spell.)"
//
// Beam me up is BeamMeUp (trek_b8_helpers.go): flashback's cast from the
// graveyard and exile, with the return of a creature you control as the
// rest of the price. The creature is returned as the cost is paid, so it
// is in hand before the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b08c5b86-4f84-46a3-877f-bbf71ec17adb",
		Name:             "Open Communications",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Draws: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{BeamMeUp("{2}{U}")},
		OnResolve:        ResolveSteps(DrawCards{N: 1}),
	})
}
