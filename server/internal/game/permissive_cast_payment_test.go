package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// #2033 and #3844 (ADR 0068 amendments 2026-10-12 and 2026-10-15): a
// cast spends what the pool can legally pay in every mode. A funded
// cast is paid in full. A waived one (an insufficient permissive pool,
// or ForceCast) spends the share the pool can pay, records it, and
// waives only the shortfall, which is what OnPaper and the cost warning
// say. Mana the cost cannot use stays in the pool.
func TestWaivedCastSpendsWhatThePoolCanPay(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cost    string
		pool    ManaPool
		params  CastSpellParams
		spent   int
		onPaper bool
		left    []string // the colours still in the pool afterwards
	}{
		{"funded", "{1}{G}", ManaPool{{Color: "G"}, {Color: "C"}, {Color: "U"}}, CastSpellParams{}, 2, false, []string{"U"}},
		{"insufficient", "{1}{G}", ManaPool{{Color: "G"}}, CastSpellParams{}, 1, true, nil},
		{"insufficient, generic only", "{3}", ManaPool{{Color: "R"}, {Color: "C"}}, CastSpellParams{}, 2, true, nil},
		{"wrong color", "{G}", ManaPool{{Color: "U"}}, CastSpellParams{}, 0, true, []string{"U"}},
		{"colour short, generic paid", "{1}{U}", ManaPool{{Color: "G"}}, CastSpellParams{}, 1, true, nil},
		{"part usable", "{R}{R}", ManaPool{{Color: "G"}, {Color: "R"}}, CastSpellParams{}, 1, true, []string{"G"}},
		{"restricted", "{G}", ManaPool{{Color: "G", Restrictions: []string{ManaRestrictType("Creature")}}}, CastSpellParams{}, 0, true, []string{"G"}},
		{"forced funded", "{G}", ManaPool{{Color: "G"}}, CastSpellParams{Strict: true, ForceCast: true}, 1, false, nil},
		{"forced permissive", "{G}", ManaPool{{Color: "G"}}, CastSpellParams{ForceCast: true}, 1, false, nil},
		{"forced short", "{1}{G}", ManaPool{{Color: "G"}}, CastSpellParams{Strict: true, ForceCast: true}, 1, true, nil},
		{"forced empty", "{1}{G}", nil, CastSpellParams{Strict: true, ForceCast: true}, 0, true, nil},
		{"zero cost", "{0}", nil, CastSpellParams{}, 0, false, nil},
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
			if got := TotalMana(item.Paid.Mana); got != tc.spent || item.Paid.OnPaper != tc.onPaper {
				t.Fatalf("payment = %+v (%d spent), want %d spent, onPaper %v", item.Paid, got, tc.spent, tc.onPaper)
			}
			// #3855: nothing reads this payment, so the waived share
			// is recorded as an amount asserted paid on paper, and the
			// total is the whole cost.
			if tc.onPaper && item.Paid.PaperMana == 0 {
				t.Errorf("a waived payment recorded no paper mana: %+v", item.Paid)
			}
			if !tc.onPaper && item.Paid.PaperMana != 0 {
				t.Errorf("a paid cast recorded paper mana: %+v", item.Paid)
			}
			// What was spent came out of the pool, and nothing else did.
			if p.ManaPool.Total() != before.Total()-tc.spent {
				t.Fatalf("pool has %d mana, want %d", p.ManaPool.Total(), before.Total()-tc.spent)
			}
			if tc.spent == 0 && !reflect.DeepEqual(p.ManaPool, before) {
				t.Fatalf("a payment that spent nothing changed the pool: %+v -> %+v", before, p.ManaPool)
			}
			var left []string
			for _, tok := range p.ManaPool.Units() {
				left = append(left, tok.Color)
			}
			if !reflect.DeepEqual(left, tc.left) {
				t.Errorf("pool left %v, want %v", left, tc.left)
			}
			// The recorded tokens are the pool's own, sources intact.
			for _, tok := range item.Paid.Mana {
				found := false
				for _, b := range before {
					if b.Source == tok.Source {
						found = true
					}
				}
				if !found {
					t.Errorf("recorded token %+v did not come from the pool", tok)
				}
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

// #3844: a permissive activation whose pool cannot cover the cost pays
// the share it can, records it, and waives the rest, exactly as a cast
// does. The {G} cannot pay {R} and stays in the pool.
func TestWaivedActivationSpendsWhatThePoolCanPay(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	p.ManaPool = ManaPool{{Color: "R"}, {Color: "G"}}
	cost, err := ParseCost("{R}{R}")
	if err != nil {
		t.Fatal(err)
	}
	paid, err := g.payAbilityManaCostLocked(p, uuid.New(), "Test Source", cost, ActivateAbilityParams{}, ManaSpendContext{}, nil)
	if err != nil {
		t.Fatalf("pay: %v", err)
	}
	if !paid.OnPaper || TotalMana(paid.Mana) != 1 || paid.Spent().Count("R") != 1 || paid.PaperMana != 1 {
		t.Fatalf("payment = %+v, want the {R} spent and the rest waived", paid)
	}
	if left := p.ManaPool.Units(); len(left) != 1 || left[0].Color != "G" {
		t.Errorf("pool left %+v, want the {G}", left)
	}
}

// #3844: a waived payment's record is a lower bound. The readers that
// grow with the mana spent count the tokens that left the pool; "no
// mana was spent" stays false, because the rest was paid on paper.
// #3855: the AMOUNT is exact, since the one-click path asserts the
// shortfall was paid on paper; the colours stay a lower bound.
func TestWaivedPaymentRecordReadsAsALowerBound(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	p.ManaPool = ManaPool{{Color: "U"}, {Color: "R", SourceKinds: ManaSourceTreasure}}
	item := castForTest(t, g, p, "{3}{U}{R}", CastSpellParams{})

	spent := item.Paid.Spent()
	if spent.Known() {
		t.Fatal("a waived payment reads as known")
	}
	if got := spent.Total(); got != 5 {
		t.Errorf("Total = %d, want the 2 that left the pool and the 3 paid on paper", got)
	}
	if got := spent.Colors(); !reflect.DeepEqual(got, []string{"U", "R"}) {
		t.Errorf("Colors = %v, want [U R]", got)
	}
	if got := spent.Count("R"); got != 1 {
		t.Errorf("Count(R) = %d, want 1", got)
	}
	if !spent.FromTreasure() {
		t.Error("the Treasure mana that left the pool does not count as spent")
	}
	if spent.None() {
		t.Error("a waived payment reads as \"no mana was spent\"")
	}
	if len(spent.Tokens()) != 2 {
		t.Errorf("Tokens = %+v, want the two pool tokens", spent.Tokens())
	}
}
