package game

import (
	"testing"
)

// optional_life_cost_test.go — ADR 0100's 2026-11-20 amendment (#5782):
// an optional additional cost that pays a fixed amount of life
// ("Buyback—Pay 4 life", Slaughter). The life is paid with the total
// cost (CR 601.2h), and only by a caster whose life total is at least
// the amount (CR 119.4); declining it pays nothing.

const testBuybackLifeOracle = "test-optional-buyback-life"

func TestOptionalLifeCostIsPaidOnlyWhenAnnouncedAndAffordable(t *testing.T) {
	buyback := []AdditionalCost{
		{Optional: true, Key: BuybackKey, PayLife: 4, Label: "Buyback—Pay 4 life"},
	}
	for _, tc := range []struct {
		name     string
		life     int
		claim    []int
		wantErr  bool
		wantLife int
	}{
		{"declined pays no life", 20, nil, false, 20},
		{"announced pays the life", 20, []int{0}, false, 16},
		{"exactly the amount may be paid", 4, []int{0}, false, 0},
		{"below the amount is refused", 3, []int{0}, true, 3},
		{"declined is legal at any life total", 3, nil, false, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			stubOptionalCosts(t, testBuybackLifeOracle, buyback)
			active := g.Seats[g.Turn.ActiveSeat]
			active.Life = tc.life

			id, err := castOptional(t, g, "Life Buyback", "Instant", testBuybackLifeOracle, tc.claim)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CastSpell = %v, wantErr %v", err, tc.wantErr)
			}
			if active.Life != tc.wantLife {
				t.Errorf("life after the cast = %d, want %d", active.Life, tc.wantLife)
			}
			if tc.wantErr {
				if !active.Hand.Contains(id) {
					t.Errorf("a refused cast did not leave the card in hand")
				}
				return
			}
			if tc.wantLife == 0 {
				// The cast was accepted; the state-based check then takes the
				// caster out of the game (CR 704.5a) with their spell.
				return
			}
			item := g.StackMeta[id]
			if item == nil {
				t.Fatalf("no stack item for the cast spell")
			}
			if paid := item.Paid.PaidOptionalCost(0); paid != (len(tc.claim) > 0) {
				t.Errorf("PaidOptionalCost(0) = %v, want %v", paid, len(tc.claim) > 0)
			}
		})
	}
}
