package game

import (
	"testing"

	"github.com/google/uuid"
)

// alternative_cost_test.go — S22. Three things worth pinning, in
// descending order of "a bug here would be invisible":
//
//  1. INSTEAD, not alongside. The printed cost has to stop being
//     charged. A version that added the alternative cost to the mana
//     cost would pass every test that only checks the cast
//     succeeded.
//  2. The clause rewrite. Overload deletes the target clause, cleave
//     swaps it — and both have to hold at announce AND at the CR
//     608.2b re-check, or an overloaded spell fizzles the moment the
//     board changes.
//  3. Additional costs survive the swap. Nothing about paying an
//     alternative cost excuses "discard a card".

func withCatalogAlternativeCosts(t *testing.T, fn func(oracleID string) []AlternativeCost) {
	t.Helper()
	prev := CatalogAlternativeCosts
	CatalogAlternativeCosts = fn
	t.Cleanup(func() { CatalogAlternativeCosts = prev })
}

// altCostFor wires a single card's offers and returns the hook body.
func altCostFor(oracle string, costs ...AlternativeCost) func(string) []AlternativeCost {
	return func(id string) []AlternativeCost {
		if id == oracle {
			return costs
		}
		return nil
	}
}

// blastInHand seeds a Vandalblast-shaped sorcery — cheap printed
// cost, expensive overload — in the active seat's hand at a main
// phase.
func blastInHand(t *testing.T, g *Game, me *Player, oracle string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Blast", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{R}"
	c.OracleID = oracle
	me.Hand.PushTop(c)
	return c.InstanceID
}

// The whole point: paying the alternative cost replaces the printed
// one. Under strict mana a pool that covers {R} must NOT cover the
// {4}{R} overload, and vice versa — a pool of five must not be told
// it also owes the printed {R}.
func TestAlternativeCostReplacesTheManaCost(t *testing.T) {
	const oracle = "test-blast"

	t.Run("too poor to overload", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
			Key: "overload", Label: "Overload {4}{R}", ManaCost: "{4}{R}",
		}))
		id := blastInHand(t, g, me, oracle)
		me.ManaPool.AddMana(ManaToken{Color: "R"})

		err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "overload"})
		var im *InsufficientManaError
		if !errorsAs(err, &im) {
			t.Fatalf("overload on one red mana: got %v, want *InsufficientManaError", err)
		}
		if !me.Hand.Contains(id) {
			t.Errorf("rejected cast left the card out of hand")
		}
		if len(me.ManaPool) != 1 {
			t.Errorf("rejected cast spent mana: %+v", me.ManaPool)
		}
		// The very same pool pays the printed cost.
		if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("printed cost on one red mana: %v", err)
		}
	})

	t.Run("rich enough, and charged only once", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
			Key: "overload", Label: "Overload {4}{R}", ManaCost: "{4}{R}",
		}))
		id := blastInHand(t, g, me, oracle)
		me.ManaPool.AddMana(
			ManaToken{Color: "R"},
			ManaToken{Color: "C"}, ManaToken{Color: "C"},
			ManaToken{Color: "C"}, ManaToken{Color: "C"},
		)

		if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "overload"}); err != nil {
			t.Fatalf("overload on exactly {4}{R}: %v", err)
		}
		// Exactly five spent. Six would mean the printed {R} was
		// charged on top — the bug this whole file exists to catch.
		if len(me.ManaPool) != 0 {
			t.Errorf("post-spend pool: %+v, want empty", me.ManaPool)
		}
		item := g.StackMeta[id]
		if item == nil || item.AltCost != "overload" {
			t.Errorf("stack item did not record the cost paid: %+v", item)
		}
	})
}

// Declining is a first-class answer: the offer changes nothing until
// it's claimed.
func TestAlternativeCostDeclinedKeepsThePrintedCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-blast"
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "overload", ManaCost: "{4}{R}",
	}))
	id := blastInHand(t, g, me, oracle)
	me.ManaPool.AddMana(ManaToken{Color: "R"})

	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("declining the offer should cast for the printed cost: %v", err)
	}
	if item := g.StackMeta[id]; item == nil || item.AltCost != "" {
		t.Errorf("declined cast recorded an alternative cost: %+v", item)
	}
}

// A key the card doesn't offer is rejected outright rather than
// falling back to the printed cost — a silent fallback would charge
// full price for a cast the player meant to overload.
func TestCastSpellRejectsUnknownAlternativeCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-blast"
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "overload", ManaCost: "{4}{R}",
	}))
	id := blastInHand(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{AlternativeCost: "kicker"}); err != ErrInvalidParam {
		t.Errorf("unknown key: got %v, want ErrInvalidParam", err)
	}
	// And a card that offers nothing at all.
	plain := blastInHand(t, g, me, "no-such-card")
	if err := g.CastSpell(me.ID, plain, CastSpellParams{AlternativeCost: "overload"}); err != ErrInvalidParam {
		t.Errorf("offer-less card: got %v, want ErrInvalidParam", err)
	}
	if !me.Hand.Contains(id) || !me.Hand.Contains(plain) {
		t.Errorf("rejected casts must leave both cards in hand")
	}
	if g.Stack.Size() != 0 {
		t.Errorf("rejected cast reached the stack")
	}
}

// Overload deletes the target clause. The spell announces with no
// targets — which also means it can't be fizzled by removing "the"
// target in response.
func TestOverloadClearsTheTargetClause(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-blast"
	victim := pushColoredCreature(g, opp, "White Knight", "{W}{W}")
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "overload", ManaCost: "{4}{R}", ClearsTargets: true,
	}))
	id := blastInHand(t, g, me, oracle)

	// Hard-cast, the clause is live: no target is a rejection.
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != ErrInvalidParam {
		t.Errorf("targeted cast with no targets: got %v, want ErrInvalidParam", err)
	}
	// Overloaded, sending one anyway is the rejection instead.
	err := g.CastSpell(me.ID, id, CastSpellParams{
		AlternativeCost: "overload",
		Targets:         []TargetRef{{Kind: TargetCard, ID: victim}},
	})
	if err != ErrInvalidParam {
		t.Errorf("overload with targets: got %v, want ErrInvalidParam", err)
	}
	// Overloaded with none is the legal shape.
	if err := g.CastSpell(me.ID, id, CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload with no targets: %v", err)
	}
	if item := g.StackMeta[id]; item == nil || len(item.Targets) != 0 {
		t.Errorf("overloaded item carries targets: %+v", item)
	}
	// CR 608.2b: with no targeted slot there is nothing to fizzle,
	// so the spell survives the board emptying underneath it.
	if fizzles(g, oracle, id) {
		t.Errorf("overloaded spell fizzled before the board even changed")
	}
	g.WithWriteLock(func() {
		g.Battlefield.Cards = nil
	})
	if fizzles(g, oracle, id) {
		t.Errorf("overloaded spell fizzled with no targets to lose")
	}
}

// fizzles runs the CR 608.2b re-check the resolution path runs, under
// the same lock, against the clause the item was actually cast with.
func fizzles(g *Game, oracleID string, itemID uuid.UUID) bool {
	out := false
	g.WithWriteLock(func() {
		item := g.StackMeta[itemID]
		out = spellAllTargetsIllegalLocked(g, item, castTargetSpecForItem(oracleID, item))
	})
	return out
}

// Cleave is the other rewrite: a different clause, not an absent
// one. The alternative cost's spec governs announce and the CR
// 608.2b re-check alike.
func TestAlternativeCostSwapsTheTargetClause(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-wash"
	white := pushColoredCreature(g, opp, "White Knight", "{W}{W}")
	black := pushColoredCreature(g, opp, "Black Knight", "{B}{B}")
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	// The cleaved clause drops the "non-black" restriction.
	anyCreature := &TargetSpec{
		Mode:  "creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "cleave", ManaCost: "{1}{U}{U}", Targets: anyCreature,
	}))
	id := blastInHand(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: black}},
	}); err != ErrIllegalTarget {
		t.Errorf("printed clause on a black creature: got %v, want ErrIllegalTarget", err)
	}
	// The printed clause is otherwise unchanged by the offer's
	// existence — the white knight is still fair game for {R}.
	lt := g.LegalTargetsForEffect(me.ID, nonBlackCreatureSpec())
	found := false
	for _, c := range lt.Cards {
		if c == white {
			found = true
		}
	}
	if !found {
		t.Errorf("white knight missing from the printed clause's legal set: %v", lt.Cards)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{
		AlternativeCost: "cleave",
		Targets:         []TargetRef{{Kind: TargetCard, ID: black}},
	}); err != nil {
		t.Fatalf("cleaved clause on a black creature: %v", err)
	}
	// The re-check reads the cleaved clause back off the item, not
	// the printed one — otherwise the spell would fizzle on
	// resolution against the target it was legally cast at.
	if fizzles(g, oracle, id) {
		t.Errorf("cleaved spell fizzles against its own legal target")
	}
}

// An alternative cost replaces the MANA cost. It is not a discount
// on the additional costs, which CR 601.2f evaluates independently.
func TestAlternativeCostDoesNotWaiveAdditionalCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-both"
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "overload", ManaCost: "{4}{R}", ClearsTargets: true,
	}))
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == oracle {
			return &AdditionalCost{DiscardCards: 1, Label: "Discard a card"}
		}
		return nil
	})
	advanceTo(t, g, StepPrecombatMain)
	me.Hand.Cards = nil
	spell := NewCard("Both", me.ID)
	spell.TypeLine = "Sorcery"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	fodder := NewCard("Fodder", me.ID)
	fodder.TypeLine = "Sorcery"
	me.Hand.PushTop(fodder)

	// Overloading without paying the discard is still a rejection.
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		AlternativeCost: "overload",
	}); err != ErrInvalidParam {
		t.Errorf("overload without the discard: got %v, want ErrInvalidParam", err)
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("rejected cast paid the additional cost anyway")
	}
	// Paying both is the legal shape, and the discard really happens.
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		AlternativeCost: "overload",
		DiscardIDs:      []uuid.UUID{fodder.InstanceID},
	}); err != nil {
		t.Fatalf("overload with the discard: %v", err)
	}
	if !me.Graveyard.Contains(fodder.InstanceID) {
		t.Errorf("additional cost was waived by the alternative cost")
	}
}

// The cost paid survives a Clone, so an undo restores a spell to the
// same card it was cast as rather than to its printed self.
func TestAlternativeCostSurvivesClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-blast"
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "overload", ManaCost: "{4}{R}", ClearsTargets: true,
	}))
	id := blastInHand(t, g, me, oracle)
	if err := g.CastSpell(me.ID, id, CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	clone := g.Clone()
	item := clone.StackMeta[id]
	if item == nil || item.AltCost != "overload" {
		t.Errorf("clone lost the alternative cost: %+v", item)
	}
	if item.CastFromZone != ZoneHand {
		t.Errorf("clone lost the cast source zone: %q", item.CastFromZone)
	}
}

// Wash Away's fact: where a spell was cast from is stamped at
// announce, because CR 601.2a moves the card a moment later and
// nothing on it remembers.
func TestCastRecordsSourceZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := blastInHand(t, g, me, "plain-card")
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	item := g.StackMeta[id]
	if item == nil {
		t.Fatalf("cast produced no stack item")
	}
	if item.CastFromZone != ZoneHand {
		t.Fatalf("hand cast: got %q, want %q", item.CastFromZone, ZoneHand)
	}
}
