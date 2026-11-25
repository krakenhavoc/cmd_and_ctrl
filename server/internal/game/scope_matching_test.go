package game

import (
	"testing"

	"github.com/google/uuid"
)

// scope_matching_test.go — #6237: a mass "can't block this turn" over a
// FILTERED set (Goblin Locksmith's creatures with defender, Rumbling
// Ruin's "creatures your opponents control with power less than or equal
// to that number"). The filter is data on the record's mods; the SET is
// read live (CR 611.2c), so a creature that arrives this turn is
// covered, and the NUMBER is whatever was stored.

// defenderQuery is "creatures with defender", read through the layered
// view, so a granted defender counts.
func defenderQuery() PermanentQuery {
	return PermanentQuery{Types: []string{"creature"}, Keyword: "defender"}
}

// TestMatchingScopeCoversTheFilteredSetOnly — Goblin Locksmith: every
// creature with defender can't block, whoever controls it, and nothing
// else is touched.
func TestMatchingScopeCoversTheFilteredSetOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushCombatant(t, g, opp, "Their Wall", 0, 4, "defender")
	bear := pushCombatant(t, g, opp, "Their Bear", 2, 2)
	myWall := pushCombatant(t, g, me, "My Wall", 0, 4, "defender")

	ok := false
	g.WithWriteLock(func() {
		ok = g.CantBlockThisTurnMatchingForEffect(uuid.Nil, me.ID,
			[]PermanentQuery{defenderQuery()}, false, "creatures with defender can't block this turn")
	})
	if !ok {
		t.Fatal("CantBlockThisTurnMatchingForEffect registered nothing")
	}
	// A creature that arrives after the ability resolved is covered too
	// (CR 611.2c: "can't block" changes no characteristic).
	late := pushCombatant(t, g, opp, "Their Late Wall", 0, 4, "defender")

	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want bool
	}{
		{wall, "their Wall", true},
		{myWall, "my Wall", true},
		{late, "their late Wall", true},
		{bear, "their Bear", false},
	} {
		if got := scopedEffectChar(t, g, tc.id).Restrictions.Has(CantBlock); got != tc.want {
			t.Errorf("%s: CantBlock = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestMatchingScopeOpponentsOnly — Rumbling Ruin: the opponents' form
// spares the controller's own creatures, and the power bound is the one
// the record stored.
func TestMatchingScopeOpponentsOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := pushCombatant(t, g, opp, "Their 2/2", 2, 2)
	big := pushCombatant(t, g, opp, "Their 4/4", 4, 4)
	mine := pushCombatant(t, g, me, "My 1/1", 1, 1)

	g.WithWriteLock(func() {
		if !g.CantBlockThisTurnMatchingForEffect(uuid.Nil, me.ID,
			[]PermanentQuery{PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(3)},
			true, "creatures your opponents control with power 3 or less can't block this turn") {
			t.Fatal("registered nothing")
		}
	})

	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want bool
	}{
		{small, "their 2/2", true},
		{big, "their 4/4", false},
		{mine, "my 1/1", false},
	} {
		if got := scopedEffectChar(t, g, tc.id).Restrictions.Has(CantBlock); got != tc.want {
			t.Errorf("%s: CantBlock = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestMatchingScopeReadsThePowerLive — the SET is live even though the
// number is locked: a creature pumped above the stored bound may block,
// and one shrunk under it may not (CR 509.1b, judged as blockers are
// declared).
func TestMatchingScopeReadsThePowerLive(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := pushCombatant(t, g, opp, "Their 2/2", 2, 2)
	g.WithWriteLock(func() {
		g.CantBlockThisTurnMatchingForEffect(uuid.Nil, me.ID,
			[]PermanentQuery{PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2)},
			true, "power 2 or less can't block this turn")
	})
	if !scopedEffectChar(t, g, creature).Restrictions.Has(CantBlock) {
		t.Fatal("a 2/2 under a power-2 bound may block")
	}
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, []AffectedObject{{ID: creature}},
			[]Mod{{Kind: ModModifyPT, Power: 3}}, g.UntilEndOfTurnDuration(), "pump")
	})
	if scopedEffectChar(t, g, creature).Restrictions.Has(CantBlock) {
		t.Error("a 5/2 is still under a power-2 bound")
	}
}

// TestMatchingScopeNeedsAQuery — the registration guard: the scope and
// its filter travel together, so a record with one and not the other is
// a programming error (matchingScopeProblem).
func TestMatchingScopeNeedsAQuery(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		defer func() {
			if recover() == nil {
				t.Error("a query-filtered record with no query registered without a panic")
			}
		}()
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeCreaturesMatching, me.ID,
			[]Mod{AddRestrictionsMod(CantBlock)}, g.UntilEndOfTurnDuration(), "no filter")
	})
	g.WithWriteLock(func() {
		if g.CantBlockThisTurnMatchingForEffect(uuid.Nil, me.ID, nil, false, "no filter") {
			t.Error("no queries registered a record")
		}
	})
}
