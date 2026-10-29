package legal_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// search_graveyard_test.go — #3217, #4519: the enumerator over an
// "and/or" search ("your library and/or graveyard", "your hand and/or
// library"). It asks the zones first, as three moves and no card pick (the
// library's candidates are not shown until the library is chosen), with
// searching both the always-legal answer. After that it offers a take from
// either chosen zone, and "fail to find" beside them ("and/or" lets the
// searcher look only in the library, so a graveyard match removes
// nothing).

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

// zonesOf is the zones a move answers the zone question with, nil for a
// move that does not.
func zonesOf(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		Zones []string `json:"search_zones"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params of %q: %v", m.Label, err)
	}
	return p.Zones
}

// chooseZones plays the move that answers the zone question with exactly
// `zones` and returns the seat's moves after it.
func chooseZones(t *testing.T, g *game.Game, me *game.Player, zones ...string) []legal.Move {
	t.Helper()
	for _, m := range legal.EnumerateFor(g, me.ID) {
		got := slices.Clone(zonesOf(t, m))
		slices.Sort(got)
		want := slices.Clone(zones)
		slices.Sort(want)
		if slices.Equal(got, want) {
			a := actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}
			if err := actions.Dispatch(g, a); err != nil {
				t.Fatalf("move %q rejected: %v", m.Label, err)
			}
			return legal.EnumerateFor(g, me.ID)
		}
	}
	t.Fatalf("no move answers the zone question with %v: %v", zones, labels(legal.EnumerateFor(g, me.ID)))
	return nil
}

// The first thing offered is the zone question: three answers, each
// accepted, none a card pick, none naming a card, searching both the
// always-legal one.
func TestEnumeratorAsksTheZonesBeforeOfferingAnyCard(t *testing.T) {
	g, me, _, _ := openCombinedSearch(t, true)
	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	var asked [][]string
	alwaysLegal := 0
	for _, m := range moves {
		if m.Kind != legal.KindChoice {
			continue
		}
		if len(pickedCardIDs(t, m)) != 0 {
			t.Errorf("a card pick %q is offered before the zones are chosen", m.Label)
		}
		z := zonesOf(t, m)
		if z == nil {
			t.Errorf("a choice move %q that is not a zone answer is offered first", m.Label)
			continue
		}
		slices.Sort(z)
		asked = append(asked, z)
		if m.AlwaysLegal {
			alwaysLegal++
			if !slices.Equal(z, []string{"graveyard", "library"}) {
				t.Errorf("the always-legal answer is %v, want searching both zones", z)
			}
		}
		if strings.Contains(m.Label, "Wanted") {
			t.Errorf("label %q names a card before the library was searched", m.Label)
		}
	}
	slices.SortFunc(asked, func(a, b []string) int { return strings.Compare(strings.Join(a, ","), strings.Join(b, ",")) })
	want := [][]string{{"graveyard"}, {"graveyard", "library"}, {"library"}}
	if !slices.EqualFunc(asked, want, slices.Equal[[]string]) {
		t.Errorf("zone answers offered = %v, want %v", asked, want)
	}
	if alwaysLegal != 1 {
		t.Errorf("%d always-legal answers, want exactly one", alwaysLegal)
	}
}

func TestEnumeratorOffersBothZonesAndFailToFind(t *testing.T) {
	g, me, libID, gyID := openCombinedSearch(t, true)
	moves := chooseZones(t, g, me, "library", "graveyard")
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

// Graveyard only: the library's card is never offered or named.
func TestEnumeratorOffersOnlyTheGraveyardWhenOnlyItIsChosen(t *testing.T) {
	g, me, libID, gyID := openCombinedSearch(t, true)
	moves := chooseZones(t, g, me, "graveyard")
	dispatchAll(t, g, me.ID, moves)
	offered := map[uuid.UUID]bool{}
	for _, m := range moves {
		if m.Kind != legal.KindChoice {
			continue
		}
		for _, id := range pickedCardIDs(t, m) {
			offered[id] = true
		}
	}
	if offered[libID] {
		t.Errorf("the library card is offered to a graveyard-only search: %v", labels(moves))
	}
	if !offered[gyID] {
		t.Errorf("the graveyard card is not offered: %v", labels(moves))
	}
}

func TestEnumeratorOffersFailToFindWhenOnlyTheLibraryMatches(t *testing.T) {
	g, me, _, _ := openCombinedSearch(t, false)
	moves := chooseZones(t, g, me, "library", "graveyard")
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

// "Search your hand and/or library" asks the same question with the hand.
func TestEnumeratorAsksTheHandZoneToo(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Wanted", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Wanted" },
			Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, AlsoHand: true,
		}); err != nil {
			t.Fatalf("search: %v", err)
		}
	})
	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	var asked []string
	for _, m := range moves {
		if m.Kind == legal.KindChoice {
			asked = append(asked, strings.Join(zonesOf(t, m), ","))
		}
	}
	slices.Sort(asked)
	if want := []string{"hand", "library", "library,hand"}; !slices.Equal(asked, want) {
		t.Errorf("zone answers offered = %v, want %v", asked, want)
	}
}
