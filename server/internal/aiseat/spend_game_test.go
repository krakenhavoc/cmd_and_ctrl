package aiseat_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spend_game_test.go is #735's end-to-end half: a whole four-seat game
// against the fake client, with the per-game spend read off the
// runners the way the Manager reads it.
//
// What it can and cannot prove is worth stating, because the issue it
// closes is specifically about a number nobody has measured. Against
// a FakeClient the TOKEN counts are the fake's invention; what this
// test establishes is that the measurement PATH is whole — every call
// a seat makes is counted, counted once, attributed to the right seat
// and the right purpose, and reaches Runner.Stats and the game record
// without being lost or double-counted on the way. The number S31's
// exit criterion 4 wants is the same plumbing pointed at a keyed
// endpoint; see docs/bot.md.

// Every model call a seat makes is on that seat's Spend, and the
// game's record is the sum of the seats.
func TestPerGameSpendIsMeasuredAcrossAWholeGame(t *testing.T) {
	requireGameTests(t)
	const (
		turnBudget = 35
		wall       = 120 * time.Second
	)
	var meter rules.Meter
	client := model.EchoesFallback()
	policies, funnels := modelSeats(t, 4, client, &meter, nil)

	res := playGame(t, 7350, policies, turnBudget, wall)
	if res.state != game.StateEnded {
		t.Logf("game did not finish inside %d turns (state %s); the spend assertions still hold",
			turnBudget, res.state)
	}
	if len(res.stats) != len(funnels) {
		t.Fatalf("%d runner stats for %d seats", len(res.stats), len(funnels))
	}

	// 1. Per seat: the runner's Spend is the funnel's own counters,
	//    which is the claim that the projection lost nothing.
	var table aiseat.Spend
	for i, st := range res.stats {
		fs := funnels[i].Stats()
		if st.Spend.Decision.Calls != fs.ModelCalls {
			t.Errorf("seat %d: runner reports %d decision calls, the funnel counted %d",
				i, st.Spend.Decision.Calls, fs.ModelCalls)
		}
		if st.Spend.Decision.Usage.InputTokens != fs.Usage.InputTokens ||
			st.Spend.Decision.Usage.OutputTokens != fs.Usage.OutputTokens {
			t.Errorf("seat %d: runner reports %+v tokens, the funnel counted in %d / out %d",
				i, st.Spend.Decision.Usage, fs.Usage.InputTokens, fs.Usage.OutputTokens)
		}
		if st.Spend.Improvisation.Calls != fs.ImprovCalls {
			t.Errorf("seat %d: runner reports %d improvisation calls, the funnel counted %d",
				i, st.Spend.Improvisation.Calls, fs.ImprovCalls)
		}
		table = table.Add(st.Spend)
	}

	// 2. The table: every call the CLIENT saw is accounted for
	//    exactly once. This is the double-counting guard — the one
	//    failure that would make a measured number worse than no
	//    number at all.
	if got, want := table.Total().Calls, int64(client.Calls()); got != want {
		t.Errorf("the seats' spend adds up to %d model calls; the client was dialled %d times", got, want)
	}
	if table.Decision.Calls == 0 {
		t.Fatal("no seat spent anything; the model path never ran and this test proves nothing")
	}
	if table.Decision.Usage.InputTokens == 0 {
		t.Error("model calls were made and no input tokens were recorded — " +
			"the provider's usage fields are not reaching the measurement")
	}

	// 3. The record the Manager writes, built the same way.
	gameID := uuid.New()
	t.Logf("per-game spend (fake client): %d calls (%d decision, %d improvisation), "+
		"in %d / out %d tokens, %v inside model calls",
		table.Total().Calls, table.Decision.Calls, table.Improvisation.Calls,
		table.Total().Usage.InputTokens, table.Total().Usage.OutputTokens,
		table.Total().Latency.Round(time.Millisecond))
	gs := aiseat.GameSpend{Game: gameID}
	for i, st := range res.stats {
		gs.Seats = append(gs.Seats, aiseat.SeatSpend{
			Seat: uuid.New(), Policy: policies[i].Name(), Spend: st.Spend,
		})
	}
	if gs.Total() != table {
		t.Errorf("GameSpend.Total() = %+v, want %+v", gs.Total(), table)
	}
	if gs.Empty() {
		t.Error("a game with model calls in it reports an empty spend")
	}
	if got := gs.Tiers(); got != "assisted x4" {
		t.Errorf("Tiers() = %q, want %q", got, "assisted x4")
	}
}

// A four-heuristic table spends nothing, and says so. The zero is a
// measurement: it is the evidence that the free tiers are free, which
// is the other half of ADR 0033 §5's cost argument.
func TestAFreeTableSpendsNothing(t *testing.T) {
	requireGameTests(t)
	res := playGame(t, 7354, heuristicSeats(4), 35, 120*time.Second)
	for i, st := range res.stats {
		if !st.Spend.Empty() {
			t.Errorf("heuristic seat %d spent %+v", i, st.Spend)
		}
	}
}
