package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turtles Forever — Instant {3}{W} (#4681, found during #4457):
//
//	"Search your library and/or outside the game for exactly four
//	 legendary creature cards you own with different names, then reveal
//	 those cards. An opponent chooses two of them. Put the chosen cards
//	 into your hand and shuffle the rest into your library."
//
// One search over two places, asked as the prompts the engine already
// has, in this order:
//
//  1. Where to look: the library and outside the game, the library
//     only, or outside the game only ("and/or" is the searcher's
//     choice, asked before anything is shown, as #4519 asks "library
//     and/or graveyard"). Looking through the library is searching it
//     (CR 701.23a), so the library is searched, and "whenever a player
//     searches" sees it, only when the player chose it.
//  2. The library half: an ordinary library search for up to four
//     legendary creature cards with different names. A library is a
//     hidden zone, so the player may find fewer than four, or none,
//     however many are there (CR 701.23b): "exactly four" is the most
//     they find, not a floor.
//  3. The outside half: a pick of the cards the player owns outside the
//     game (their sideboard, or the card database for a deck with none,
//     ADR 0191 and ADR 0198), up to four minus what the library gave,
//     whose names differ from each other and from the library's finds.
//     Searching outside the game the player "may choose" a card (CR
//     701.23j), so this half may come up short too.
//  4. Every card found is revealed (CR 701.20a), the caster names the
//     opponent who chooses (skipped when there is one), and that
//     opponent chooses two (fewer when fewer were found, CR 609.3).
//  5. The chosen cards go to the hand, from wherever they were found.
//     The rest go into the library, the ones from outside the game
//     included, and the library is shuffled, even when nothing was found
//     (CR 701.24d: a set of no objects is still shuffled in).
//
// The library's finds wait on top of the library (the search's ToTop,
// Gifts Ungiven's shape) until the opponent has chosen, which is where
// they would be in paper, set aside and unshuffled. The outside half is
// asked after the library half because its names must differ from the
// library's: the found names ride the outside pick's word (MatchArg), so
// the filter that lists and checks it, on the sideboard and on the
// database, reads them after a restore too.
//
// Every prompt is keyed, so a table waiting on any of them is a restore
// point. Nothing is stronger than printed.
func init() {
	Register(Spec{
		OracleID:     "cff5aca2-72c5-4053-a32c-c0bd3174c52d",
		Name:         "Turtles Forever",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return turtlesForeverResolve(ctx)
		},
	})
}

// turtlesForeverMax is "exactly four".
const turtlesForeverMax = 4

// The zone question's options, in the order offered. The first is the
// always-legal one the enumerator relies on.
const (
	turtlesForeverBoth = iota
	turtlesForeverLibraryOnly
	turtlesForeverOutsideOnly
)

// turtlesForeverResolve asks where the caster searches.
//
// Caller holds g.mu.
func turtlesForeverResolve(ctx *Context) error {
	return PickOption{
		Question: "Turtles Forever — search your library and/or outside the game for four legendary creature cards with different names",
		Options: []game.ChoiceOption{
			turtlesForeverBoth:        {Label: "Search your library and outside the game"},
			turtlesForeverLibraryOnly: {Label: "Search only your library"},
			turtlesForeverOutsideOnly: {Label: "Search only outside the game"},
		},
		ThenKey: turtlesForeverZones,
	}.Apply(ctx)
}

// turtlesForeverZones is the answer to "where do you search?". An on-disk
// key. Nobody answered (the caster left the game) does nothing.
var turtlesForeverZones = OptionPickThen("option-pick/turtles-forever-zones",
	func(ctx *Context, r game.OptionPicked) error {
		switch r.Index {
		case turtlesForeverBoth, turtlesForeverLibraryOnly:
			alsoOutside := 0
			if r.Index == turtlesForeverBoth {
				alsoOutside = 1
			}
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player: ctx.Controller(),
				Source: ctx.Source(),
				Pred:   MatchLegendaryCreature,
				Dest:   game.ZoneLibrary,
				Limit:  turtlesForeverMax,
				// The reveal is of every card found, the library's and
				// outside the game's together, so it waits for both.
				Reveal:  false,
				Shuffle: false,
				// Gifts Ungiven's shape: ToTop makes the search report what
				// it found, and the shuffle at the end wipes the order.
				ToTop:    true,
				Reason:   "Turtles Forever — search your library for up to four legendary creature cards with different names",
				Rule:     searchRuleDifferentNames,
				ThenTail: ctx.TailN(turtlesForeverLibraryFound, alsoOutside),
			})
		case turtlesForeverOutsideOnly:
			return turtlesForeverAskOutside(ctx.Game, ctx.Controller(), ctx.Source(), nil)
		}
		return nil
	})

// turtlesForeverLibraryFound is the library search's continuation, told
// the cards it found; Tail.N is 1 when the caster also searches outside
// the game. An on-disk key.
var turtlesForeverLibraryFound = CardTail("turtles-forever/library-found", func(ctx *Context, o game.TailOutcome) error {
	found := turtlesForeverInLibrary(ctx.Game, ctx.Controller(), o.IDs)
	if o.Tail.N == 1 {
		return turtlesForeverAskOutside(ctx.Game, ctx.Controller(), ctx.Source(), found)
	}
	return turtlesForeverReveal(ctx.Game, ctx.Controller(), ctx.Source(), found, nil)
})

// turtlesForeverAskOutside asks for the outside half: as many cards as
// the library left room for, none sharing a name with the library's
// finds. With no room, a closed pool or nothing that matches, nothing is
// asked and the found cards are revealed at once.
//
// Caller holds g.mu.
func turtlesForeverAskOutside(g *game.Game, player, source uuid.UUID, fromLibrary []uuid.UUID) error {
	room := turtlesForeverMax - len(fromLibrary)
	if room > 0 && g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:   player,
		Source:   source,
		Question: "Turtles Forever — choose legendary creature cards you own from outside the game with different names",
		Min:      0,
		Max:      room,
		Keyed:    turtlesForeverOutside,
		MatchArg: turtlesForeverTakenNames(g, fromLibrary),
		Carry:    fromLibrary,
	}) {
		return nil
	}
	return turtlesForeverReveal(g, player, source, fromLibrary, nil)
}

// turtlesForeverOutside is the outside half, keyed: legendary creature
// cards whose names are not among the library's finds (the word), with
// different names from each other (the set rule). Its continuation is
// handed the library's finds as the carry. An on-disk key.
var turtlesForeverOutside = game.RegisterOutsidePickRule("choose-cards/turtles-forever-outside",
	turtlesForeverOutsideMatch, DifferentNames, func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered {
			// The caster left the game with the prompt open.
			return nil
		}
		fromLibrary := turtlesForeverInLibrary(g, r.Chooser, r.Carry)
		return turtlesForeverReveal(g, r.Chooser, r.Source, fromLibrary, g.StillOutsideForEffect(r.Chooser, r.Picked))
	})

// turtlesForeverOutsideMatch is the outside half's filter: a legendary
// creature card whose name is not one of `taken`, the library's finds as
// turtlesForeverTakenNames wrote them.
func turtlesForeverOutsideMatch(c game.Card, taken string) bool {
	if !MatchLegendaryCreature(c) {
		return false
	}
	name := turtlesForeverNameKey(c.Name)
	for _, t := range strings.Split(taken, "\n") {
		if t != "" && t == name {
			return false
		}
	}
	return true
}

// turtlesForeverTakenNames is the names of `ids` (cards in a library),
// one per line, compared the way DifferentNames compares them.
//
// Caller holds g.mu.
func turtlesForeverTakenNames(g *game.Game, ids []uuid.UUID) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok {
			if key := turtlesForeverNameKey(c.Name); key != "" {
				names = append(names, key)
			}
		}
	}
	return strings.Join(names, "\n")
}

// turtlesForeverNameKey is a name as DifferentNames compares it.
func turtlesForeverNameKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// turtlesForeverInLibrary is the cards of `ids` still in `player`'s
// library, in order: a found card that has gone elsewhere (an Opposition
// Agent's exile) is no longer one of "those cards".
//
// Caller holds g.mu.
func turtlesForeverInLibrary(g *game.Game, player uuid.UUID, ids []uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	var out []uuid.UUID
	for _, id := range ids {
		if p.Library.Contains(id) {
			out = append(out, id)
		}
	}
	return out
}

// turtlesForeverReveal reveals everything found and has the caster name
// the opponent who chooses. Nothing found shuffles the library and ends;
// nobody to choose leaves every found card unchosen.
//
// Caller holds g.mu.
func turtlesForeverReveal(g *game.Game, player, source uuid.UUID, fromLibrary, fromOutside []uuid.UUID) error {
	if len(fromLibrary)+len(fromOutside) == 0 {
		return g.ShuffleLibraryForEffect(player)
	}
	const reason = "Turtles Forever — reveal the legendary creature cards found"
	g.RevealForEffect(game.RevealSpec{Player: player, Source: source, Reason: reason, Cards: fromLibrary})
	g.RevealOutsideForEffect(game.RevealSpec{Player: player, Source: source, Reason: reason, Cards: fromOutside})
	found := append(append([]uuid.UUID(nil), fromLibrary...), fromOutside...)
	ctx := keyedContext(g, player, source)
	// The caster rides first on the carry (OptionPickThen's contract), the
	// revealed cards after.
	if g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  player,
		Among:    Opponents.seats(ctx, player),
		Source:   source,
		Question: "Turtles Forever — choose an opponent to choose two of the revealed cards",
		ThenKey:  turtlesForeverOpponent,
		Carry:    append([]uuid.UUID{player}, found...),
	}) != uuid.Nil {
		return nil
	}
	return turtlesForeverSettle(g, player, source, nil, found)
}

// turtlesForeverOpponent is the answer to "which opponent chooses?", the
// revealed cards riding as the carry. An on-disk key.
var turtlesForeverOpponent = OptionPickThen("option-pick/turtles-forever-opponent",
	func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil || r.Option.Player == uuid.Nil {
			return turtlesForeverSettle(ctx.Game, ctx.Controller(), r.Source, nil, r.Carry)
		}
		asked, err := ctx.Game.RevealPickThenForEffect(game.RevealPickPrompt{
			Chooser:  r.Option.Player,
			Owner:    ctx.Controller(),
			Source:   r.Source,
			Question: "Turtles Forever — choose two of these cards; they go into their hand and the rest are shuffled into their library",
			Cards:    r.Carry,
			Min:      2,
			Max:      2,
			ThenKey:  turtlesForeverTake,
		})
		if err != nil || asked > 0 {
			return err
		}
		return turtlesForeverSettle(ctx.Game, ctx.Controller(), r.Source, nil, r.Carry)
	})

// turtlesForeverTake is the opponent's choice: the chosen cards and the
// rest. An on-disk key.
var turtlesForeverTake = game.RegisterChooseCardsThen("choose-cards/turtles-forever-take",
	func(g *game.Game, r game.CardsChosen) error {
		_, left := revealPartition(r.Candidates, r.Picked)
		return turtlesForeverSettle(g, r.From, r.Source, r.Picked, left)
	})

// turtlesForeverSettle is the last sentence: the chosen cards into the
// hand and the rest into the library, then the shuffle. A card is moved
// from where it is now, outside the game or in the library; one that is
// in neither any more is not moved.
//
// Caller holds g.mu.
func turtlesForeverSettle(g *game.Game, player, source uuid.UUID, chosen, rest []uuid.UUID) error {
	outside := map[uuid.UUID]bool{}
	for _, id := range g.StillOutsideForEffect(player, append(append([]uuid.UUID(nil), chosen...), rest...)) {
		outside[id] = true
	}
	var toHand, restOutside []uuid.UUID
	for _, id := range chosen {
		if outside[id] {
			if err := g.BringFromOutsideForEffect(player, id); err != nil {
				return err
			}
			continue
		}
		toHand = append(toHand, id)
	}
	for _, id := range rest {
		if outside[id] {
			restOutside = append(restOutside, id)
		}
	}
	if len(restOutside) > 0 {
		if _, err := g.PutFromOutsideIntoLibraryForEffect(player, restOutside); err != nil {
			return err
		}
	}
	// The library's chosen cards leave through the one exit path, which
	// may pause (a commander's CR 903.9 offer, an "instead" replacement);
	// the shuffle is its keyed tail.
	return g.BounceCardsToHandKeyedForEffect(turtlesForeverInLibrary(g, player, toHand),
		game.KeyedTail{Ref: thenShuffle, Player: player, Source: source})
}
