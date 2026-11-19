package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Run Ashore — Instant {4}{U}{U} (Foundations Jumpstart):
//
//	"Choose one or both —
//	 • The owner of target nonland permanent puts it on their choice of
//	   the top or bottom of their library.
//	 • Return target nonland permanent to its owner's hand."
//
// The bullets resolve in printed order (CR 608.2c), each on its own
// target and skipped if that target has gone (CR 608.2b); both may name
// the same permanent, in which case the first bullet has already moved it
// and the second finds nothing. The first bullet is Ice Magic's Blizzara:
// the OWNER picks the lane, and a commander is offered the command zone
// (CR 903.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b3bcd57-bf3f-4d72-810a-fdb5c8ea1551",
		Name:         "Run Ashore",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("The owner of target nonland permanent puts it on their choice of the top or bottom of their library.",
				TargetPermanent("target nonland permanent", Nonland()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return PutIntoLibraryAtDepthOrBottom{Card: t.ID, Depth: 1, Label: "Run Ashore — put it on your choice of the top or bottom of your library"}.Apply(ctx)
				}),
			ModeDoing("Return target nonland permanent to its owner's hand.",
				TargetPermanent("target nonland permanent", Nonland()),
				tyrantsScornBounce),
		),
	})
}
