package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dematerialize — Sorcery {3}{U} (Odyssey, #5613):
//
//	"Return target permanent to its owner's hand.
//	 Flashback {5}{U}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The shared bounceTheTarget body: the permanent goes to its OWNER's
// hand, whoever controls it, and a target gone by resolution does
// nothing (CR 608.2b).
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d2fbad7d-d05a-4321-9aff-a93bd102a3f4",
		Name:             "Dematerialize",
		Completeness:     CompletenessFull,
		Targets:          TargetPermanent("target permanent"),
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}{U}")},
		OnResolve:        bounceTheTarget,
	})
}
