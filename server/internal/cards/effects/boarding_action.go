package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boarding Action — Sorcery {1}{R} (Star Trek, #4973):
//
//	"Create a 2/1 red Klingon Warrior creature token with haste.
//	 Beam me up {2}{R} (You may cast this card from your graveyard for
//	 {2}{R} if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// The token is the printed ttrk Klingon Warrior. Beam me up is
// trekB14BeamMeUp: a graveyard cast for {2}{R} plus the return of a
// creature you control as a cost, exiled as the spell leaves the stack,
// resolved or countered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f9f70fb6-953e-40e8-a15d-ea3ee62d09e3",
		Name:             "Boarding Action",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{trekB14BeamMeUp("{2}{R}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("2/1 red Klingon Warrior with haste"), N: 1}),
	})
}
