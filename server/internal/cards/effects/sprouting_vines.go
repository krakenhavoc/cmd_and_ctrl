package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sprouting Vines — Instant {2}{G} (Scourge):
//
//	"Search your library for a basic land card, reveal that card, put it
//	 into your hand, then shuffle.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn.)"
//
// The shared reveal-and-shuffle tutor over basic land cards, plus the storm
// keyword (CR 702.40a). Each copy is its own search, so a storm count of two
// finds up to three basic lands, one at a time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a38ad1aa-fc12-49be-b9f7-fe287597a5ca",
		Name:         "Sprouting Vines",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve:    b06TutorToHand("Sprouting Vines — a basic land card", IsBasicLand),
	})
}
