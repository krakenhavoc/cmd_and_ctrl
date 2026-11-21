package testsg

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func furyCastFromExile(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"})
}

// Every card of mana value 5 or less may be cast, one after another: the
// first cast does not spend the permission, and there is no running total
// (5 + 2 is over any budget of 5).
func TestHazoretsUndyingFuryCastsAnyNumberEachAtMostFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	five, six, two, forest := effectstest.FuryLibrary(me)
	effectstest.FuryCast(t, g)
	for _, id := range []uuid.UUID{five, six, two, forest} {
		if !effectstest.InExile(g, id) {
			t.Fatalf("%v was among the top four and is not in exile", id)
		}
	}
	if err := furyCastFromExile(g, me, six); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("cast the mana value 6 card: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	if err := furyCastFromExile(g, me, five); err != nil {
		t.Fatalf("cast the mana value 5 card: %v", err)
	}
	if err := furyCastFromExile(g, me, two); err != nil {
		t.Fatalf("cast the mana value 2 card after the 5: %v", err)
	}
	if err := furyCastFromExile(g, me, five); err == nil {
		t.Error("the mana value 5 card was cast twice")
	}
	if err := furyCastFromExile(g, me, forest); err == nil {
		t.Error("a land was cast")
	}
}

// Only the top four leave the library, and the library is shuffled first.
func TestHazoretsUndyingFuryExilesExactlyFourAfterAShuffle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	for i := 0; i < 9; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Artifact", ManaCost: "{1}"})
	}
	effectstest.FuryCast(t, g)
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Name == "Filler" {
			exiled++
		}
	}
	if exiled != 4 || me.Library.Size() != 5 {
		t.Errorf("exiled %d, library %d; want 4 and 5", exiled, me.Library.Size())
	}
}

// The window is the resolving spell's (#4250: its resolution_cast prompt):
// each cast is made during the resolution, the decline closes it, and the
// cards not cast stay in exile.
func TestHazoretsUndyingFuryWindowClosesOnPass(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	five, _, two, _ := effectstest.FuryLibrary(me)
	effectstest.FuryCast(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, five, two)
	if err := furyCastFromExile(g, me, five); err != nil {
		t.Fatal(err)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("one cast closed the \"any number\" offer")
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.CastInsideTheResolution(t, g, five)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, two) {
		t.Error("the card not cast stays in exile")
	}
	if err := furyCastFromExile(g, me, two); err == nil {
		t.Error("a card was cast after the window closed")
	}
	if len(me.CastPermissions) != 0 {
		t.Errorf("a permission survived the pass: %+v", me.CastPermissions)
	}
}

// What a bot picks from is judged by the same gate, and a cast does not
// withdraw the other offers.
func TestHazoretsUndyingFuryEnumeratorOffersEachCardUnderTheCap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	five, six, two, _ := effectstest.FuryLibrary(me)
	effectstest.FuryCast(t, g)
	offered := func(id uuid.UUID) bool {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Kind == legal.KindCast && strings.Contains(string(m.Params), id.String()) {
				return true
			}
		}
		return false
	}
	if !offered(five) || !offered(two) {
		t.Error("a card at or under mana value 5 is not offered")
	}
	if offered(six) {
		t.Error("the mana value 6 card is offered")
	}
	if err := furyCastFromExile(g, me, five); err != nil {
		t.Fatal(err)
	}
	if !offered(two) {
		t.Error("the other card is no longer offered after a cast")
	}
	if offered(five) {
		t.Error("the cast card is still offered")
	}
}

// What is left of the permission is part of the snapshot.
func TestHazoretsUndyingFurySurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	five, six, two, _ := effectstest.FuryLibrary(me)
	effectstest.FuryCast(t, g)
	if err := furyCastFromExile(g, me, five); err != nil {
		t.Fatal(err)
	}
	r := effectstest.RestoreRoundTrip(t, g, false)
	rme := r.PlayerByIDForEffect(me.ID)
	if err := furyCastFromExile(r, rme, six); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("restored: the 6: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	if err := furyCastFromExile(r, rme, two); err != nil {
		t.Fatalf("restored: the 2 after the 5: %v", err)
	}
}

// The permission names no cards when the library is empty.
func TestHazoretsUndyingFuryWithAnEmptyLibraryGrantsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	effectstest.FuryCast(t, g)
	if len(me.CastPermissions) != 0 {
		t.Errorf("permissions = %+v, want none", me.CastPermissions)
	}
}

// Lands you control don't untap during your next untap step; an opponent's
// do.
func TestHazoretsUndyingFuryHoldsTheCastersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	myLand := effectstest.UpsTapped(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirLand := effectstest.UpsTapped(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.FuryLibrary(me)
	effectstest.FuryCast(t, g)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if effectstest.UpsIsTapped(t, g, theirLand) {
		t.Error("the opponent's land stayed tapped")
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if !effectstest.UpsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
}
