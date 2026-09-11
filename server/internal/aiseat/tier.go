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
// when a policy lands. Only TierRandom is buildable today —
// everything else reports Available:false and NewPolicy refuses it.
// The refusal is the point. A bot silently downgraded to random while
// the seat label says "strong" is worse than no bot at all.
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
	// ErrTierUnavailable is returned by NewPolicy for a tier this
	// build declares but cannot yet play. Kept distinct from
	// ErrUnknownTier so the HTTP layer can tell "that is not a tier"
	// from "not yet — it arrives with sub-PR 6".
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
	// Available reports whether a policy exists for it. Unavailable
	// tiers are still listed so the picker can grey them out and say
	// what is coming, rather than pretend the slider has one notch.
	Available bool `json:"available"`
}

// tierCatalog is the declared order, weakest-first, which is also the
// order the picker renders.
var tierCatalog = []TierInfo{
	{
		Tier:        TierRandom,
		Label:       "Random",
		Description: "Picks uniformly among its legal moves. Not an opponent — a warm body, and the engine's fuzzer.",
		Available:   true,
	},
	{
		Tier:        TierHeuristic,
		Label:       "Heuristic",
		Description: "Scores the board and plays the best move it can see. Arrives with S31 sub-PR 6.",
	},
	{
		Tier:        TierAssisted,
		Label:       "Assisted",
		Description: "Heuristic play with a model consulted on the close calls. Arrives with S31 sub-PR 7.",
	},
	{
		Tier:        TierStrong,
		Label:       "Strong",
		Description: "Wider candidate set and a one-ply simulation on the shortlist. Not yet built.",
	},
}

// Tiers returns the declared tiers in picker order, including the
// ones that are not available yet.
func Tiers() []TierInfo {
	return append([]TierInfo(nil), tierCatalog...)
}

// AvailableTiers returns only the tiers a bot can actually be seated
// with. This is what the lobby validates an add request against.
func AvailableTiers() []Tier {
	var out []Tier
	for _, t := range tierCatalog {
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

// NewPolicy builds a fresh policy for a tier. A declared-but-unbuilt
// tier returns ErrTierUnavailable; an unrecognised string returns
// ErrUnknownTier. Neither falls back to random — see the Tier doc.
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
// waits on a model for longer than that.
func ConfigFor(t Tier) Config {
	cfg := DefaultConfig()
	if t == TierStrong {
		cfg.MaxThink = 5 * time.Second
	}
	return cfg
}
