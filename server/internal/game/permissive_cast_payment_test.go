package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestPermissiveCastSpendsOnlyAFullyFundedPool(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cost    string
		pool    ManaPool
		params  CastSpellParams
		spent   int
		onPaper bool
	}{
		{"funded", "{1}{G}", ManaPool{{Color: "G"}, {Color: "C"}, {Color: "U"}}, CastSpellParams{}, 2, false},
		{"insufficient", "{1}{G}", ManaPool{{Color: "G"}}, CastSpellParams{}, 0, true},
		{"wrong color", "{G}", ManaPool{{Color: "U"}}, CastSpellParams{}, 0, true},
		{"restricted", "{G}", ManaPool{{Color: "G", Restrictions: []string{ManaRestrictType("Creature")}}}, CastSpellParams{}, 0, true},
		{"forced funded", "{G}", ManaPool{{Color: "G"}}, CastSpellParams{Strict: true, ForceCast: true}, 0, true},
		{"forced permissive", "{G}", ManaPool{{Color: "G"}}, CastSpellParams{ForceCast: true}, 0, true},
		{"zero cost", "{0}", nil, CastSpellParams{}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			p := g.Seats[0]
			p.ManaPool = tc.pool
			for i := range p.ManaPool {
				p.ManaPool[i].Source = uuid.New()
			}
			before := append(ManaPool(nil), p.ManaPool...)
			item := castForTest(t, g, p, tc.cost, tc.params)
			if len(item.Paid.Mana) != tc.spent || item.Paid.OnPaper != tc.onPaper {
				t.Fatalf("payment = %+v, want %d spent, onPaper %v", item.Paid, tc.spent, tc.onPaper)
			}
			if p.ManaPool.Total() != len(before)-tc.spent {
				t.Fatalf("pool has %d tokens, want %d", p.ManaPool.Total(), len(before)-tc.spent)
			}
			if tc.onPaper && !reflect.DeepEqual(p.ManaPool, before) {
				t.Fatalf("waived charge changed pool: %+v -> %+v", before, p.ManaPool)
			}
			warned := false
			for _, ev := range g.Events {
				if ev.Kind == EventCostWarning && ev.Source == item.ID {
					warned = true
				}
			}
			if warned != tc.onPaper {
				t.Errorf("cost warning = %v, want %v", warned, tc.onPaper)
			}
		})
	}
}
