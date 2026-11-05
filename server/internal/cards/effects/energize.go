package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Energize — Sorcery {1}{U} (Star Trek, #4973):
//
//	"Return target creature an opponent controls to its owner's hand.
//	 Beam me up {3}{U} (You may cast this card from your graveyard for
//	 {3}{U} if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// Beam me up is trekB6BeamMeUp: a graveyard cast for {3}{U} plus the
// return of a creature you control as a cost (named in alt_cost_ids and
// paid with the spell on the stack), with flashback's exile as the spell
// leaves the stack, resolved or countered. The body is Void Squall's
// bounce of the first still-legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "275257ea-3c72-4ca9-b2fc-e50efa071bdb",
		Name:             "Energize",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          TargetCreature("target creature an opponent controls", OpponentControls()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{trekB6BeamMeUp("{3}{U}")},
		OnResolve:        bounceTheTarget,
	})
}
