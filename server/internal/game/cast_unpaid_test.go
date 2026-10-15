package game

import "testing"

// ADR 0118 §2 (#2188), amended by #3855: what a caster says about a
// waived payment is announced on its EventCast, so the log can tell the
// table. A one-click Cast anyway names the shortfall it asserts was paid
// on paper (PaperPaid); a declaration names the declared mana
// (PaperDeclared) or, declared free, is Unpaid. A permissive cast nobody
// asked about and a paid cast announce nothing. Since #3844 a forced
// cast spends what the pool can pay, so one the pool covered in full was
// paid and announces nothing either.
func TestForcedCastIsAnnouncedUnpaid(t *testing.T) {
	cases := []struct {
		name      string
		params    CastSpellParams
		pool      []ManaToken
		unpaid    bool
		paper     string
		declared  bool
		wantPaper int
	}{
		{"forced, strict", CastSpellParams{Strict: true, ForceCast: true}, nil, false, "{U}{U}", false, 2},
		{"forced, pool pays part", CastSpellParams{Strict: true, ForceCast: true}, []ManaToken{{Color: "U"}}, false, "{U}", false, 1},
		{"forced, pool pays all", CastSpellParams{Strict: true, ForceCast: true}, []ManaToken{{Color: "U"}, {Color: "U"}}, false, "", false, 0},
		{"forced, declared free", CastSpellParams{Strict: true, ForceCast: true, PaperPayment: &PaperPayment{Free: true}}, nil, true, "", false, 0},
		{"forced, declared paid", CastSpellParams{Strict: true, ForceCast: true, PaperPayment: &PaperPayment{Mana: []PaperMana{{Color: "U"}, {Color: "U"}}}}, nil, false, "{U}{U}", true, 0},
		{"permissive", CastSpellParams{}, nil, false, "", false, 2},
		{"permissive, declared", CastSpellParams{PaperPayment: &PaperPayment{Mana: []PaperMana{{Color: "U"}}}}, []ManaToken{{Color: "U"}}, false, "{U}", true, 0},
		{"strict, paid", CastSpellParams{Strict: true}, []ManaToken{{Color: "U"}, {Color: "U"}}, false, "", false, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			p := g.Seats[0]
			id := pushTypedCardToHandWithCost(p, "Counterspell", "Instant", "{U}{U}")
			p.ManaPool.AddMana(tc.pool...)

			if err := g.CastSpell(p.ID, id, tc.params); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			ev, ok := lastEvent(g, EventCast)
			if !ok {
				t.Fatal("no EventCast")
			}
			if ev.CardID != id {
				t.Fatalf("EventCast names %v, want %v", ev.CardID, id)
			}
			if ev.Unpaid != tc.unpaid || ev.PaperPaid != tc.paper || ev.PaperDeclared != tc.declared {
				t.Errorf("EventCast = {Unpaid:%v PaperPaid:%q PaperDeclared:%v}, want {%v %q %v}",
					ev.Unpaid, ev.PaperPaid, ev.PaperDeclared, tc.unpaid, tc.paper, tc.declared)
			}
			if got := g.StackMeta[id].Paid.PaperMana; got != tc.wantPaper {
				t.Errorf("PaperMana = %d, want %d", got, tc.wantPaper)
			}
		})
	}
}
