package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// samut_tyrant_of_naktamun_test.go — #3502, CR 702.61 and CR 611.3a.
// "Instant and sorcery spells you control have split second", read live
// off the battlefield rather than stamped when a spell is cast.

const samutOracle = "a275f7ac-7dae-4422-875a-f34f24b6f28a"

func pushSamut(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Samut, Tyrant of Naktamun",
		TypeLine: "Legendary Creature — Human Wizard", OracleID: samutOracle,
		Owner: controller, Controller: controller,
	})
}

// hasSplitSecond is what the table reads for one spell: the item's own
// fact or a battlefield grant, the same answer the stack chip shows.
func hasSplitSecond(g *game.Game, spell uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.SpellHasSplitSecondForEffect(spell) })
	return out
}

func castBoltAt(t *testing.T, g *game.Game, caster *game.Player, victim uuid.UUID) error {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: caster.ID, Controller: caster.ID})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}},
	})
}

// The grant applies: a Bolt its controller casts has split second with
// no flag and no printed keyword on it, the table is shut to responses
// (but not to mana abilities, CR 702.61b), and the gate opens when it
// resolves.
func TestSamutGivesYourInstantsSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pushSamut(g, me.ID)
	ring := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: solRingOracle, Owner: opp.ID, Controller: opp.ID,
	})
	toMain(t, g)

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if !g.SplitSecondActive || !hasSplitSecond(g, bolt) {
		t.Fatalf("a Samut controller's Bolt has no split second (table=%v)", g.SplitSecondActive)
	}
	if g.StackMeta[bolt].SplitSecond {
		t.Error("the grant was stamped on the stack item: it must be read live, not locked in at cast")
	}

	err := castBoltAt(t, g, opp, me.ID)
	if !errors.Is(err, game.ErrSplitSecondActive) {
		t.Errorf("an answer to a Samut-granted split second spell: got %v, want ErrSplitSecondActive", err)
	}
	if err := g.ActivateManaAbility(opp.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("CR 702.61b: a mana ability was refused under granted split second: %v", err)
	}

	passPriorityAroundTable(t, g)
	if g.SplitSecondActive {
		t.Error("split second outlived the spell")
	}
	if err := castBoltAt(t, g, opp, me.ID); err != nil {
		t.Errorf("a cast once the spell had resolved: %v", err)
	}
}

// Sorceries get it too, and a spell outside the filter does not: Samut
// says instant and sorcery, not "spells".
func TestSamutReachesSorceriesAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushSamut(g, me.ID)
	toMain(t, g)

	for _, tc := range []struct {
		name, typeLine string
		want           bool
	}{
		{"Test Sorcery", "Sorcery", true},
		{"Test Creature", "Creature — Bear", false},
		{"Test Enchantment", "Enchantment", false},
	} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine,
			Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", tc.name, err)
		}
		if got := hasSplitSecond(g, id); got != tc.want {
			t.Errorf("%s: split second = %v, want %v", tc.name, got, tc.want)
		}
		if g.SplitSecondActive != tc.want {
			t.Errorf("%s: the table's gate = %v, want %v", tc.name, g.SplitSecondActive, tc.want)
		}
		passPriorityAroundTable(t, g)
	}
}

// An opponent's spells are unaffected: the grant is "spells YOU control".
func TestSamutDoesNotGiveAnOpponentsSpellsSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pushSamut(g, me.ID)
	toMain(t, g)

	if err := castBoltAt(t, g, opp, me.ID); err != nil {
		t.Fatalf("the opponent's Bolt: %v", err)
	}
	var theirs uuid.UUID
	for id, it := range g.StackMeta {
		if it.Controller == opp.ID {
			theirs = id
		}
	}
	if theirs == uuid.Nil {
		t.Fatal("the opponent's Bolt is not on the stack")
	}
	if g.SplitSecondActive || hasSplitSecond(g, theirs) {
		t.Fatal("Samut gave an opponent's Bolt split second")
	}
	// And Samut's controller can answer it.
	if err := castBoltAt(t, g, me, opp.ID); err != nil {
		t.Errorf("answering the opponent's Bolt with Samut out: %v", err)
	}
}

// CR 611.3a: the grant is a static ability's continuous effect, not
// locked in. A Samut that leaves while the spell is on the stack takes
// split second with her (a mana-ability sacrifice, which split second
// does not stop, is the way that happens), and the table can answer.
func TestSamutLeavingTakesSplitSecondOffTheSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	samut := pushSamut(g, me.ID)
	toMain(t, g)

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if !g.SplitSecondActive {
		t.Fatal("setup: the Bolt has no split second")
	}

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(samut); err != nil {
			t.Fatalf("sacrifice Samut: %v", err)
		}
	})
	if g.SplitSecondActive || hasSplitSecond(g, bolt) {
		t.Fatal("split second stayed on the spell after Samut left")
	}
	if err := castBoltAt(t, g, opp, me.ID); err != nil {
		t.Errorf("an answer after Samut left: %v", err)
	}
}

// And the other direction: a Samut that arrives under a spell already
// on the stack gives it split second from that moment.
func TestSamutArrivingGivesASpellOnTheStackSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	toMain(t, g)

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if g.SplitSecondActive {
		t.Fatal("setup: a plain Bolt has split second")
	}
	pushSamut(g, me.ID)
	if !g.SplitSecondActive || !hasSplitSecond(g, bolt) {
		t.Fatal("a Samut that arrived did not give the spell on the stack split second")
	}
}

// A spell that PRINTS split second keeps it whatever Samut does, and a
// card with no grant anywhere is unchanged (the opt-in invariant).
func TestPrintedSplitSecondIsIndependentOfSamut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	samut := pushSamut(g, me.ID)
	ring := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: solRingOracle, Owner: opp.ID, Controller: opp.ID,
	})
	toMain(t, g)
	grip := castCatalogSpell(t, g, "Krosan Grip", "Instant", krosanGripOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ring}})
	requireSplitSecondOn(t, g, grip)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(samut); err != nil {
			t.Fatalf("sacrifice Samut: %v", err)
		}
	})
	requireSplitSecondOn(t, g, grip)
}
