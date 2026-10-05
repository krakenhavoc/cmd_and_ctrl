package model

import (
	"context"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// totals_test.go is ADR 0123 §11's check on the model half of §3's bot
// totals: one FakeClient-driven decision moves the process-wide
// cmdctrl_bot_model_* families exactly as the funnel's own Stats move,
// with no network.

func botValue(t *testing.T, name string, labels metricstest.L) float64 {
	t.Helper()
	return metricstest.Value(t, name, labels)
}

// usageClient answers index 1 and reports a fixed usage, Anthropic
// cache tokens included.
func usageClient() *FakeClient {
	return &FakeClient{Reply: func(int, Request) (Response, error) {
		return Response{
			Text:  `{"index": 1, "why": "fake"}`,
			Usage: Usage{InputTokens: 100, OutputTokens: 7, CacheReadTokens: 20, CacheWriteTokens: 3},
		}, nil
	}}
}

// A window the model answers: one ok call, one call observed, its
// tokens, no fallback, and the layer noted as C for the runner.
func TestAModelDecisionMovesTheTotals(t *testing.T) {
	b := &stubB{index: 2, reason: "stub"}
	p := testPolicy(t, usageClient(), b, nil) // DefaultConfig: tier assisted
	tier := metrics.BotTierAssisted

	calls := botValue(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier, "result": "ok"})
	seconds := botValue(t, "cmdctrl_bot_model_call_seconds", metricstest.L{"tier": tier})
	prompt := botValue(t, "cmdctrl_bot_model_tokens_total", metricstest.L{"tier": tier, "direction": "prompt"})
	completion := botValue(t, "cmdctrl_bot_model_tokens_total", metricstest.L{"tier": tier, "direction": "completion"})
	fallbacks := botValue(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": tier})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctx, noted := aiseat.WithLayerNote(ctx)
	if _, err := p.Decide(ctx, castWindow()); err != nil {
		t.Fatal(err)
	}

	if got := noted(); got != LayerC {
		t.Errorf("noted layer = %q, want C", got)
	}
	if got := botValue(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier, "result": "ok"}) - calls; got != 1 {
		t.Errorf("ok calls rose by %v, want 1", got)
	}
	if got := botValue(t, "cmdctrl_bot_model_call_seconds", metricstest.L{"tier": tier}) - seconds; got != 1 {
		t.Errorf("call seconds observed %v, want 1", got)
	}
	if got := botValue(t, "cmdctrl_bot_model_tokens_total", metricstest.L{"tier": tier, "direction": "prompt"}) - prompt; got != 123 {
		t.Errorf("prompt tokens rose by %v, want 123 (input plus cache reads and writes)", got)
	}
	if got := botValue(t, "cmdctrl_bot_model_tokens_total", metricstest.L{"tier": tier, "direction": "completion"}) - completion; got != 7 {
		t.Errorf("completion tokens rose by %v, want 7", got)
	}
	if got := botValue(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": tier}) - fallbacks; got != 0 {
		t.Errorf("fallbacks rose by %v on an answered window, want 0", got)
	}
}

// Each way a call can end lands on its result, and each fallback on its
// model- cause, once. No client and no budget are fallbacks without a
// call.
func TestEveryModelOutcomeIsCountedOnce(t *testing.T) {
	cases := []struct {
		name     string
		client   Client
		tweak    func(*Config)
		budget   time.Duration
		result   string // "" when no call is made
		cause    string
		attempts float64
	}{
		{"error", AlwaysFails(ErrOutage), nil, 2 * time.Second, "error", "model-error", 1},
		{"malformed", AlwaysText("cast the bear"), nil, 2 * time.Second, "malformed", "model-malformed", 1},
		{"out of range", AlwaysIndex(99), nil, 2 * time.Second, "out_of_range", "model-out-of-range", 1},
		{"timeout", NeverAnswers(), func(c *Config) {
			c.Reserve = 100 * time.Millisecond
			c.MinBudget = 10 * time.Millisecond
		}, 400 * time.Millisecond, "timeout", "model-error", 1},
		{"no client", nil, nil, 2 * time.Second, "", "model-no-client", 0},
		{"no budget", AlwaysIndex(1), func(c *Config) {
			c.Reserve = 250 * time.Millisecond
			c.MinBudget = 200 * time.Millisecond
		}, 300 * time.Millisecond, "", "model-no-budget", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tier := metrics.BotTierStrong
			p := testPolicy(t, c.client, &stubB{index: 2, reason: "stub"}, func(cfg *Config) {
				cfg.Tier = TierStrong
				if c.tweak != nil {
					c.tweak(cfg)
				}
			})
			calls := botValue(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier})
			result := botValue(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier, "result": c.result})
			cause := botValue(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": tier, "cause": c.cause})

			ctx, noted := aiseat.WithLayerNote(context.Background())
			ctx, cancel := context.WithTimeout(ctx, c.budget)
			defer cancel()
			if _, err := p.Decide(ctx, castWindow()); err != nil {
				t.Fatal(err)
			}

			if got := noted(); got != LayerB {
				t.Errorf("noted layer = %q, want B (the heuristic's answer was played)", got)
			}
			if got := botValue(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier}) - calls; got != c.attempts {
				t.Errorf("calls rose by %v, want %v", got, c.attempts)
			}
			if c.result != "" {
				if got := botValue(t, "cmdctrl_bot_model_calls_total", metricstest.L{"tier": tier, "result": c.result}) - result; got != 1 {
					t.Errorf("result %q rose by %v, want 1", c.result, got)
				}
			}
			if got := botValue(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": tier, "cause": c.cause}) - cause; got != 1 {
				t.Errorf("cause %q rose by %v, want 1", c.cause, got)
			}
		})
	}
	if err := metrics.CheckClosedLabels(metrics.Registry); err != nil {
		t.Error(err)
	}
}

// Every funnel fallback cause is its own value of the cause label,
// prefixed model-; none falls to "other".
func TestEveryFunnelFallbackCauseIsALabel(t *testing.T) {
	for _, c := range []string{
		FallbackNoClient, FallbackNoBudget, FallbackError, FallbackMalformed, FallbackOutOfRange, FallbackPolicyError,
	} {
		want := "model-" + c
		before := botValue(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": "other", "cause": want})
		metrics.RecordModelFallback("cause-check", c)
		if got := botValue(t, "cmdctrl_bot_fallbacks_total", metricstest.L{"tier": "other", "cause": want}) - before; got != 1 {
			t.Errorf("funnel cause %q did not land on %q (rose by %v)", c, want, got)
		}
	}
}

// Layer A's answer is noted as A, and costs no call.
func TestAnAbsorbedWindowIsNotedA(t *testing.T) {
	p := testPolicy(t, AlwaysIndex(1), &stubB{index: 0}, nil)
	in := aiseat.Input{Seat: meSeat, View: view(), Moves: castWindow().Moves[:1]}
	ctx, noted := aiseat.WithLayerNote(context.Background())
	if _, err := p.Decide(ctx, in); err != nil {
		t.Fatal(err)
	}
	if got := noted(); got != LayerA {
		t.Errorf("noted layer = %q, want A", got)
	}
}
