package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Silver Scrutiny
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestOfferXCeilingViewCarriesTheClaimedOffersCeiling is the #3225 view
// half: an alternative cost with its own X ceiling ("as though it had
// flash if X is 3 or less", CR 601.3c) ships that ceiling on the OFFER
// as `x_max`, and leaves the card's own `x_max` absent — the cast for
// the printed cost is not bound by it. A printed ceiling on the card is
// folded in by taking the smaller.
func TestOfferXCeilingViewCarriesTheClaimedOffersCeiling(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Silver Scrutiny", TypeLine: "Sorcery",
			OracleID: "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e",
			ManaCost: "{X}{U}{U}", Owner: me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.XMax != nil {
		t.Errorf("the card carries x_max %d; the ceiling belongs to the offer", *c.XMax)
	}
	var found bool
	for _, a := range c.AlternativeCosts {
		if a.Key != "flash-x-at-most" {
			continue
		}
		found = true
		if a.XMax == nil || *a.XMax != 3 {
			t.Errorf("offer x_max = %v, want 3", a.XMax)
		}
	}
	if !found {
		t.Fatalf("the flash offer is not on the card: %+v", c.AlternativeCosts)
	}
}

// The folding rule on its own: the offer's ceiling and a printed one,
// the smaller wins, and an offer without one answers as the card does.
func TestOfferXCeilingFoldsWithThePrintedCeiling(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	offer := &game.AlternativeCost{Key: "k", ManaCost: "{X}{U}", XCeiling: 3}
	var got int
	var ok bool
	g.WithWriteLock(func() { got, ok = g.SpellXCeilingForOfferEffect(me.ID, "no-printed-ceiling", offer) })
	if !ok || got != 3 {
		t.Errorf("offer alone: %d, %v; want 3, true", got, ok)
	}
	g.WithWriteLock(func() { got, ok = g.SpellXCeilingForOfferEffect(me.ID, "no-printed-ceiling", nil) })
	if ok {
		t.Errorf("no offer and no printed ceiling reported %d", got)
	}
}
