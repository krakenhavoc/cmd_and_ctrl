package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fact or Fiction — Instant {3}{U}:
//
//	"Reveal the top five cards of your library. An opponent separates
//	 those cards into two piles. Put one pile into your hand and the
//	 other into your graveyard."
//
// The card #568 is named for, and the one a draw player notices
// missing. Three engine pieces, all of them now present:
//
//   - the reveal (#549) makes the five cards public, which is what
//     entitles an opponent to look at cards off the top of somebody
//     else's library at all;
//   - the pile split is a PendingChoiceChooseCards addressed to that
//     opponent — floor zero, ceiling five, so an empty pile is a legal
//     and frequently correct split;
//   - the pile pick is a PendingChoiceOptionPick addressed back to the
//     controller, each option carrying its pile (#568), chained off the
//     split's answer (#552).
//
// The redaction pass is what makes the middle prompt safe: the
// splitter is being shown cards out of a library that is not theirs,
// and protocol shows them exactly the cards the reveal made public and
// nothing else (redactChoiceCards). Without the reveal they would see
// an empty prompt rather than a library.
//
// # Declared sandbox simplification: WHICH opponent separates
//
// CR gives the choice of opponent to the spell's controller in a
// multiplayer game. There is no "choose an opponent" prompt yet (the
// engine-seams row of that name), so the seat after the controller in
// turn order does the splitting. In a two-player game there is no
// difference at all.
func init() {
	Register(Spec{
		OracleID:     "437b2dab-15e0-4b9a-a204-58622d37a3b3",
		Name:         "Fact or Fiction",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With more than one opponent, the player who splits the piles is the one after you in turn order rather than an opponent you pick.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return factOrFictionReveal(ctx)
		},
	})
}

// factOrFictionReveal is the whole card: reveal five, hand the split
// to an opponent, and chain the pile pick back to the controller.
//
// Caller holds g.mu.
func factOrFictionReveal(ctx *Context) error {
	controller := ctx.Controller()
	opponents := ctx.Opponents()
	revealed := ctx.Game.RevealTopOfLibraryForEffect(controller, ctx.Source(), 5,
		"Fact or Fiction — reveal the top five cards of your library")
	if len(revealed) == 0 {
		// An empty library reveals nothing, and there is nothing to
		// separate. Not an error: revealing is not drawing.
		return nil
	}
	if len(opponents) == 0 {
		// Nobody left to separate them (CR 800.4a). The cards were
		// revealed and the spell still resolves; everything is one
		// pile and the controller takes it, which is the only
		// outcome left.
		return factOrFictionSettle(ctx.Game, revealed, nil)
	}
	return PileSplit{
		Splitter:      opponents[0],
		Chooser:       controller,
		Owner:         controller,
		SplitQuestion: "Fact or Fiction — separate these five cards into two piles",
		PickQuestion:  "Fact or Fiction — take one pile into your hand; the other goes to your graveyard",
		Cards:         revealed,
		Then:          factOrFictionTake,
	}.Apply(ctx)
}

// factOrFictionTake is the continuation both prompts lead to: the
// chosen pile goes to hand, the other to the graveyard.
//
// Caller holds g.mu.
func factOrFictionTake(ctx *Context, taken, left []uuid.UUID) error {
	return factOrFictionSettle(ctx.Game, taken, left)
}

// factOrFictionSettle moves the two piles. A card that is no longer
// where the split found it is skipped rather than erroring — the
// prompts are asynchronous and the library can move under them, and
// half a Fact or Fiction is better than a resolution that stops at
// the first missing card.
//
// Caller holds g.mu.
func factOrFictionSettle(g *game.Game, toHand, toGraveyard []uuid.UUID) error {
	for _, id := range toHand {
		if g.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := g.BounceToHandForEffect(id); err != nil {
			return err
		}
	}
	for _, id := range toGraveyard {
		if g.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := g.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
}
