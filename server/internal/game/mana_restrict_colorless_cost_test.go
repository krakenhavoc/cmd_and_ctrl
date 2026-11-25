package game

import "testing"

// mana_restrict_colorless_cost_test.go — #6259: Cultivator Drone's
// "Spend this mana only to cast a colorless spell, activate an ability of
// a colorless permanent, or pay a cost that contains {C}." CR 106.6: the
// restriction changes what the mana may be spent on, never its type.

func cultivatorDroneMana() ManaToken {
	return ManaToken{Color: "C", Restrictions: []string{ManaRestrictAnyOf(
		[]string{ManaRestrictCast, ManaRestrictColorless},
		[]string{ManaRestrictActivate, ManaRestrictColorless},
		[]string{ManaRestrictColorlessCost},
	)}}
}

// The acceptance: the mana pays a blue spell whose cost contains {C}
// (all of it, generic included), but not a blue spell costing {1}{U}.
func TestColorlessCostManaPaysOnlyACostWithC(t *testing.T) {
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[0]
	withC := pushHandSpell(me, "Blue Seer", "Instant", "{1}{C}{U}")
	without := pushHandSpell(me, "Blue Bolt", "Instant", "{1}{U}")

	me.ManaPool.AddMana(cultivatorDroneMana(), ManaToken{Color: "U"})
	wantInsufficient(t, g.CastSpell(me.ID, without, CastSpellParams{Strict: true}), "drone mana for a {1}{U} spell")
	me.ManaPool = nil
	me.ManaPool.AddMana(cultivatorDroneMana(), cultivatorDroneMana(), ManaToken{Color: "U"})
	if err := g.CastSpell(me.ID, withC, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("drone mana for {1}{C}{U}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// The other two clauses still work: a colourless spell takes it.
func TestColorlessCostManaStillPaysAColorlessSpell(t *testing.T) {
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[0]
	rock := pushHandSpell(me, "Rock", "Artifact", "{2}")
	me.ManaPool.AddMana(cultivatorDroneMana(), cultivatorDroneMana())
	if err := g.CastSpell(me.ID, rock, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("drone mana for a colourless {2} spell: %v", err)
	}
}

func TestColorlessCostRestrictionReadsTheContext(t *testing.T) {
	r := []string{ManaRestrictColorlessCost}
	cases := []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"cast, cost has {C}", ManaSpendContext{Purpose: SpendPurposeCast, CostHasC: true}, true},
		{"activate, cost has {C}", ManaSpendContext{Purpose: SpendPurposeActivate, CostHasC: true}, true},
		{"cast, no {C}", ManaSpendContext{Purpose: SpendPurposeCast}, false},
		{"unknown payment", ManaSpendContext{CostHasC: true}, false},
	}
	for _, c := range cases {
		if got := c.ctx.allows(r); got != c.want {
			t.Errorf("%s: allows = %v, want %v", c.name, got, c.want)
		}
	}
	for cost, want := range map[string]bool{"{3}{C}": true, "{C/W}": true, "{2}{S}": false, "{2}{U}": false, "": false} {
		if got := CostStringHasColorless(cost); got != want {
			t.Errorf("CostStringHasColorless(%q) = %v, want %v", cost, got, want)
		}
	}
	if got := ManaSpendForCast(Card{ManaCost: "{4}{C}", TypeLine: "Creature — Eldrazi"}).CostHasC; !got {
		t.Error("ManaSpendForCast of a {4}{C} spell: CostHasC false")
	}
}
