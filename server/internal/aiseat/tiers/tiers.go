// Package tiers builds the four bot difficulties ADR 0033 §6 defines
// and maps each one to the runner pacing it needs.
//
//	| Tier        | Layers                       | Cost |
//	| random      | uniform over Moves           | none |
//	| heuristic   | A + B                        | none |
//	| assisted    | A + B + C                    | low  |
//	| strong      | A + C, wider candidates      | high |
//
// It exists as its own package for a mechanical reason worth writing
// down, because it looks like over-decomposition until you try the
// obvious thing: the constructor cannot live in `aiseat`. Every
// policy package imports aiseat (for Input, Decision, Decline), so
// aiseat importing a policy package back is an import cycle. The
// factory has to sit above all of them, and this is it.
//
// Like every other package under aiseat/, it may not import
// internal/game.
package tiers

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
)

// Tier is a bot difficulty. The string is the wire value: it reaches
// the lobby API as `{tier}` and the client as PlayerView.bot_tier.
type Tier string

// The four tiers.
const (
	// Random picks uniformly. Not a joke tier — a four-Random table
	// playing unattended is the cheapest rules-engine fuzzer this
	// project will ever get.
	Random Tier = "random"
	// Heuristic is Layer A plus the rule-based scorer. Free,
	// deterministic, and the fallback under every model failure.
	Heuristic Tier = "heuristic"
	// Assisted is the default: Layer A, the heuristic, and a model on
	// the windows that survive both.
	Assisted Tier = "assisted"
	// Strong escalates every surviving window to the frontier model
	// with a wider candidate list, and pays for it with a longer
	// think.
	Strong Tier = "strong"
)

// All returns the tiers in difficulty order.
func All() []Tier { return []Tier{Random, Heuristic, Assisted, Strong} }

// Parse validates a wire value.
func Parse(s string) (Tier, error) {
	for _, t := range All() {
		if string(t) == s {
			return t, nil
		}
	}
	names := make([]string, 0, 4)
	for _, t := range All() {
		names = append(names, string(t))
	}
	sort.Strings(names)
	return "", fmt.Errorf("aiseat/tiers: unknown tier %q (want one of %v)", s, names)
}

// NeedsModel reports whether the tier will try to call a model. A
// tier that needs one still WORKS without one — it degrades to
// Layer A + B — so this is a signal for the lobby ("this table will
// cost money"), not a precondition.
func (t Tier) NeedsModel() bool { return t == Assisted || t == Strong }

// MaxThink is the hard deadline ADR 0033 §10 gives this tier: 2s for
// everything up to `assisted`, 5s for `strong`, which is buying a
// deeper frontier answer and has to pay for it in wall clock.
func (t Tier) MaxThink() time.Duration {
	if t == Strong {
		return 5 * time.Second
	}
	return 2 * time.Second
}

// RunnerConfig is the production runner pacing for this tier:
// DefaultConfig with the tier's own MaxThink.
func (t Tier) RunnerConfig() aiseat.Config {
	cfg := aiseat.DefaultConfig()
	cfg.MaxThink = t.MaxThink()
	return cfg
}

// Options are the per-seat inputs a tier needs. The zero value builds
// working `random` and `heuristic` seats; `assisted` and `strong`
// want a Client and a Deck, and run as Layer A + B without them.
type Options struct {
	// Rand seeds the random tier. Nil uses the process generator.
	Rand rand.Source
	// Meter collects Layer A absorption. Share one across a table so
	// the rate describes the game. Optional.
	Meter *rules.Meter
	// Client is the model transport for the model tiers. Nil is
	// legal and means "no model on this VPS": the seat plays on
	// Layer A + B under the tier's own name.
	Client model.Client
	// Deck is the static, prompt-cached half of the model prompt.
	Deck model.DeckProfile
	// Config overrides the model funnel's tuning. Nil takes the
	// tier's default.
	Config *model.Config
}

// New builds a policy for the tier.
func New(t Tier, opt Options) (aiseat.Policy, error) {
	switch t {
	case Random:
		// Deliberately NOT wrapped in the Layer A filter. The random
		// tier's job is to hit the engine with legal moves nobody
		// sensible would make; a filter that answered the trivial
		// windows "correctly" would narrow the fuzzer for no gain.
		return aiseat.NewRandomPolicy(opt.Rand), nil

	case Heuristic:
		f := rules.NewFilter(heuristic.New(), opt.Meter)
		f.Tier = string(Heuristic)
		return f, nil

	case Assisted, Strong:
		cfg := model.DefaultConfig()
		if t == Strong {
			cfg = model.StrongConfig()
		}
		if opt.Config != nil {
			cfg = *opt.Config
			cfg.Tier = string(t)
		}
		cfg.Client = opt.Client
		cfg.Deck = opt.Deck
		cfg.Meter = opt.Meter
		if cfg.Fallback == nil {
			cfg.Fallback = heuristic.New()
		}
		return model.New(cfg), nil
	}
	return nil, fmt.Errorf("aiseat/tiers: unknown tier %q", t)
}
