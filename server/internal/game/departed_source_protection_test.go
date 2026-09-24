package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// departed_source_protection_test.go covers #1417 (ADR 0056 amendment
// 2026-09-24 "Still not covered", closed by its 2026-09-25 follow-up;
// ADR 0072 §3 amendment): the characteristics CR 702.16e protection
// reads off a damage source that has LEFT the battlefield are the ones
// it had as it last existed there (CR 608.2h), taken from the same
// #1379 per-object record #1396 reads lifelink and deathtouch from —
// not the card's characteristics in the zone it went to.
//
// Every scenario here makes the two answers DIFFER, with a layer 5
// colour change that applies on the battlefield and nowhere else: a
// printed-colourless creature painted red, and a printed-red creature
// bleached colourless. Without the change the graveyard card's printed
// colour decides, so the painted one's damage gets through protection
// from red and the bleached one's is prevented.

const (
	lkiPainterOracle  = "lki-1417-painter"  // "Painted creatures are red."
	lkiBleacherOracle = "lki-1417-bleacher" // "Bleached creatures are colorless."
)

// withLKIColourPainters installs the two layer 5 sources. Each applies
// only to the instance IDs in its set, and only while its source
// permanent is on the battlefield — the ordinary shape of a static.
func withLKIColourPainters(t *testing.T, painted, bleached map[uuid.UUID]bool) {
	t.Helper()
	withStaticAbilities(t, func(key string) []StaticAbility {
		switch key {
		case lkiPainterOracle:
			return []StaticAbility{{
				Layer:     Layer5Color,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return painted[target.InstanceID] },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Colors = []string{"R"} },
			}}
		case lkiBleacherOracle:
			return []StaticAbility{{
				Layer:     Layer5Color,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return bleached[target.InstanceID] },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Colors = nil },
			}}
		}
		return nil
	})
}

// pushLKISource puts a 2/2 on the battlefield with the given printed
// colours and brings the layer cache up to date, so the departure
// record (which reads Effective() without recomputing) sees the
// painted colour.
func pushLKISource(g *Game, owner uuid.UUID, colors []string) uuid.UUID {
	id := pushTypedTestCard(g, Card{
		Name: "LKI Source", TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		Colors: colors, Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return id
}

// pushLKIPainter puts one of the two layer 5 sources on the battlefield.
func pushLKIPainter(g *Game, owner uuid.UUID, oracle string) uuid.UUID {
	id := pushTypedTestCard(g, Card{
		Name: oracle, TypeLine: "Enchantment", OracleID: oracle, Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return id
}

// liveColours reads a card's colours in whatever zone holds it, the way
// the pre-#1417 damage tail did.
func liveColours(t *testing.T, g *Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if ch := g.damageSourceLKILocked(id); ch != nil {
			out = ch.Colors
		}
	})
	return out
}

// A printed-colourless creature painted red dies. Its damage — through
// the instance-ID entry point (a "when this dies" trigger) and through
// the object-named one — is RED damage, and protection from red
// prevents it, even though the card in the graveyard is colourless. A
// creature with no protection takes it, so the damage itself is real.
func TestDepartedRedSourceIsPreventedByProtectionFromRed(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	painted := map[uuid.UUID]bool{}
	withLKIColourPainters(t, painted, nil)

	proRedByID := pushColouredCreature(g, opp, "Pro-Red by ID", []string{"W"}, "protection from red")
	proRedByRef := pushColouredCreature(g, opp, "Pro-Red by ref", []string{"W"}, "protection from red")
	plain := pushColouredCreature(g, opp, "Plain", []string{"W"})
	src := pushLKISource(g, me.ID, nil)
	painted[src] = true
	pushLKIPainter(g, me.ID, lkiPainterOracle)
	if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
		t.Fatalf("setup: the painted source is %v on the battlefield, want red", got)
	}
	ref := refOf(t, g, src)
	destroy(t, g, src)
	if got := liveColours(t, g, src); len(got) != 0 {
		t.Fatalf("setup: the card in the graveyard is %v; it must be colourless "+
			"there, or this test proves nothing", got)
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, proRedByID, 2); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
		if err := g.DealDamageFromObjectForEffect(ref, proRedByRef, 2); err != nil {
			t.Fatalf("DealDamageFromObjectForEffect: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(src, plain, 2); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	if got := damageOn(g, proRedByID); got != 0 {
		t.Errorf("pro-red creature took %d damage from a departed source that was "+
			"red as it last existed (by instance ID, CR 608.2h / 702.16e, #1417)", got)
	}
	if got := damageOn(g, proRedByRef); got != 0 {
		t.Errorf("pro-red creature took %d damage from a departed source that was "+
			"red as it last existed (by object, #1417)", got)
	}
	if got := damageOn(g, plain); got != 2 {
		t.Errorf("unprotected creature took %d damage, want 2", got)
	}
}

// The reverse: a printed-red creature made colourless dies. It last
// existed as colourless, so protection from red does NOT prevent its
// damage, even though the card in the graveyard is red again.
func TestDepartedColourlessSourceIsNotPreventedByProtectionFromRed(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bleached := map[uuid.UUID]bool{}
	withLKIColourPainters(t, nil, bleached)

	proRed := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	src := pushLKISource(g, me.ID, []string{"R"})
	bleached[src] = true
	pushLKIPainter(g, me.ID, lkiBleacherOracle)
	if got := liveColours(t, g, src); len(got) != 0 {
		t.Fatalf("setup: the bleached source is %v on the battlefield, want colourless", got)
	}
	destroy(t, g, src)
	if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
		t.Fatalf("setup: the card in the graveyard is %v, want red", got)
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, proRed, 1); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	if got := damageOn(g, proRed); got != 1 {
		t.Errorf("pro-red creature took %d damage, want 1: the source was colourless "+
			"as it last existed (CR 608.2h, #1417)", got)
	}
}

// A live source is read live, exactly as before: the painted creature
// on the battlefield deals red damage and the bleached one colourless
// damage, by instance ID and by a ref naming the live object.
func TestLiveSourceProtectionColourStillReadsLive(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	painted, bleached := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	withLKIColourPainters(t, painted, bleached)

	proRed := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	red := pushLKISource(g, me.ID, nil)
	painted[red] = true
	colourless := pushLKISource(g, me.ID, []string{"R"})
	bleached[colourless] = true
	pushLKIPainter(g, me.ID, lkiPainterOracle)
	pushLKIPainter(g, me.ID, lkiBleacherOracle)
	redRef, colourlessRef := refOf(t, g, red), refOf(t, g, colourless)

	g.WithWriteLock(func() {
		for _, err := range []error{
			g.DealDamageToCreatureForEffect(red, proRed, 4),
			g.DealDamageFromObjectForEffect(redRef, proRed, 4),
			g.DealDamageToCreatureForEffect(colourless, proRed, 1),
			g.DealDamageFromObjectForEffect(colourlessRef, proRed, 1),
		} {
			if err != nil {
				t.Fatalf("deal: %v", err)
			}
		}
	})
	if got := damageOn(g, proRed); got != 2 {
		t.Errorf("pro-red creature took %d damage, want 2: the live red source's 8 "+
			"prevented, the live colourless source's 2 dealt", got)
	}
}

// CR 400.7: the card died colourless, came back, and the NEW object is
// painted red. Damage named by the OLD object is colourless damage and
// gets through; damage named by the bare instance ID is the live new
// object's, red, and is prevented.
func TestReturnedObjectsColourIsNotTheDepartedObjects(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	painted := map[uuid.UUID]bool{}
	withLKIColourPainters(t, painted, nil)

	proRedOld := pushColouredCreature(g, opp, "Pro-Red (old object)", []string{"W"}, "protection from red")
	proRedNew := pushColouredCreature(g, opp, "Pro-Red (new object)", []string{"W"}, "protection from red")
	src := pushLKISource(g, me.ID, nil)
	first := refOf(t, g, src)
	destroy(t, g, src)
	returnUnderOwner(t, g, src, me.ID)
	painted[src] = true
	pushLKIPainter(g, me.ID, lkiPainterOracle)
	if second := refOf(t, g, src); second == first {
		t.Fatal("setup: the returned card should be a new object")
	}
	if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
		t.Fatalf("setup: the new object is %v, want red", got)
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageFromObjectForEffect(first, proRedOld, 2); err != nil {
			t.Fatalf("DealDamageFromObjectForEffect: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(src, proRedNew, 2); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	if got := damageOn(g, proRedOld); got != 2 {
		t.Errorf("pro-red creature took %d damage from the departed COLOURLESS "+
			"object, want 2: the returned card's red is a new object's (CR 400.7)", got)
	}
	if got := damageOn(g, proRedNew); got != 0 {
		t.Errorf("pro-red creature took %d damage from the live red object, want 0", got)
	}
}

// #662 stands: a card that left and has moved on is not the departed
// permanent. Painted red, bounced, then put on the stack, it is a
// colourless SPELL now, and its damage is the spell's — no record
// answers for a card two zone changes past it.
func TestMovedOnSourceKeepsItsCurrentZoneColour(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	painted := map[uuid.UUID]bool{}
	withLKIColourPainters(t, painted, nil)

	proRed := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	src := pushLKISource(g, me.ID, nil)
	painted[src] = true
	pushLKIPainter(g, me.ID, lkiPainterOracle)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(src); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, err := MoveCard(me.Hand, g.Stack, src); err != nil {
			t.Fatalf("MoveCard to the stack: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(src, proRed, 2); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	if got := damageOn(g, proRed); got != 2 {
		t.Errorf("pro-red creature took %d damage from a colourless spell, want 2: "+
			"the spell is not the red permanent the card used to be", got)
	}
}
