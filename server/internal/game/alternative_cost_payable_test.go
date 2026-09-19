package game

import (
	"testing"

	"github.com/google/uuid"
)

// alternative_cost_payable_test.go — #695. One predicate decides
// whether an alternative cost is on the table, and the three things
// it has to get right are the three the view used to get wrong:
//
//  1. CR 119.4's life. A player at 3 was shown Snuff Out's "pay 4
//     life", chose it, and had the cast refused with ErrInvalidParam.
//     Exactly N is payable — paying down to zero is legal.
//  2. CR 601.2b's card component. Force of Will with no other blue
//     card in hand is a picker with nothing in it.
//  3. The offer's own Condition, which was the ONLY thing asked
//     before and still has to be asked.
//
// Mana is deliberately not among them: CR 601.2g lets the caster tap
// for it after the cost is chosen, so "you cannot afford it yet" is
// not a reason to hide the offer. That is the auto-tapper's job.

// blueCardSpec matches a blue card, the shape of Force of Will's
// "exile a blue card from your hand".
func blueCardSpec() *TargetSpec {
	return &TargetSpec{
		Zones: []ZoneKind{ZoneHand},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			for _, col := range c.Colors {
				if col == "U" {
					return true
				}
			}
			return false
		},
	}
}

// handCard seeds a card in a seat's hand and returns it.
func altCostHandCard(p *Player, name string, colors ...string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Instant"
	c.Colors = colors
	p.Hand.PushTop(c)
	return c.InstanceID
}

// CR 119.4, all three sides of the boundary.
func TestAlternativeCostLifePayableAtTheBoundary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	self := altCostHandCard(me, "Test Snuff Out")
	offer := &AlternativeCost{Key: "pay_life", Label: "Pay 4 life", Life: 4}

	for _, tc := range []struct {
		life int
		want bool
	}{
		{life: 3, want: false},
		// Exactly the payment: legal, and the state-based action that
		// follows is a different rule (CR 704.5a).
		{life: 4, want: true},
		{life: 20, want: true},
		{life: 0, want: false},
	} {
		me.Life = tc.life
		var got bool
		g.WithWriteLock(func() { got = g.AlternativeCostPayableLocked(me.ID, self, offer) })
		if got != tc.want {
			t.Errorf("at %d life, a 4-life offer payable = %v, want %v", tc.life, got, tc.want)
		}
		// The predicate and the announce validator are the same rule,
		// so they must answer alike on every one of these.
		var accepted bool
		g.WithWriteLock(func() {
			accepted = g.validateAlternativeCostPaymentLocked(me.ID, self, offer, nil) == nil
		})
		if accepted != tc.want {
			t.Errorf("at %d life, the announce validator accepted = %v but the offer predicate said %v",
				tc.life, accepted, tc.want)
		}
	}
	// An offer with no life component is payable at any life total,
	// including zero — most offers are this.
	me.Life = 0
	free := &AlternativeCost{Key: "overload", ManaCost: "{6}{U}"}
	var got bool
	g.WithWriteLock(func() { got = g.AlternativeCostPayableLocked(me.ID, self, free) })
	if !got {
		t.Errorf("an offer with no life component was withheld at 0 life")
	}
}

// CR 601.2b's card component: the offer is on the table only when the
// board can pay it. Force of Will cannot pitch itself (CR 601.2a has
// already moved it to the stack), which is why the spell being cast
// is excluded from the count.
func TestAlternativeCostCardComponentNeedsACandidate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	self := altCostHandCard(me, "Test Force of Will", "U")
	offer := &AlternativeCost{
		Key: "pitch", Label: "Pay 1 life, exile a blue card", Life: 1,
		ExileFromHand: blueCardSpec(),
	}

	payable := func() bool {
		var ok bool
		g.WithWriteLock(func() { ok = g.AlternativeCostPayableLocked(me.ID, self, offer) })
		return ok
	}

	// Only the spell itself is in hand, and it is blue: the offer is
	// still unpayable, because a spell cannot pitch itself.
	if payable() {
		t.Errorf("the offer was shown with no card in hand but the spell itself")
	}
	// A red card is in hand, and it does not match the clause.
	altCostHandCard(me, "Test Mountain Bolt", "R")
	if payable() {
		t.Errorf("the offer was shown with only a card the clause refuses")
	}
	// A blue one does.
	altCostHandCard(me, "Test Brainstorm", "U")
	if !payable() {
		t.Errorf("the offer was withheld with a legal pitch in hand")
	}
	// …and the life half still gates it independently.
	me.Life = 0
	if payable() {
		t.Errorf("the offer survived a life total below its life component")
	}
}

// Escape's "exile N OTHER cards from your graveyard" is a COUNT, so
// the predicate counts rather than asking whether any card qualifies.
func TestAlternativeCostCountsEveryCardTheComponentDemands(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	self := NewCard("Test Uro", me.ID)
	self.TypeLine = "Creature — Giant"
	me.Graveyard.PushTop(self)
	offer := &AlternativeCost{
		Key: "escape", Label: "Escape — {1}{G}{U}, exile three other cards",
		ManaCost: "{1}{G}{U}", FromZone: ZoneGraveyard,
		ExileFromGraveyard: escapeExileSpec(3),
	}

	payable := func() bool {
		var ok bool
		g.WithWriteLock(func() { ok = g.AlternativeCostPayableLocked(me.ID, self.InstanceID, offer) })
		return ok
	}

	for i, name := range []string{"Fuel A", "Fuel B"} {
		if payable() {
			t.Fatalf("escape offered with %d other cards in the graveyard, want three", i)
		}
		c := NewCard(name, me.ID)
		c.TypeLine = "Instant"
		me.Graveyard.PushTop(c)
	}
	if payable() {
		t.Errorf("escape offered with two other cards in the graveyard, want three")
	}
	third := NewCard("Fuel C", me.ID)
	third.TypeLine = "Instant"
	me.Graveyard.PushTop(third)
	if !payable() {
		t.Errorf("escape withheld with exactly three other cards in the graveyard")
	}
}

// The Condition is still asked, and asked FIRST — Snuff Out with no
// Swamp is not an offer however much life its controller has.
func TestAlternativeCostConditionStillGatesTheOffer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	self := altCostHandCard(me, "Test Snuff Out")
	swamps := false
	offer := &AlternativeCost{
		Key: "pay_life", Life: 4,
		Condition: func(*Game, uuid.UUID) bool { return swamps },
	}
	var got bool
	g.WithWriteLock(func() { got = g.AlternativeCostPayableLocked(me.ID, self, offer) })
	if got {
		t.Errorf("an offer whose condition is false was on the table")
	}
	swamps = true
	g.WithWriteLock(func() { got = g.AlternativeCostPayableLocked(me.ID, self, offer) })
	if !got {
		t.Errorf("an offer whose condition is true was withheld")
	}
	// Nil is never an offer.
	g.WithWriteLock(func() { got = g.AlternativeCostPayableLocked(me.ID, self, nil) })
	if got {
		t.Errorf("a nil offer reported payable")
	}
}
