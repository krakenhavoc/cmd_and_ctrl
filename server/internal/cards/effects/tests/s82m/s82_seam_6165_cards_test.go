package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_seam_6165_cards_test.go — #6165: Scroll Rack exiles cards from the
// hand face down, puts that many from the top of the library into the hand
// (not a draw), then puts the exiled cards back on top in the chosen order.

const (
	seam6165ScrollRackOracle = "648bd03d-247e-4d6f-87a9-12a67688cbf5"
)

func TestSeam6165ScrollRackIsFull(t *testing.T) {
	spec, ok := effects.Lookup(seam6165ScrollRackOracle)
	if !ok || spec.Name != "Scroll Rack" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Scroll Rack: %+v %v", spec, ok)
	}
}

// seam6165Known reports whether viewer knows the card id, wherever it is.
func seam6165Known(g *game.Game, id, viewer uuid.UUID) bool {
	known := false
	g.ReadSnapshot(func() {
		for _, z := range []*game.Zone{g.Exile, g.Seats[0].Library, g.Seats[1].Library, g.Seats[2].Library, g.Seats[3].Library} {
			for i := range z.Cards {
				if z.Cards[i].InstanceID == id {
					known = z.Cards[i].IsKnownTo(viewer)
				}
			}
		}
	})
	return known
}

// CR 406.3: three cards exiled face down; the top three go to the hand
// without a draw, and the three exiled cards go back on top in the chosen
// order, never seen by an opponent.
func TestScrollRackSwapsThreeCardsInTheChosenOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rack := effectstest.SeedPermanentWithOracle(g, me.ID, "Scroll Rack", "Artifact", seam6165ScrollRackOracle)
	var hand, top []uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		for _, n := range []string{"Hand A", "Hand B", "Hand C", "Hand D"} {
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
			hand = append(hand, id)
		}
		for _, n := range []string{"Lib 3", "Lib 2", "Lib 1"} {
			id := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
			top = append([]uuid.UUID{id}, top...)
		}
	})
	drawsBefore := effectstest.CountEvents(g, game.EventDrawCard)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, rack, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no hand pick: %+v stack=%d hand=%d step=%v", g.PendingChoices, g.Stack.Size(), len(me.Hand.Cards), g.Turn.Step)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 4 {
		t.Fatalf("hand pick bounds %d..%d, want any number (0..4)", pick.ChooseMin, pick.ChooseMax)
	}
	exiled := hand[:3]
	if err := g.ResolveChooseCards(pick.ID, me.ID, exiled); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range exiled {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not exiled", id)
		}
		if seam6165Known(g, id, opp.ID) {
			t.Errorf("an opponent can see exiled card %s", id)
		}
	}
	for _, id := range top {
		if !me.Hand.Contains(id) {
			t.Errorf("top card %s did not go to the hand", id)
		}
	}
	if got := effectstest.CountEvents(g, game.EventDrawCard); got != drawsBefore {
		t.Errorf("%d draw events, want none: the cards are put into the hand, not drawn", got-drawsBefore)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatalf("no ordering prompt: %+v", g.PendingChoices)
	}
	want := []uuid.UUID{exiled[2], exiled[0], exiled[1]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	n := len(me.Library.Cards)
	for i, id := range want {
		if got := me.Library.Cards[n-1-i].InstanceID; got != id {
			t.Errorf("library position %d from the top = %s, want %s", i, got, id)
		}
		if seam6165Known(g, id, opp.ID) {
			t.Errorf("an opponent knows the returned card %s", id)
		}
	}
	if !me.Hand.Contains(hand[3]) {
		t.Error("the card kept in hand left it")
	}
}

// Exiling nothing puts nothing into the hand and asks nothing more.
func TestScrollRackExilingNothingDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rack := effectstest.SeedPermanentWithOracle(g, me.ID, "Scroll Rack", "Artifact", seam6165ScrollRackOracle)
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Kept", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	handBefore, libBefore := len(me.Hand.Cards), len(me.Library.Cards)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, rack, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no hand pick: %+v stack=%d hand=%d step=%v", g.PendingChoices, g.Stack.Size(), len(me.Hand.Cards), g.Turn.Step)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if len(me.Hand.Cards) != handBefore || len(me.Library.Cards) != libBefore {
		t.Errorf("hand %d→%d, library %d→%d: nothing should move", handBefore, len(me.Hand.Cards), libBefore, len(me.Library.Cards))
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("an ordering prompt was raised for no cards")
	}
}
