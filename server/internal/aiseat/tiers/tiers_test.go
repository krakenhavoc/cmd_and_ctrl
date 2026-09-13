package tiers_test

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

var seat = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func window() aiseat.Input {
	return aiseat.Input{
		Seat: seat,
		View: protocol.GameView{
			Seats: []protocol.PlayerView{{ID: seat.String(), Name: "Bot0", Life: 40}},
			Turn:  protocol.TurnView{Number: 1, Step: "precombat_main"},
		},
		Moves: []legal.Move{
			{Kind: legal.KindPass, Label: "Pass priority", Player: seat},
			{Kind: legal.KindCast, Label: "Cast Bear", Player: seat, Params: []byte(`{"instance_id":"b"}`)},
		},
	}
}

func TestParse(t *testing.T) {
	for _, want := range tiers.All() {
		got, err := tiers.Parse(string(want))
		if err != nil || got != want {
			t.Errorf("Parse(%q) = %q, %v", want, got, err)
		}
	}
	if _, err := tiers.Parse("expert"); err == nil {
		t.Error("Parse accepted a tier that does not exist")
	}
	if _, err := tiers.Parse(""); err == nil {
		t.Error("Parse accepted an empty tier")
	}
}

// Every tier builds, names itself on the wire value, and answers a
// window without a model endpoint anywhere.
func TestEveryTierBuildsAndDecidesWithNoModel(t *testing.T) {
	for _, tier := range tiers.All() {
		t.Run(string(tier), func(t *testing.T) {
			p, err := tiers.New(tier, tiers.Options{Rand: rand.NewPCG(1, 2)})
			if err != nil {
				t.Fatalf("New(%q): %v", tier, err)
			}
			if p.Name() != string(tier) {
				t.Errorf("Name() = %q, want %q — the tier name reaches the lobby and the player", p.Name(), tier)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			d, err := p.Decide(ctx, window())
			if err != nil {
				t.Fatalf("Decide: %v", err)
			}
			if d.Index < 0 || d.Index >= len(window().Moves) {
				t.Errorf("index = %d, which is not a move", d.Index)
			}
		})
	}
}

func TestUnknownTierIsAnError(t *testing.T) {
	if _, err := tiers.New(tiers.Tier("expert"), tiers.Options{}); err == nil {
		t.Error("New built a tier that does not exist")
	}
}

// ADR 0033 §10 gives `strong` a longer deadline because it is buying
// a deeper frontier answer; everything else gets two seconds.
func TestMaxThinkPerTier(t *testing.T) {
	for _, tier := range []tiers.Tier{tiers.Random, tiers.Heuristic, tiers.Assisted} {
		if got := tier.MaxThink(); got != 2*time.Second {
			t.Errorf("%s MaxThink = %v, want 2s", tier, got)
		}
	}
	if got := tiers.Strong.MaxThink(); got != 5*time.Second {
		t.Errorf("strong MaxThink = %v, want 5s", got)
	}
	if got := tiers.Strong.RunnerConfig().MaxThink; got != 5*time.Second {
		t.Errorf("RunnerConfig MaxThink = %v, want 5s", got)
	}
	// The rest of the production pacing has to survive the override.
	if cfg := tiers.Assisted.RunnerConfig(); cfg.MinThink != aiseat.DefaultConfig().MinThink {
		t.Errorf("MinThink = %v, want the production pacing", cfg.MinThink)
	}
}

func TestNeedsModel(t *testing.T) {
	for tier, want := range map[tiers.Tier]bool{
		tiers.Random: false, tiers.Heuristic: false, tiers.Assisted: true, tiers.Strong: true,
	} {
		if got := tier.NeedsModel(); got != want {
			t.Errorf("%s.NeedsModel() = %v, want %v", tier, got, want)
		}
	}
}

// The random tier is the fuzzer and must NOT be wrapped in Layer A: a
// filter that answered the trivial windows correctly would narrow the
// input space the fuzzer exists to widen.
func TestRandomIsNotFiltered(t *testing.T) {
	var m rules.Meter
	p, err := tiers.New(tiers.Random, tiers.Options{Rand: rand.NewPCG(1, 2), Meter: &m})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		if _, err := p.Decide(context.Background(), window()); err != nil {
			t.Fatal(err)
		}
	}
	if st := m.Stats(); st.Windows != 0 {
		t.Errorf("the random tier went through the Layer A filter: %+v", st)
	}
}

// The heuristic tier IS "A + B" per ADR 0033 §6, so its windows have
// to reach the meter.
func TestHeuristicTierIsFiltered(t *testing.T) {
	var m rules.Meter
	p, err := tiers.New(tiers.Heuristic, tiers.Options{Meter: &m})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Decide(context.Background(), window()); err != nil {
		t.Fatal(err)
	}
	if st := m.Stats(); st.Windows != 1 {
		t.Errorf("meter = %+v, want one window", st)
	}
}

// A model tier with no Client keeps its name and plays on Layer A + B.
// This is the deployment with no API key, and it has to be a working
// bot rather than a broken one.
func TestModelTiersWithoutAClientAreCompletePolicies(t *testing.T) {
	for _, tier := range []tiers.Tier{tiers.Assisted, tiers.Strong} {
		p, err := tiers.New(tier, tiers.Options{})
		if err != nil {
			t.Fatal(err)
		}
		mp, ok := p.(*model.Policy)
		if !ok {
			t.Fatalf("%s is not a model policy: %T", tier, p)
		}
		if _, err := mp.Decide(context.Background(), window()); err != nil {
			t.Fatalf("%s: %v", tier, err)
		}
		if st := mp.Stats(); st.ByFallback[model.FallbackNoClient] != 1 {
			t.Errorf("%s: fallbacks = %v, want one %q", tier, st.ByFallback, model.FallbackNoClient)
		}
	}
}

// The model tiers route through the Client they were given, and the
// tier picks its own default config.
func TestModelTiersUseTheClient(t *testing.T) {
	fake := model.AlwaysIndex(1)
	p, err := tiers.New(tiers.Assisted, tiers.Options{Client: fake, Deck: model.DeckProfile{Name: "test"}})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Decide(context.Background(), window())
	if err != nil {
		t.Fatal(err)
	}
	if fake.Calls() != 1 {
		t.Errorf("calls = %d, want 1", fake.Calls())
	}
	if d.Index != 1 {
		t.Errorf("index = %d, want the model's 1", d.Index)
	}
	if reqs := fake.Requests(); len(reqs) == 0 || len(reqs[0].System) == 0 {
		t.Error("the deck profile never reached the prompt")
	}
}

// `strong` escalates every surviving window; `assisted` does not.
func TestStrongEscalatesWhereAssistedDoesNot(t *testing.T) {
	strongFake := model.AlwaysIndex(0)
	sp, _ := tiers.New(tiers.Strong, tiers.Options{Client: strongFake})
	if _, err := sp.Decide(context.Background(), window()); err != nil {
		t.Fatal(err)
	}
	reqs := strongFake.Requests()
	if len(reqs) != 1 {
		t.Fatalf("calls = %d", len(reqs))
	}
	if reqs[0].Model != model.StrongConfig().Frontier.ID {
		t.Errorf("strong asked %q, want the frontier model", reqs[0].Model)
	}

	assistedFake := model.AlwaysIndex(0)
	ap, _ := tiers.New(tiers.Assisted, tiers.Options{Client: assistedFake})
	if _, err := ap.Decide(context.Background(), window()); err != nil {
		t.Fatal(err)
	}
	if r := assistedFake.Requests(); len(r) != 1 || r[0].Model != model.DefaultConfig().Routine.ID {
		t.Errorf("assisted asked %+v, want the routine model on a quiet window", r)
	}
}

// An explicit Config override still gets the tier's name and the
// caller's client, so the lobby cannot accidentally build an
// `assisted` seat that reports itself as something else.
func TestConfigOverrideKeepsTheTierIdentity(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Tier = "whatever"
	cfg.MaxCandidates = 3
	p, err := tiers.New(tiers.Strong, tiers.Options{Config: &cfg, Client: model.AlwaysIndex(0)})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != string(tiers.Strong) {
		t.Errorf("Name() = %q, want %q", p.Name(), tiers.Strong)
	}
}
