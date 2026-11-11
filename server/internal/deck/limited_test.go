package deck

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func limitedCards(n int, name string, basic bool) []LimitedCard {
	out := make([]LimitedCard, n)
	for i := range out {
		out[i] = LimitedCard{Name: name, Basic: basic}
	}
	return out
}

// limitedPool is 23 different spells, two copies of "Twin", and one
// basic land a pack held.
func limitedPool() []LimitedCard {
	var pool []LimitedCard
	for i := range 23 {
		pool = append(pool, LimitedCard{Name: "Spell " + string(rune('A'+i))})
	}
	pool = append(pool, limitedCards(2, "Twin", false)...)
	pool = append(pool, LimitedCard{Name: "Forest", Basic: true})
	return pool
}

func limitedCodes(t *testing.T, err error) []string {
	t.Helper()
	if err == nil {
		return nil
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("error is %T, want *ValidationError: %v", err, err)
	}
	var codes []string
	for _, v := range ve.Violations {
		codes = append(codes, v.Code)
	}
	return codes
}

func TestValidateLimited(t *testing.T) {
	pool := limitedPool()
	spells := pool[:23]
	tests := []struct {
		name string
		main []LimitedCard
		want []string
	}{
		{
			name: "23 drafted cards and 17 basics is legal",
			main: append(append([]LimitedCard(nil), spells...), limitedCards(17, "Island", true)...),
		},
		{
			name: "basic lands are free in any number, whatever the pool holds",
			main: append(append([]LimitedCard(nil), spells[:1]...), limitedCards(60, "Swamp", true)...),
		},
		{
			name: "more than 40 cards is legal",
			main: append(append(append([]LimitedCard(nil), spells...), limitedCards(2, "Twin", false)...), limitedCards(20, "Plains", true)...),
		},
		{
			name: "too few cards",
			main: append(append([]LimitedCard(nil), spells...), limitedCards(16, "Island", true)...),
			want: []string{CodeWrongCardCount},
		},
		{
			name: "a card that is not in the pool",
			main: append(append(append([]LimitedCard(nil), spells...), LimitedCard{Name: "Black Lotus"}), limitedCards(17, "Island", true)...),
			want: []string{CodeNotInPool},
		},
		{
			name: "more copies than the pool holds",
			main: append(append(append([]LimitedCard(nil), spells...), limitedCards(3, "Twin", false)...), limitedCards(17, "Island", true)...),
			want: []string{CodeMoreThanPool},
		},
		{
			name: "every problem is named",
			main: append(limitedCards(3, "Twin", false), LimitedCard{Name: "Black Lotus"}),
			want: []string{CodeWrongCardCount, CodeNotInPool, CodeMoreThanPool},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := limitedCodes(t, ValidateLimited(tc.main, pool))
			if len(got) != len(tc.want) {
				t.Fatalf("violations = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("violations = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// CR 100.4b: the sideboard is the pool less the deck. A basic the deck
// took from outside the pool takes nothing out of it.
func TestLimitedSideboardIsTheRestOfThePool(t *testing.T) {
	pool := limitedPool()
	main := append(append([]LimitedCard(nil), pool[:20]...), LimitedCard{Name: "Twin"})
	main = append(main, limitedCards(19, "Island", true)...)
	side := LimitedSideboard(main, pool)
	// 26 in the pool, 21 of them in the deck.
	if len(side) != 5 {
		t.Fatalf("sideboard has %d cards, want 5", len(side))
	}
	names := map[string]int{}
	for _, i := range side {
		names[pool[i].Name]++
	}
	if names["Twin"] != 1 || names["Forest"] != 1 || names["Spell U"] != 1 {
		t.Fatalf("sideboard = %v", names)
	}
}

// A list alone cannot be a limited deck: the check needs the pool.
func TestValidateRefusesALimitedListWithoutItsPool(t *testing.T) {
	if err := Validate(&List{}, game.FormatLimited); err == nil {
		t.Fatal("Validate accepted a limited deck with no pool to check it against")
	}
}
