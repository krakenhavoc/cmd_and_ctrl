package game

import (
	"testing"

	"github.com/google/uuid"
)

// bands_with_other_test.go — #6142: "bands with other [quality]" (CR
// 702.22b-c, j-k). The CR text is quoted in bands_with_other.go.

// pushTypedCombatant is pushCombatant with a type line of its own.
func pushTypedCombatant(t *testing.T, g *Game, owner *Player, name, typeLine string, power, toughness int, keywords ...string) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.Power = power
	c.Toughness = toughness
	c.Keywords = append([]string(nil), keywords...)
	g.Battlefield.PushTop(c)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == c.InstanceID {
			g.Battlefield.Cards[i].effective = printedEffectiveWith(c)
		}
	}
	return c.InstanceID
}

const legendLine = "Legendary Creature — Human Knight"

// CR 702.22c: two legendary creatures, one with "bands with other
// legendary creatures" (Cathedral of Serra's grant), attack as a band; a
// non-legendary creature can't join it.
func TestBandsWithOtherLegendaryFormsABandOfLegends(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	anchor := pushTypedCombatant(t, g, me, "Sir Anchor", legendLine, 2, 2, BandsWithOtherLegendary)
	legend := pushTypedCombatant(t, g, me, "Lady Legend", legendLine, 2, 2)
	other := pushTypedCombatant(t, g, me, "Lady Other", legendLine, 1, 1)
	if err := declareBand(t, g, anchor, legend, other); err != nil {
		t.Fatalf("a band of legends anchored by bands-with-other was refused: %v", err)
	}
	var mates []uuid.UUID
	g.ReadSnapshot(func() { mates = g.BandMatesForEffect(anchor) })
	if len(mates) != 2 {
		t.Fatalf("band mates %v, want the two other legends", mates)
	}
}

func TestBandsWithOtherLegendaryRefusesANonLegend(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	anchor := pushTypedCombatant(t, g, me, "Sir Anchor", legendLine, 2, 2, BandsWithOtherLegendary)
	legend := pushTypedCombatant(t, g, me, "Lady Legend", legendLine, 2, 2)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	bandRefusal(t, declareBand(t, g, anchor, legend, bear), BandRefusalTwoWithout)
}

// The anchor must itself be a [quality] creature: a non-legendary
// creature with "bands with other legendary creatures" anchors nothing.
func TestBandsWithOtherNeedsAQualityAnchor(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	plain := pushCombatant(t, g, me, "Squire", 1, 1, BandsWithOtherLegendary)
	l1 := pushTypedCombatant(t, g, me, "Lady Legend", legendLine, 2, 2)
	l2 := pushTypedCombatant(t, g, me, "Sir Legend", legendLine, 2, 2)
	bandRefusal(t, declareBand(t, g, plain, l1, l2), BandRefusalTwoWithout)
}

// Wolves of the Hunt band with each other only: tokens named Wolves of
// the Hunt with "bands with other creatures named Wolves of the Hunt"
// band; a differently named wolf can't join.
func TestBandsWithOtherNamedBandsOnlyThatName(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	kw := BandsWithOtherNamed("Wolves of the Hunt")
	w1 := pushTypedCombatant(t, g, me, "Wolves of the Hunt", "Token Creature — Wolf", 1, 1, kw)
	w2 := pushTypedCombatant(t, g, me, "Wolves of the Hunt", "Token Creature — Wolf", 1, 1, kw)
	w3 := pushTypedCombatant(t, g, me, "Wolves of the Hunt", "Token Creature — Wolf", 1, 1, kw)
	if err := declareBand(t, g, w1, w2, w3); err != nil {
		t.Fatalf("three Wolves of the Hunt were refused as a band: %v", err)
	}

	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	a := pushTypedCombatant(t, g2, me2, "Wolves of the Hunt", "Token Creature — Wolf", 1, 1, kw)
	b := pushTypedCombatant(t, g2, me2, "Wolves of the Hunt", "Token Creature — Wolf", 1, 1, kw)
	stranger := pushTypedCombatant(t, g2, me2, "Timber Wolves", "Creature — Wolf", 1, 1)
	other := pushTypedCombatant(t, g2, me2, "Gray Wolf", "Creature — Wolf", 1, 1)
	bandRefusal(t, declareBand(t, g2, a, b, stranger, other), BandRefusalTwoWithout)
}

// CR 702.22j: an attacker blocked by two legendary creatures, one with
// "bands with other legendary creatures", has its damage divided by the
// defending player.
func TestBandsWithOtherBlockersHandTheDivisionToTheDefender(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := pushCombatant(t, g, me, "Trampling Beast", 4, 4, "trample")
	anchor := pushTypedCombatant(t, g, opp, "Sir Anchor", legendLine, 1, 1, BandsWithOtherLegendary)
	legend := pushTypedCombatant(t, g, opp, "Lady Legend", legendLine, 2, 2)
	declareAttacks(t, g, beast)
	if err := block(t, g, anchor, beast, legend, beast); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil {
		t.Fatal("no prompt for the beast's damage")
	}
	if choice.Chooser != opp.ID || !choice.DamageAssignment.Banding || choice.DamageAssignment.AllowTrample {
		t.Fatalf("chooser %s banding %v; want the defending player dividing", choice.Chooser, choice.DamageAssignment.Banding)
	}
}

// Without a second [quality] creature among the blockers nothing changes
// hands: the attacker's controller divides as usual (CR 510.1c).
func TestBandsWithOtherNeedsAnotherQualityBlocker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := pushCombatant(t, g, me, "Beast", 4, 4)
	anchor := pushTypedCombatant(t, g, opp, "Sir Anchor", legendLine, 1, 1, BandsWithOtherLegendary)
	bear := pushCombatant(t, g, opp, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, beast)
	if err := block(t, g, anchor, beast, bear, beast); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if c := damageChoice(g); c != nil && (c.Chooser != me.ID || c.DamageAssignment.Banding) {
		t.Fatalf("chooser %s banding %v; want the attacker's controller", c.Chooser, c.DamageAssignment.Banding)
	}
}

// CR 702.22b: losing banding (Tolaria) removes every "bands with other"
// ability too; "loses all bands with other abilities" (Shelkin Brownie)
// is the prefix removal; both until end of turn.
func TestLosingBandingLosesBandsWithOther(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	anchor := pushTypedCombatant(t, g, me, "Sir Anchor", legendLine, 2, 2, KeywordBanding, BandsWithOtherLegendary, "flying")
	registerScopedEffectForTest(t, g, anchor, []Mod{RemoveKeywordsMod(KeywordBanding)}, g.UntilEndOfTurnDuration())
	var c *Card
	g.ReadSnapshot(func() { c = findBattlefieldCard(g, anchor) })
	if HasKeyword(c, KeywordBanding) || HasKeyword(c, BandsWithOtherLegendary) {
		t.Fatal("a creature that lost banding kept banding or bands with other")
	}
	if !HasKeyword(c, "flying") {
		t.Fatal("losing banding took flying too")
	}
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.ClearEndOfTurnScopedStaticsLocked()
	})
	g.ReadSnapshot(func() { c = findBattlefieldCard(g, anchor) })
	if !HasKeyword(c, KeywordBanding) || !HasKeyword(c, BandsWithOtherLegendary) {
		t.Fatal("the abilities did not come back after end of turn")
	}

	registerScopedEffectForTest(t, g, anchor, []Mod{RemoveKeywordPrefixesMod(KeywordBandsWithOtherPrefix)}, g.UntilEndOfTurnDuration())
	g.ReadSnapshot(func() { c = findBattlefieldCard(g, anchor) })
	if HasKeyword(c, BandsWithOtherLegendary) || !HasKeyword(c, KeywordBanding) {
		t.Fatal("losing bands with other took banding, or left bands with other")
	}
}

func TestParseBandsWithOther(t *testing.T) {
	if !IsBandsWithOther(BandsWithOtherLegendary) || !IsBandsWithOther(BandsWithOtherNamed("Wolves of the Hunt")) {
		t.Error("the two qualities are not read")
	}
	for _, tok := range []string{"bands with other", "bands with other dinosaurs", "bands with other creatures named ", "banding"} {
		if IsBandsWithOther(tok) {
			t.Errorf("%q read as bands with other", tok)
		}
	}
}
