package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// search_graveyard_test.go — #3217: the enumerator over a "search your
// library and/or graveyard" prompt. It offers a take from either zone,
// offers "fail to find" beside them ("and/or" lets the searcher look
// only in the library, so a graveyard match removes nothing).

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

func TestEnumeratorOffersBothZonesAndFailToFind(t *testing.T) {
	g, me, libID, gyID := openCombinedSearch(t, true)
	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	offered := map[uuid.UUID]bool{}
	failToFind := false
	for _, m := range moves {
		if m.Kind != legal.KindChoice {
			continue
		}
		ids := pickedCardIDs(t, m)
		if len(ids) == 0 {
			failToFind = m.AlwaysLegal
		}
		for _, id := range ids {
			offered[id] = true
		}
	}
	if !failToFind {
		t.Errorf("no always-legal fail to find with a graveyard match present: %v", labels(moves))
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
