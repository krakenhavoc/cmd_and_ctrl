package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// --- #1831: a choose_cards prompt over the bot's own library ---------

// withLibrary puts these cards on top of the seat's library, as the
// seat's own filtered view carries them after a look: every card of the
// library is listed by instance ID, and the looked-at ones are known.
func withLibrary(cards ...protocol.CardView) seatOpt {
	return func(p *protocol.PlayerView) {
		p.Library = protocol.ZoneView{Kind: "library", Owner: p.ID, Count: 60, Cards: cards}
	}
}

// libraryChoice is a choose_cards prompt over seat 0's own library —
// the shape effects.TakeFromLibraryToHand, PutFromLibraryOntoBattlefield
// and hideaway queue.
func libraryChoice(id, reason string, lo, hi int, opts ...protocol.CardView) protocol.PendingChoiceView {
	return protocol.PendingChoiceView{
		ID: id, Kind: "choose_cards", Chooser: seatID(0).String(),
		FromPlayer: seatID(0).String(), Reason: reason,
		Options: opts, ChooseMin: lo, ChooseMax: hi, Count: hi,
	}
}

func ids(xs ...string) map[string]any { return map[string]any{"card_ids": xs} }

// #1831. Every library-take prompt scored a flat 0.5, the enumerator
// offers "take nothing" first for a "you may", and the selection loop
// keeps the first maximum — so the bot threw away Horn of the Mark's
// card, and Explore the Vastlands' two, every time. Naming a card in
// its own library is TAKING it, so the answer is scored by what it
// takes. The winning answer is never the enumerator's first or last
// offer here, so a test that passed on ordering would fail.
func TestLibraryTakeTakesTheBestCard(t *testing.T) {
	const choiceID = "choice-look"
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	shock := spell(cardID(3), 0, "Shock", "{R}")

	t.Run("may take one: takes the best, not nothing", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withLibrary(shock, bear, dragon)), newSeat(1)},
			withBattlefield(sixLands()...),
			withChoice(libraryChoice(choiceID, "Horn of the Mark — you may put a creature card into your hand", 0, 1, dragon, bear)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "take nothing", ids()),
			choiceMove(t, 0, choiceID, "take the dragon", ids(cardID(1))),
			choiceMove(t, 0, choiceID, "take the bear", ids(cardID(2))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take the dragon" {
			t.Fatalf("chose %q, want the Dragon", got)
		}
	})

	// A card cardValue prices at zero is still a card the bot did not
	// have. Without the floor it ties with "take nothing", which the
	// enumerator offers first.
	t.Run("a zero-valued card still beats nothing", func(t *testing.T) {
		free := spell(cardID(4), 0, "Gitaxian Probe", "")
		v := newView([]protocol.PlayerView{newSeat(0, withLibrary(free)), newSeat(1)},
			withChoice(libraryChoice(choiceID, "You may put it into your hand", 0, 1, free)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "take nothing", ids()),
			choiceMove(t, 0, choiceID, "take it", ids(cardID(4))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take it" {
			t.Fatalf("chose %q", got)
		}
	})

	// "Any number of them": every card is worth taking, so the
	// largest set wins.
	t.Run("any number takes them all", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withLibrary(shock, bear, dragon)), newSeat(1)},
			withChoice(libraryChoice(choiceID, "Put any number of them into your hand", 0, 3, dragon, bear, shock)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "take nothing", ids()),
			choiceMove(t, 0, choiceID, "take the dragon", ids(cardID(1))),
			choiceMove(t, 0, choiceID, "take all three", ids(cardID(1), cardID(2), cardID(3))),
			choiceMove(t, 0, choiceID, "take dragon + bear", ids(cardID(1), cardID(2))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take all three" {
			t.Fatalf("chose %q", got)
		}
	})

	// A mandatory pick (hideaway's "exile one of them face down", or a
	// must-take) has no empty answer: the best card still wins rather
	// than the enumerator's first.
	t.Run("must take one: takes the best", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withLibrary(shock, bear, dragon)), newSeat(1)},
			withBattlefield(sixLands()...),
			withChoice(libraryChoice(choiceID, "Exile one of them face down", 1, 1, shock, bear, dragon)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "the shock", ids(cardID(3))),
			choiceMove(t, 0, choiceID, "the dragon", ids(cardID(1))),
			choiceMove(t, 0, choiceID, "the bear", ids(cardID(2))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "the dragon" {
			t.Fatalf("chose %q", got)
		}
	})
}

// The signal is the bot's OWN library and nothing else. A pick over an
// opponent's library, or over cards that are not all in the bot's
// library, says nothing on the wire about what naming a card does, so
// it keeps the flat score and the enumerator's order — the rule
// TestChooseCardsOverOtherZonesKeepsTheEnumeratorsOrder states for the
// battlefield.
func TestLibraryTakeOnlyReadsTheBotsOwnLibrary(t *testing.T) {
	const choiceID = "choice-other"
	theirDragon := creature(cardID(1), 1, "Dragon", 6, 6)
	theirBear := creature(cardID(2), 1, "Bear", 2, 2)

	t.Run("an opponent's library keeps the enumerator's order", func(t *testing.T) {
		opp := newSeat(1, withLibrary(theirBear, theirDragon))
		ch := libraryChoice(choiceID, "Choose a card from among them", 0, 1, theirBear, theirDragon)
		ch.FromPlayer = opp.ID
		v := newView([]protocol.PlayerView{newSeat(0), opp}, withChoice(ch))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "name nothing", ids()),
			choiceMove(t, 0, choiceID, "name the dragon", ids(cardID(1))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "name nothing" {
			t.Fatalf("chose %q, want the enumerator's first answer", got)
		}
	})

	t.Run("a candidate outside the library keeps the enumerator's order", func(t *testing.T) {
		mine := creature(cardID(3), 0, "Dragon", 6, 6)
		elsewhere := creature(cardID(4), 0, "Bear", 2, 2)
		v := newView([]protocol.PlayerView{newSeat(0, withLibrary(mine)), newSeat(1)},
			withChoice(libraryChoice(choiceID, "Choose a card", 0, 1, mine, elsewhere)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "name nothing", ids()),
			choiceMove(t, 0, choiceID, "name the dragon", ids(cardID(3))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "name nothing" {
			t.Fatalf("chose %q, want the enumerator's first answer", got)
		}
	})
}

// The hand rule (#798) is untouched by the library rule: a discard over
// the bot's own hand still names its worst card, even with a known
// library beside it.
func TestHandPickStillNamesTheWorstWithALibraryInView(t *testing.T) {
	const choiceID = "choice-discard"
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	mountain := land(cardID(3), 0)
	topCard := creature(cardID(5), 0, "Bear", 2, 2)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(dragon, mountain), withLibrary(topCard)), newSeat(1)},
		withBattlefield(sixLands()...),
		withChoice(handChoice(choiceID, "Faithless Looting — discard a card", 1, 1, dragon, mountain)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "pitch the dragon", ids(cardID(1))),
		choiceMove(t, 0, choiceID, "pitch the mountain", ids(cardID(3))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch the mountain" {
		t.Fatalf("chose %q, want the answer that keeps the Dragon", got)
	}
}
