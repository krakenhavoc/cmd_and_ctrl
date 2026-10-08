package botarena_test

import (
	"context"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// monolith_test.go is #2500's arena check: with the Monolith deck, no
// heuristic table stalls on "{3}: Untap Basalt Monolith" (the CR 732
// breaker and the runner's #810 hold parked the seat before the fix),
// and the untap is offered and never taken.

func TestMonolithBattleDeck(t *testing.T) {
	deck := botarena.MonolithBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
	}
	n := 0
	for _, c := range deck {
		if c.Name == "Basalt Monolith" || c.Name == "Grim Monolith" {
			n++
		}
	}
	if n != 10 {
		t.Errorf("%d Monoliths, want 10", n)
	}
	if !botarena.IsSynthetic(botarena.MonolithDeckID) {
		t.Error("monolith-battle needs no Scryfall dump")
	}
}

func TestNoTableStallsOnAMonolithUntap(t *testing.T) {
	deck := botarena.MonolithDeckID
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{
			{Tier: tiers.Heuristic, Deck: deck}, {Tier: tiers.Heuristic, Deck: deck},
		},
		Games: 4, Seed: 2500, TurnBudget: 14, Wall: 3 * time.Minute, Lockstep: true, Rotate: true,
	}
	sum, err := botarena.Run(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if sum.Stalls() != 0 {
		t.Errorf("%d of %d games stalled: a heuristic seat untapped a Monolith for no net mana (#2500)", sum.Stalls(), len(sum.Games))
	}
	for _, cc := range sum.Cards {
		for _, c := range cc.Cards {
			if c.Action == botarena.ActionActivate && (c.Name == "Basalt Monolith" || c.Name == "Grim Monolith") && c.Taken != 0 {
				t.Errorf("%s untapped %d times", c.Name, c.Taken)
			}
		}
	}
}
