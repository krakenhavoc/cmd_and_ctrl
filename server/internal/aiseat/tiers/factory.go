package tiers

import (
	"fmt"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
)

// factory.go is the injection this package's doc comment describes,
// made concrete: an aiseat.PolicyFactory that aiseat.Manager can hold
// without aiseat importing any policy package.
//
// The direction of the dependency is the whole point. aiseat defines
// the interface and holds one; tiers implements it and imports
// aiseat; main.go — the only place above both — constructs the
// implementation and hands it down. Nothing in aiseat ever names
// heuristic, model or rules, so the cycle that would otherwise exist
// does not, and the Manager still seats all four tiers.
//
// What the Manager gets out of it, beyond a policy, is the truth
// about availability: the factory is the only thing that knows
// whether a model transport was configured, so it is the only thing
// that can honestly answer "can this server seat an `assisted` bot".

// Factory builds seat policies for aiseat.Manager. Construct one at
// boot with NewFactory and pass it to Manager.SetPolicyFactory. Safe
// for concurrent use: everything it holds is read-only after
// construction, and the Meter has its own lock.
type Factory struct {
	opt FactoryOptions
}

// DeckProfileFunc resolves a curated-deck ID to the static, prompt-
// cached half of a model prompt. Returning false means "no profile
// for that deck" — a supported case, not an error: the seat then
// plays with the board and the move list but no decklist in the
// prompt.
//
// It is a func rather than an interface, and lives here rather than
// in internal/decks, so that this package does not depend on the deck
// catalog (which depends on the card index, which depends on a
// Scryfall dump). main.go closes over deckprofile.Build and the
// loaded index and passes the result down.
type DeckProfileFunc func(deckID string) (model.DeckProfile, bool)

// FactoryOptions configure every seat this factory builds.
type FactoryOptions struct {
	// Client is the model transport — an Anthropic client, an
	// OpenAI-compatible one pointed at a local model, or nil.
	//
	// Nil is a supported deployment, and it is what decides the
	// availability of the two model tiers: see TierStatus.
	Client model.Client
	// Models overrides the funnel's two model ids. A single-model
	// deployment sets Routine alone.
	Models Models
	// MaxThink overrides the model tiers' hard deadline. Zero takes
	// each tier's default (ADR 0033 §10: 2s, 5s for `strong`).
	MaxThink time.Duration
	// DeckProfile resolves SeatSpec.Deck for the model tiers. Nil
	// means no seat gets a decklist in its prompt.
	DeckProfile DeckProfileFunc
	// NoImprovise turns ADR 0033 §8 improvisation off for every seat
	// this factory builds (CMDCTRL_BOT_IMPROVISE=0). The model tiers
	// have it on by default since #686.
	NoImprovise bool
	// Meter is the shared Layer A absorption meter. Nil allocates
	// one, so Factory.Meter is always readable.
	Meter *rules.Meter
}

// NewFactory builds the factory. The zero FactoryOptions is valid and
// yields a server that can seat `random` and `heuristic` — the two
// tiers that need nothing but the binary.
func NewFactory(opt FactoryOptions) *Factory {
	if opt.Meter == nil {
		opt.Meter = &rules.Meter{}
	}
	return &Factory{opt: opt}
}

// Meter is the shared Layer A absorption meter, for whatever reports
// it. One per server rather than one per table: ADR 0033 §5's
// acceptance number is a property of the funnel, not of a game.
func (f *Factory) Meter() *rules.Meter { return f.opt.Meter }

// Models is the resolved model-id pair this factory builds seats
// with, for the boot log.
func (f *Factory) Models() Models { return f.opt.Models.resolve() }

// HasClient reports whether a model transport is configured — the one
// fact that decides whether the model tiers are on offer.
func (f *Factory) HasClient() bool { return f.opt.Client != nil }

// TierStatus answers whether this server can seat a bot at t.
//
// `random` and `heuristic` need nothing but the binary. `assisted`
// and `strong` need a model transport, AND THIS IS THE DELIBERATE
// PART: without one they are reported unavailable rather than
// offered.
//
// tiers.Options says plainly that a nil Client is legal and that the
// seat "plays on Layer A + B under the tier's own name", and that is
// still true of the POLICY — it is a complete policy, and the outage
// drill depends on it being one. What is not acceptable is OFFERING
// that seat in the picker. A player who chooses "assisted" and gets
// heuristic play under the assisted label has been told something
// false about the game they are playing, and aiseat/tier.go names
// that exact failure as worse than no bot at all. The nil-Client path
// therefore remains fully supported for a seat that is ALREADY
// running when a transport goes away mid-game (which is what the
// funnel's fallback is for), and is not something a picker offers up
// front.
func (f *Factory) TierStatus(t aiseat.Tier) aiseat.TierStatus {
	tier, err := Parse(string(t))
	if err != nil {
		return aiseat.TierStatus{Reason: fmt.Sprintf("%q is not a tier this build knows", t)}
	}
	if tier.NeedsModel() && f.opt.Client == nil {
		return aiseat.TierStatus{Reason: "needs a model: set CMDCTRL_OPENAI_ENDPOINT to a local LLM (Ollama, LM Studio, llama.cpp, vLLM) or CMDCTRL_ANTHROPIC_API_KEY to a hosted one, then restart the server"}
	}
	return aiseat.TierStatus{Available: true}
}

// RunnerConfig is the pacing a runner gets for this tier: the tier's
// own MaxThink, widened by the deployment's override.
func (f *Factory) RunnerConfig(t aiseat.Tier) aiseat.Config {
	tier, err := Parse(string(t))
	if err != nil {
		return aiseat.ConfigFor(t)
	}
	return tier.RunnerConfigWith(f.opt.MaxThink)
}

// NewPolicy builds the policy for one seat.
//
// A tier this server cannot seat is an error wrapping
// aiseat.ErrTierUnavailable — never a weaker policy under the
// requested name. The Manager logs it and leaves the seat idle, which
// is loud; a downgrade would be silent.
func (f *Factory) NewPolicy(seat aiseat.SeatSpec) (aiseat.Policy, error) {
	tier, err := Parse(seat.Tier)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", aiseat.ErrUnknownTier, seat.Tier)
	}
	if st := f.TierStatus(aiseat.Tier(tier)); !st.Available {
		return nil, fmt.Errorf("%w: %q: %s", aiseat.ErrTierUnavailable, tier, st.Reason)
	}
	opt := Options{
		Meter:       f.opt.Meter,
		Client:      f.opt.Client,
		Models:      f.opt.Models,
		MaxThink:    f.opt.MaxThink,
		NoImprovise: f.opt.NoImprovise,
	}
	if tier.NeedsModel() && f.opt.DeckProfile != nil && seat.Deck != "" {
		if profile, ok := f.opt.DeckProfile(seat.Deck); ok {
			opt.Deck = profile
		}
	}
	return New(tier, opt)
}
