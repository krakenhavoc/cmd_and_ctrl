package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selvala's Stampede — Sorcery {4}{G}{G}:
//
//	"Council's dilemma — Starting with you, each player votes for wild
//	 or free. Reveal cards from the top of your library until you
//	 reveal a creature card for each wild vote. Put those creature
//	 cards onto the battlefield, then shuffle the rest into your
//	 library. You may put a permanent card from your hand onto the
//	 battlefield for each free vote."
//
// A council's dilemma (CR 701.38, ADR 0146): the tally is read per
// option, and both halves scale with it.
//
//   - Wild. The top of the library is revealed until the Nth creature
//     card, N being the wild votes (revealUntilCount: the same
//     reveal-until the single-hit family uses, with a count). Every
//     creature card revealed enters together, then the library is
//     shuffled, which sends the rest of the revealed cards back into it
//     at random. A library that runs out first reveals all of itself and
//     puts the creatures it had. With no wild vote nothing is revealed,
//     and the library is not shuffled either: there is no "rest" to
//     shuffle into it.
//   - Free. The controller may put up to one permanent card from hand
//     per free vote onto the battlefield, all at once
//     (PutFromHandOntoBattlefield.Max, one batch: every entry is judged
//     against the board before any of them arrived). A hand with fewer
//     permanent cards than free votes is asked about the cards it has.
//
// The free half follows the wild half, as printed, and waits for it: a
// creature that enters and asks its own question is answered first.
//
// A bot casting it votes wild; an opponent's bot votes free, which asks
// for a card in a hand that may well have none worth putting down.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d2aea05-f55c-4355-950f-bfdc4e985e80",
		Name:         "Selvala's Stampede",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Selvala's Stampede — vote for wild or free",
				Words:         []string{"wild", "free"},
				ForController: []int{2, 1},
				ForOpponents:  []int{1, 2},
				Then:          selvalasStampedeVoted,
			}.Apply(ctx)
		},
	})
}

var selvalasStampedeVoted = VoteResultThen("vote/selvalas-stampede", func(ctx *Context, r game.VoteResult) error {
	wild, free := r.Votes(0), r.Votes(1)
	if wild < 1 {
		return selvalasStampedeFree(ctx, free)
	}
	player := ctx.Controller()
	run, hits := revealUntilCount(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		wild, "Selvala's Stampede — reveal until a creature card for each wild vote")
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return selvalasStampedeIsHit(hits, c.InstanceID) },
		All:      true,
		ThenTail: selvalasStampedeShuffleThenFree,
		ThenN:    free,
	}.Apply(ctx)
})

// selvalasStampedeIsHit reports whether id is one of the creature cards
// the reveal stopped on.
func selvalasStampedeIsHit(hits []uuid.UUID, id uuid.UUID) bool {
	for _, h := range hits {
		if h == id {
			return true
		}
	}
	return false
}

// selvalasStampedeFree is "you may put a permanent card from your hand
// onto the battlefield for each free vote": up to n, at once.
func selvalasStampedeFree(ctx *Context, n int) error {
	if n < 1 {
		return nil
	}
	return PutFromHandOntoBattlefield{
		Optional: true,
		Max:      n,
		Label:    "Selvala's Stampede — you may put a permanent card from your hand onto the battlefield for each free vote",
	}.Apply(ctx)
}

// selvalasStampedeShuffleThenFree is the rider as a keyed tail (#4253
// slice 6): shuffle the rest in, then offer the free votes' hand puts
// (N). An on-disk key.
var selvalasStampedeShuffleThenFree = PutFromLibraryTail("selvalas-stampede/shuffle-then-free",
	func(ctx *Context, res PutFromLibraryResult) error {
		g := ctx.Game
		// "Then shuffle the rest into your library": the rest are
		// still in it, so shuffling is putting them in.
		if err := g.ShuffleLibraryForEffect(res.Player); err != nil {
			return err
		}
		// Rebuilt from the live game: this runs after any entry
		// question, and undo restores a clone.
		return selvalasStampedeFree(NewContext(g, &game.StackItem{
			Kind:         game.StackItemSpell,
			Controller:   res.Player,
			Owner:        res.Player,
			SourceCardID: res.Source,
		}), res.N)
	})
