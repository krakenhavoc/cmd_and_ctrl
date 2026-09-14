package tiers_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
)

func botSeat(tier, deck string) aiseat.SeatSpec {
	return aiseat.SeatSpec{PlayerID: uuid.New(), Tier: tier, Deck: deck}
}

// The two free tiers need nothing but the binary; the two model tiers
// need a transport, and without one they are reported UNAVAILABLE
// rather than quietly offered as Layer A + B under a stronger name.
func TestTierStatusTracksTheTransport(t *testing.T) {
	bare := tiers.NewFactory(tiers.FactoryOptions{})
	for _, tier := range []aiseat.Tier{aiseat.TierRandom, aiseat.TierHeuristic} {
		if st := bare.TierStatus(tier); !st.Available {
			t.Errorf("%s unavailable with no transport: %q", tier, st.Reason)
		}
	}
	for _, tier := range []aiseat.Tier{aiseat.TierAssisted, aiseat.TierStrong} {
		st := bare.TierStatus(tier)
		if st.Available {
			t.Errorf("%s reported available with no model transport — that seat would play the heuristic under a model tier's name", tier)
		}
		// The reason has to name the fix. An operator reading the
		// picker should not have to read this package to enable it.
		if !strings.Contains(st.Reason, "CMDCTRL_OPENAI_ENDPOINT") || !strings.Contains(st.Reason, "CMDCTRL_ANTHROPIC_API_KEY") {
			t.Errorf("%s reason does not name both ways to configure a model: %q", tier, st.Reason)
		}
	}

	wired := tiers.NewFactory(tiers.FactoryOptions{Client: model.AlwaysIndex(0)})
	for _, tier := range []aiseat.Tier{aiseat.TierRandom, aiseat.TierHeuristic, aiseat.TierAssisted, aiseat.TierStrong} {
		if st := wired.TierStatus(tier); !st.Available {
			t.Errorf("%s unavailable with a transport configured: %q", tier, st.Reason)
		}
	}
	if st := wired.TierStatus("galaxy-brain"); st.Available {
		t.Error("a string that is not a tier reported available")
	}
}

// NewPolicy builds the policy the tier names, and refuses the ones it
// cannot build. It never substitutes.
func TestNewPolicyBuildsTheNamedTier(t *testing.T) {
	f := tiers.NewFactory(tiers.FactoryOptions{Client: model.AlwaysIndex(0)})
	for tier, want := range map[string]string{
		"random": "random", "heuristic": "heuristic",
		"assisted": "assisted", "strong": "strong",
	} {
		p, err := f.NewPolicy(botSeat(tier, ""))
		if err != nil {
			t.Fatalf("%s: %v", tier, err)
		}
		if got := p.Name(); got != want {
			t.Errorf("tier %q built the %q policy", tier, got)
		}
	}

	bare := tiers.NewFactory(tiers.FactoryOptions{})
	for _, tier := range []string{"assisted", "strong"} {
		p, err := bare.NewPolicy(botSeat(tier, ""))
		if !errors.Is(err, aiseat.ErrTierUnavailable) {
			t.Errorf("%s with no transport: err = %v, want ErrTierUnavailable", tier, err)
		}
		if p != nil {
			t.Errorf("%s with no transport returned the %q policy — an unavailable tier must not fall back", tier, p.Name())
		}
	}
	if _, err := bare.NewPolicy(botSeat("galaxy-brain", "")); !errors.Is(err, aiseat.ErrUnknownTier) {
		t.Errorf("unknown tier: err = %v, want ErrUnknownTier", err)
	}
}

// The seat's deck ID is resolved to a prompt profile, and only for
// the tiers that have a prompt.
func TestNewPolicyResolvesTheSeatsDeckProfile(t *testing.T) {
	var asked []string
	f := tiers.NewFactory(tiers.FactoryOptions{
		Client: model.AlwaysIndex(0),
		DeckProfile: func(id string) (model.DeckProfile, bool) {
			asked = append(asked, id)
			return model.DeckProfile{Name: "Raid and Ransack"}, true
		},
	})

	if _, err := f.NewPolicy(botSeat("assisted", "izzet-aggro")); err != nil {
		t.Fatal(err)
	}
	if len(asked) != 1 || asked[0] != "izzet-aggro" {
		t.Fatalf("deck profile lookups: %v, want [izzet-aggro]", asked)
	}
	// A free tier has no prompt, so it must not go looking for one,
	// and a seat with a pasted decklist has no ID to look up.
	if _, err := f.NewPolicy(botSeat("heuristic", "izzet-aggro")); err != nil {
		t.Fatal(err)
	}
	if _, err := f.NewPolicy(botSeat("assisted", "")); err != nil {
		t.Fatal(err)
	}
	if len(asked) != 1 {
		t.Errorf("deck profile lookups: %v, want the model seat's alone", asked)
	}
}

// The think deadline is a deployment decision once a self-hosted
// model is in play, and it has to reach the RUNNER — a policy allowed
// to think for 30s inside a runner that gives up at 2s has not been
// given anything.
func TestMaxThinkOverrideReachesTheRunnerConfig(t *testing.T) {
	f := tiers.NewFactory(tiers.FactoryOptions{Client: model.AlwaysIndex(0), MaxThink: 30 * time.Second})
	if got := f.RunnerConfig(aiseat.TierAssisted).MaxThink; got != 30*time.Second {
		t.Errorf("assisted MaxThink = %s, want 30s", got)
	}
	if got := f.RunnerConfig(aiseat.TierStrong).MaxThink; got != 30*time.Second {
		t.Errorf("strong MaxThink = %s, want 30s", got)
	}
	// The tiers that never call a model keep ADR 0033's pacing: the
	// override is there to wait for a slow model, and there is no
	// model to wait for.
	if got := f.RunnerConfig(aiseat.TierHeuristic).MaxThink; got != aiseat.ConfigFor(aiseat.TierHeuristic).MaxThink {
		t.Errorf("heuristic MaxThink = %s, want the tier default", got)
	}
	// A value smaller than the tier's own default is ignored rather
	// than applied: the override exists to give a slow model more
	// room, never to make the table wait less than the tier promises.
	short := tiers.NewFactory(tiers.FactoryOptions{Client: model.AlwaysIndex(0), MaxThink: time.Millisecond})
	if got := short.RunnerConfig(aiseat.TierAssisted).MaxThink; got != aiseat.ConfigFor(aiseat.TierAssisted).MaxThink {
		t.Errorf("assisted MaxThink = %s, want the tier default", got)
	}
}

// One local model in both funnel slots is a configuration, not a
// mistake.
func TestSingleModelConfigurationIsExpressible(t *testing.T) {
	f := tiers.NewFactory(tiers.FactoryOptions{
		Client: model.AlwaysIndex(0),
		Models: tiers.Models{Routine: "qwen3:8b"},
	})
	got := f.Models()
	if got.Routine != "qwen3:8b" || got.Frontier != "qwen3:8b" {
		t.Fatalf("models = %+v, want both slots on qwen3:8b", got)
	}
	if !got.SingleModel() {
		t.Error("SingleModel() is false for one model in both slots")
	}
	split := tiers.Models{Routine: "small", Frontier: "big"}
	if split.SingleModel() {
		t.Error("SingleModel() is true for two different models")
	}
}

// The model id a seat actually asks for is the configured one. This
// is the difference between a working local deployment and a 404 per
// window, and the only place it can be checked offline is the request
// the transport was handed.
func TestConfiguredModelIDReachesTheRequest(t *testing.T) {
	fake := &model.FakeClient{
		Label: "recorder",
		Reply: func(int, model.Request) (model.Response, error) {
			return model.Response{Text: `{"index": 0, "why": "ok"}`}, nil
		},
	}
	p, err := tiers.New(tiers.Assisted, tiers.Options{
		Client: fake,
		Models: tiers.Models{Routine: "llama3.1:8b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := p.Decide(ctx, window()); err != nil {
		t.Fatal(err)
	}
	reqs := fake.Requests()
	if len(reqs) != 1 {
		t.Fatalf("requests: %d, want 1", len(reqs))
	}
	if reqs[0].Model != "llama3.1:8b" {
		t.Errorf("asked for model %q, want llama3.1:8b", reqs[0].Model)
	}
}
