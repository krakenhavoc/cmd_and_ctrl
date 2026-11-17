package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temporal Fissure — Sorcery {4}{U} (Scourge):
//
//	"Return target permanent to its owner's hand.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Void Squall's bounce (bounceTheTarget) over any permanent, plus the storm keyword (CR
// 702.40a). Each copy may name a different permanent (CR 707.10c); a copy
// whose permanent has already left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a9525f4-74a6-44f0-a1b1-d57937d7783b",
		Name:         "Temporal Fissure",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target permanent"),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve:    bounceTheTarget,
	})
}
