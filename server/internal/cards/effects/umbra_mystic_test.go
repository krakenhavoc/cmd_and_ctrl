package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const umbraMysticOracle = "c0a4f47c-c7f9-4b95-aa1b-30580e4cf030"

// mysticBoard puts an Umbra Mystic and a 2/2 under `owner`'s control and
// returns both.
func mysticBoard(g *game.Game, owner uuid.UUID) (mystic, bear uuid.UUID) {
	mystic = b12Push(g, owner, "Umbra Mystic", "Creature — Human Wizard", umbraMysticOracle, 2, 2)
	bear = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return mystic, bear
}

func mysticDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
}

// A plain Aura on my creature has umbra armor, whoever controls the Aura.
func TestUmbraMysticGivesAnAuraUmbraArmor(t *testing.T) {
	for _, auraMine := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		_, bear := mysticBoard(g, me.ID)
		auraOwner := me
		if !auraMine {
			auraOwner = opp
		}
		aura := pushAuraOnLand(t, g, auraOwner.ID, "Plain Aura", "", bear)
		src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
		pr6Damage(t, g, src, bear, 5)
		if !g.Battlefield.Contains(bear) {
			t.Fatalf("aura mine %v: the creature should survive lethal damage", auraMine)
		}
		if got := pr6Marked(g, bear); got != 0 {
			t.Errorf("aura mine %v: damage marked %d, want 0", auraMine, got)
		}
		if !umbraInGraveyard(auraOwner, aura) {
			t.Fatalf("aura mine %v: the Aura should be destroyed instead", auraMine)
		}
		// The Aura is gone: the next destruction is not replaced.
		mysticDestroy(t, g, bear)
		if g.Battlefield.Contains(bear) {
			t.Errorf("aura mine %v: a creature with no Aura left was saved", auraMine)
		}
	}
}

// Only permanents the Mystic's controller controls, and only destruction.
func TestUmbraMysticCoversOnlyYourPermanentsAndOnlyDestruction(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, bear := mysticBoard(g, me.ID)
	theirs := pr7Creature(g, opp.ID, "Their Bear", 2, "G")
	theirAura := pushAuraOnLand(t, g, me.ID, "My Aura On Theirs", "", theirs)
	mysticDestroy(t, g, theirs)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("an opponent's enchanted creature was saved")
	}
	g.RunStateChecksForTest()
	if !umbraInGraveyard(me, theirAura) {
		t.Fatal("the Aura on the destroyed creature should go to the graveyard with nothing to enchant")
	}
	aura := pushAuraOnLand(t, g, me.ID, "Plain Aura", "", bear)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(aura) {
		t.Fatal("a sacrifice is not a destruction: the creature goes, and its Aura after it")
	}
}

// Without the Mystic the same Aura does nothing.
func TestAPlainAuraHasNoUmbraArmorWithoutUmbraMystic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mystic, bear := mysticBoard(g, me.ID)
	pushAuraOnLand(t, g, me.ID, "Plain Aura", "", bear)
	mysticDestroy(t, g, mystic)
	if g.Battlefield.Contains(mystic) {
		t.Fatal("an unenchanted Mystic was saved")
	}
	mysticDestroy(t, g, bear)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Mystic has left: the Aura has no umbra armor")
	}
}

// Two Auras: the creature's controller chooses which is destroyed (CR
// 616.1), and the other stays to save it again.
func TestUmbraMysticTwoAurasAreOrderedByTheController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, bear := mysticBoard(g, me.ID)
	a := pushAuraOnLand(t, g, me.ID, "Plain Aura A", "", bear)
	b := pushAuraOnLand(t, g, opp.ID, "Plain Aura B", "", bear)
	mysticDestroy(t, g, bear)
	order := b06ReplacementOrderFor(g, me.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want a CR 616 ordering prompt for the creature's controller, got %+v", order)
	}
	ids := order.ReplacementEffectIDs
	var label string
	var firstSrc uuid.UUID
	g.ReadSnapshot(func() { label, firstSrc = g.ReplacementOptionMetaForEffect(ids[1]) })
	if label != game.UmbraArmorLabel || (firstSrc != a && firstSrc != b) {
		t.Fatalf("option %q from %s, want umbra armor from one of the Auras", label, firstSrc)
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{ids[1], ids[0]}); err != nil {
		t.Fatal(err)
	}
	kept := a
	if firstSrc == a {
		kept = b
	}
	if !g.Battlefield.Contains(bear) || g.Battlefield.Contains(firstSrc) || !g.Battlefield.Contains(kept) {
		t.Fatalf("bear %v (want true), chosen Aura %v (want false), other Aura %v (want true)",
			g.Battlefield.Contains(bear), g.Battlefield.Contains(firstSrc), g.Battlefield.Contains(kept))
	}
	mysticDestroy(t, g, bear)
	if !g.Battlefield.Contains(bear) || g.Battlefield.Contains(kept) {
		t.Fatal("the second Aura should save the creature from the second destruction")
	}
}

// An Aura that prints umbra armor gains nothing: one Bear Umbra is one
// replacement, with no prompt. Two Mystics give a plain Aura one.
func TestUmbraMysticDoesNotDoubleUmbraArmor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, bear := mysticBoard(g, me.ID)
	b12Push(g, me.ID, "Umbra Mystic", "Creature — Human Wizard", umbraMysticOracle, 2, 2)
	umbra := pushAuraOnLand(t, g, me.ID, "Bear Umbra", bearUmbraOracleID, bear)
	mysticDestroy(t, g, bear)
	if order := b06ReplacementOrderFor(g, me.ID); order != nil {
		t.Fatalf("one Aura with umbra armor asked for an order: %+v", order)
	}
	if !g.Battlefield.Contains(bear) || !umbraInGraveyard(me, umbra) {
		t.Fatal("Bear Umbra should be destroyed in the creature's place")
	}
	plain := pushAuraOnLand(t, g, me.ID, "Plain Aura", "", bear)
	mysticDestroy(t, g, bear)
	if order := b06ReplacementOrderFor(g, me.ID); order != nil {
		t.Fatalf("two Mystics over one Aura asked for an order: %+v", order)
	}
	if !g.Battlefield.Contains(bear) || !umbraInGraveyard(me, plain) {
		t.Fatal("the plain Aura should be destroyed in the creature's place")
	}
}

// A wipe that destroys the Mystic and an enchanted creature at once is
// still replaced for the creature (CR 614.4), whichever leaves first, and
// an enchanted Mystic saves itself.
func TestUmbraMysticStillGrantsInTheWipeThatDestroysIt(t *testing.T) {
	for _, mysticFirst := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		mystic, bear := mysticBoard(g, me.ID)
		aura := pushAuraOnLand(t, g, me.ID, "Plain Aura", "", bear)
		ids := []uuid.UUID{bear, mystic}
		if mysticFirst {
			ids = []uuid.UUID{mystic, bear}
		}
		g.WithWriteLock(func() { g.DestroyPermanentsForEffect(ids) })
		g.RunStateChecksForTest()
		if g.Battlefield.Contains(mystic) || !g.Battlefield.Contains(bear) || !umbraInGraveyard(me, aura) {
			t.Fatalf("mystic first %v: mystic %v (want false), bear %v (want true), aura in graveyard %v (want true)",
				mysticFirst, g.Battlefield.Contains(mystic), g.Battlefield.Contains(bear), umbraInGraveyard(me, aura))
		}
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	mystic, _ := mysticBoard(g, me.ID)
	aura := pushAuraOnLand(t, g, me.ID, "Plain Aura", "", mystic)
	mysticDestroy(t, g, mystic)
	if !g.Battlefield.Contains(mystic) || !umbraInGraveyard(me, aura) {
		t.Fatal("an enchanted Mystic should be saved by its own Aura")
	}
}
