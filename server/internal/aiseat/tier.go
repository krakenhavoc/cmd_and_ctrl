package aiseat

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

// Tier is the difficulty slider (ADR 0033 §6). A tier names a policy
// stack; the wire value is the lowercase string.
//
// All four names are declared here rather than one per sub-PR so the
// API shape is settled up front: the picker, the `tier` field on POST
// /games/{id}/seats/bot and `bot_tier` on PlayerView do not change
// when a policy lands.
//
// WHICH of them a given server can actually seat is not decided here,
// and cannot be. Every policy package imports aiseat (for Input,
// Decision, Decline), so aiseat importing a policy package back is an
// import cycle — the factory that builds all four lives in
// aiseat/tiers, above them all, and is INJECTED into the Manager from
// main.go (Manager.SetPolicyFactory). The builtin factory in this
// file can only build `random`, which is what a build with nothing
// injected honestly reports.
//
// The refusal is the point, at every layer. A bot silently downgraded
// to random while the seat label says "strong" is worse than no bot
// at all — and so is one downgraded to heuristic while the label says
// "assisted". An unbuildable tier reports Available:false with a
// reason, and asking for it anyway is an error rather than a weaker
// policy under a stronger name.
type Tier string

const (
	// TierRandom picks uniformly among the legal moves. Shipped with
	// sub-PR 3: the engine fuzzer, and the baseline every other
	// policy is measured against.
	TierRandom Tier = "random"
	// TierHeuristic is ADR 0033's layers A+B — sub-PR 6.
	TierHeuristic Tier = "heuristic"
	// TierAssisted is layers A+B+C — sub-PR 7.
	TierAssisted Tier = "assisted"
	// TierStrong is A+C with wider candidates and a 1-ply sim.
	TierStrong Tier = "strong"
)

var (
	// ErrTierUnavailable is returned for a tier this build declares
	// but this server cannot play — because no policy factory is
	// wired in, or because the factory is missing something the tier
	// needs (a model client, say). Kept distinct from ErrUnknownTier
	// so the HTTP layer can tell "that is not a tier" from "not on
	// this server".
	ErrTierUnavailable = errors.New("aiseat: tier is not available in this build")
	// ErrUnknownTier is returned for a string that is not one of the
	// declared tier names at all.
	ErrUnknownTier = errors.New("aiseat: unknown tier")
)

// TierInfo describes one tier to a picker.
type TierInfo struct {
	Tier Tier `json:"tier"`
	// Label is the short human name for the option.
	Label string `json:"label"`
	// Description is one sentence on how it plays.
	Description string `json:"description"`
	// Available reports whether THIS server can seat a bot at this
	// tier. Unavailable tiers are still listed so the picker can grey
	// them out and say why, rather than pretend the difficulty slider
	// has one notch.
	Available bool `json:"available"`
	// Reason is why an unavailable tier is unavailable, in one line a
	// player can read ("needs a model API key…"). Empty when
	// Available: there is nothing to explain.
	Reason string `json:"reason,omitempty"`
}

// TierStatus is a factory's verdict on one tier: whether this server
// can build it, and if not, why not.
type TierStatus struct {
	// Available is whether NewPolicy will return a policy for it.
	Available bool
	// Reason is the one-line explanation a player sees when it will
	// not. Required when Available is false — an unexplained grey row
	// in the picker is a bug report waiting to happen.
	Reason string
}

// PolicyFactory builds the policy for one bot seat and reports which
// tiers it can build.
//
// It is an injection point rather than a function in this package for
// the import-cycle reason in the Tier doc above: aiseat cannot name
// the packages that implement the policies. main.go constructs
// aiseat/tiers' factory — which can build all four — and hands it to
// Manager.SetPolicyFactory. A Manager with nothing injected falls
// back to builtinFactory and offers `random` alone.
//
// Implementations must be safe for concurrent use: the Manager builds
// one policy per seat and several games may start at once.
type PolicyFactory interface {
	// NewPolicy builds a fresh policy for one seat. It must NOT
	// substitute a weaker tier for one it cannot build: return an
	// error wrapping ErrTierUnavailable instead.
	NewPolicy(seat SeatSpec) (Policy, error)
	// TierStatus reports whether NewPolicy would succeed for t.
	// Callers use it to grey out the picker, so it has to agree with
	// NewPolicy: a tier reported available and then refused is the
	// same lie as a downgrade, told later.
	TierStatus(t Tier) TierStatus
	// RunnerConfig is the pacing a runner gets for t. The factory
	// owns it rather than ConfigFor because a deployment may have
	// raised the think deadline (a self-hosted model is slower than
	// a hosted one), and the runner's deadline and the policy's own
	// call budget have to be raised by the same amount or the seat
	// spends the extra time failing.
	RunnerConfig(t Tier) Config
}

// builtinFactory is what a Manager uses when nothing is injected: the
// random policy and nothing else. It is what `cmd/server` would offer
// if the wiring in main.go were deleted, and it is what the
// package-level Tiers / AvailableTiers / NewPolicy helpers describe.
type builtinFactory struct{}

func (builtinFactory) NewPolicy(seat SeatSpec) (Policy, error) {
	t, ok := LookupTier(seat.Tier)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownTier, seat.Tier)
	}
	return NewPolicy(t)
}

func (builtinFactory) RunnerConfig(t Tier) Config { return ConfigFor(t) }

func (builtinFactory) TierStatus(t Tier) TierStatus {
	switch t {
	case TierRandom:
		return TierStatus{Available: true}
	case TierHeuristic, TierAssisted, TierStrong:
		return TierStatus{Reason: "this server has no bot policy factory wired in; only `random` can be seated"}
	default:
		return TierStatus{Reason: fmt.Sprintf("%q is not a tier", t)}
	}
}

// tierCatalog is the declared order, weakest-first, which is also the
// order the picker renders. Availability is deliberately NOT a field
// of the literal: it is a property of the server's factory, filled in
// by tiersFor.
var tierCatalog = []TierInfo{
	{
		Tier:        TierRandom,
		Label:       "Random",
		Description: "Picks uniformly among its legal moves. Not an opponent — a warm body, and the engine's fuzzer.",
	},
	{
		Tier:        TierHeuristic,
		Label:       "Heuristic",
		Description: "Scores the board and plays the best move it can see. Free, fast and deterministic.",
	},
	{
		Tier:        TierAssisted,
		Label:       "Assisted",
		Description: "Heuristic play with a model consulted on the close calls. Needs a model API key on the server.",
	},
	{
		Tier:        TierStrong,
		Label:       "Strong",
		Description: "Every close call goes to the frontier model with a wider candidate list, and a longer think to pay for it.",
	},
}

// tiersFor returns the declared catalog with each entry's
// availability (and, when unavailable, its reason) answered by f.
func tiersFor(f PolicyFactory) []TierInfo {
	if f == nil {
		f = builtinFactory{}
	}
	out := append([]TierInfo(nil), tierCatalog...)
	for i := range out {
		st := f.TierStatus(out[i].Tier)
		out[i].Available = st.Available
		if st.Available {
			out[i].Reason = ""
		} else {
			out[i].Reason = st.Reason
		}
	}
	return out
}

// Tiers returns the declared tiers in picker order as the BUILTIN
// factory sees them: every name, with `random` available. A server
// that injected a factory should ask its Manager (Manager.TierInfo)
// instead — that is the one that knows what this deployment can play.
func Tiers() []TierInfo {
	return tiersFor(builtinFactory{})
}

// AvailableTiers returns the tiers the builtin factory can seat. Same
// caveat as Tiers: Manager.Tiers is the deployment's answer.
func AvailableTiers() []Tier {
	var out []Tier
	for _, t := range tiersFor(builtinFactory{}) {
		if t.Available {
			out = append(out, t.Tier)
		}
	}
	return out
}

// LookupTier resolves a wire string to a declared Tier.
func LookupTier(s string) (Tier, bool) {
	for _, t := range tierCatalog {
		if string(t.Tier) == s {
			return t.Tier, true
		}
	}
	return "", false
}

// NewPolicy builds a fresh policy for a tier out of nothing but this
// package, which means `random` and nothing else. Every other
// declared tier returns ErrTierUnavailable, because the policy that
// implements it lives in a package aiseat may not import — see
// PolicyFactory. An unrecognised string returns ErrUnknownTier.
// Neither falls back to random.
func NewPolicy(t Tier) (Policy, error) {
	switch t {
	case TierRandom:
		return NewRandomPolicy(rand.NewPCG(rand.Uint64(), rand.Uint64())), nil
	case TierHeuristic, TierAssisted, TierStrong:
		return nil, fmt.Errorf("%w: %q", ErrTierUnavailable, t)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownTier, t)
	}
}

// ConfigFor is the pacing a tier gets. ADR 0033 §10: 2s of think for
// everything up to assisted, 5s for strong, because the table never
// waits on a model for longer than that. Kept in step with
// tiers.Tier.RunnerConfig, which says the same thing one layer up.
func ConfigFor(t Tier) Config {
	cfg := DefaultConfig()
	if t == TierStrong {
		cfg.MaxThink = 5 * time.Second
	}
	return cfg
}
