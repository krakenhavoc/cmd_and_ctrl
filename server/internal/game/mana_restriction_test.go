package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_restriction_test.go — the spend half of restricted mana (#352).
//
// Every test in this file exists because the production half alone is
// a #259 bug. "Add {C}{C}, spend only on colorless Eldrazi" with no
// spend-time check is plain {C}{C} on a land with no drawback, and
// the card ships strictly better than printed. The first test is
// therefore the one that matters most: restricted mana must be
// REFUSED for a spell it does not name.

func creatureSpendContext() ManaSpendContext {
	return ManaSpendContext{
		Purpose: SpendPurposeCast,
		Types:   []string{"Creature"},
		Colors:  []string{"G"},
	}
}

func instantSpendContext() ManaSpendContext {
	return ManaSpendContext{
		Purpose: SpendPurposeCast,
		Types:   []string{"Instant"},
		Colors:  []string{"R"},
	}
}

// The headline guarantee. Ancient Ziggurat's {G} is not general mana.
func TestRestrictedManaCannotPayForAnUnmatchedSpell(t *testing.T) {
	pool := ManaPool{{
		Color:        "G",
		Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
	}}
	cost, err := ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if pool.CanPayFor(cost, 0, instantSpendContext()) {
		t.Fatal("creature-only mana paid for an instant — the #259 direction")
	}
	if !pool.CanPayFor(cost, 0, creatureSpendContext()) {
		t.Error("creature-only mana refused a creature spell")
	}
}

// A refused payment must not consume the mana. Failing "loudly but
// destructively" would be its own bug.
func TestRefusedRestrictedSpendLeavesThePoolIntact(t *testing.T) {
	pool := ManaPool{{
		Color:        "G",
		Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
	}}
	cost, _ := ParseCost("{G}")
	if _, ok := pool.SpendManaFor(cost, 0, instantSpendContext()); ok {
		t.Fatal("spend succeeded on a restriction mismatch")
	}
	if len(pool) != 1 {
		t.Fatalf("pool = %v, want the token untouched", pool)
	}
}

// The zero context is the conservative default: a caller that has not
// said what it is paying for gets no restricted mana. CanPay /
// SpendMana / Missing are the shorthands that take it.
func TestZeroSpendContextRefusesEveryRestriction(t *testing.T) {
	pool := ManaPool{{Color: "G", Restrictions: []string{ManaRestrictCast}}}
	cost, _ := ParseCost("{G}")
	if pool.CanPay(cost, 0) {
		t.Error("CanPay's zero context admitted restricted mana")
	}
	if got := pool.Missing(cost, 0); len(got) != 1 || got[0] != "{G}" {
		t.Errorf("Missing = %v, want [{G}]", got)
	}
}

// Restricted mana is use-it-or-lose-it, so when it is legal here it
// should be spent BEFORE equivalent unrestricted mana. Getting this
// backwards is not a correctness bug but it wastes the token every
// time, which in play is indistinguishable from the card not working.
func TestRestrictedManaIsSpentBeforeUnrestrictedMana(t *testing.T) {
	restricted := ManaToken{Color: "G", Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")}}
	free := ManaToken{Color: "G", Source: uuid.New()}
	pool := ManaPool{free, restricted}
	cost, _ := ParseCost("{G}")

	if _, ok := pool.SpendManaFor(cost, 0, creatureSpendContext()); !ok {
		t.Fatal("spend failed")
	}
	if len(pool) != 1 {
		t.Fatalf("pool = %v, want one token left", pool)
	}
	if len(pool[0].Restrictions) != 0 {
		t.Error("the unrestricted token was spent first; the restricted one will be wasted at end of step")
	}
}

// Generic costs go through the same filter as coloured ones. This is
// the loophole that would otherwise exist: a restricted token refused
// for {G} but silently accepted as generic.
func TestRestrictedManaIsFilteredForGenericCostsToo(t *testing.T) {
	pool := ManaPool{{Color: "G", Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")}}}
	cost, _ := ParseCost("{1}")
	if pool.CanPayFor(cost, 0, instantSpendContext()) {
		t.Fatal("restricted mana paid a generic cost for a spell it does not name")
	}
	if !pool.CanPayFor(cost, 0, creatureSpendContext()) {
		t.Error("restricted mana refused a generic cost it should cover")
	}
}

// Multiple tags are an AND. Eldrazi Temple's mana needs the object to
// be colorless AND an Eldrazi.
func TestMultipleRestrictionsAllMustHold(t *testing.T) {
	pool := ManaPool{{
		Color:        "C",
		Restrictions: []string{ManaRestrictColorless, ManaRestrictSubtype("Eldrazi")},
	}}
	cost, _ := ParseCost("{C}")

	colorlessEldrazi := ManaSpendContext{
		Purpose:  SpendPurposeCast,
		Types:    []string{"Creature"},
		Subtypes: []string{"Eldrazi"},
	}
	coloredEldrazi := ManaSpendContext{
		Purpose:  SpendPurposeCast,
		Types:    []string{"Creature"},
		Subtypes: []string{"Eldrazi"},
		Colors:   []string{"B"},
	}
	colorlessGolem := ManaSpendContext{
		Purpose:  SpendPurposeCast,
		Types:    []string{"Artifact", "Creature"},
		Subtypes: []string{"Golem"},
	}

	if !pool.CanPayFor(cost, 0, colorlessEldrazi) {
		t.Error("refused a colorless Eldrazi")
	}
	if pool.CanPayFor(cost, 0, coloredEldrazi) {
		t.Error("a BLACK Eldrazi is not a colorless Eldrazi")
	}
	if pool.CanPayFor(cost, 0, colorlessGolem) {
		t.Error("a colorless non-Eldrazi is not an Eldrazi")
	}
}

// Eldrazi Temple's mana funds a cast OR an activation, so it carries
// no purpose tag; Shrine of the Forsaken Gods' names casting only.
func TestPurposeTagsSeparateCastingFromActivating(t *testing.T) {
	castOnly := ManaPool{{Color: "C", Restrictions: []string{ManaRestrictCast, ManaRestrictColorless}}}
	eitherWay := ManaPool{{Color: "C", Restrictions: []string{ManaRestrictColorless}}}
	cost, _ := ParseCost("{C}")

	activating := ManaSpendContext{Purpose: SpendPurposeActivate, Types: []string{"Creature"}}
	if castOnly.CanPayFor(cost, 0, activating) {
		t.Error("cast-only mana paid an activation cost")
	}
	if !eitherWay.CanPayFor(cost, 0, activating) {
		t.Error("untagged-purpose mana refused an activation")
	}
}

// An unrecognised tag makes the token unspendable rather than
// unrestricted. A card file that invents a tag the matcher doesn't
// know must fail closed.
func TestUnknownRestrictionTagDenies(t *testing.T) {
	pool := ManaPool{{Color: "G", Restrictions: []string{"spend_on_tuesdays"}}}
	cost, _ := ParseCost("{G}")
	if pool.CanPayFor(cost, 0, creatureSpendContext()) {
		t.Error("an unknown restriction tag was treated as no restriction")
	}
}

// Type matching is case-insensitive so a card file that writes
// "creature" does not silently mint dead mana.
func TestRestrictionMatchingIsCaseInsensitive(t *testing.T) {
	pool := ManaPool{{Color: "G", Restrictions: []string{ManaRestrictCast, ManaRestrictType("creature")}}}
	cost, _ := ParseCost("{G}")
	if !pool.CanPayFor(cost, 0, creatureSpendContext()) {
		t.Error("lowercase type tag failed to match a Creature")
	}
}

// Ordinary mana is unaffected by any of this. The overwhelming
// majority of every pool has no restrictions at all and must keep
// paying for everything.
func TestUnrestrictedManaPaysForAnything(t *testing.T) {
	pool := ManaPool{{Color: "R"}, {Color: "C"}}
	cost, _ := ParseCost("{1}{R}")
	for _, ctx := range []ManaSpendContext{{}, creatureSpendContext(), instantSpendContext()} {
		if !pool.CanPayFor(cost, 0, ctx) {
			t.Errorf("unrestricted pool refused cost under context %+v", ctx)
		}
	}
}

// ManaSpendForCast reads the card's characteristics so a card file
// never has to build a context by hand.
func TestManaSpendForCastReadsTheCard(t *testing.T) {
	c := NewCard("Ulamog, the Ceaseless Hunger", uuid.Nil)
	c.TypeLine = "Legendary Creature — Eldrazi"
	c.ManaCost = "{10}"
	ctx := ManaSpendForCast(c)
	if ctx.Purpose != SpendPurposeCast {
		t.Errorf("purpose = %q, want cast", ctx.Purpose)
	}
	pool := ManaPool{{Color: "C", Restrictions: []string{ManaRestrictColorless, ManaRestrictSubtype("Eldrazi")}}}
	cost, _ := ParseCost("{C}")
	if !pool.CanPayFor(cost, 0, ctx) {
		t.Error("Eldrazi Temple mana refused a colorless Eldrazi built from a real Card")
	}
}

// TestChangelingSatisfiesASubtypeRestriction â CR 702.73a. Cavern of
// Souls named for Elf pays for Universal Automaton, which prints
// Shapeshifter and no Elf anywhere.
//
// The spell is built through ManaSpendForCast rather than by setting
// the context field by hand, because the thing under test is the
// projection: a changeling's Subtypes list really does say only
// "Shapeshifter", and the context has to notice the keyword.
func TestChangelingSatisfiesASubtypeRestriction(t *testing.T) {
	changeling := Card{
		Name: "Universal Automaton", TypeLine: "Artifact Creature — Shapeshifter",
		Keywords: []string{KeywordChangeling},
	}
	ctx := ManaSpendForCast(changeling)
	cavernElf := []string{ManaRestrictCast, ManaRestrictType("Creature"), ManaRestrictSubtype("Elf")}
	if !ctx.allows(cavernElf) {
		t.Error("Cavern of Souls named Elf refused a changeling creature spell")
	}
	// Still not every NON-creature subtype: changeling grants creature
	// types, not Equipment or Aura.
	if ctx.allows([]string{ManaRestrictSubtype("Equipment")}) {
		t.Error("changeling satisfied a non-creature subtype restriction")
	}
	// And an ordinary Shapeshifter with no changeling does not.
	plain := ManaSpendForCast(Card{Name: "Plain", TypeLine: "Creature — Shapeshifter"})
	if plain.allows(cavernElf) {
		t.Error("a non-changeling Shapeshifter was paid for by Elf-named Cavern mana")
	}
}

// TestRestrictionsFuncWinsOverDeclaredRestrictions pins the precedence
// Cavern of Souls depends on, and the empty-tribe case: before the
// type is named the ability must produce UNSPENDABLE mana, not
// unrestricted mana.
func TestRestrictionsFuncWinsOverDeclaredRestrictions(t *testing.T) {
	g := NewGame()
	ab := ManaAbilityShape{
		Restrictions: []string{ManaRestrictSupertype("Legendary")},
		RestrictionsFunc: func(*Game, uuid.UUID, uuid.UUID) []string {
			return []string{ManaRestrictSubtype("Elf")}
		},
	}
	got := restrictionsFor(g, &ab, uuid.Nil, uuid.Nil)
	if len(got) != 1 || got[0] != ManaRestrictSubtype("Elf") {
		t.Fatalf("restrictionsFor = %v, want the computed list", got)
	}
	// The declared list is used when there is no func.
	plain := ManaAbilityShape{Restrictions: []string{ManaRestrictCast}}
	if got := restrictionsFor(g, &plain, uuid.Nil, uuid.Nil); len(got) != 1 || got[0] != ManaRestrictCast {
		t.Errorf("restrictionsFor with no func = %v, want the declared list", got)
	}
	// An unnamed tribe yields "subtype:" with an empty value, which the
	// matcher refuses — unspendable, not free.
	unnamed := ManaSpendForCast(Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid"})
	if unnamed.allows([]string{ManaRestrictSubtype("")}) {
		t.Error("an empty subtype tag was treated as no restriction")
	}
}
