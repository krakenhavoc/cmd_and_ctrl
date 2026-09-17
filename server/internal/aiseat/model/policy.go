package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
)

// ModelProfile is one model and how to ask it. Two of these make the
// cheap/frontier split ADR 0033 §5 asks for.
//
// Effort and Thinking are strings rather than bools because the
// options differ by model and a wrong one is a 400: some models
// reject output_config.effort outright, and disabling thinking is
// accepted on some and refused on others at high effort. Empty omits
// the field, which is always valid.
type ModelProfile struct {
	// ID is the provider's model id.
	ID string
	// Effort maps to output_config.effort. Empty omits it.
	Effort string
	// Thinking is "", "adaptive" or "disabled". Empty omits it and
	// takes the model's own default.
	Thinking string
	// MaxTokens caps the reply. The reply is one small JSON object.
	MaxTokens int
}

// Config tunes the funnel. Use DefaultConfig or StrongConfig; the
// zero value has no models in it.
type Config struct {
	// Tier is the reported policy name: "assisted" or "strong".
	Tier string
	// Fallback is Layer B. Nil installs heuristic.New(), which is
	// what every tier should use — ADR 0033 §6 defines the heuristic
	// as the fallback under every model failure, and a policy
	// configured without one is a policy with nothing underneath it.
	Fallback aiseat.Policy
	// Client is the model transport. **Nil is legal**: the policy
	// then runs as Layer A + Layer B and says so in every record.
	// A deployment with no API key is a supported deployment.
	Client Client
	// Routine answers the ordinary windows; Frontier answers the
	// escalated ones.
	Routine  ModelProfile
	Frontier ModelProfile
	// Deck is the static, prompt-cached half of the prompt.
	Deck DeckProfile
	// Meter is the shared Layer A absorption meter. Optional.
	Meter *rules.Meter

	// MaxCandidates caps how many moves the model is shown. The pass
	// and Layer B's own pick are always among them.
	MaxCandidates int
	// MaxZoneCards caps how much of a zone the prompt renders.
	MaxZoneCards int

	// Epsilon is the margin under which the heuristic's top two
	// candidates count as a tie and the window escalates.
	Epsilon float64
	// ThreatPower is the total creature power one opponent must have
	// on board before "removal is available" counts as a high-stakes
	// window, and ThreatCreature is the same for a single creature.
	ThreatPower    int
	ThreatCreature int
	// DangerLife is the life total at or below which every removal
	// window is a high-stakes window.
	DangerLife int
	// WideTargetCount is ADR 0033's ">4 candidates" on a pick_target
	// choice.
	WideTargetCount int
	// AlwaysEscalate makes every window that survives Layer A go to
	// the frontier model. This is the `strong` tier.
	AlwaysEscalate bool

	// Reserve is held back from the runner's deadline so that a
	// model call which runs long still leaves time to return Layer
	// B's answer and dispatch it. ADR 0033 §10: the table never
	// waits on a bot.
	Reserve time.Duration
	// MinBudget is the least time worth starting a call with. Below
	// it the policy does not dial at all.
	MinBudget time.Duration
	// MaxCall bounds a call when the caller supplied no deadline —
	// which the runner always does, but a Policy is a public
	// interface and a test may not.
	MaxCall time.Duration

	// RecordsKept bounds the per-decision record ring. Default 256.
	RecordsKept int
	// Log receives one line per model call. Nil uses slog.Default.
	Log *slog.Logger
}

// Tier names (ADR 0033 §6).
const (
	TierAssisted = "assisted"
	TierStrong   = "strong"
)

// DefaultConfig is the `assisted` tier: Layer A, Layer B, a cheap
// model on routine windows and a frontier model on the escalation
// triggers. This is ADR 0033's default tier.
//
// The model ids are the shipped default and are meant to be
// overridden from configuration; they are named here rather than
// left blank so that a caller who supplies only a Client gets a
// working policy instead of a 404.
//
// # The latency budget is tight and nobody has measured it yet
//
// ADR 0033 §10 sets MaxThink at 2s for `assisted`. Reserve takes
// 250ms of that, MaxCall caps the rest at 1.5s, and a frontier model
// with thinking on can take longer than 1.5s on a hard position. The
// failure is safe — the window falls back to the heuristic and the
// table never waits — but a funnel that pays for frontier calls and
// then discards most of them is worse than not making them.
//
// This cannot be settled offline: there is no endpoint in CI, so the
// real distribution of call latencies is unknown. The knobs for when
// somebody has one are all here — lower Frontier.Effort, set
// Frontier.Thinking to "disabled", raise the tier's MaxThink, or
// point Frontier at the routine model and keep the escalation only
// for the wider candidate list. Watch Stats.MaxModelLatency and
// Stats.ByFallback[FallbackError] to decide which.
func DefaultConfig() Config {
	return Config{
		Tier: TierAssisted,
		// Cheap and fast: the routine model answers windows where the
		// heuristic already has a defensible answer and the model is
		// being asked to do better, not to rescue anything.
		Routine: ModelProfile{ID: "claude-haiku-4-5", MaxTokens: 128},
		// The frontier model, at low effort. Low is not a cost
		// dodge — MaxThink is two seconds, and a deep reasoning pass
		// that lands after the deadline is worth exactly as much as
		// no answer at all.
		Frontier: ModelProfile{ID: "claude-opus-5", Effort: "low", MaxTokens: 256},

		MaxCandidates:   24,
		MaxZoneCards:    24,
		Epsilon:         0.75,
		ThreatPower:     8,
		ThreatCreature:  5,
		DangerLife:      12,
		WideTargetCount: 4,

		Reserve:     250 * time.Millisecond,
		MinBudget:   200 * time.Millisecond,
		MaxCall:     1500 * time.Millisecond,
		RecordsKept: 256,
	}
}

// StrongConfig is the `strong` tier: ADR 0033 §6's "A + C, wider
// candidates".
//
// One line of that row is not implemented and saying so is better
// than pretending: **"1-ply sim on top-K" is not here.** Simulating a
// move means cloning a game, cloning a game means holding a
// *game.Game, and a policy may not have one (ADR 0033 §3; the import
// test fails the build over it). This is the same collision sub-PR 6
// hit with "Δscore on a cloned game", and it resolves the same way —
// the guarantee is worth more than the lookahead. What `strong`
// actually buys is a wider candidate list and the frontier model on
// every window that Layer A did not settle, with a larger MaxThink
// (5s, per ADR 0033 §10) to pay for it.
func StrongConfig() Config {
	c := DefaultConfig()
	c.Tier = TierStrong
	c.AlwaysEscalate = true
	c.MaxCandidates = 40
	c.MaxZoneCards = 40
	c.Frontier.Effort = "medium"
	c.MaxCall = 4 * time.Second
	return c
}

func (c Config) withDefaults() Config {
	if c.Tier == "" {
		c.Tier = TierAssisted
	}
	if c.Fallback == nil {
		c.Fallback = heuristic.New()
	}
	if c.MaxCandidates <= 0 {
		c.MaxCandidates = 24
	}
	if c.MaxZoneCards <= 0 {
		c.MaxZoneCards = 24
	}
	if c.Epsilon <= 0 {
		c.Epsilon = 0.75
	}
	if c.ThreatPower <= 0 {
		c.ThreatPower = 8
	}
	if c.ThreatCreature <= 0 {
		c.ThreatCreature = 5
	}
	if c.DangerLife <= 0 {
		c.DangerLife = 12
	}
	if c.WideTargetCount <= 0 {
		c.WideTargetCount = 4
	}
	if c.Reserve <= 0 {
		c.Reserve = 250 * time.Millisecond
	}
	if c.MinBudget <= 0 {
		c.MinBudget = 200 * time.Millisecond
	}
	if c.MaxCall <= 0 {
		c.MaxCall = 1500 * time.Millisecond
	}
	if c.RecordsKept <= 0 {
		c.RecordsKept = 256
	}
	if c.Log == nil {
		c.Log = slog.Default()
	}
	return c
}

// ranker is the optional Layer B extension that exposes the scorer's
// working. aiseat/heuristic implements it; a policy that does not is
// still a complete fallback, it just cannot fire the close-call
// escalation trigger.
type ranker interface {
	Rank(ctx context.Context, in aiseat.Input) []heuristic.Candidate
}

// Policy is the funnel: Layer A, then Layer B, then — for the windows
// that earn it — Layer C. Construct one per bot seat.
type Policy struct {
	cfg    Config
	static []Block
	rec    *recorder
}

// New returns a funnel policy. cfg.Client may be nil, in which case
// this is Layer A + Layer B wearing the tier's name.
func New(cfg Config) *Policy {
	cfg = cfg.withDefaults()
	return &Policy{
		cfg:    cfg,
		static: cfg.Deck.staticBlocks(),
		rec:    newRecorder(cfg.RecordsKept),
	}
}

// Name is the tier name (ADR 0033 §6).
func (p *Policy) Name() string { return p.cfg.Tier }

// Stats snapshots the per-decision instrumentation.
func (p *Policy) Stats() Stats { return p.rec.snapshot() }

// Records returns the retained per-decision records, oldest first.
func (p *Policy) Records() []DecisionRecord { return p.rec.records() }

// ShouldConcede forwards to Layer B. Conceding is a judgement about
// the position and the heuristic already makes it conservatively; a
// model call to decide whether to scoop would be the most expensive
// possible way to answer the question least often asked.
func (p *Policy) ShouldConcede(in aiseat.Input) bool {
	c, ok := p.cfg.Fallback.(aiseat.Conceder)
	return ok && c.ShouldConcede(in)
}

// Decide runs the funnel.
//
// The order is the whole design. Layer B is computed BEFORE the model
// is called, so that by the time anything can go wrong there is
// already an answer in hand: an outage, a timeout, a reply that is
// not JSON and a number that is not a move all take the same path and
// all cost nothing but the latency already spent.
func (p *Policy) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	d, _, err := p.decideTraced(ctx, in)
	return d, err
}

// Compile-time assertion: the funnel is a Tracer.
var _ aiseat.Tracer = (*Policy)(nil)

// DecideTraced is Decide with the funnel's working attached: which
// layer answered, what Layer B ranked, the exact prompt the model was
// shown, its raw reply, the index parsed out of it, and what the call
// cost.
//
// It is the same decision, not a second one — Decide is a thin
// wrapper around the same code — so a caller must use one or the
// other and never both on a window: two calls are two model calls.
func (p *Policy) DecideTraced(ctx context.Context, in aiseat.Input) (aiseat.Decision, aiseat.Trace, error) {
	return p.decideTraced(ctx, in)
}

func (p *Policy) decideTraced(ctx context.Context, in aiseat.Input) (aiseat.Decision, aiseat.Trace, error) {
	started := time.Now()
	tr := aiseat.Trace{HeuristicIndex: aiseat.Decline}
	if len(in.Moves) == 0 {
		return aiseat.Decision{}, tr, aiseat.ErrNoMoves
	}

	// --- Layer A ---------------------------------------------------
	v := rules.Resolve(in)
	p.cfg.Meter.Observe(v)
	if v.Absorbed() {
		p.rec.record(DecisionRecord{
			Layer: LayerA, Rule: v.Rule, Index: v.Index,
			Reason: v.Reason, Latency: time.Since(started),
		})
		tr.Layer, tr.Rule = LayerA, v.Rule
		return aiseat.Decision{Index: v.Index, Reason: v.Reason}, tr, nil
	}

	// --- Layer B ---------------------------------------------------
	var cands []heuristic.Candidate
	if r, ok := p.cfg.Fallback.(ranker); ok {
		cands = r.Rank(ctx, in)
	}
	tr.Candidates = traceCandidates(cands)
	base, err := p.cfg.Fallback.Decide(ctx, in)
	if err != nil {
		// Nothing is left underneath. Hand the error up; the runner
		// takes the pass.
		p.rec.record(DecisionRecord{
			Layer: LayerB, Fallback: FallbackPolicyError,
			Index: base.Index, Latency: time.Since(started),
		})
		tr.Layer, tr.Fallback = LayerB, FallbackPolicyError
		return base, tr, err
	}

	rec := DecisionRecord{
		Layer:       LayerB,
		Escalations: p.escalationReasons(in, cands),
		Index:       base.Index,
		Reason:      base.Reason,
	}
	tr.Layer, tr.HeuristicIndex, tr.Escalations = LayerB, base.Index, rec.Escalations
	finish := func(d aiseat.Decision) (aiseat.Decision, aiseat.Trace) {
		rec.Latency = time.Since(started)
		p.rec.record(rec)
		tr.Layer, tr.Fallback = rec.Layer, rec.Fallback
		return d, tr
	}

	// --- Layer C ---------------------------------------------------
	if p.cfg.Client == nil {
		rec.Fallback = FallbackNoClient
		d, tr := finish(base)
		return d, tr, nil
	}
	budget := p.budget(ctx)
	if budget < p.cfg.MinBudget {
		rec.Fallback = FallbackNoBudget
		d, tr := finish(base)
		return d, tr, nil
	}
	profile := p.cfg.Routine
	if len(rec.Escalations) > 0 {
		profile = p.cfg.Frontier
	}
	rec.Model = profile.ID
	tr.Model = profile.ID

	delta, _ := p.buildDelta(in, cands, base.Index)
	req := Request{
		Model:     profile.ID,
		System:    p.static,
		User:      delta,
		MaxTokens: profile.MaxTokens,
		Effort:    profile.Effort,
		Thinking:  profile.Thinking,
	}
	tr.Prompt = tracePrompt(req)

	callCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	callStarted := time.Now()
	resp, cerr := p.cfg.Client.Complete(callCtx, req)
	rec.Attempted = true
	rec.ModelLatency = time.Since(callStarted)
	rec.Usage = resp.Usage
	tr.ModelLatency, tr.Usage, tr.Reply = rec.ModelLatency, traceUsage(resp.Usage), resp.Text
	if cerr != nil {
		rec.Fallback = FallbackError
		// A deadline miss is a different operational problem from an
		// outage or a 500, and on a self-hosted model it is the
		// LIKELY one: the machine is simply slower than this tier's
		// MaxThink. It stays under FallbackError — it is still a call
		// that failed — but it is counted and named separately,
		// because "your model is too slow, raise the deadline" and
		// "your endpoint is down" have different fixes and the seat
		// looks identical from the table either way.
		if errors.Is(cerr, context.DeadlineExceeded) {
			rec.TimedOut = true
			tr.TimedOut = true
			p.cfg.Log.Warn("bot model call TIMED OUT; playing the heuristic's move — the model is slower than this tier's deadline (raise CMDCTRL_BOT_MAX_THINK)",
				"tier", p.cfg.Tier, "model", profile.ID, "took", rec.ModelLatency, "budget", budget)
			d, tr := finish(base)
			return d, tr, nil
		}
		p.cfg.Log.Warn("bot model call failed; playing the heuristic's move",
			"tier", p.cfg.Tier, "model", profile.ID, "took", rec.ModelLatency, "err", cerr)
		d, tr := finish(base)
		return d, tr, nil
	}

	idx, why, perr := parseAnswer(resp.Text)
	if perr == nil {
		parsed := idx
		tr.ParsedIndex = &parsed
	}
	switch {
	case perr != nil:
		rec.Fallback = FallbackMalformed
		p.cfg.Log.Warn("bot model reply was not an index; playing the heuristic's move",
			"tier", p.cfg.Tier, "model", profile.ID, "reply", truncate(resp.Text, 200))
		d, tr := finish(base)
		return d, tr, nil
	case idx < 0 || idx >= len(in.Moves):
		// The one failure the closed move list makes harmless: a
		// number that is not a move is not a move, and there is
		// nothing to validate beyond the bounds.
		//
		// -1 is caught here too, deliberately. It is aiseat.Decline
		// on the Go side, but the model was never told that and
		// "none of these" is not one of the things it is being asked.
		// Declining belongs to Layer B, which owns the judgement and
		// whose answer this falls back to anyway.
		rec.Fallback = FallbackOutOfRange
		p.cfg.Log.Warn("bot model chose a move that was not offered; playing the heuristic's move",
			"tier", p.cfg.Tier, "model", profile.ID, "index", idx, "moves", len(in.Moves))
		d, tr := finish(base)
		return d, tr, nil
	}

	rec.Layer = LayerC
	rec.Index = idx
	rec.Reason = modelReason(profile.ID, why)
	d, tr := finish(aiseat.Decision{Index: idx, Reason: rec.Reason})
	return d, tr, nil
}

// BuildRequest assembles exactly the model call this window WOULD
// make, and makes none: Layer A, then Layer B's ranking and pick,
// then the prompt.
//
// It exists for the tools that need the prompt without the cost — the
// position suite's labelling screen renders it, `boteval probe` sends
// one of them by hand, and the prompt-size metric measures it. Pulling
// the same bytes out of a Trace would mean paying for a call to see
// what the call would have said.
//
// The returned Verdict is Layer A's. When it absorbed the window
// there is no request to build and the zero Request comes back: the
// model would never have been asked.
func (p *Policy) BuildRequest(ctx context.Context, in aiseat.Input) (Request, []heuristic.Candidate, rules.Verdict) {
	// Resolve, deliberately WITHOUT Meter.Observe: this is not a
	// window the seat played, and counting it would move the
	// absorption rate with calls nobody made.
	v := rules.Resolve(in)
	if len(in.Moves) == 0 || v.Absorbed() {
		return Request{}, nil, v
	}
	var cands []heuristic.Candidate
	if r, ok := p.cfg.Fallback.(ranker); ok {
		cands = r.Rank(ctx, in)
	}
	fallback := 0
	if base, err := p.cfg.Fallback.Decide(ctx, in); err == nil {
		fallback = base.Index
	}
	profile := p.cfg.Routine
	if len(p.escalationReasons(in, cands)) > 0 {
		profile = p.cfg.Frontier
	}
	delta, _ := p.buildDelta(in, cands, fallback)
	return Request{
		Model:     profile.ID,
		System:    p.static,
		User:      delta,
		MaxTokens: profile.MaxTokens,
		Effort:    profile.Effort,
		Thinking:  profile.Thinking,
	}, cands, v
}

// --- trace projections ----------------------------------------------

func traceCandidates(cands []heuristic.Candidate) []aiseat.Candidate {
	if len(cands) == 0 {
		return nil
	}
	out := make([]aiseat.Candidate, 0, len(cands))
	for _, c := range cands {
		out = append(out, aiseat.Candidate{Index: c.Index, Value: c.Value, Reason: c.Reason})
	}
	return out
}

func tracePrompt(req Request) *aiseat.Prompt {
	pr := &aiseat.Prompt{User: req.User}
	for _, b := range req.System {
		pr.System = append(pr.System, b.Text)
	}
	return pr
}

func traceUsage(u Usage) aiseat.TokenUsage {
	return aiseat.TokenUsage{
		InputTokens:        u.InputTokens,
		OutputTokens:       u.OutputTokens,
		CacheReadTokens:    u.CacheReadTokens,
		CacheWriteTokens:   u.CacheWriteTokens,
		CachedPromptTokens: u.CachedPromptTokens,
	}
}

// budget is how long a model call may take: whatever the runner's
// deadline leaves, less Reserve so there is still time to come back
// with Layer B's answer and dispatch it.
func (p *Policy) budget(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return p.cfg.MaxCall
	}
	left := time.Until(deadline) - p.cfg.Reserve
	if left > p.cfg.MaxCall {
		left = p.cfg.MaxCall
	}
	return left
}

// --- parsing the reply ---------------------------------------------

type answer struct {
	Index *int   `json:"index"`
	Why   string `json:"why"`
}

// ParseAnswerIndex is the funnel's own reply parser, exported for the
// tools that have to score a raw reply exactly the way a live seat
// would — `boteval probe` sends one request by hand and has to
// classify the answer identically or it is measuring its own parser.
func ParseAnswerIndex(text string) (int, error) {
	i, _, err := parseAnswer(text)
	return i, err
}

// parseAnswer pulls an index out of the reply.
//
// It is deliberately tolerant of packaging and strict about content:
// a fenced block, leading prose or a trailing sentence are all
// survivable, because none of them changes which move was chosen and
// the alternative is throwing away a correct answer over a code
// fence. Anything that is not ultimately an integer is malformed and
// goes to Layer B — which is the specified behaviour and a cheap
// failure.
func parseAnswer(text string) (int, string, error) {
	s := strings.TrimSpace(text)
	if s == "" {
		return 0, "", errors.New("empty reply")
	}
	// A bare integer is a valid answer and costs the fewest tokens.
	if n, err := strconv.Atoi(s); err == nil {
		return n, "", nil
	}
	// Otherwise the first balanced JSON object in the reply.
	if obj := firstJSONObject(s); obj != "" {
		var a answer
		if err := json.Unmarshal([]byte(obj), &a); err == nil && a.Index != nil {
			return *a.Index, a.Why, nil
		}
	}
	return 0, "", fmt.Errorf("no index in reply %q", truncate(s, 120))
}

// firstJSONObject returns the first balanced {...} run in s, ignoring
// braces inside strings. Empty when there is none.
func firstJSONObject(s string) string {
	start, depth := -1, 0
	inStr, escaped := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					return s[start : i+1]
				}
			}
		}
	}
	return ""
}

func modelReason(model, why string) string {
	why = oneLine(strings.TrimSpace(why))
	if why == "" {
		return model
	}
	return model + ": " + truncate(why, 120)
}

// truncate cuts at a rune boundary. A Decision.Reason reaches the
// chat log behind the "show bot reasoning" setting, and half a rune
// on the wire is a rendering bug in somebody else's code.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}
