package effects

import (
	"github.com/google/uuid"
)

// face_down_piles.go — "look at the top N cards of your library and
// separate them into a face-down pile and a face-up pile. Put one pile
// into your hand and the other into your graveyard" (#2147, ADR 0114
// PR 5's last open row). Sauron's Ransom, Fortune's Favor and Atris
// hand the separating to an opponent; Riddles in the Dark and Curator
// of Destinies keep it and hand the CHOICE to an opponent.
//
// Fact or Fiction's split reveals the cards. These are only looked at,
// so the honesty comes from who knows what:
//
//   - the separator is the one knower (LookAtTopOfPlayersLibraryForEffect),
//     so the split prompt reaches them and nobody else;
//   - the face-up pile is revealed when the split is answered, to the
//     whole table;
//   - the face-down pile is never revealed. The pick that follows is an
//     option pick whose second option carries it, and the wire's one
//     redaction pass (redactChoiceCards) drops it for every seat that is
//     not a knower, keeping answerable backs only for a chooser who owns
//     the cards. Every seat still reads the pile's size off its label.
//
// A library with fewer than N cards looks at what it has; an empty one
// asks nothing and the rest of the card still happens.

// faceDownPileSplit is the shared shape: `splitter` looks at the top n
// cards of the controller's library and separates them; `chooser` takes
// one pile and the other goes to the controller's graveyard.
//
// Caller holds g.mu.
func faceDownPileSplit(ctx *Context, source string, n int, splitter, chooser uuid.UUID, then func(ctx *Context, taken, left []uuid.UUID) error) error {
	owner := ctx.Controller()
	looked := ctx.Game.LookAtTopOfPlayersLibraryForEffect(splitter, owner, n)
	return PileSplit{
		Splitter:      splitter,
		Chooser:       chooser,
		Owner:         owner,
		SplitQuestion: source + " — choose the cards for the face-down pile; the rest go face up",
		PickQuestion:  source + " — take one pile into your hand; the other goes to your graveyard",
		Cards:         looked,
		FaceDown:      true,
		Then:          then,
	}.Apply(ctx)
}

// faceDownPilesToHand moves the taken pile to the controller's hand and
// the other to their graveyard, the settlement Fact or Fiction uses.
// "Taken" is the pile the chooser picked, whoever the chooser is.
func faceDownPilesToHand(ctx *Context, taken, left []uuid.UUID) error {
	return factOrFictionSettle(ctx.Game, taken, left)
}

// opponentSeparatesFaceDownPiles is "an opponent looks at the top n
// cards of your library and separates them into a face-down pile and a
// face-up pile; you take one": the splitter is the opponent, the
// chooser the controller. With no opponent left nothing is separated
// and nothing is taken, never more than printed.
func opponentSeparatesFaceDownPiles(ctx *Context, source string, n int, splitter uuid.UUID, then func(ctx *Context, taken, left []uuid.UUID) error) error {
	if splitter == uuid.Nil {
		return nil
	}
	return faceDownPileSplit(ctx, source, n, splitter, ctx.Controller(), then)
}

// youSeparateFaceDownPiles is "look at the top n cards of your library
// and separate them into a face-down pile and a face-up pile. An
// opponent chooses one of the piles": the controller looks and splits,
// and a chosen opponent picks a pile without seeing the face-down
// cards. The pile they pick is the one the controller puts into their
// hand.
func youSeparateFaceDownPiles(ctx *Context, source string, n int) error {
	// The look comes first, as printed; the opponent is named after.
	looked := ctx.Game.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), ctx.Controller(), n)
	return ChoosePlayer{
		Among:    Opponents,
		Question: source + " — choose an opponent to choose one of the piles",
		Then: func(ctx *Context) error {
			chooser := ctx.ChosenPlayer()
			if chooser == uuid.Nil {
				return nil
			}
			return PileSplit{
				Splitter:      ctx.Controller(),
				Chooser:       chooser,
				Owner:         ctx.Controller(),
				SplitQuestion: source + " — choose the cards for the face-down pile; the rest go face up",
				PickQuestion:  source + " — choose a pile; it goes into their hand and the other into their graveyard",
				Cards:         looked,
				FaceDown:      true,
				Then:          faceDownPilesToHand,
			}.Apply(ctx)
		},
	}.Apply(ctx)
}

// opponentTarget is the item's still-legal player target, or uuid.Nil
// when it has gone (the spell has fizzled and nothing is looked at).
func opponentTarget(ctx *Context) uuid.UUID {
	id, _ := firstLegalPlayerTarget(ctx)
	return id
}
