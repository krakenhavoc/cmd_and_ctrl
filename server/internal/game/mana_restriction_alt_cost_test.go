package game

import (
	"errors"
	"testing"
)

// #5680: "Spend this mana only to pay a disturb cost or cast an instant
// or sorcery spell" (Unblinking Observer). ManaRestrictAltCost names the
// alternative cost a spell is cast for, which no other tag can: a
// disturb cast and a flashback cast are both casts from a graveyard.

func observerTags() []string {
	return []string{ManaRestrictAnyOf(
		[]string{ManaRestrictAltCost("disturb")},
		[]string{ManaRestrictCast, ManaRestrictAnyType("Instant", "Sorcery")},
	)}
}

// observerPool is three of the Observer's {U}.
func observerPool() ManaPool {
	return ManaPool{
		{Color: "U", Restrictions: observerTags()},
		{Color: "U", Restrictions: observerTags()},
		{Color: "U", Restrictions: observerTags()},
	}
}

func withAltCost(ctx ManaSpendContext, key string) ManaSpendContext {
	ctx.AltCost = key
	return ctx
}

func TestAltCostRestriction(t *testing.T) {
	disturbOnly := []string{ManaRestrictAltCost("disturb")}
	creature := withCastFrom(creatureSpendContext(), ZoneGraveyard)

	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"a cast for the disturb cost", withAltCost(creature, "disturb"), true},
		{"a graveyard cast for the mana cost", creature, false},
		{"a cast for flashback", withAltCost(creature, "flashback"), false},
		{"an instant cast for escape", withAltCost(instantSpendContext(), "escape"), false},
		{"an activation that claims the key", ManaSpendContext{Purpose: SpendPurposeActivate, AltCost: "disturb"}, false},
		{"an unknown purpose that claims the key", ManaSpendContext{AltCost: "disturb"}, false},
	} {
		if got := tc.ctx.allows(disturbOnly); got != tc.want {
			t.Errorf("%s: allows = %v, want %v", tc.name, got, tc.want)
		}
	}
	if (ManaSpendContext{Purpose: SpendPurposeCast}).allows([]string{ManaRestrictAltCost("")}) {
		t.Error("an empty key admitted a cast at the mana cost")
	}

	// The Observer's whole clause: the OR with "an instant or sorcery
	// spell".
	both := observerTags()
	if !withAltCost(creature, "disturb").allows(both) {
		t.Error("the Observer's mana refused a disturb cast")
	}
	if !instantSpendContext().allows(both) {
		t.Error("the Observer's mana refused an instant")
	}
	if creature.allows(both) || withAltCost(creature, "flashback").allows(both) {
		t.Error("the Observer's mana paid for a creature spell that is not disturbed")
	}
}

// The engine path: a cast that claims disturb spends the mana, through
// ManaSpendForCastParams.
func TestAltCostManaPaysADisturbCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedDisturb(t, g, me)

	me.ManaPool = observerPool()
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Strict: true}); err != nil {
		t.Fatalf("disturb cast with disturb-only mana: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the disturb cast left %v floating, want {1}{U}{U} spent", me.ManaPool)
	}
	if spell := stackCardForTest(t, g, id); spell.ActiveFace != 1 {
		t.Errorf("the spell is face %d, want the back face", spell.ActiveFace)
	}
}

// The same card cast for its mana cost is a creature spell and no
// disturb cost is paid, so the mana stays in the pool.
func TestAltCostManaRefusesTheManaCostOfADisturbCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(disturbFixtureOracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(disturbFixtureOracle, disturbOffer("{1}{U}{U}")))
	advanceTo(t, g, StepPrecombatMain)
	c := disturbFixture(me.ID)
	me.Hand.PushTop(c)

	me.ManaPool = observerPool()
	err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("mana-cost cast with disturb-only mana: err = %v, want insufficient mana", err)
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("the refused cast spent mana: %v", me.ManaPool)
	}
}

// Flashback is a graveyard cast for an alternative cost too, and not a
// disturb cost. The fixture is a sorcery, so the bare disturb tag is
// what is tested here, not the Observer's whole clause.
func TestAltCostManaRefusesAFlashbackCast(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedFlashbackCard(t, g, me, oracle)

	me.ManaPool = ManaPool{
		{Color: "R", Restrictions: []string{ManaRestrictAltCost("disturb")}},
		{Color: "R", Restrictions: []string{ManaRestrictAltCost("disturb")}},
		{Color: "R", Restrictions: []string{ManaRestrictAltCost("disturb")}},
	}
	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("flashback cast with disturb-only mana: err = %v, want insufficient mana", err)
	}
	// The matching key is admitted.
	for i := range me.ManaPool {
		me.ManaPool[i].Restrictions = []string{ManaRestrictAltCost("flashback")}
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback cast with flashback-only mana: %v", err)
	}
}

// A spend context is stored on an assist prompt (AssistPrompt.Spend), so
// the claimed cost must survive a restore: the helper's answer is
// checked under it.
func TestAltCostSurvivesARestoreOnAStoredSpendContext(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { assistPrompt(t, g).Assist.Spend.AltCost = "disturb" })

	_, restored := roundTrip(t, g)
	spend := assistPrompt(t, restored).Assist.Spend
	if spend.AltCost != "disturb" {
		t.Fatalf("restored spend context claims %q, want disturb", spend.AltCost)
	}
	if !spend.allows([]string{ManaRestrictAltCost("disturb")}) {
		t.Error("the restored context refused disturb-only mana")
	}
}
