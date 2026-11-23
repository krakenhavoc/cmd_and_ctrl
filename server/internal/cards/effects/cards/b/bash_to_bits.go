package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bash to Bits — Instant {3}{R} (Odyssey, #5613):
//
//	"Destroy target artifact.
//	 Flashback {4}{R}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Shatter's sentence (CR 701.8a) on the shared destroyTheTargetPermanent
// body; a target gone by resolution does nothing (CR 608.2b).
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "894b8375-5d55-4138-8448-c38f83a17e4f",
		Name:             "Bash to Bits",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetPermanent("target artifact", effects.Artifact()),
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{R}{R}")},
		OnResolve:        effects.DestroyTheTargetPermanent,
	})
}
