package game

import "testing"

// compleated_provenance_test.go — ADR 0183: PaidCost.PhyrexianLife counts
// only PRINTED Phyrexian symbols paid with life, never a {B} that a
// life-for-mana grant (K'rrik, ADR 0131) let be paid with life.
func TestPhyrexianLifeCountsOnlyPrintedSymbols(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[0]
	for _, tc := range []struct {
		name          string
		claimed       int
		wantPhyrexian int
		wantLife      int
	}{
		// {B} (granted) comes first in printed order, so one claim
		// strikes it: life is paid, no Phyrexian symbol is.
		{"granted symbol only", 1, 0, 2},
		{"both", 2, 1, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cost, err := ParseCost("{B}{B/P}")
			if err != nil {
				t.Fatal(err)
			}
			cost.Required[0].LifeGranted = true
			var paid PaidCost
			g.WithWriteLock(func() {
				_, _, err = g.strikePhyrexianLifeLocked(p, "Test", cost, ManaSpendContext{}, tc.claimed, &paid)
			})
			if err != nil {
				t.Fatalf("strike: %v", err)
			}
			if paid.PhyrexianLife != tc.wantPhyrexian || paid.LifePaid != tc.wantLife {
				t.Errorf("PhyrexianLife %d, LifePaid %d; want %d and %d", paid.PhyrexianLife, paid.LifePaid, tc.wantPhyrexian, tc.wantLife)
			}
		})
	}
}
