package legal_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// search_three_zones_test.go — #5408: "search your graveyard, hand, and/or
// library" (Say Its Name). The enumerator answers the zone question with
// every non-empty subset of the three zones, searching all three the
// always-legal answer, and after the answer offers only the chosen zones'
// cards.

func openThreeZoneSearch(t *testing.T) (g *game.Game, me *game.Player, lib, gy, hand uuid.UUID) {
	t.Helper()
	g = newTable(t)
	me = g.Seats[g.Turn.ActiveSeat]
	lib, gy, hand = uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		me.Hand.Cards = nil
		card := func(id uuid.UUID) game.Card {
			return game.Card{InstanceID: id, Name: "Wanted", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID}
		}
		me.Library.PushTop(card(lib))
		me.Graveyard.PushTop(card(gy))
		me.Hand.PushTop(card(hand))
		if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Wanted" },
			Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, AlsoGraveyard: true, AlsoHand: true,
		}); err != nil {
			t.Fatalf("search: %v", err)
		}
	})
	return g, me, lib, gy, hand
}

func TestEnumeratorOffersEverySubsetOfTheThreeZones(t *testing.T) {
	g, me, _, _, _ := openThreeZoneSearch(t)
	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	var asked []string
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
		asked = append(asked, strings.Join(z, ","))
		if m.AlwaysLegal {
			alwaysLegal++
			if len(z) != 3 {
				t.Errorf("the always-legal answer is %v, want all three zones", z)
			}
		}
	}
	slices.Sort(asked)
	want := []string{
		"graveyard", "graveyard,hand", "graveyard,hand,library", "graveyard,library",
		"hand", "hand,library", "library",
	}
	if !slices.Equal(asked, want) {
		t.Errorf("zone answers offered = %v, want %v", asked, want)
	}
	if alwaysLegal != 1 {
		t.Errorf("%d always-legal answers, want exactly one", alwaysLegal)
	}
}

func TestEnumeratorAfterGraveyardAndHandOffersNoLibraryCard(t *testing.T) {
	g, me, lib, gy, hand := openThreeZoneSearch(t)
	moves := chooseZones(t, g, me, "graveyard", "hand")
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
	if offered[lib] || !offered[gy] || !offered[hand] {
		t.Errorf("offered %v: want the graveyard and hand matches, never the library's", offered)
	}
}
