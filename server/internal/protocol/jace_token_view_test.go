package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jace_token_view_test.go — the wire half of ADR 0139 (#2796). The Jace
// planeswalker token has no printing and no oracle ID, so everything a
// player needs to use it — its loyalty, its two loyalty abilities and
// their costs, the text it prints — has to reach the board through the
// token key, exactly as a printed planeswalker's does through its
// oracle ID. Nothing on the wire is new; this pins that the existing
// fields are filled for a token.
func TestJaceTokenShipsItsLoyaltyAbilitiesAndText(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	var made []string
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me, effects.JaceToken(), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("create the Jace token: %v %v", ids, err)
		}
		if err := g.AddCounterForEffect(ids[0], game.CounterLoyalty, 2); err != nil {
			t.Fatalf("put loyalty on it: %v", err)
		}
		made = append(made, ids[0].String())
	})
	var c CardView
	for _, v := range ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == made[0] {
			c = v
		}
	}
	if c.InstanceID == "" {
		t.Fatal("the Jace token is missing from the view")
	}
	if c.TypeLine != "Token Planeswalker — Jace" || c.Name != "Jace" {
		t.Errorf("token reads %q / %q", c.Name, c.TypeLine)
	}
	if c.Counters[game.CounterLoyalty] != 2 {
		t.Errorf("loyalty on the wire %d, want 2", c.Counters[game.CounterLoyalty])
	}
	if c.TokenText == "" {
		t.Error("the token's printed text did not reach the board")
	}
	if len(c.ActivatedAbilities) != 2 {
		t.Fatalf("%d activated abilities on the wire, want 2", len(c.ActivatedAbilities))
	}
	for i, want := range []int{-1, -3} {
		a := c.ActivatedAbilities[i]
		if a.LoyaltyCost == nil || *a.LoyaltyCost != want {
			t.Errorf("ability %d (%q) loyalty_cost %v, want %d", i, a.Label, a.LoyaltyCost, want)
		}
	}
}
