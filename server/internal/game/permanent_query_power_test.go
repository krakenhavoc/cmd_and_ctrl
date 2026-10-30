package game

import (
	"testing"

	"github.com/google/uuid"
)

// permanent_query_power_test.go — #4659: PermanentQuery.PowerAtMost and
// PowerAtLeast, the bound in Queen Mother Ramonda's "creatures with power 2
// or less".

func TestPermanentQueryPowerBoundsReadEffectivePowerAndCounters(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	p := g.Seats[1].ID
	small := pushTypedTestCard(g, Card{Name: "Mouse", TypeLine: "Creature — Mouse", Power: 1, Toughness: 1, Owner: p, Controller: p})
	two := pushTypedTestCard(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: p, Controller: p})
	grown := pushTypedTestCard(g, Card{Name: "Grown Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: p, Controller: p,
		Counters: map[string]int{"+1/+1": 1}})
	zero := pushTypedTestCard(g, Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Owner: p, Controller: p})
	land := pushLandForTest(g, p, "Island", "Basic Land — Island")

	atMost2 := PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2)
	atLeast2 := PermanentQuery{Types: []string{"creature"}}.WithPowerAtLeast(2)
	atMost0 := PermanentQuery{}.WithPowerAtMost(0)
	cases := []struct {
		name string
		q    PermanentQuery
		yes  []uuid.UUID
		no   []uuid.UUID
	}{
		{"power 2 or less", atMost2, []uuid.UUID{small, two, zero}, []uuid.UUID{grown, land}},
		{"power 2 or greater", atLeast2, []uuid.UUID{two, grown}, []uuid.UUID{small, zero, land}},
		{"zero is a real bound", atMost0, []uuid.UUID{zero, land}, []uuid.UUID{small, two}},
	}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, tc := range cases {
			for _, id := range tc.yes {
				if !tc.q.matchesLocked(g, findBattlefieldCard(g, id)) {
					t.Errorf("%s: does not match %s", tc.name, findBattlefieldCard(g, id).Name)
				}
			}
			for _, id := range tc.no {
				if tc.q.matchesLocked(g, findBattlefieldCard(g, id)) {
					t.Errorf("%s: matches %s", tc.name, findBattlefieldCard(g, id).Name)
				}
			}
		}
	})
}

func TestPermanentQueryPowerBoundsAreInEqualNounAndClone(t *testing.T) {
	a := PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2)
	b := PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2)
	if !a.Equal(b) {
		t.Error("two queries with the same bound are not equal")
	}
	for name, other := range map[string]PermanentQuery{
		"another bound": PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(3),
		"no bound":      {Types: []string{"creature"}},
		"the other end": PermanentQuery{Types: []string{"creature"}}.WithPowerAtLeast(2),
	} {
		if a.Equal(other) {
			t.Errorf("a query bounded at most 2 equals one with %s", name)
		}
	}
	if got := clonePermanentQueries([]PermanentQuery{a}); len(got) != 1 || !got[0].Equal(a) {
		t.Errorf("clonePermanentQueries dropped the bound: %+v", got)
	}

	for _, tc := range []struct {
		q    PermanentQuery
		noun string
	}{
		{a, "creature with power 2 or less"},
		{PermanentQuery{Types: []string{"creature"}}.WithPowerAtLeast(3), "creature with power 3 or greater"},
		{PermanentQuery{Types: []string{"creature"}, Keyword: "flying"}.WithPowerAtMost(2), "creature with flying and power 2 or less"},
		{PermanentQuery{Types: []string{"creature"}}.WithPowerAtLeast(1).WithPowerAtMost(3), "creature with power 1 to 3"},
	} {
		if got := tc.q.Noun(); got != tc.noun {
			t.Errorf("noun = %q, want %q", got, tc.noun)
		}
	}
}

// A characteristic snapshot has no counters, so it can't say a power: a
// bounded query never matches one (like Enchanted).
func TestPermanentQueryPowerBoundNeverMatchesACharacteristicSnapshot(t *testing.T) {
	ch := &Characteristic{Types: []string{"creature"}, Power: 1}
	if (PermanentQuery{Types: []string{"creature"}}).matchesCharacteristic(ch) != true {
		t.Fatal("setup: the unbounded query should match")
	}
	if (PermanentQuery{}.WithPowerAtMost(2)).matchesCharacteristic(ch) {
		t.Error("a bounded query matched a characteristic snapshot")
	}
}
