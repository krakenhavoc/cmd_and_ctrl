package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// named_card_copy_test.go — #6155: a copy of a card defined by name is
// created outside the game (CR 707.13), offered for a cast from there, and
// ceases to exist once nothing offers it (CR 707.10a, 704.5e).

func init() {
	RegisterConjurable(Card{Name: "Test Named Copy Bolt", OracleID: "test-named-copy-bolt", TypeLine: "Instant", ManaCost: "{R}"})
}

func TestCopyNamedCardIsCreatedOutsideTheGameAndSweptWhenNotOffered(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.CopyNamedCardAndOfferCastForEffect(CardCopyOffer{Player: me.ID, SourceName: "Test"}, "Test Named Copy Bolt")
		if err != nil {
			t.Fatalf("CopyNamedCardAndOfferCastForEffect: %v", err)
		}
	})
	cards := me.SideboardCards()
	if len(cards) != 1 || cards[0].InstanceID != id || !cards[0].CardCopy || cards[0].Name != "Test Named Copy Bolt" {
		t.Fatalf("the outside container holds %+v; want the one copy", cards)
	}
	if _, ok := g.LookupCardForEffect(id); ok {
		t.Error("the copy outside the game is reachable by the card index (CR 400.11c)")
	}
	g.WithWriteLock(func() {
		c, zone, ok := g.ResolutionCastCandidateLocked(me.ID, id)
		if !ok || zone != ZoneOutside || c.InstanceID != id {
			t.Errorf("the copy is not a resolution-cast candidate outside the game: %v %q", ok, zone)
		}
		if _, _, ok := g.ResolutionCastCandidateLocked(g.Seats[1].ID, id); ok {
			t.Error("another player's prompt sees the copy")
		}
		// Still offered: the sweep leaves it.
		g.cardCopySweepSBALocked()
		if len(me.SideboardCards()) != 1 {
			t.Fatal("an offered copy was swept")
		}
		g.withdrawCardCopyOffersLocked(me.ID, []uuid.UUID{id}, nil)
		g.cardCopySweepSBALocked()
	})
	if n := len(me.SideboardCards()); n != 0 {
		t.Errorf("a copy nobody may cast is still outside the game (%d cards)", n)
	}
}

func TestCopyNamedCardRefusesAnUnknownName(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() {
		if _, err := g.CopyNamedCardAndOfferCastForEffect(CardCopyOffer{Player: g.Seats[0].ID}, "No Such Card Name"); !errors.Is(err, ErrNotConjurable) {
			t.Errorf("err = %v, want ErrNotConjurable", err)
		}
	})
}
