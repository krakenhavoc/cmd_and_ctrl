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
//
// Those numbers were sized for a HOSTED cheap model. A model running
// on the deployment's own hardware is routinely slower than either,
// and a model tier that misses its deadline on every window is Layer
// B wearing a stronger name — the downgrade aiseat/tier.go forbids.
// So the deadline is overridable per deployment (Options.MaxThink,
// from CMDCTRL_BOT_MAX_THINK); this is the default, not a ceiling.
func (t Tier) MaxThink() time.Duration {
	if t == Strong {
		return 5 * time.Second
	}
	return 2 * time.Second
}

// RunnerConfig is the production runner pacing for this tier:
// DefaultConfig with the tier's own MaxThink.
func (t Tier) RunnerConfig() aiseat.Config {
	return t.RunnerConfigWith(0)
}

// RunnerConfigWith is RunnerConfig with the deployment's own think
// deadline. Zero, or a value below the tier's default, takes the
// default: the override exists to give a slow local model MORE room,
// never to make the table wait less than the tier promises.
func (t Tier) RunnerConfigWith(maxThink time.Duration) aiseat.Config {
	cfg := aiseat.DefaultConfig()
	cfg.MaxThink = t.MaxThink()
	if t.NeedsModel() && maxThink > cfg.MaxThink {
		cfg.MaxThink = maxThink
	}
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
	// Models overrides the two model ids the funnel asks for. Empty
	// fields keep the shipped defaults.
	Models Models
	// MaxThink overrides the model tiers' hard deadline — see
	// Tier.MaxThink. Zero takes the tier's default. It also widens
	// the funnel's per-call budget, because a deadline the call is
	// not allowed to use is not a deadline that was raised.
	MaxThink time.Duration
	// NoImprovise turns ADR 0033 §8 improvisation OFF for the model
	// tiers, which have it on by default since #686.
	//
	// Spelled as a negative so that the zero Options keeps the tier's
	// own default. A seat with it set is a complete seat that never
	// improvises: it casts an uncatalogued card, the card resolves
	// into silence, and a human applies the text by hand — which is
	// what every tier did before #686 and is a perfectly reasonable
	// posture for a table that would rather not have a model touching
	// the board.
	NoImprovise bool
	// Config overrides the model funnel's tuning. Nil takes the
	// tier's default.
	Config *model.Config
}

// Models names the model ids for the two funnel slots.
//
// Pointing BOTH at the same id is a legitimate configuration, not a
// mistake, and the code treats it as one. The cheap/frontier split is
// an economy: it exists so that routine windows do not pay frontier
// prices. A deployment running one local model has no prices to pay
// and nothing to split, and the escalation still earns its keep —
// what an escalated window buys there is the wider candidate list and
// `strong`'s longer think, not a different model.
type Models struct {
	// Routine answers the ordinary windows. Empty keeps the default.
	Routine string
	// Frontier answers the escalated ones. Empty falls back to
	// Routine when Routine is set, and to the default otherwise.
	Frontier string
}

// resolve fills Frontier from Routine, so that setting one id is all
// a single-model deployment has to do.
func (m Models) resolve() Models {
	if m.Frontier == "" {
		m.Frontier = m.Routine
	}
	return m
}

// SingleModel reports whether both slots name the same model — true
// for the one-local-model deployment. Callers use it to log the
// collapsed funnel once at boot instead of leaving an operator to
// wonder why "frontier" and "routine" are the same line.
func (m Models) SingleModel() bool {
	r := m.resolve()
	return r.Routine != "" && r.Routine == r.Frontier
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
		if ids := opt.Models.resolve(); ids.Routine != "" {
			cfg.Routine.ID = ids.Routine
			cfg.Frontier.ID = ids.Frontier
			// Improvisation asks the frontier slot (ADR 0033 §8,
			// amended #686), so a deployment that named its own
			// models must not be left dialling the shipped default
			// for the one call that reaches an opponent's board.
			cfg.Improv.ID = ids.Frontier
		}
		if opt.NoImprovise {
			cfg.Improvise = false
		}
		if think := opt.MaxThink; think > t.MaxThink() {
			// The runner's deadline moved, so the call budget inside
			// it has to move too — everything less the reserve the
			// funnel keeps to come back with Layer B's answer.
			reserve := cfg.Reserve
			if reserve <= 0 {
				reserve = 250 * time.Millisecond
			}
			if call := think - reserve; call > cfg.MaxCall {
				cfg.MaxCall = call
			}
		}
		if cfg.Fallback == nil {
			cfg.Fallback = heuristic.New()
		}
		return model.New(cfg), nil
	}
	return nil, fmt.Errorf("aiseat/tiers: unknown tier %q", t)
}
