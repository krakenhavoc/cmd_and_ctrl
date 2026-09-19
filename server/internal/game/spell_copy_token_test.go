package game

import (
	"reflect"
	"testing"
)

// spell_copy_token_test.go — CR 608.3f's token, and the line between
// what a copy COPIES and what merely travels with it.
//
// CR 707.2 says the copiable values are the printed characteristics,
// as modified by other copy effects, by face-down status and by
// "as … enters". Three things that landed alongside this work are
// emphatically NOT in that set, and each is pinned below, because the
// cheap mistake in every one of them is to reach for PrintedValues:
//
//	StackItem.Foretold      (#987) — how the spell was CAST
//	Card.FaceDownKind       (#987) — an exile status, not a permanent's
//	PaidCost.OptionalCosts  (#988) — what was PAID
//
// The third one still reaches the token, because CR 707.10b copies
// the choices made when casting and CR 400.7d carries them onto the
// permanent — but it gets there as per-instance state stamped at
// entry, the way the ordinary permanent branch stamps it, and never
// as a characteristic.

// TestAForetoldSpellsCopyIsNotForetold — CR 707.10. The copy was
// never cast, so it was not cast for a foretell cost: nothing that
// reads "was this cast from a foretold card" may see the copy.
func TestAForetoldSpellsCopyIsNotForetold(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "BBB")
	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{Strict: true})

	g.mu.Lock()
	item.Foretold = true
	_ = g.CopySpellForEffect(item.ID, me.ID, false, nil)
	var copyItem *StackItem
	for id, it := range g.StackMeta {
		if id != item.ID {
			copyItem = it
		}
	}
	g.mu.Unlock()

	if copyItem == nil {
		t.Fatal("no copy was created")
	}
	if copyItem.Foretold {
		t.Error("the copy reports as foretold — a copy is created, not cast (CR 707.10)")
	}
	if !item.Foretold {
		t.Error("the ORIGINAL stopped being foretold")
	}
}

// TestACopysExceptClauseDoesNotReachTheOriginal — the except clause
// edits a value copy taken off the stack, so the spell being copied
// is untouched however the clause is written.
func TestACopysExceptClauseDoesNotReachTheOriginal(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "BBB")
	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{Strict: true})

	g.mu.Lock()
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == item.ID {
			g.Stack.Cards[i].TypeLine = "Legendary Creature — Bear"
		}
	}
	_ = g.CopySpellForEffect(item.ID, me.ID, false, func(v *PrintedValues) {
		v.RemoveSupertype("Legendary")
	})
	var original, copied Card
	for _, c := range g.Stack.Cards {
		if c.InstanceID == item.ID {
			original = c
		} else {
			copied = c
		}
	}
	g.mu.Unlock()

	if copied.InstanceID == original.InstanceID {
		t.Fatal("no copy was created")
	}
	if copied.TypeLine != "Creature — Bear" {
		t.Errorf("the copy's type line = %q, want the except clause applied", copied.TypeLine)
	}
	if original.TypeLine != "Legendary Creature — Bear" {
		t.Errorf("the ORIGINAL's type line = %q — the except clause reached the copied spell", original.TypeLine)
	}
}

// TestTheTokenACopyBecomesCarriesNoCastState is the CR 707.2 line,
// stated as an assertion. The token a resolving copy becomes is built
// from the copiable values and the Token supertype and nothing else:
// no face-down status, no foretell, no cost record.
func TestTheTokenACopyBecomesCarriesNoCastState(t *testing.T) {
	src := Card{
		Name:         "Grizzly Bears",
		OracleID:     "oracle-bears",
		TypeLine:     "Legendary Creature — Bear",
		ManaCost:     "{1}{G}",
		Power:        2,
		Toughness:    2,
		FaceDown:     true,
		FaceDownKind: FaceDownForetold,
		Provenance:   CastProvenance{OptionalCosts: []int{0, 0}},
		Tapped:       true,
		DamageMarked: 3,
		Counters:     map[string]int{"+1/+1": 2},
	}

	tok := tokenCopyOfSpell(src)

	if !tok.IsToken() {
		t.Fatalf("type line = %q, want the Token supertype", tok.TypeLine)
	}
	if tok.TypeLine != "Token Legendary Creature — Bear" {
		t.Errorf("type line = %q", tok.TypeLine)
	}
	if tok.FaceDown || tok.FaceDownKind != FaceDownNone {
		t.Errorf("the token is face down (%v / %q) — face-down status is not carried by CopiableValuesOf",
			tok.FaceDown, tok.FaceDownKind)
	}
	if tok.Provenance.Any() {
		t.Errorf("the token template carries cast provenance %+v — CR 707.2, that is not a characteristic",
			tok.Provenance)
	}
	if tok.Tapped || tok.DamageMarked != 0 || len(tok.Counters) != 0 {
		t.Error("status, damage or counters came across; none of them is a copiable value")
	}
	if tok.Power != 2 || tok.Toughness != 2 || tok.Name != "Grizzly Bears" {
		t.Errorf("the printed characteristics did not: %d/%d %q", tok.Power, tok.Toughness, tok.Name)
	}
}

// TestTheTokenIsNotBornACopyOfItself — PrintedSelf is the pre-copy
// stash a PERMANENT reverts to on the way out (CR 400.7). A token has
// nothing to revert to and never leaves the battlefield as an object,
// so it must not carry one.
func TestTheTokenIsNotBornACopyOfItself(t *testing.T) {
	tok := tokenCopyOfSpell(Card{
		Name: "Grizzly Bears", OracleID: "oracle-bears",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	if tok.IsCopy() {
		t.Error("the token reports as a copy — it IS the thing, there is nothing to revert to")
	}
	// And its own copiable values are what a later Clone would take.
	v := CopiableValuesOf(tok)
	if v.Name != "Grizzly Bears" || v.Power != 2 {
		t.Errorf("copiable values off the token: %+v", v)
	}
	if !reflect.DeepEqual(v.GrantedAbilities, []string(nil)) {
		t.Errorf("the token carries grants it was never given: %v", v.GrantedAbilities)
	}
}
