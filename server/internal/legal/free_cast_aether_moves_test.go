package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_cast_aether_moves_test.go — the enumerator half of #2110, Aetherflux
// Conduit and Aetherworks Marvel (ADR 0033 §1): a bot is offered exactly the
// casts the engine accepts, and nothing it could not cast.

// "Any number of spells": one grant per hand card, so the bot is offered a
// free cast of each, and casting one leaves the other on offer.
func TestEnumeratorOffersEveryHandCardAnyNumberGrantOpens(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepEnd)

	a := handCard(seat, game.Card{Name: "First Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{U}", Layout: "normal"})
	b := handCard(seat, game.Card{Name: "Second Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{R}", Layout: "normal"})
	land := handCard(seat, game.Card{Name: "A Swamp", TypeLine: "Basic Land — Swamp", Layout: "normal"})
	free := func(id uuid.UUID) bool {
		for _, m := range castMovesFor(legal.EnumerateFor(g, seat.ID), id) {
			if strings.Contains(string(m.Params), game.AltCostKeyGrantedCost) {
				return true
			}
		}
		return false
	}
	g.WithWriteLock(func() {
		g.FreeCastFromHandForEffect(seat.ID, uuid.Nil, a, "Aetherflux Conduit")
		g.FreeCastFromHandForEffect(seat.ID, uuid.Nil, b, "Aetherflux Conduit")
	})
	if !free(a) || !free(b) {
		t.Fatalf("both sorceries should be offered a free cast in the end step: %v", labels(legal.EnumerateFor(g, seat.ID)))
	}
	if free(land) {
		t.Error("a land was offered as a free cast")
	}
	dispatchAll(t, g, seat.ID, legal.EnumerateFor(g, seat.ID))

	if err := g.CastSpell(seat.ID, a, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast the first: %v", err)
	}
	if !free(b) {
		t.Error("the second grant did not survive the first cast")
	}
}

// A card looked at on top of the library and picked for a free cast is
// offered as a cast from the library, at any timing, until the holder passes.
func TestEnumeratorOffersALookedAtLibraryFreeCastUntilThePass(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepEnd)

	seat.Library.Cards = nil
	pick := game.NewCard("Looked At Sorcery", seat.ID)
	pick.TypeLine = "Sorcery"
	pick.ManaCost = "{6}{R}"
	pick.Layout = "normal"
	pick.KnownBy = map[uuid.UUID]bool{seat.ID: true}
	seat.Library.PushTop(pick)
	under := game.NewCard("Under It", seat.ID)
	under.TypeLine = "Sorcery"
	under.ManaCost = "{R}"
	under.Layout = "normal"
	seat.Library.PushBottom(under)

	offered := func(id uuid.UUID) bool {
		return len(castMovesFor(legal.EnumerateFor(g, seat.ID), id)) > 0
	}
	if offered(pick.InstanceID) {
		t.Fatal("a library card is offered before anything grants it")
	}
	g.WithWriteLock(func() {
		g.FreeCastFromLibraryForEffect(seat.ID, uuid.Nil, pick.InstanceID, "Aetherworks Marvel")
	})
	if !offered(pick.InstanceID) {
		t.Fatalf("no cast offered for the looked-at pick: %v", labels(legal.EnumerateFor(g, seat.ID)))
	}
	if offered(under.InstanceID) {
		t.Error("a card the grant does not name was offered")
	}
	dispatchAll(t, g, seat.ID, legal.EnumerateFor(g, seat.ID))

	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if offered(pick.InstanceID) {
		t.Fatal("the cast is still offered after the holder passed")
	}
}
