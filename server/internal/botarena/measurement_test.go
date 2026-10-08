package botarena_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// measurement_test.go pins ADR 0126 §1, the measurement PR: the
// per-contestant Play table, the Cards section, and the arena-only
// `heuristic-baseline` contestant.

func TestParseContestant(t *testing.T) {
	for in, want := range map[string]botarena.SeatSpec{
		"heuristic":          {Tier: tiers.Heuristic},
		"random":             {Tier: tiers.Random},
		"heuristic-baseline": {Tier: tiers.Heuristic, Variant: botarena.VariantBaseline},
		"heuristic-noplan":   {Tier: tiers.Heuristic, Variant: botarena.VariantNoPlan},
	} {
		got, err := botarena.ParseContestant(in)
		if err != nil {
			t.Fatalf("ParseContestant(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("ParseContestant(%q) = %+v, want %+v", in, got, want)
		}
		if got.Contestant() != in || got.Label() != in {
			t.Errorf("%q round-trips as contestant %q, label %q", in, got.Contestant(), got.Label())
		}
	}
	_, err := botarena.ParseContestant("assisted-baseline")
	if err == nil || !strings.Contains(err.Error(), "heuristic-baseline") {
		t.Errorf("an unknown contestant was accepted, or the error does not name the arena-only one: %v", err)
	}
}

// heuristic-baseline is an ARENA name, not a tier: tiers.Parse, which
// the lobby and GET /bot/options read, must not know it.
func TestBaselineIsNotATier(t *testing.T) {
	for _, name := range []string{botarena.BaselineContestant, botarena.NoPlanContestant} {
		if _, err := tiers.Parse(name); err == nil {
			t.Fatalf("tiers.Parse accepts %q; it would be offered in the lobby", name)
		}
		for _, tier := range tiers.All() {
			if string(tier) == name {
				t.Fatalf("tiers.All lists %q", tier)
			}
		}
	}
}

func TestValidateRefusesAVariantOfTheWrongTier(t *testing.T) {
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{{Tier: tiers.Random, Variant: botarena.VariantBaseline}, {Tier: tiers.Heuristic}},
		Games: 1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("a random-baseline seat was accepted")
	}
	cfg.Seats[0] = botarena.SeatSpec{Tier: tiers.Heuristic, Variant: "frozen"}
	if err := cfg.Validate(); err == nil {
		t.Error("an unknown variant was accepted")
	}
	cfg.Seats[0] = botarena.SeatSpec{Tier: tiers.Heuristic, Variant: botarena.VariantBaseline}
	if err := cfg.Validate(); err != nil {
		t.Errorf("heuristic-baseline refused: %v", err)
	}
}

// A short real run fills the per-contestant Play table and the Cards
// section, in the Summary and in the Markdown.
func TestRunReportsContestantsAndCards(t *testing.T) {
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{
			{Tier: tiers.Heuristic},
			{Tier: tiers.Heuristic, Variant: botarena.VariantBaseline},
		},
		Games: 2, Seed: 66, TurnBudget: 5, Wall: 2 * time.Minute, Lockstep: true, Rotate: true,
	}
	sum, err := botarena.Run(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sum.PerContestant) != 2 {
		t.Fatalf("PerContestant has %d rows, want 2: %+v", len(sum.PerContestant), sum.PerContestant)
	}
	for _, row := range sum.PerContestant {
		pol := sum.PerPolicy[row.Policy]
		if pol == nil {
			t.Fatalf("contestant %q has no per-policy row", row.Policy)
		}
		// One deck per policy here, so the two tables agree.
		if row.Games != pol.Games || row.Wins != pol.Wins || row.Decided != pol.Decided {
			t.Errorf("%s: contestant row %+v disagrees with policy row %+v", row.Policy, row, pol)
		}
	}
	if len(sum.Cards) != 2 {
		t.Fatalf("Cards has %d contestants, want 2", len(sum.Cards))
	}
	casts := 0
	for _, cc := range sum.Cards {
		if cc.SeatGames != 2 {
			t.Errorf("%s played %d seat-games, want 2", cc.Contestant(), cc.SeatGames)
		}
		for _, c := range cc.Cards {
			if c.Taken > c.Windows || c.GamesUsed > c.GamesOffered || c.GamesOffered > cc.SeatGames {
				t.Errorf("%s: impossible tally %+v", cc.Contestant(), c)
			}
			if c.Action == botarena.ActionCast {
				casts += c.Taken
			}
		}
	}
	if casts == 0 {
		t.Error("five turns of the battle deck cast nothing, or the Cards section did not see it")
	}
	for _, row := range sum.PerContestant {
		tm := row.TurnMana
		if tm.Turns == 0 || tm.Stranded > tm.Idle || tm.Idle > tm.Turns {
			t.Errorf("%s: impossible or empty turn-mana tally %+v", row.Policy, tm)
		}
		// No plan is made before ADR 0136 PR 4.
		if tm.Planned != 0 || tm.Checked != 0 || tm.Misses != 0 {
			t.Errorf("%s: plan numbers before there is a plan: %+v", row.Policy, tm)
		}
	}
	md := sum.Markdown()
	for _, want := range []string{"### Play by contestant", "### Cards", "#### Acceptance-bar cards", "| A3 | Sol Ring |", "not offered", "### Turn mana (ADR 0136 §8)"} {
		if !strings.Contains(md, want) {
			t.Errorf("Markdown() is missing %q:\n%s", want, md)
		}
	}
}
