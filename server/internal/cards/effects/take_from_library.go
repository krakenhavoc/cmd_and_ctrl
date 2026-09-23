package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// take_from_library.go — "put it into your hand", where "it" is a card
// NAMED out of a look at or a reveal from the top of a library (#952).
//
// The sibling of put_from_library.go, clause for clause, and split the
// same way. The MOVE is engine:
// game.TakeFromLibraryToHandThenForEffect routes library -> hand like
// every other zone change, so the CR 614 window opens, a commander is
// offered CR 903.9, and the continuation is told what ARRIVED. The
// PICK is card text — which of the looked-at cards qualify, "a" or
// "any number", "you may", whether the taken cards are revealed, and
// where the rest go — and it lives here, on a choose_cards prompt with
// Zone: ZoneLibrary.
//
// Hidden information is the caller's first line, not this file's. A
// card that LOOKS calls game.LookAtTopOfLibraryForEffect (only the
// looker becomes a knower); a card that REVEALS the whole top calls
// game.RevealTopOfLibraryForEffect (every seat does). Horn of the Mark
// does both in one sentence — it LOOKS at five and REVEALS only the
// card it takes — which is what TakeFromLibraryToHand.Reveal is for.
//
// Why this exists rather than one more BounceToHand: the issue is in
// game/library_to_hand.go's header. In short, "return it to its
// owner's hand" happening to move a card out of a library is an
// accident of the zone router, and Goblin Ringleader shipped on it.

// TakeFromLibraryToHand offers `Player` the cards in Cards that pass
// Match and puts the chosen ones into their owner's hand.
//
// # The shapes a card prints
//
//   - "Put all Goblin cards revealed this way into your hand"
//     (Goblin Ringleader): All. No prompt — there is no choice.
//   - "You may reveal a creature card from among them and put it into
//     your hand" (Horn of the Mark): Optional, Max 1, Reveal.
//   - "Put any number of them into your hand" : Optional, Max 0,
//     meaning no ceiling.
//
// A mandatory clause whose only legal answer is every candidate is not
// asked either: a prompt with one possible answer is a click, not a
// choice.
//
// # It queues; it does not finish
//
// Apply returns once the prompt is queued. Anything sequenced after it
// in the same Do(…) runs BEFORE the player answers, so "put the rest
// on the bottom of your library in a random order" — which must know
// which cards were NOT taken — goes in Then. Then runs in every case:
// after the answer, immediately when there was nothing to take,
// immediately when All or a forced answer made the choice, and even
// when the move itself returned an error, so the rest are never left
// on top because one card failed. (On that last path `Taken` is empty
// rather than partial — a route that fails a leg never reports what
// landed — while `Rest` is read off the live board and is right
// regardless, which is what "the rest" needs.)
type TakeFromLibraryToHand struct {
	// Player chooses, and is the player taking the cards. uuid.Nil
	// means the resolving item's controller. The cards go to their
	// OWNER's hand, which is the same player for every printed case
	// and is the rule when it is not.
	Player uuid.UUID

	// Cards are the looked-at or revealed cards, top card first — the
	// "them" of "from among them". A card that has since left its
	// library is ignored.
	Cards []uuid.UUID

	// Match is the clause's filter ("a creature card", "Goblin
	// cards"). Nil means any card. Tokens are never candidates (CR
	// 108.2) whatever Match says.
	Match CardPredicate

	// Max is "a" (1) or "any number" (0). Ignored with All.
	Max int

	// Optional is the printed "you MAY". It sets the prompt's floor to
	// zero.
	Optional bool

	// All is "put ALL [matching] cards … into your hand": no prompt.
	All bool

	// Reveal is the printed "you may REVEAL a creature card from among
	// them and put it into your hand": the cards that are TAKEN become
	// public, while the rest of the look stays private to the looker.
	// It is revealed before the move, while the cards are still in the
	// library, which is where the reveal happens and where the rest of
	// the engine's reveals of library cards happen.
	Reveal bool

	// Validate is the clause's rule about the picked SET, as opposed
	// to Match's rule about each card — "any number of cards with
	// different names", "up to two cards with total mana value 4 or
	// less". Nil means the bounds and Match are the whole rule,
	// which is every card in the catalog that prints this sentence
	// today.
	//
	// It is game.ChooseCardsPrompt.Validate forwarded verbatim, with
	// that field's whole contract, and it is the same field
	// PutFromLibraryOntoBattlefield carries — see there for the long
	// version, including why it switches off the forced-answer
	// shortcut below and why All ignores it.
	Validate func([]game.Card) bool

	// Label is the prompt header, "<card> — <the printed clause>".
	Label string

	// Then runs after the move with what happened. It is the only
	// correct place for "the rest" and for any rider on the result. It
	// runs with g.mu held, may queue further prompts, and must capture
	// only scalars.
	Then func(g *game.Game, res TakeFromLibraryResult) error
}

// TakeFromLibraryResult is what Then is told about the take.
type TakeFromLibraryResult struct {
	// Source is the card whose clause this was.
	Source uuid.UUID
	// Player is who was asked.
	Player uuid.UUID
	// Taken are the cards that reached a hand, in the order given.
	// Empty when none did: a declined "you may", no candidate, or a
	// commander that took CR 903.9's offer instead.
	Taken []uuid.UUID
	// Rest are the cards of Cards that are still in a library — "the
	// rest", "the cards revealed this way that weren't put into your
	// hand" — in the order Cards gave them.
	Rest []uuid.UUID
}

func (p TakeFromLibraryToHand) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source := ctx.Source()
	cards := append([]uuid.UUID(nil), p.Cards...)
	then := p.Then
	reveal, label := p.Reveal, p.Label

	finish := func(g *game.Game, picked []uuid.UUID) error {
		reported := false
		report := func(g *game.Game, taken []uuid.UUID) error {
			reported = true
			if then == nil {
				return nil
			}
			return then(g, TakeFromLibraryResult{
				Source: source,
				Player: player,
				Taken:  taken,
				Rest:   cardsStillInALibrary(g, cards),
			})
		}
		if len(picked) == 0 {
			return report(g, nil)
		}
		if reveal {
			// The taken cards become public while they are still in
			// the library — a reveal is not a move, and doing it after
			// the move would be a reveal of cards in a hand.
			g.RevealForEffect(game.RevealSpec{
				Player: player,
				Source: source,
				Reason: label,
				Cards:  cardsStillInALibrary(g, picked),
			})
		}
		moveErr := g.TakeFromLibraryToHandThenForEffect(player, picked, report)
		if moveErr != nil && !reported {
			// A leg failed and the route returned before reaching its
			// continuation, so "the rest" has not been dealt with.
			// Run it anyway: the looked-at cards must not be left on
			// TOP of the library — with the looker still a knower of
			// them — because one card of the batch failed. This is the
			// rule put_from_library.go states and keeps, and the doc
			// above promises.
			//
			// Taken is empty rather than partial: the route reports
			// what landed only through the continuation it did not
			// reach. Rest is read off the live board, so it is right
			// either way, which is what the rest leg actually needs.
			thenErr := report(g, nil)
			_ = thenErr // the move's error is the one worth returning
		}
		return moveErr
	}

	candidates := libraryCardsTakeable(ctx.Game, player, cards, p.Match)
	if len(candidates) == 0 {
		return finish(ctx.Game, nil)
	}
	if p.All {
		return finish(ctx.Game, candidates)
	}
	hi := p.Max
	if hi <= 0 || hi > len(candidates) {
		hi = len(candidates)
	}
	lo := hi
	if p.Optional {
		lo = 0
	}
	if lo == len(candidates) && p.Validate == nil {
		// The only legal answer is every candidate. Not so with a
		// set rule: it is the rule, not the count, that decides
		// which subsets are answers, and taking the shortcut would
		// perform a set the prompt would have refused.
		return finish(ctx.Game, candidates)
	}
	question := label
	if question == "" {
		question = "Put a card from among them into your hand"
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: question,
		Cards:    candidates,
		Min:      lo,
		Max:      hi,
		// Re-checked on submit: every pick must still be in a library
		// when the answer arrives.
		Zone:     game.ZoneLibrary,
		Validate: p.Validate,
		Then:     finish,
	})
	return nil
}

// libraryCardsTakeable filters `ids` to the cards still in a library
// that pass `pred`, in the order given. Characteristics are read off
// the card in the library, where they are printed.
//
// A token is never a candidate: it is not a "card" (CR 108.2), and a
// token that has left the battlefield cannot be put into a hand (CR
// 111.8). One only reaches a library because the engine has no CR
// 704.5d sweep.
//
// libraryCardsMatching's twin, and separate from it on purpose: that
// one drops every NONPERMANENT card, because "put it onto the
// battlefield" can only mean a permanent. A hand takes anything.
//
// Caller holds g.mu.
func libraryCardsTakeable(g *game.Game, player uuid.UUID, ids []uuid.UUID, pred CardPredicate) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.IsToken() {
			continue
		}
		if pred != nil && !pred(g, player, c) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// TakeRestOnBottomInRandomOrder is the Then for "put the rest on the
// bottom of your library in a random order" — Horn of the Mark.
//
// Only for cards that PRINT "a random order". A card that prints "in
// any order" uses TakeRestOnBottomInAnyOrder (library_order.go), which
// asks the player (#996, ADR 0088).
func TakeRestOnBottomInRandomOrder(g *game.Game, res TakeFromLibraryResult) error {
	return g.PutOnBottomInRandomOrderForEffect(res.Player, game.ZoneLibrary, res.Rest)
}

// TakeRestIntoGraveyard is the Then for "put the rest into your
// graveyard". Not a mill (CR 701.17a is a count off the TOP), so it
// emits an ordinary zone move — and it goes through
// game.PutIntoGraveyardForEffect, which routes, so Rest in Peace,
// Leyline of the Void and CR 903.9 all see the arrival.
//
// A token among the rest stays where it is (CR 111.8). An error on one
// card does not keep the others out of the graveyard: the loop carries
// on and the first error is returned.
func TakeRestIntoGraveyard(g *game.Game, res TakeFromLibraryResult) error {
	return restIntoGraveyard(g, res.Rest)
}

// LookAtTopThenMayTakeToHand is the whole sentence: "look at the top N
// cards of your library. You may reveal a [Match] card from among them
// and put it into your hand. Put the rest on the bottom of your library
// in a random order" — Horn of the Mark.
//
// A LOOK, so only the looker sees the five; the card taken is revealed,
// which is the half the table is entitled to. Max is 1 for "a" and 0
// for "any number".
func LookAtTopThenMayTakeToHand(n int, match CardPredicate, max int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player:   player,
			Cards:    g.LookAtTopOfLibraryForEffect(player, n),
			Match:    match,
			Max:      max,
			Optional: true,
			Reveal:   true,
			Label:    label,
			Then:     TakeRestOnBottomInRandomOrder,
		}.Apply(ctx)
	}
}

// RevealTopThenTakeToHand is the whole sentence for the public
// version: "reveal the top N cards of your library. Put all [Match]
// cards revealed this way into your hand and the rest on the bottom of
// your library in any order" — Goblin Ringleader, Sylvan Messenger,
// Garruk, Caller of Beasts. The rest are ordered by the player (#996).
//
// Mandatory and unbounded, which is why it raises no prompt at all:
// "all of them" is the only answer.
func RevealTopThenTakeToHand(ctx *Context, player uuid.UUID, n int, match CardPredicate, label string) error {
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), n, label),
		Match:  match,
		All:    true,
		Label:  label,
		Then:   TakeRestOnBottomInAnyOrder,
	}.Apply(ctx)
}
