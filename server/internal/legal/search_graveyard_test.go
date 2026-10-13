package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// search_graveyard_test.go — #3217: the enumerator over a "search your
// library and/or graveyard" prompt. It offers a take from either zone,
// never offers "fail to find" while a public card is on offer (the
// engine would refuse it), and offers it again once only library
// candidates remain.

func openCombinedSearch(t *testing.T, graveyardMatch bool) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var libID, gyID uuid.UUID
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		libID, gyID = uuid.New(), uuid.New()
		me.Library.PushTop(game.Card{InstanceID: libID, Name: "Wanted", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		if graveyardMatch {
			me.Graveyard.PushTop(game.Card{InstanceID: gyID, Name: "Wanted", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		}
		if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Wanted" },
			Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
		}); err != nil {
			t.Fatalf("search: %v", err)
		}
	})
	return g, me, libID, gyID
}

func TestEnumeratorOffersBothZonesAndNoFailToFindWhilePublicCardIsThere(t *testing.T) {
	g, me, libID, gyID := openCombinedSearch(t, true)
	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	offered := map[uuid.UUID]bool{}
	for _, m := range moves {
		if m.Kind != legal.KindChoice {
			continue
		}
		if m.AlwaysLegal {
			t.Errorf("%q is marked always-legal, but the engine refuses a short answer while the graveyard holds a match", m.Label)
		}
		ids := pickedCardIDs(t, m)
		if len(ids) == 0 {
			t.Errorf("offered fail to find: %q", m.Label)
		}
		for _, id := range ids {
			offered[id] = true
		}
	}
	if !offered[libID] || !offered[gyID] {
		t.Errorf("offered library=%v graveyard=%v, want both zones (%v)", offered[libID], offered[gyID], labels(moves))
	}
}

func TestEnumeratorOffersFailToFindWhenOnlyTheLibraryMatches(t *testing.T) {
	g, me, _, _ := openCombinedSearch(t, false)
	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	found := false
	for _, m := range moves {
		if m.Kind == legal.KindChoice && m.AlwaysLegal && len(pickedCardIDs(t, m)) == 0 {
			found = true
		}
	}
	if !found {
		t.Errorf("no always-legal fail to find on a library-only prompt: %v", labels(moves))
	}
}
