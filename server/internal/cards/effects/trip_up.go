package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trip Up — Instant {3}{U} (Aetherdrift, #4464):
//
//	"Target nonland permanent's owner puts it on their choice of the top
//	 or bottom of their library.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Guidance Failure's tuck (ownerTucksFirstTargetTopOrBottom) with the
// target any nonland permanent: the OWNER is asked top or bottom, not the
// caster. A target gone by resolution means the spell does nothing
// (CR 608.2b). Cycling is the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bac52ae6-24ee-4c4f-a32c-3eb9859dc025",
		Name:         "Trip Up",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ownerTucksFirstTargetTopOrBottom(ctx, "Trip Up — put it on your choice of the top or bottom of your library")
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
