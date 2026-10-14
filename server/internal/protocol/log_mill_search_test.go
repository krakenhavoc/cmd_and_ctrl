package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_mill_search_test.go pins two public log lines that said nothing
// about what happened: a mill named no card, and a library search that
// was not narrated at all. A mill names the cards (they are public once
// in a graveyard). A search says the library was searched and shuffled,
// and never names the card it took, because that card is not public.

// logEntriesOfKind returns every entry of kind k in a log.
func logEntriesOfKind(entries []LogEvent, k LogKind) []LogEvent {
	var out []LogEvent
	for _, e := range entries {
		if e.Kind == k {
			out = append(out, e)
		}
	}
	return out
}

func TestPublicLogMillNamesTheMilledCardsInOneLine(t *testing.T) {
	g := buildActiveGame(t)
	milled := g.Seats[0]
	want := topNames(g, milled, 3)

	var err error
	g.WithWriteLock(func() {
		_, err = g.MillToZoneForEffect(milled.ID, 3, game.ZoneGraveyard)
	})
	if err != nil {
		t.Fatalf("MillToZoneForEffect: %v", err)
	}

	wantText := "P1 milled " + want[0] + ", " + want[1] + " and " + want[2]
	for _, viewer := range allViewers(g) {
		mills := logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogMill)
		if len(mills) != 1 {
			t.Fatalf("viewer %q: %d mill entries, want 1 line for one mill instruction: %+v", viewer, len(mills), mills)
		}
		e := mills[0]
		if e.Amount != 3 || e.Seat != milled.Seat {
			t.Errorf("viewer %q: mill entry fields: %+v", viewer, e)
		}
		if e.Text != wantText {
			t.Errorf("viewer %q: text %q, want %q", viewer, e.Text, wantText)
		}
	}
}

func TestPublicLogSearchSaysSearchedAndShuffledWithNoCard(t *testing.T) {
	g := buildActiveGame(t)
	searcher := g.Seats[0]
	library := topNames(g, searcher, 2)

	var err error
	g.WithWriteLock(func() {
		// A search that finds nothing: no prompt, no card taken, so the
		// only lines are the search and the shuffle.
		err = g.SearchLibraryForEffect(searcher.ID, func(game.Card) bool { return false }, game.ZoneHand, 1, false, true)
	})
	if err != nil {
		t.Fatalf("SearchLibraryForEffect: %v", err)
	}

	want := []string{"P1 searched their library and shuffled it"}
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		searches := logEntriesOfKind(log, LogSearch)
		if len(searches) != len(want) {
			t.Fatalf("viewer %q: %d search entries, want %d: %+v", viewer, len(searches), len(want), searches)
		}
		for i, e := range searches {
			if e.Text != want[i] {
				t.Errorf("viewer %q: entry %d text %q, want %q", viewer, i, e.Text, want[i])
			}
			if e.CardID != "" || e.Target != "" {
				t.Errorf("viewer %q: search entry carries a card reference: %+v", viewer, e)
			}
			for _, name := range library {
				if strings.Contains(e.Text, name) {
					t.Errorf("viewer %q: search line names library card %q: %q", viewer, name, e.Text)
				}
			}
		}
	}
}

// TestPublicLogFetchSearchIsNotSaidTwice pins the fold's rule: a
// library search that fetched a card was already told by the reveal the
// engine emits before it, so the search line is dropped. A plain shuffle
// after it still reads, since it is not a search.
func TestPublicLogFetchSearchIsNotSaidTwice(t *testing.T) {
	g := buildActiveGame(t)
	searcher := g.Seats[0]
	var topID uuid.UUID
	g.ReadSnapshot(func() {
		lib := searcher.Library.Cards
		topID = lib[len(lib)-1].InstanceID
	})

	g.WithWriteLock(func() {
		seq := nextEventSeq(g)
		g.EmitEvent(game.Event{
			Kind:      game.EventRevealCards,
			Actor:     searcher.ID,
			CardID:    topID,
			OldZone:   game.ZoneLibrary,
			RevealTo:  game.ZoneBattlefield,
			RevealSeq: seq,
		})
		g.EmitEvent(game.Event{Kind: game.EventSearchLibrary, Actor: searcher.ID, Target: searcher.ID, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventSearchLibrary, Actor: searcher.ID, Target: searcher.ID, Label: "shuffle"})
	})

	log := ViewOfGameFor(g, "").Log
	if got := logEntriesOfKind(log, LogSearch); len(got) != 1 || got[0].Text != "P1 shuffled their library" {
		t.Fatalf("search entries: %+v, want only the shuffle line", got)
	}
	reveals := revealLogEntries(log)
	if len(reveals) != 1 || !strings.HasPrefix(reveals[0].Text, "P1 fetched ") {
		t.Fatalf("reveal entries: %+v, want one fetched line", reveals)
	}
}
