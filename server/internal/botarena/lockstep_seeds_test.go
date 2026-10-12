package botarena_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// #2904: seeds 13 and 52 of a 64-game, four-curated-deck, rotated
// lockstep run (seed 1) did not replay. ScoreEval summed the opposition
// in seat-ID order, and a seat ID is a fresh UUID for every game, so
// the float rounding of the sum, and with it a tie between equally
// priced moves (Loran of the Third Path's three targets), differed from
// one run to the next. TestScoreEvalDoesNotDependOnSeatIDs
// (aiseat/heuristic) pins the cause cheaply; this plays the two seeds
// as the report did, in the chairs `--seed 1 --rotate` gives them
// (games 12 and 51), twice each, and compares the move logs.
//
// Gated on CMDCTRL_SCRYFALL_DUMP like the other curated-deck tests: the
// four decks are built from the real catalog, and four games of ~2,000
// to ~4,000 moves each are far too slow for the PR gate (about ten
// seconds a game without -race). The nightly sets the dump.
func TestLockstepRotatedSeedsReplay(t *testing.T) {
	dump := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if dump == "" {
		t.Skip("needs CMDCTRL_SCRYFALL_DUMP: the curated decks are built from the real catalog")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(dump); err != nil {
		t.Fatalf("load %s: %v", dump, err)
	}
	var seats []botarena.SeatSpec
	for _, d := range []string{"izzet-aggro", "mono-black-aristocrats", "simic-ramp", "esper-control"} {
		seats = append(seats, botarena.SeatSpec{Tier: tiers.Heuristic, Deck: d})
	}
	cfg := botarena.Config{
		Seats: seats, Games: 64, Seed: 1, Rotate: true, TurnBudget: 60,
		Wall: 10 * time.Minute, Lockstep: true, Index: idx,
	}
	for _, g := range []struct{ seed uint64 }{{13}, {52}} {
		game := int(g.seed - cfg.Seed)
		play := func() (botarena.GameResult, []string) {
			log := &moveLog{}
			c := cfg
			c.Runner.Observer = log
			res, err := botarena.Play(context.Background(), c, g.seed, botarena.Order(len(seats), game, true))
			if err != nil {
				t.Fatalf("Play seed %d: %v", g.seed, err)
			}
			return res, log.lines()
		}
		a, logA := play()
		b, logB := play()
		if a.Stalled || b.Stalled {
			t.Fatalf("seed %d stalled:\n%s\n%s", g.seed, a.StallDump, b.StallDump)
		}
		if len(logA) < 50 {
			t.Fatalf("seed %d logged %d moves; too few to mean anything", g.seed, len(logA))
		}
		for i := 0; i < len(logA) && i < len(logB); i++ {
			if logA[i] != logB[i] {
				t.Fatalf("seed %d diverged at move %d of %d/%d:\n  run 1: %s\n  run 2: %s",
					g.seed, i, len(logA), len(logB), logA[i], logB[i])
			}
		}
		if len(logA) != len(logB) || a.Turns != b.Turns || a.Winner != b.Winner {
			t.Fatalf("seed %d: moves %d/%d, turns %d/%d, winner %d/%d",
				g.seed, len(logA), len(logB), a.Turns, b.Turns, a.Winner, b.Winner)
		}
	}
}
