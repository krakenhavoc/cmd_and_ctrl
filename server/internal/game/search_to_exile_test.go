package game

import (
	"testing"

	"github.com/google/uuid"
)

// search_to_exile_test.go — #1230: SearchLibrarySpec.Dest now accepts
// ZoneExile (game.searchDestZoneLocked used to return ErrZoneNotFound
// for it, which failed "search your library for ... cards, exile
// them, then shuffle" — Ugin, Eye of the Storms — on its first line
// and took the shuffle down with it). The take is routed through
// searchRoute -> routeCardToZoneLocked exactly like every other
// search destination, so CR 614 and CR 903.9 apply unchanged; this
// file is the CR 903.9 half, mirroring the exile / bounce / tuck /
// mill coverage in commander_zone_routes_test.go for the SEARCH path
// specifically. seatCommander, expectCommanderPrompt and assertOnlyIn
// are declared there.

// TestCommanderSearchedToExileOffersCommandZone: a commander found by
// a library search whose destination is exile still gets the CR
// 903.9 offer before it lands anywhere.
func TestCommanderSearchedToExileOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, owner.Library, owner)

	g.mu.Lock()
	err := g.SearchLibraryThenForEffect(SearchLibrarySpec{
		Player: owner.ID,
		Pred:   func(c Card) bool { return c.InstanceID == cmdID },
		Dest:   ZoneExile,
		Limit:  1,
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("SearchLibraryThenForEffect: %v", err)
	}
	if g.Exile.Contains(cmdID) {
		t.Fatalf("searched-to-exile commander hit exile before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, g.Exile, owner.Library)
}

// TestCommanderSearchedToExileDeclineGoesToExile: declining still
// exiles the card where the search asked.
func TestCommanderSearchedToExileDeclineGoesToExile(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, owner.Library, owner)

	g.mu.Lock()
	err := g.SearchLibraryThenForEffect(SearchLibrarySpec{
		Player: owner.ID,
		Pred:   func(c Card) bool { return c.InstanceID == cmdID },
		Dest:   ZoneExile,
		Limit:  1,
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("SearchLibraryThenForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, g.Exile, owner.Command, owner.Library)
}

// TestSearchToExileNonCommanderIsUnchanged: an ordinary card searched
// to exile just goes there, no prompt.
func TestSearchToExileNonCommanderIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := uuid.MustParse("00000000-0000-0000-0000-0000000000e1")
	me.Library.PushTop(Card{InstanceID: id, Name: "A Card", Owner: me.ID, Controller: me.ID})

	g.mu.Lock()
	err := g.SearchLibraryThenForEffect(SearchLibrarySpec{
		Player: me.ID,
		Pred:   func(c Card) bool { return c.InstanceID == id },
		Dest:   ZoneExile,
		Limit:  1,
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("SearchLibraryThenForEffect: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("no commander involved: want no pending choices, got %d", len(g.PendingChoices))
	}
	if !g.Exile.Contains(id) {
		t.Error("the card should be in exile")
	}
	if me.Library.Contains(id) {
		t.Error("the card should have left the library")
	}
}
