package game

import "testing"

// ADR 0118 §2 (#2188): a cast made with ForceCast — the client's "Cast
// anyway (don't pay)" row and the dock's Cast anyway after a refused
// cast — is announced as Unpaid on its EventCast, so the log can tell
// the table. A permissive cast and a paid strict cast are not.
func TestForcedCastIsAnnouncedUnpaid(t *testing.T) {
	cases := []struct {
		name   string
		params CastSpellParams
		pool   []ManaToken
		want   bool
	}{
		{"forced, strict", CastSpellParams{Strict: true, ForceCast: true}, nil, true},
		{"forced, pool could pay", CastSpellParams{Strict: true, ForceCast: true}, []ManaToken{{Color: "U"}, {Color: "U"}}, true},
		{"permissive", CastSpellParams{}, nil, false},
		{"strict, paid", CastSpellParams{Strict: true}, []ManaToken{{Color: "U"}, {Color: "U"}}, false},
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
			if ev.Unpaid != tc.want {
				t.Errorf("EventCast.Unpaid = %v, want %v", ev.Unpaid, tc.want)
			}
		})
	}
}
