package game

import "testing"

// damage_dealers_types_test.go — "dealt combat damage by three or more
// Pirates this turn" (Admiral Beckett Brass, #5368): DamageDealtRecord
// keeps the dealer's creature types as it dealt the damage.

func TestCountCreaturesOfTypeThatDealtCombatDamageToCountsDistinctDealersAsTheyWere(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	pirate := func(name string) *Card {
		return &Card{Name: name, TypeLine: "Creature — Human Pirate", Power: 1, Toughness: 1, Owner: opp, Controller: opp}
	}
	a := pushPermanent(g, *pirate("A"))
	b := pushPermanent(g, *pirate("B"))
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp, Controller: opp})

	count := func() int {
		n := 0
		g.WithWriteLock(func() { n = g.CountCreaturesOfTypeThatDealtCombatDamageToThisTurn(me, "Pirate") })
		return n
	}
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: a, Target: me, Amount: 1, Combat: true})
	// First strike then regular damage: one Pirate, not two.
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: a, Target: me, Amount: 1, Combat: true})
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: bear, Target: me, Amount: 2, Combat: true})
	if got := count(); got != 1 {
		t.Fatalf("one Pirate dealt damage twice and a Bear once: counted %d, want 1", got)
	}
	// Noncombat damage does not count.
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: b, Target: me, Amount: 1})
	if got := count(); got != 1 {
		t.Errorf("noncombat damage counted: %d", got)
	}
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: b, Target: me, Amount: 1, Combat: true})
	if got := count(); got != 2 {
		t.Errorf("two Pirates: counted %d", got)
	}
	// The record outlives the dealers, and a changeling counts as any type.
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(a); err != nil {
			t.Fatal(err)
		}
	})
	ling := pushPermanent(g, Card{Name: "Ling", TypeLine: "Creature — Shapeshifter", Keywords: []string{KeywordChangeling}, Power: 1, Toughness: 1, Owner: opp, Controller: opp})
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: ling, Target: me, Amount: 1, Combat: true})
	if got := count(); got != 3 {
		t.Errorf("an exiled Pirate stays counted and a changeling joins: counted %d, want 3", got)
	}
	g.WithWriteLock(func() {
		if got := g.CountCreaturesOfTypeThatDealtCombatDamageToThisTurn(g.Seats[2].ID, "Pirate"); got != 0 {
			t.Errorf("the count is per victim: %d", got)
		}
	})
}
