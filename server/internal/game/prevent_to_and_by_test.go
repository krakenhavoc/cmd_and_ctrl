package game

import (
	"testing"

	"github.com/google/uuid"
)

// prevent_to_and_by_test.go — ADR 0108 Delivery PR 7 (#1904): "prevent
// all [combat] damage that would be dealt to and dealt by <it>" as ONE
// preventFromSource record (Mod.AndDealtBy), and a shield protecting
// several permanents in one record (DamageShield.ProtectPermanents).

func toAndByShield(t *testing.T, g *Game, controller uuid.UUID, combat bool, then BodyRef, ids ...uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		s := DamageShield{Controller: controller, ProtectPermanents: ids, AndDealtBy: true, CombatOnly: combat, Then: then, Label: "Maze"}
		if !g.PreventDamageFromSourceThisTurnForEffect(s) {
			t.Fatal("no shield registered")
		}
	})
}

func markedOn(g *Game, id uuid.UUID) int {
	if c := findBattlefieldCard(g, id); c != nil {
		return c.DamageMarked
	}
	return -1
}

// Maze of Ith: the attacker's combat damage and the blocker's combat
// damage to it are both prevented, by one record, and a CR 615.5
// follow-up runs once for the step's one instance (CR 615.13) with the
// total of both directions.
func TestToAndByShieldIsOneEffect(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushColouredCreature(g, opp, "Attacker", []string{"R"})
	blocker := pushColouredCreature(g, me, "Blocker", []string{"G"})
	var calls []followUpCall
	toAndByShield(t, g, me.ID, true, testFollowUp(&calls), attacker)
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("%d records, want one", n)
	}
	g.WithWriteLock(func() {
		g.markCombatDamageOnCardLocked(blocker, 2, attacker, "")
		g.markCombatDamageOnCardLocked(attacker, 1, blocker, "")
		g.flushPreventionFollowUpsLocked()
	})
	if markedOn(g, blocker) != 0 || markedOn(g, attacker) != 0 {
		t.Fatalf("blocker %d, attacker %d damage: want both prevented", markedOn(g, blocker), markedOn(g, attacker))
	}
	if len(calls) != 1 || calls[0].amount != 3 {
		t.Fatalf("follow-ups %+v, want one with 3: one effect applied to one instance", calls)
	}
}

// The combat-only shield leaves non-combat damage to and from the creature
// alone, and other creatures' combat damage too.
func TestToAndByShieldCombatOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushColouredCreature(g, opp, "Attacker", []string{"R"})
	other := pushColouredCreature(g, opp, "Other", []string{"R"})
	blocker := pushColouredCreature(g, me, "Blocker", []string{"G"})
	toAndByShield(t, g, me.ID, true, BodyRef{}, attacker)
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(blocker, attacker, 1)
		_ = g.DealDamageToPlayerForEffect(attacker, me.ID, 1)
		g.markCombatDamageToPlayerLocked(me.ID, other, 2, "")
		g.markCombatDamageToPlayerLocked(me.ID, attacker, 4, "")
	})
	if got := markedOn(g, attacker); got != 1 {
		t.Errorf("attacker has %d damage, want the non-combat 1", got)
	}
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Errorf("life %d, want %d: the ping and the other creature's combat damage dealt", got, start-3)
	}
}

// "Dealt to and dealt by": all damage, either way, for every pinned
// object (Energy Arc's "those creatures"), and nothing between two
// creatures neither of which is pinned.
func TestToAndByShieldOverSeveralCreatures(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	b := pushColouredCreature(g, me, "B", []string{"G"})
	x := pushColouredCreature(g, opp, "X", []string{"R"})
	y := pushColouredCreature(g, me, "Y", []string{"G"})
	toAndByShield(t, g, me.ID, false, BodyRef{}, a, b)
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("%d records, want one", n)
	}
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(a, b, 2)
		_ = g.DealDamageToCreatureForEffect(x, a, 2)
		_ = g.DealDamageToPlayerForEffect(a, me.ID, 2)
		_ = g.DealDamageToCreatureForEffect(x, y, 1)
	})
	if markedOn(g, a) != 0 || markedOn(g, b) != 0 || lifeOf(g, me.ID) != start {
		t.Fatalf("a %d, b %d, life %d (want 0, 0, %d)", markedOn(g, a), markedOn(g, b), lifeOf(g, me.ID), start)
	}
	if got := markedOn(g, y); got != 1 {
		t.Errorf("y has %d damage, want 1: neither end is pinned", got)
	}
}

// A pinned object that leaves and returns is a new object (CR 400.7): the
// shield does not follow it.
func TestToAndByShieldDoesNotFollowANewObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	b := pushColouredCreature(g, opp, "B", []string{"R"})
	toAndByShield(t, g, me.ID, false, BodyRef{}, a, b)
	g.WithWriteLock(func() {
		findBattlefieldCard(g, a).EnteredBattlefieldAt += 1000
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(a, me.ID, 2) })
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d: A is a new object", got, start-2)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(b, me.ID, 2) })
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d: B is still the pinned object", got, start-2)
	}
}

// Registration refuses a to-and-by shield with nothing pinned, or with a
// source, a player or a charge; the mod check refuses the field anywhere
// else.
func TestToAndByShieldRefusesItsOtherFields(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(a)
		for _, s := range []DamageShield{
			{Controller: me.ID, AndDealtBy: true},
			{Controller: me.ID, AndDealtBy: true, ProtectPermanent: a, Source: ref, SourceZone: zone},
			{Controller: me.ID, AndDealtBy: true, ProtectPermanent: a, ProtectPlayer: me.ID},
			{Controller: me.ID, AndDealtBy: true, ProtectPermanent: a, Queries: []PermanentQuery{{Types: []string{"creature"}}}},
		} {
			if g.PreventDamageFromSourceThisTurnForEffect(s) {
				t.Errorf("registered %+v", s)
			}
		}
	})
	for _, m := range []Mod{
		{Kind: ModPreventFromSource, AndDealtBy: true, Objects: []ObjectRef{{ID: uuid.New()}}},
		{Kind: ModPreventFromSource, AndDealtBy: true, Player: me.ID},
		{Kind: ModPreventNextFromSource, AndDealtBy: true, Objects: []ObjectRef{{ID: uuid.New()}}},
		{Kind: ModPreventDamage, Amount: 1, AndDealtBy: true},
	} {
		if nextFromSourceModProblem(m) == "" {
			t.Errorf("mod %+v passed the check", m)
		}
	}
	if p := nextFromSourceModProblem(Mod{Kind: ModPreventFromSource, AndDealtBy: true, CombatOnly: true}); p != "" {
		t.Errorf("a sound to-and-by mod was refused: %s", p)
	}
}

// Redeem: "Prevent all damage that would be dealt this turn to up to two
// target creatures" is one record protecting both, and nothing else.
func TestShieldProtectingSeveralPermanents(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushColouredCreature(g, me, "A", []string{"W"})
	b := pushColouredCreature(g, me, "B", []string{"W"})
	c := pushColouredCreature(g, me, "C", []string{"W"})
	src := pushColouredCreature(g, opp, "Src", []string{"R"})
	g.WithWriteLock(func() {
		if !g.PreventDamageFromSourceThisTurnForEffect(DamageShield{Controller: me.ID, ProtectPermanents: []uuid.UUID{a, b}, Label: "Redeem"}) {
			t.Fatal("no shield")
		}
	})
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("%d records, want one", n)
	}
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(src, a, 2)
		_ = g.DealDamageToCreatureForEffect(src, b, 2)
		_ = g.DealDamageToCreatureForEffect(src, c, 1)
		_ = g.DealDamageToCreatureForEffect(a, src, 1)
	})
	if markedOn(g, a) != 0 || markedOn(g, b) != 0 || markedOn(g, c) != 1 || markedOn(g, src) != 1 {
		t.Fatalf("a %d, b %d, c %d, src %d (want 0, 0, 1, 1)", markedOn(g, a), markedOn(g, b), markedOn(g, c), markedOn(g, src))
	}
}
