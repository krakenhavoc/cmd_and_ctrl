package botarena_test

import (
	"context"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// exert_test.go is ADR 0130 §9's small arena check: the exert battle
// deck needs no dump, `heuristic-noexert` is an arena contestant, and a
// short run tallies the exert attacks in the Cards section, taken by
// `heuristic` and never by `heuristic-noexert`.

func TestParseNoExertContestant(t *testing.T) {
	got, err := botarena.ParseContestant(botarena.NoExertContestant)
	if err != nil {
		t.Fatal(err)
	}
	if got != (botarena.SeatSpec{Tier: tiers.Heuristic, Variant: botarena.VariantNoExert}) || got.Contestant() != "heuristic-noexert" {
		t.Errorf("parsed %+v (%q)", got, got.Contestant())
	}
	if _, err := tiers.Parse(botarena.NoExertContestant); err == nil {
		t.Error("heuristic-noexert is a tier; it would be offered in the lobby")
	}
}

func TestExertBattleDeck(t *testing.T) {
	deck := botarena.ExertBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
	}
	exert := 0
	for _, c := range deck {
		if c.OracleID != "" && c.Name != "Lightning Bolt" {
			exert++
		}
	}
	if exert != 15 {
		t.Errorf("%d exert cards, want 15", exert)
	}
}

func TestArenaTalliesExertAttacks(t *testing.T) {
	// Parallel (#2766): a lockstep arena run owns its games, rooms and
	// seeds, so it replays the same tables beside other tests.
	t.Parallel()
	deck := botarena.ExertDeckID
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{
			{Tier: tiers.Heuristic, Deck: deck},
			{Tier: tiers.Heuristic, Variant: botarena.VariantNoExert, Deck: deck},
		},
		Games: 4, Seed: 130, TurnBudget: 14, Wall: 3 * time.Minute, Lockstep: true, Rotate: true,
	}
	sum, err := botarena.Run(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	offered := map[string]int{}
	taken := map[string]int{}
	for _, cc := range sum.Cards {
		for _, c := range cc.Cards {
			if c.Action == botarena.ActionExert {
				offered[cc.Policy] += c.Windows
				taken[cc.Policy] += c.Taken
			}
		}
	}
	if offered["heuristic"] == 0 || offered["heuristic-noexert"] == 0 {
		t.Fatalf("no exert attack offered in four games of the exert deck: %v", offered)
	}
	if taken["heuristic-noexert"] != 0 {
		t.Errorf("heuristic-noexert exerted %d times", taken["heuristic-noexert"])
	}
	t.Logf("exert attacks offered %v, taken %v", offered, taken)
}
