package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Silver Scrutiny
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestFlashIfXCeilingViewBoundsOnlyTheCastThatNeedsIt is the #5207 view
// half of Silver Scrutiny's "as though it had flash if X is 3 or less"
// (CR 601.3b): in a main phase the card's `x_max` is absent, and at
// instant speed, where only the card's own permission opens the window,
// it is 3. There is no alternative-cost offer for it.
func TestFlashIfXCeilingViewBoundsOnlyTheCastThatNeedsIt(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Silver Scrutiny", TypeLine: "Sorcery",
			OracleID: "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e",
			ManaCost: "{X}{U}{U}", Owner: me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.XMax != nil {
		t.Errorf("main phase: the card carries x_max %d, want none", *c.XMax)
	}
	if len(c.AlternativeCosts) != 0 {
		t.Errorf("alternative costs %+v, want none", c.AlternativeCosts)
	}
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.XMax == nil || *c.XMax != 3 {
		t.Errorf("beginning of combat: x_max = %v, want 3", c.XMax)
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
