package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sauron's Ransom — Instant {3}{U}{B}:
//
//	"Choose an opponent. They look at the top four cards of your
//	 library and separate them into a face-down pile and a face-up
//	 pile. Put one pile into your hand and the other into your
//	 graveyard. The Ring tempts you."
//
// The first card on the face-down-pile split (face_down_piles.go): the
// chosen opponent is the only seat that sees all four cards, the
// face-up pile is revealed when they answer, and the caster picks a
// pile knowing only the face-up cards and the face-down pile's size.
// An empty pile is a legal split. The Ring tempts you runs after the
// piles have moved, and still runs when the library was empty.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe04a81b-413c-4bc8-bb83-29978da9a43d",
		Name:         "Sauron's Ransom",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePlayer{
				Among:    Opponents,
				Question: "Sauron's Ransom — choose an opponent to separate the top four cards of your library",
				Then: func(ctx *Context) error {
					splitter := ctx.ChosenPlayer()
					if splitter == uuid.Nil {
						// Nobody to separate them: no piles move, the
						// rest of the sentence still happens.
						return TheRingTemptsYou{}.Apply(ctx)
					}
					return opponentSeparatesFaceDownPiles(ctx, "Sauron's Ransom", 4, splitter, sauronsRansomTake)
				},
			}.Apply(ctx)
		},
	})
}

// sauronsRansomTake settles the piles and then tempts. A package-level
// function so the continuation survives an undo.
func sauronsRansomTake(ctx *Context, taken, left []uuid.UUID) error {
	if err := faceDownPilesToHand(ctx, taken, left); err != nil {
		return err
	}
	return TheRingTemptsYou{}.Apply(ctx)
}
