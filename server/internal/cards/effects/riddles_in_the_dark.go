package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riddles in the Dark — Instant {2}{U}:
//
//	"Look at the top four cards of your library and separate them
//	 into a face-down pile and a face-up pile. An opponent chooses one
//	 of the piles. Put that pile into your hand and the other into your
//	 graveyard."
//
// The mirror of Sauron's Ransom (face_down_piles.go): the caster
// looks and separates, and an opponent they choose picks a pile seeing
// only the face-up cards and the face-down pile's size.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78bed291-0ec7-466f-bc7a-9dae7d0b56ae",
		Name:         "Riddles in the Dark",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return youSeparateFaceDownPiles(ctx, "Riddles in the Dark", 4)
		},
	})
}
