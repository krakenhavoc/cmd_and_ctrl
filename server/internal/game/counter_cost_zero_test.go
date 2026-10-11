package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_cost_zero_test.go — #2951. "Remove any number of counters"
// prints a floor of zero, and zero is a real announcement: Black Mana
// Battery adds {B} for it. Nothing comes off, the source still taps,
// and the paid count the produced mana reads is zero.

func zeroRemovalSource(g *Game, me *Player, counters map[string]int, floor int) uuid.UUID {
	return pushManaSource(g, me, ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "storage", N: floor, Variable: true},
		ProducedForPaid: func(_ *Game, _, _ uuid.UUID, paid PaidCost) string {
			return "{B}" + map[bool]string{true: "{B}", false: ""}[paid.CountersRemoved > 0]
		},
	}, counters)
}

func TestVariableCounterCostAcceptsAnAnnouncementOfZero(t *testing.T) {
	for _, tc := range []struct {
		name     string
		counters map[string]int
		params   ManaAbilityParams
	}{
		{"no payment fields, none held", nil, ManaAbilityParams{}},
		{"no payment fields, some held", map[string]int{"storage": 3}, ManaAbilityParams{}},
		{"an explicit zero", map[string]int{"storage": 3}, ManaAbilityParams{CounterCounts: []int{0}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			land := zeroRemovalSource(g, me, tc.counters, 0)

			if err := g.ActivateManaAbility(me.ID, land, 0, tc.params); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if got := counterOf(g, land, "storage"); got != tc.counters["storage"] {
				t.Errorf("storage counters %d, want %d untouched", got, tc.counters["storage"])
			}
			if len(me.ManaPool) != 1 {
				t.Errorf("pool has %d tokens, want the one {B} for removing none", len(me.ManaPool))
			}
			if c := findBattlefieldCard(g, land); c == nil || !c.Tapped {
				t.Error("the source did not tap")
			}
		})
	}
}

// Zero is only an announcement when the printed floor is zero, and
// only for the self form; every other shape still refuses it.
func TestZeroAnnouncementIsRefusedWhereTheFloorIsAboveZero(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	land := zeroRemovalSource(g, me, map[string]int{"storage": 3}, 1)

	for _, params := range []ManaAbilityParams{{}, {CounterCounts: []int{0}}} {
		err := g.ActivateManaAbility(me.ID, land, 0, params)
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("zero under a floor of 1 with %+v: %v, want ErrInvalidParam", params, err)
		}
	}
	if got := counterOf(g, land, "storage"); got != 3 {
		t.Errorf("a refused activation changed the counters: %d", got)
	}
}
