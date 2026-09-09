package game

import (
	"testing"

	"github.com/google/uuid"
)

// additional_cost_test.go — S21 sub-PR 5. The behaviour worth
// pinning is the ORDER: the discard is paid at announce, with the
// spell on the stack, so a discard payoff triggers above it.

func withCatalogAdditionalCost(t *testing.T, fn func(oracleID string) *AdditionalCost) {
	t.Helper()
	prev := CatalogAdditionalCost
	CatalogAdditionalCost = fn
	t.Cleanup(func() { CatalogAdditionalCost = prev })
}

// thrillInHand seeds a "discard a card" instant plus `filler` spare
// cards in hand, and walks the turn to a main phase.
func thrillInHand(t *testing.T, g *Game, me *Player, oracle string, filler int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	// Start from an empty hand so the opening seven can't be
	// mistaken for the cards the test discards.
	me.Hand.Cards = nil
	spell := NewCard("Thrill of Possibility", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	spares := make([]uuid.UUID, 0, filler)
	for i := 0; i < filler; i++ {
		c := NewCard("Spare", me.ID)
		c.TypeLine = "Sorcery"
		me.Hand.PushTop(c)
		spares = append(spares, c.InstanceID)
	}
	return spell.InstanceID, spares
}

func TestCastSpellRejectsBadAdditionalCostPayment(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-thrill"
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == oracle {
			return &AdditionalCost{DiscardCards: 1, Label: "Discard a card"}
		}
		return nil
	})
	spell, spares := thrillInHand(t, g, me, oracle, 2)

	oppCard := NewCard("Theirs", opp.ID)
	oppCard.TypeLine = "Sorcery"
	opp.Hand.PushTop(oppCard)

	cases := []struct {
		name     string
		discards []uuid.UUID
		want     error
	}{
		{"none", nil, ErrInvalidParam},
		{"too many", []uuid.UUID{spares[0], spares[1]}, ErrInvalidParam},
		{"the spell itself", []uuid.UUID{spell}, ErrInvalidParam},
		{"not in hand", []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"another player's card", []uuid.UUID{oppCard.InstanceID}, ErrCardNotFound},
	}
	for _, tc := range cases {
		err := g.CastSpell(me.ID, spell, CastSpellParams{DiscardIDs: tc.discards})
		if err != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, err, tc.want)
		}
	}
	// Every rejection must be total: nothing discarded, spell still
	// in hand.
	if !me.Hand.Contains(spell) {
		t.Errorf("rejected cast left the spell out of hand")
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("rejected cast paid the cost anyway: graveyard has %d", me.Graveyard.Size())
	}
	// A duplicate pick can't pay a two-card cost twice.
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == oracle {
			return &AdditionalCost{DiscardCards: 2}
		}
		return nil
	})
	if err := g.CastSpell(me.ID, spell, CastSpellParams{DiscardIDs: []uuid.UUID{spares[0], spares[0]}}); err != ErrInvalidParam {
		t.Errorf("duplicate discard: %v, want ErrInvalidParam", err)
	}
}

// A card with no additional cost that arrives with discard IDs is a
// client bug, and rejected rather than silently ignored.
func TestCastSpellRejectsUnexpectedDiscardIDs(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogAdditionalCost(t, func(string) *AdditionalCost { return nil })
	spell, spares := thrillInHand(t, g, me, "no-such-cost", 1)
	if err := g.CastSpell(me.ID, spell, CastSpellParams{DiscardIDs: []uuid.UUID{spares[0]}}); err != ErrInvalidParam {
		t.Fatalf("unexpected discard_ids: %v, want ErrInvalidParam", err)
	}
}

func TestCastSpellPaysAdditionalCostWithSpellOnStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-thrill"
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == oracle {
			return &AdditionalCost{DiscardCards: 1, Label: "Discard a card"}
		}
		return nil
	})
	spell, spares := thrillInHand(t, g, me, oracle, 1)

	// Watch the order of the two events the cast produces.
	probe := &castCostProbe{}
	g.RegisterListener(probe)

	if err := g.CastSpell(me.ID, spell, CastSpellParams{DiscardIDs: []uuid.UUID{spares[0]}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if len(probe.order) != 2 || probe.order[0] != EventDiscardCard || probe.order[1] != EventCast {
		t.Errorf("event order = %v, want [discard_card cast]", probe.order)
	}
	if probe.stackAtDiscard != 1 {
		t.Errorf("stack size when the cost was paid = %d, want 1 (the spell)", probe.stackAtDiscard)
	}
	if me.Hand.Contains(spares[0]) {
		t.Errorf("discarded card still in hand")
	}
	if !me.Graveyard.Contains(spares[0]) {
		t.Errorf("discarded card never reached the graveyard")
	}
	if !g.Stack.Contains(spell) {
		t.Errorf("spell not on the stack")
	}
	if me.Hand.Size() != 0 {
		t.Errorf("hand size = %d, want 0 (spell cast, spare discarded)", me.Hand.Size())
	}
}

// The cost is paid on the way to the stack, so countering the spell
// afterwards doesn't give the card back.
func TestAdditionalCostSurvivesTheSpellFizzling(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-thrill"
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == oracle {
			return &AdditionalCost{DiscardCards: 1}
		}
		return nil
	})
	spell, spares := thrillInHand(t, g, me, oracle, 1)
	if err := g.CastSpell(me.ID, spell, CastSpellParams{DiscardIDs: []uuid.UUID{spares[0]}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.CounterSpell(spell, nil); err != nil {
		t.Fatalf("counter: %v", err)
	}
	if !me.Graveyard.Contains(spares[0]) {
		t.Errorf("countering the spell refunded the additional cost")
	}
}

// castCostProbe records the discard / cast event order and the stack
// depth at the moment the cost was paid.
type castCostProbe struct {
	order          []EventKind
	stackAtDiscard int
}

func (p *castCostProbe) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventDiscardCard:
		p.order = append(p.order, ev.Kind)
		// CR 601.2a puts the spell on the stack before costs are
		// paid — the whole reason this is an additional cost and not
		// part of the spell's effect.
		p.stackAtDiscard = g.Stack.Size()
	case EventCast:
		p.order = append(p.order, ev.Kind)
	}
}
