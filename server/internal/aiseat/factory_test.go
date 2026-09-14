package aiseat_test

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// factory_test.go covers the injection: a Manager takes its policies,
// its tier list and its pacing from a PolicyFactory handed down from
// main, because aiseat cannot build the policies itself without an
// import cycle.
//
// The test that matters most here is the LAST one, and it is worth
// saying why. A seat whose tier says `heuristic` and whose policy is
// actually the random one plays the whole game without a single
// visible symptom: it takes legal moves, it commits, it passes, it
// loses. Nothing on the wire carries the difference. That failure is
// exactly what shipping a placeholder factory produced for the whole
// of S31, so it gets an assertion rather than a comment.

// stubFactory is a PolicyFactory with a scripted availability table.
// It stands in for whatever main.go injects.
type stubFactory struct {
	available map[aiseat.Tier]bool
	reason    string
	built     []aiseat.SeatSpec
}

func (f *stubFactory) TierStatus(t aiseat.Tier) aiseat.TierStatus {
	if f.available[t] {
		return aiseat.TierStatus{Available: true}
	}
	return aiseat.TierStatus{Reason: f.reason}
}

func (f *stubFactory) RunnerConfig(t aiseat.Tier) aiseat.Config { return aiseat.ConfigFor(t) }

func (f *stubFactory) NewPolicy(seat aiseat.SeatSpec) (aiseat.Policy, error) {
	t, ok := aiseat.LookupTier(seat.Tier)
	if !ok {
		return nil, errors.New("unknown tier")
	}
	if !f.available[t] {
		return nil, aiseat.ErrTierUnavailable
	}
	f.built = append(f.built, seat)
	return aiseat.NewPolicy(aiseat.TierRandom)
}

func allTiers() map[aiseat.Tier]bool {
	return map[aiseat.Tier]bool{
		aiseat.TierRandom: true, aiseat.TierHeuristic: true,
		aiseat.TierAssisted: true, aiseat.TierStrong: true,
	}
}

// With nothing injected a Manager offers what this package alone can
// build, which is `random` — and says so about the rest rather than
// leaving the picker to guess.
func TestManagerWithoutAFactoryOffersRandomOnly(t *testing.T) {
	m := aiseat.NewManager(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if got := m.Tiers(); len(got) != 1 || got[0] != string(aiseat.TierRandom) {
		t.Fatalf("tiers: %v, want [random]", got)
	}
	for _, info := range m.TierInfo() {
		if info.Tier == aiseat.TierRandom {
			continue
		}
		if info.Available {
			t.Errorf("%s reported available with no factory injected", info.Tier)
		}
		if info.Reason == "" {
			t.Errorf("%s is unavailable with no reason — the picker has nothing to show", info.Tier)
		}
	}
}

// Injecting a factory that can build all four makes all four
// available, which is the whole point of the injection.
func TestManagerTiersFollowTheInjectedFactory(t *testing.T) {
	m := aiseat.NewManager(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	m.SetPolicyFactory(&stubFactory{available: allTiers()})

	if got := len(m.Tiers()); got != 4 {
		t.Fatalf("tiers: %v, want all four", m.Tiers())
	}
	for _, info := range m.TierInfo() {
		if !info.Available {
			t.Errorf("%s unavailable although the factory builds it", info.Tier)
		}
		if info.Reason != "" {
			t.Errorf("%s is available but carries a reason %q", info.Tier, info.Reason)
		}
	}
}

// A factory that cannot build the model tiers must say so WITH a
// reason. "Greyed out and unexplained" is the state that produces a
// bug report instead of a configuration change.
func TestManagerReportsWhyAModelTierIsMissing(t *testing.T) {
	f := &stubFactory{
		available: map[aiseat.Tier]bool{aiseat.TierRandom: true, aiseat.TierHeuristic: true},
		reason:    "needs a model endpoint",
	}
	m := aiseat.NewManager(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	m.SetPolicyFactory(f)

	got := strings.Join(m.Tiers(), ",")
	if got != "random,heuristic" {
		t.Fatalf("tiers: %q, want random,heuristic", got)
	}
	for _, tier := range []string{string(aiseat.TierAssisted), string(aiseat.TierStrong)} {
		if reason := m.TierReason(tier); reason != "needs a model endpoint" {
			t.Errorf("TierReason(%q) = %q", tier, reason)
		}
	}
	if reason := m.TierReason(string(aiseat.TierRandom)); reason != "" {
		t.Errorf("an available tier carries a reason: %q", reason)
	}
}

// The seat's curated-deck ID reaches the factory. The model tiers
// build the static half of their prompt from it, and it is the one
// per-seat input that is not the tier itself.
func TestStartBotsPassesTheDeckToTheFactory(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	f := &stubFactory{available: allTiers()}
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	host := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, log)
	host.SetPolicyFactory(f)
	l.SetBotHost(host)

	meta, err := l.Create("Deck plumbing")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, _, err := l.AddBot(meta.ID, "Bot", string(aiseat.TierRandom), "izzet-aggro", "Raid and Ransack", monoRedDeck(uuid.Nil)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	defer host.Shutdown()

	if len(f.built) != 2 {
		t.Fatalf("factory built %d policies, want 2", len(f.built))
	}
	for _, seat := range f.built {
		if seat.Deck != "izzet-aggro" {
			t.Errorf("seat reached the factory with deck %q, want izzet-aggro", seat.Deck)
		}
	}
}

// THE ONE THAT MATTERS. A seat added at `heuristic` must be driven by
// the heuristic policy. This is invisible from the table — a random
// bot in a heuristic seat plays every window and commits every move —
// so the assertion is on the runner's own policy name.
func TestSeatsRunThePolicyTheirTierNames(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	host := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, log)
	// The production factory, with a fake transport standing in for
	// the model endpoint CI does not have.
	host.SetPolicyFactory(tiers.NewFactory(tiers.FactoryOptions{
		Client: model.AlwaysIndex(0),
	}))
	l.SetBotHost(host)

	want := map[string]string{
		string(aiseat.TierRandom):    "random",
		string(aiseat.TierHeuristic): "heuristic",
		string(aiseat.TierAssisted):  "assisted",
		string(aiseat.TierStrong):    "strong",
	}

	meta, err := l.Create("One of each")
	if err != nil {
		t.Fatal(err)
	}
	seats := map[uuid.UUID]string{}
	for _, tier := range []string{"random", "heuristic", "assisted", "strong"} {
		_, playerID, err := l.AddBot(meta.ID, "Bot "+tier, tier, "", "Mono Red", monoRedDeck(uuid.Nil))
		if err != nil {
			t.Fatalf("AddBot %s: %v", tier, err)
		}
		seats[playerID] = tier
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	defer host.Shutdown()

	runners := host.Runners(meta.ID)
	if len(runners) != 4 {
		t.Fatalf("runners: %d, want 4 — a tier the factory can build must not be skipped", len(runners))
	}
	for _, r := range runners {
		tier := seats[r.Seat()]
		if got := r.PolicyName(); got != want[tier] {
			t.Errorf("seat seated at %q is running the %q policy, want %q — a downgraded seat is invisible from the table",
				tier, got, want[tier])
		}
	}
}
