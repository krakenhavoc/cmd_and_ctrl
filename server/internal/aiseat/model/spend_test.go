package model

import (
	"context"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
)

// spend_test.go is #735's unit half: the funnel reports what it spent,
// and it never adds the two purposes together on the way out.
//
// The split is the part worth pinning. A decision call asks for one
// integer against a cached prefix; an improvisation call asks for a
// paragraph written from oracle text, with MaxTokens an order of
// magnitude larger and its own hard per-game cap. A Spend that folded
// them into one number would read as "the average call costs 500
// tokens" on a table where no call cost anything like 500 tokens.

// spendingClient reports usage a test can tell apart: decision calls
// and improvisation calls get different, recognisable token counts.
// A decision request is the one with more than one system block (the
// primer plus the deck); improvisation sends exactly one.
func spendingClient() *FakeClient {
	return &FakeClient{
		Label: "spending",
		Reply: func(_ int, req Request) (Response, error) {
			if len(req.System) == 1 {
				// Improvisation: one cache-marked primer, a bundle back.
				return Response{
					Text: `{"effect": "lose 3 life", "why": "no spec",
					  "steps": [{"verb": "change_life", "player": "` + meSeat.String() + `", "params": {"delta": -3}}]}`,
					Usage: Usage{InputTokens: 1000, OutputTokens: 100, CacheReadTokens: 10},
				}, nil
			}
			return Response{
				Text:  `{"index": 0, "why": "decision"}`,
				Usage: Usage{InputTokens: 7, OutputTokens: 3, CacheReadTokens: 1},
			}, nil
		},
	}
}

// The funnel is a Spender, and the two purposes are counted apart and
// added only by Total.
func TestSpendSplitsDecidingFromImprovising(t *testing.T) {
	client := spendingClient()
	p := improvPolicy(t, client, nil)

	if !p.Spend().Empty() {
		t.Fatalf("a seat that has done nothing has spent something: %+v", p.Spend())
	}

	// Three decision windows.
	for i := 0; i < 3; i++ {
		if _, err := p.Decide(context.Background(), castWindow()); err != nil {
			t.Fatalf("Decide: %v", err)
		}
	}
	// One improvisation: the spell crosses the stack, then resolves.
	if _, want := improvise(t, p, improvInput(castingView(grimID, "Grim Tutor", "Sorcery")), time.Second); want {
		t.Fatal("improvised a spell that is still on the stack")
	}
	if _, want := improvise(t, p, improvInput(resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")), time.Second); !want {
		t.Fatal("the resolved uncatalogued spell produced no bundle; the rest of this test proves nothing")
	}

	sp := p.Spend()
	if sp.Decision.Calls != 3 {
		t.Errorf("decision calls = %d, want 3", sp.Decision.Calls)
	}
	if sp.Improvisation.Calls != 1 {
		t.Errorf("improvisation calls = %d, want 1", sp.Improvisation.Calls)
	}
	if got, want := sp.Decision.Usage.InputTokens, 3*7; got != want {
		t.Errorf("decision input tokens = %d, want %d", got, want)
	}
	if got, want := sp.Improvisation.Usage.InputTokens, 1000; got != want {
		t.Errorf("improvisation input tokens = %d, want %d — "+
			"the two purposes have been folded together", got, want)
	}
	// Total is the only place they meet.
	total := sp.Total()
	if got, want := total.Calls, int64(4); got != want {
		t.Errorf("Total().Calls = %d, want %d", got, want)
	}
	if got, want := total.Usage.InputTokens, 3*7+1000; got != want {
		t.Errorf("Total() input tokens = %d, want %d", got, want)
	}
	if got, want := total.Usage.OutputTokens, 3*3+100; got != want {
		t.Errorf("Total() output tokens = %d, want %d", got, want)
	}

	// And the projection agrees with the counters it is projecting
	// from. Two tallies of one fact drift; this is the assertion that
	// says they have not.
	st := p.Stats()
	if sp.Decision.Calls != st.ModelCalls || sp.Improvisation.Calls != st.ImprovCalls {
		t.Errorf("Spend disagrees with Stats: %+v vs ModelCalls=%d ImprovCalls=%d",
			sp, st.ModelCalls, st.ImprovCalls)
	}
	if sp.Decision.Usage.InputTokens != st.Usage.InputTokens ||
		sp.Improvisation.Usage.InputTokens != st.ImprovUsage.InputTokens {
		t.Errorf("Spend's tokens disagree with Stats': %+v vs %+v / %+v", sp, st.Usage, st.ImprovUsage)
	}
}

// A call that FAILED still cost money, and the measurement says so.
//
// This is the one place a spend number could quietly flatter a
// deployment: a self-hosted model that times out on every window
// looks, to a report counting only successful answers, like a seat
// that never dialled at all — while the bill says otherwise.
func TestAFailedCallIsStillSpend(t *testing.T) {
	p := testPolicy(t, AlwaysText("this is not an index"), &stubB{index: 0, reason: "stub"}, nil)
	if _, err := p.Decide(context.Background(), castWindow()); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	sp := p.Spend()
	if sp.Decision.Calls != 1 {
		t.Errorf("a malformed reply was not counted as a call: %+v", sp)
	}
	if sp.Decision.Usage.InputTokens == 0 {
		t.Errorf("a malformed reply reported no input tokens: %+v", sp)
	}
}

// A seat with no transport spends nothing, and is still a Spender.
// The zero is the true answer for the whole no-endpoint deployment.
func TestASeatWithNoTransportSpendsNothing(t *testing.T) {
	p := testPolicy(t, nil, &stubB{index: 0, reason: "stub"}, nil)
	if _, err := p.Decide(context.Background(), castWindow()); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	var sp aiseat.Spender = p
	if !sp.Spend().Empty() {
		t.Errorf("a clientless seat reported spend: %+v", sp.Spend())
	}
}
