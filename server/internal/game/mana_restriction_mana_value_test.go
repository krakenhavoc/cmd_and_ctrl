package game

import (
	"errors"
	"testing"
)

// mana_restriction_mana_value_test.go — #4694, the engine half of
// Ashling, Rimebound's "Spend this mana only to cast spells with mana
// value 4 or greater" (CR 106.6, CR 202.3, CR 202.3e). The card is in
// internal/cards/effects/ashling_rekindled_test.go.

var mv4 = []string{ManaRestrictManaValueAtLeast(4)}

// The tag reads the spell's mana value, and only a cast's.
func TestManaValueRestrictionAdmitsOnlyASpellOfThatManaValue(t *testing.T) {
	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"a four-drop", ManaSpendForCast(Card{ManaCost: "{2}{G}{G}", TypeLine: "Creature — Bear"}), true},
		{"a six-drop", ManaSpendForCast(Card{ManaCost: "{6}", TypeLine: "Artifact"}), true},
		{"a three-drop", ManaSpendForCast(Card{ManaCost: "{2}{G}", TypeLine: "Creature — Bear"}), false},
		{"a hybrid four (CR 202.3f)", ManaSpendForCast(Card{ManaCost: "{2/W}{2/W}", TypeLine: "Instant"}), true},
		{"a land-cost spell of mana value 0", ManaSpendForCast(Card{ManaCost: "", TypeLine: "Sorcery"}), false},
		{"an unreadable cost", ManaSpendForCast(Card{ManaCost: "{1}{R} // {2}{U}", TypeLine: "Instant // Sorcery"}), false},
		{"an activation of a four-drop's ability", ManaSpendForAbility(Card{ManaCost: "{2}{G}{G}", TypeLine: "Creature — Bear"}), false},
		{"an activation that claims a mana value", ManaSpendContext{Purpose: SpendPurposeActivate, ManaValue: 9, ManaValueKnown: true}, false},
		{"an unlock", ManaSpendContext{Purpose: SpendPurposeUnlock, ManaValue: 9, ManaValueKnown: true}, false},
		{"a payment that says nothing", ManaSpendContext{ManaValue: 9, ManaValueKnown: true}, false},
		{"a cast whose mana value is not known", ManaSpendContext{Purpose: SpendPurposeCast, ManaValue: 9}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ctx.allows(mv4); got != tc.want {
				t.Errorf("allows = %v, want %v (mana value %d, known %v)", got, tc.want, tc.ctx.ManaValue, tc.ctx.ManaValueKnown)
			}
		})
	}
}

// The comparison generalises: "N or less", and an alternation of the
// two through the keyed tags' `|`.
func TestManaValueRestrictionComparisons(t *testing.T) {
	three := ManaSpendForCast(Card{ManaCost: "{2}{G}", TypeLine: "Sorcery"})
	five := ManaSpendForCast(Card{ManaCost: "{4}{G}", TypeLine: "Sorcery"})
	if !three.allows([]string{ManaRestrictManaValueAtMost(3)}) || five.allows([]string{ManaRestrictManaValueAtMost(3)}) {
		t.Error("ManaRestrictManaValueAtMost(3) does not admit exactly the spells of mana value 3 or less")
	}
	either := []string{ManaRestrictManaValueAtMost(1) + "|" + manaValueAtLeast + "5"}
	if three.allows(either) || !five.allows(either) {
		t.Error("an alternation of comparisons is not read as an OR")
	}
	for _, bad := range []string{"mv:4", "mv:>=", "mv:>=four", "mv:>4"} {
		if five.allows([]string{bad}) {
			t.Errorf("malformed tag %q admitted a cast; an unknown tag must refuse", bad)
		}
	}
}

// CR 202.3e: on the stack {X} counts as the value announced, so the
// context counts the X the cast announces, and WithX recounts it.
func TestManaValueRestrictionCountsTheAnnouncedX(t *testing.T) {
	fireball := Card{ManaCost: "{X}{R}", TypeLine: "Sorcery"}
	if ManaSpendForCast(fireball).allows(mv4) {
		t.Error("{X}{R} with X unannounced (0) was admitted")
	}
	if ManaSpendForCastParams(fireball, CastSpellParams{XValue: 2}).allows(mv4) {
		t.Error("{X}{R} with X = 2 (mana value 3) was admitted")
	}
	if !ManaSpendForCastParams(fireball, CastSpellParams{XValue: 3}).allows(mv4) {
		t.Error("{X}{R} with X = 3 (mana value 4) was refused")
	}
	ctx := ManaSpendForCast(fireball)
	if got := ctx.WithX(5).ManaValue; got != 6 {
		t.Errorf("WithX(5) on {X}{R}: mana value %d, want 6", got)
	}
	if got := ctx.WithX(5).WithX(1).ManaValue; got != 2 {
		t.Errorf("WithX recounts from the base: mana value %d, want 2", got)
	}
	twoX := ManaSpendForCastParams(Card{ManaCost: "{X}{X}{G}", TypeLine: "Sorcery"}, CastSpellParams{XValue: 2})
	if twoX.ManaValue != 5 || !twoX.allows(mv4) {
		t.Errorf("{X}{X}{G} with X = 2: mana value %d, want 5 and admitted", twoX.ManaValue)
	}
	ability := ManaSpendForAbility(Card{ManaCost: "{X}{R}"}).WithX(9)
	if ability.ManaValueKnown || ability.allows(mv4) {
		t.Error("WithX gave an activation a mana value")
	}
}

// CR 712.8c: a transforming card cast transformed has its front face's
// mana value; CR 708.4 / CR 202.3a: a face-down spell's is 0.
func TestManaValueRestrictionReadsTheSpellsFace(t *testing.T) {
	dfc := Card{
		Layout:     LayoutTransform,
		TypeLine:   "Creature — Werewolf",
		Faces:      []Face{{Name: "Front", ManaCost: "{3}{R}", TypeLine: "Creature — Human"}, {Name: "Back", TypeLine: "Creature — Werewolf"}},
		ActiveFace: 1,
		Name:       "Back",
	}
	if got := ManaSpendForCast(dfc); !got.ManaValueKnown || got.ManaValue != 4 {
		t.Errorf("a transformed spell: mana value %d (known %v), want its front face's 4", got.ManaValue, got.ManaValueKnown)
	}
	hidden := Card{ManaCost: "{5}{G}{G}", TypeLine: "Creature — Hydra"}
	hidden.SetFaceDown(FaceDownMorphed)
	if ManaSpendForCast(hidden).allows(mv4) {
		t.Error("a face-down spell, mana value 0, was admitted")
	}
}

// Through CastSpell in strict mode, paying from the pool: the restricted
// mana casts a four-drop and an X spell with X large enough, and refuses
// a three-drop and the X spell at a smaller X.
func TestManaValueManaThroughTheCastPath(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepPrecombatMain)
	restricted := func() {
		me.ManaPool.EmptyPool()
		for range 4 {
			me.ManaPool.AddMana(ManaToken{Color: "R", Restrictions: mv4})
		}
	}
	cast := func(name, cost string, x int) error {
		c := NewCard(name, me.ID)
		c.TypeLine, c.ManaCost = "Instant", cost
		me.Hand.PushTop(c)
		restricted()
		return g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true, XValue: x})
	}
	var short *InsufficientManaError
	if err := cast("Three Drop", "{2}{R}", 0); !errors.As(err, &short) {
		t.Fatalf("a three-drop paid with mana-value-4 mana: err = %v, want InsufficientManaError", err)
	}
	if err := cast("Small Fireball", "{X}{R}", 2); !errors.As(err, &short) {
		t.Fatalf("{X}{R} at X = 2 paid with mana-value-4 mana: err = %v, want InsufficientManaError", err)
	}
	if err := cast("Four Drop", "{2}{R}{R}", 0); err != nil {
		t.Fatalf("a four-drop with mana-value-4 mana: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the four-drop left %d restricted mana in the pool", len(me.ManaPool))
	}
	if err := cast("Big Fireball", "{X}{R}", 3); err != nil {
		t.Fatalf("{X}{R} at X = 3 with mana-value-4 mana: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the X spell left %d restricted mana in the pool", len(me.ManaPool))
	}
}

// The auto-payer plans around the restricted mana for a cheaper spell:
// it taps the lands for the three-drop and leaves the restricted mana
// floating, and it spends the restricted mana on a four-drop.
func TestManaValueManaWithTheAutoPayer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepPrecombatMain)
	for range 3 {
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	}
	me.ManaPool.AddMana(ManaToken{Color: "R", Restrictions: mv4})
	me.ManaPool.AddMana(ManaToken{Color: "R", Restrictions: mv4})

	three := NewCard("Three Drop", me.ID)
	three.TypeLine, three.ManaCost = "Instant", "{2}{R}"
	me.Hand.PushTop(three)
	if err := g.CastSpell(me.ID, three.InstanceID, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("a three-drop with three Mountains beside the restricted mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("after the three-drop the pool holds %d mana, want the 2 restricted mana untouched", got)
	}

	// Two more lands, untapped: the four-drop needs the restricted {R}{R}
	// beside them.
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	four := NewCard("Four Drop", me.ID)
	four.TypeLine, four.ManaCost = "Instant", "{2}{R}{R}"
	me.Hand.PushTop(four)
	if err := g.CastSpell(me.ID, four.InstanceID, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("a four-drop with the restricted mana and two Mountains: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("after the four-drop the pool holds %d mana, want 0", got)
	}
}

// The mana can't pay for an activated ability, whatever the source's
// mana value.
func TestManaValueManaCantPayAnAbility(t *testing.T) {
	cost, err := ParseCost("{1}")
	if err != nil {
		t.Fatal(err)
	}
	pool := ManaPool{{Color: "R", Restrictions: mv4}}
	if pool.CanPayFor(cost, 0, ManaSpendForAbility(Card{ManaCost: "{7}", TypeLine: "Artifact"})) {
		t.Error("mana-value-4 mana paid an activation of a seven-drop's ability")
	}
	if !pool.CanPayFor(cost, 0, ManaSpendForCast(Card{ManaCost: "{7}", TypeLine: "Artifact"})) {
		t.Error("mana-value-4 mana did not pay a seven-drop")
	}
}

// CR 500.5, CR 106.4: the mana empties as the step ends like any other.
func TestManaValueManaEmptiesAtTheEndOfTheStep(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R", Restrictions: mv4})
	me.ManaPool.AddMana(ManaToken{Color: "R", Restrictions: mv4})
	advanceIntoStep(t, g, StepBeginCombat)
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("the pool holds %d mana after the main phase ended, want 0", got)
	}
}

// The bot's matcher, which holds no Card: the same answers from a type
// line and a mana cost, X counted.
func TestCastAdmitsRestrictedManaReadsTheManaValue(t *testing.T) {
	for _, tc := range []struct {
		name, cost string
		x          int
		want       bool
	}{
		{"a four-drop", "{2}{R}{R}", 0, true},
		{"a three-drop", "{2}{R}", 0, false},
		{"an X spell at X = 3", "{X}{R}", 3, true},
		{"an X spell at X = 2", "{X}{R}", 2, false},
		{"an unreadable cost", "{1}{R} // {2}{U}", 0, false},
	} {
		if got := CastAdmitsRestrictedMana(mv4, "Sorcery", []string{"R"}, false, "", tc.cost, tc.x); got != tc.want {
			t.Errorf("%s: admits = %v, want %v", tc.name, got, tc.want)
		}
	}
}
