package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earth Rift — Sorcery {3}{R} (Odyssey, #5613):
//
//	"Destroy target land.
//	 Flashback {5}{R}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Stone Rain's sentence (CR 701.8a) on the shared destroyTheTargetPermanent
// body; a target gone by resolution does nothing (CR 608.2b).
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "888a5071-24b1-48ce-a97b-39f108976c6d",
		Name:             "Earth Rift",
		Completeness:     CompletenessFull,
		Targets:          TargetPermanent("target land", Land()),
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{R}{R}")},
		OnResolve:        destroyTheTargetPermanent,
	})
}
