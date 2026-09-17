package botarena

import (
	"fmt"
	"math"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// report.go is ADR 0052's report block: the thing a run is FOR.
//
// The shape is fixed on purpose. Every PR in the Part 2 sequence
// pastes one of these into its description, and the whole point of
// the exercise is that two of them can be read side by side — which
// only works if the same numbers are in the same places.
//
// Three things are reported about every policy and none of them is
// optional:
//
//   - the win rate WITH an interval and WITH the null rate. "assisted
//     won 30% of a four-seat table" is not a result; the null is 25%
//     and ten games cannot tell them apart. Wilson is the interval
//     that behaves at the small n and the extreme p this harness
//     lives at — a normal-approximation interval on 0 wins from 10
//     games is [0, 0], which is a confident statement of something
//     false.
//   - the funnel. A model tier whose every window fell back to Layer
//     B has a win rate identical to the heuristic's and a completely
//     different explanation.
//   - the tails. A p50 says nothing about a seat that hangs the table
//     twice a game.

// DefaultZ is the 95% two-sided normal quantile.
const DefaultZ = 1.96

// Wilson is the Wilson score interval for wins out of n at z
// standard deviations.
//
// Why not the textbook p ± z·sqrt(p(1−p)/n): at the sample sizes an
// arena run can afford (10 games, sometimes 4) and the proportions it
// produces (0, 1, or something near the 1/seats null), the normal
// approximation is wrong in the direction that matters. It gives a
// zero-width interval at p = 0 and p = 1, and intervals that run
// outside [0,1] near them. Wilson is derived by inverting the score
// test instead, never leaves [0,1], and has a sane width at 0 wins —
// 0 of 10 comes back [0, 0.28], which is the honest statement that
// ten games have not ruled out a 25% policy.
//
// n == 0 returns the whole range: no games is no evidence.
func Wilson(wins, n int, z float64) (lo, hi float64) {
	if n <= 0 {
		return 0, 1
	}
	if z <= 0 {
		z = DefaultZ
	}
	nf := float64(n)
	p := float64(wins) / nf
	z2 := z * z
	denom := 1 + z2/nf
	centre := (p + z2/(2*nf)) / denom
	margin := z * math.Sqrt(p*(1-p)/nf+z2/(4*nf*nf)) / denom
	lo, hi = centre-margin, centre+margin
	return clamp01(lo), clamp01(hi)
}

func clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

// TurnPercentiles is the nearest-rank distribution of a run's game
// lengths. It is not aiseat.Percentiles because a turn count is not a
// duration, and a report that prints "turns p50: 47ns" is a report
// nobody trusts the rest of.
type TurnPercentiles struct {
	Count int `json:"count"`
	P50   int `json:"p50"`
	P99   int `json:"p99"`
	Max   int `json:"max"`
}

// turnPercentiles reuses aiseat's nearest-rank implementation by
// borrowing its type for the arithmetic — one definition of "p99",
// not two that drift.
func turnPercentiles(turns []int) TurnPercentiles {
	if len(turns) == 0 {
		return TurnPercentiles{}
	}
	d := make([]time.Duration, 0, len(turns))
	for _, t := range turns {
		d = append(d, time.Duration(t))
	}
	p := aiseat.PercentilesOf(d)
	return TurnPercentiles{Count: p.Count, P50: int(p.P50), P99: int(p.P99), Max: int(p.Max)}
}

func medianInt(v []int) int {
	if len(v) == 0 {
		return 0
	}
	s := append([]int(nil), v...)
	sort.Ints(s)
	return s[(len(s)-1)/2]
}

// PolicyTotals is everything one policy did across a run.
//
// Games counts SEAT-games, not games: two heuristic seats in one
// four-seat game is two games of evidence about the heuristic, and
// Wins+Losses+Draws == Games.
type PolicyTotals struct {
	Policy  string `json:"policy"`
	Games   int    `json:"games"`
	Wins    int    `json:"wins"`
	Losses  int    `json:"losses"`
	Draws   int    `json:"draws"`
	Stalled int    `json:"stalled"`
	// WinRate is Wins/Games; CILow and CIHigh are its Wilson 95%
	// interval; Null is what a seat would win by chance at this table
	// size (1/seats). A WinRate whose interval straddles Null is a
	// measurement that has not yet said anything.
	WinRate float64 `json:"win_rate"`
	CILow   float64 `json:"ci_low"`
	CIHigh  float64 `json:"ci_high"`
	Null    float64 `json:"null"`

	Turns          TurnPercentiles    `json:"turns"`
	Decision       aiseat.Percentiles `json:"decision_latency"`
	ModelCall      aiseat.Percentiles `json:"model_call_latency"`
	PromptBytesP50 int                `json:"prompt_bytes_p50,omitempty"`

	Funnel model.Stats      `json:"funnel"`
	Runner aiseat.Stats     `json:"runner"`
	Meter  rules.MeterStats `json:"meter"`
}

// Beats reports whether the whole interval is above the null rate —
// the only honest way to say "this policy is stronger than chance"
// from a handful of games.
func (t PolicyTotals) Beats() bool { return t.Games > 0 && t.CILow > t.Null }

// GameDigest is one line about one game, for the summary. The full
// GameResult goes to games.jsonl; this is what fits in a table.
type GameDigest struct {
	Seed    uint64        `json:"seed"`
	GameID  uuid.UUID     `json:"game_id"`
	Turns   int           `json:"turns"`
	State   game.State    `json:"state"`
	Winner  string        `json:"winner,omitempty"`
	Elapsed time.Duration `json:"elapsed_ns"`
	Stalled bool          `json:"stalled,omitempty"`
}

func digest(r GameResult) GameDigest {
	return GameDigest{
		Seed: r.Seed, GameID: r.GameID, Turns: r.Turns, State: r.State,
		Winner: r.WinnerLabel(), Elapsed: r.Elapsed, Stalled: r.Stalled,
	}
}

// ConfigSummary is what the run was, recorded next to what it found.
// ADR 0052's "record with each run" list, minus the model's quant,
// which no API exposes and an operator writes into Note.
type ConfigSummary struct {
	Seats       []SeatSpec    `json:"seats"`
	Games       int           `json:"games"`
	Seed        uint64        `json:"seed"`
	Rotate      bool          `json:"rotate"`
	TurnBudget  int           `json:"turn_budget"`
	Wall        time.Duration `json:"wall_ns"`
	Stall       time.Duration `json:"stall_ns"`
	MaxThink    time.Duration `json:"max_think_ns"`
	Routine     string        `json:"routine_model,omitempty"`
	Frontier    string        `json:"frontier_model,omitempty"`
	Endpoint    string        `json:"endpoint,omitempty"`
	HasIndex    bool          `json:"has_scryfall_index"`
	DecisionLog string        `json:"decision_log_dir,omitempty"`
	ReplayDir   string        `json:"replay_dir,omitempty"`
	// Revision is the build's VCS revision, stamped by the Go
	// toolchain. It answers "which prompt was this?" without anybody
	// having to remember to write it down.
	Revision string `json:"revision,omitempty"`
	// Note is free-form: the model's quantisation, what was being
	// tested, whatever the next reader will wish had been recorded.
	Note string `json:"note,omitempty"`
}

func describeConfig(cfg Config) ConfigSummary {
	c := ConfigSummary{
		Seats: append([]SeatSpec(nil), cfg.Seats...), Games: cfg.Games, Seed: cfg.Seed,
		Rotate: cfg.Rotate, TurnBudget: cfg.TurnBudget, Wall: cfg.Wall, Stall: cfg.Stall,
		MaxThink: cfg.MaxThink, Routine: cfg.Models.Routine, Frontier: frontierOf(cfg),
		HasIndex: cfg.Index != nil, ReplayDir: cfg.ReplayDir, Revision: revision(), Note: cfg.Note,
	}
	if cfg.DecisionLog != nil {
		c.DecisionLog = cfg.DecisionLog.Dir()
	}
	if u, ok := cfg.Client.(interface{ URL() string }); ok {
		c.Endpoint = u.URL()
	}
	return c
}

// frontierOf mirrors tiers.Models' own rule: a deployment that set
// one model id is running one model, and a report that says
// "frontier: —" about it is reporting a gap that is not there.
func frontierOf(cfg Config) string {
	if cfg.Models.Frontier != "" {
		return cfg.Models.Frontier
	}
	return cfg.Models.Routine
}

func revision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			return s.Value
		}
	}
	return ""
}

// DecisionLogTotals is the decision log's own tally across a run.
//
// It is in the report because the writer is asynchronous and drops
// rather than blocks a bot seat, so a corpus can have holes in it and
// nothing else would say so. Dropped > 0 means the position suite
// harvested from this run is missing windows: DroppedQueue is the
// seats outrunning the writer, DroppedCap the per-game byte cap,
// DroppedError a marshal or write that failed.
type DecisionLogTotals struct {
	Records      int64 `json:"records"`
	Dropped      int64 `json:"dropped"`
	DroppedQueue int64 `json:"dropped_queue"`
	DroppedCap   int64 `json:"dropped_cap"`
	DroppedError int64 `json:"dropped_error"`
	Bytes        int64 `json:"bytes"`
}

// Add folds one game's stats in.
func (t *DecisionLogTotals) Add(s decisionlog.Stats) {
	t.Records += s.Records
	t.Dropped += s.Dropped
	t.DroppedQueue += s.DroppedQueue
	t.DroppedCap += s.DroppedCap
	t.DroppedError += s.DroppedError
	t.Bytes += s.Bytes
}

// Summary is a whole run.
type Summary struct {
	Config    ConfigSummary            `json:"config"`
	Started   time.Time                `json:"started"`
	Elapsed   time.Duration            `json:"elapsed_ns"`
	Games     []GameDigest             `json:"games"`
	PerPolicy map[string]*PolicyTotals `json:"per_policy"`
	// DecisionLog is the writer's tally, zero when no log was on.
	DecisionLog DecisionLogTotals `json:"decision_log"`
}

// Policies returns the tally keys in a stable order: most games
// first, then alphabetically.
func (s Summary) Policies() []string {
	out := make([]string, 0, len(s.PerPolicy))
	for k := range s.PerPolicy {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := s.PerPolicy[out[i]], s.PerPolicy[out[j]]
		if a.Games != b.Games {
			return a.Games > b.Games
		}
		return out[i] < out[j]
	})
	return out
}

// Stalls counts the games that stopped committing moves.
func (s Summary) Stalls() int {
	n := 0
	for _, g := range s.Games {
		if g.Stalled {
			n++
		}
	}
	return n
}

// WallPerGame is the mean wall clock a game took.
func (s Summary) WallPerGame() time.Duration {
	if len(s.Games) == 0 {
		return 0
	}
	var total time.Duration
	for _, g := range s.Games {
		total += g.Elapsed
	}
	return total / time.Duration(len(s.Games))
}

// accumulator folds GameResults into PolicyTotals.
type accumulator struct {
	seats  int
	per    map[string]*policyAcc
	stalls map[string]int
}

type policyAcc struct {
	totals      PolicyTotals
	turns       []int
	decision    []time.Duration
	modelCall   []time.Duration
	promptBytes []int
}

func newAccumulator(seats int) *accumulator {
	return &accumulator{seats: seats, per: map[string]*policyAcc{}, stalls: map[string]int{}}
}

func (a *accumulator) add(r GameResult) {
	for _, s := range r.Seats {
		key := s.Spec.Label()
		p := a.per[key]
		if p == nil {
			p = &policyAcc{}
			p.totals.Policy = key
			a.per[key] = p
		}
		t := &p.totals
		t.Games++
		switch {
		case r.Winner < 0:
			t.Draws++
		case s.Position == r.Winner:
			t.Wins++
		default:
			t.Losses++
		}
		if r.Stalled {
			t.Stalled++
		}
		p.turns = append(p.turns, r.Turns)
		mergeRunner(&t.Runner, s.Runner)
		mergeMeter(&t.Meter, s.Meter)
		if s.Funnel != nil {
			mergeFunnel(&t.Funnel, *s.Funnel)
		}
		if s.raw != nil {
			s.raw.mu.Lock()
			p.decision = append(p.decision, s.raw.decision...)
			p.modelCall = append(p.modelCall, s.raw.modelCall...)
			p.promptBytes = append(p.promptBytes, s.raw.promptBytes...)
			s.raw.mu.Unlock()
		}
	}
}

func (a *accumulator) totals(seats int) map[string]*PolicyTotals {
	out := make(map[string]*PolicyTotals, len(a.per))
	null := 0.0
	if seats > 0 {
		null = 1 / float64(seats)
	}
	for k, p := range a.per {
		t := p.totals
		if t.Games > 0 {
			t.WinRate = float64(t.Wins) / float64(t.Games)
		}
		t.CILow, t.CIHigh = Wilson(t.Wins, t.Games, DefaultZ)
		t.Null = null
		t.Turns = turnPercentiles(p.turns)
		t.Decision = aiseat.PercentilesOf(p.decision)
		t.ModelCall = aiseat.PercentilesOf(p.modelCall)
		t.PromptBytesP50 = medianInt(p.promptBytes)
		t.Runner.Latency = t.Decision
		total := t
		out[k] = &total
	}
	return out
}

func mergeRunner(dst *aiseat.Stats, s aiseat.Stats) {
	dst.Decisions += s.Decisions
	dst.Applied += s.Applied
	dst.Rejected += s.Rejected
	dst.Fallbacks += s.Fallbacks
	dst.Passes += s.Passes
	dst.Improvisations += s.Improvisations
	dst.ImprovRefused += s.ImprovRefused
	// Bounded: the rejection list is evidence, not a tally, and a
	// hundred-game run would otherwise carry thousands of them into
	// summary.json.
	for _, r := range s.Rejections {
		if len(dst.Rejections) >= 32 {
			break
		}
		dst.Rejections = append(dst.Rejections, r)
	}
}

func mergeMeter(dst *rules.MeterStats, s rules.MeterStats) {
	dst.Windows += s.Windows
	dst.Absorbed += s.Absorbed
	if len(s.ByRule) > 0 && dst.ByRule == nil {
		dst.ByRule = map[string]int64{}
	}
	for k, v := range s.ByRule {
		dst.ByRule[k] += v
	}
}

func mergeFunnel(dst *model.Stats, s model.Stats) {
	dst.Windows += s.Windows
	dst.Escalated += s.Escalated
	dst.ModelCalls += s.ModelCalls
	dst.ModelTimeouts += s.ModelTimeouts
	dst.ModelLatency += s.ModelLatency
	if s.MaxModelLatency > dst.MaxModelLatency {
		dst.MaxModelLatency = s.MaxModelLatency
	}
	dst.Usage.InputTokens += s.Usage.InputTokens
	dst.Usage.OutputTokens += s.Usage.OutputTokens
	dst.Usage.CacheReadTokens += s.Usage.CacheReadTokens
	dst.Usage.CacheWriteTokens += s.Usage.CacheWriteTokens
	dst.Usage.CachedPromptTokens += s.Usage.CachedPromptTokens
	dst.ByLayer = mergeCounts(dst.ByLayer, s.ByLayer)
	dst.ByEscalation = mergeCounts(dst.ByEscalation, s.ByEscalation)
	dst.ByFallback = mergeCounts(dst.ByFallback, s.ByFallback)
}

func mergeCounts(dst, src map[string]int64) map[string]int64 {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = map[string]int64{}
	}
	for k, v := range src {
		dst[k] += v
	}
	return dst
}

// --- the report -----------------------------------------------------

// Markdown renders ADR 0052's report block. It is what gets pasted
// into a PR description, so it is Markdown, it fits in a comment, and
// every table has the same columns in every run.
func (s Summary) Markdown() string {
	var b strings.Builder
	c := s.Config
	fmt.Fprintf(&b, "## Arena — %d games, %d seats\n\n", len(s.Games), len(c.Seats))
	fmt.Fprintf(&b, "- **seats**: %s\n", seatLine(c.Seats))
	fmt.Fprintf(&b, "- **games**: %d, seed %d, rotation %s, turn budget %d\n",
		c.Games, c.Seed, onOff(c.Rotate), c.TurnBudget)
	if c.Routine != "" || c.Endpoint != "" {
		fmt.Fprintf(&b, "- **model**: %s (frontier %s) at %s, max think %s\n",
			orDash(c.Routine), orDash(c.Frontier), orDash(c.Endpoint), c.MaxThink)
	} else {
		b.WriteString("- **model**: none (no model tier in this run)\n")
	}
	fmt.Fprintf(&b, "- **started**: %s, elapsed %s, %s per game\n",
		s.Started.UTC().Format(time.RFC3339), s.Elapsed.Round(time.Millisecond), s.WallPerGame().Round(time.Millisecond))
	fmt.Fprintf(&b, "- **stalls**: %d of %d games\n", s.Stalls(), len(s.Games))
	if c.Revision != "" {
		fmt.Fprintf(&b, "- **revision**: `%s`\n", c.Revision)
	}
	if c.DecisionLog != "" {
		fmt.Fprintf(&b, "- **decision log**: `%s` — %d records, %s (operator-only; never attach it to a bug report)\n",
			c.DecisionLog, s.DecisionLog.Records, droppedLine(s.DecisionLog))
	}
	if c.ReplayDir != "" {
		fmt.Fprintf(&b, "- **replays**: `%s`\n", c.ReplayDir)
	}
	if c.Note != "" {
		fmt.Fprintf(&b, "- **note**: %s\n", c.Note)
	}

	names := s.Policies()

	b.WriteString("\n### Play\n\n")
	b.WriteString("| policy | seat-games | wins | win% | 95% CI | null | draws | stalls | turns p50 | beats null |\n")
	b.WriteString("|---|---:|---:|---:|---|---:|---:|---:|---:|:--:|\n")
	for _, n := range names {
		t := s.PerPolicy[n]
		fmt.Fprintf(&b, "| %s | %d | %d | %s | %s–%s | %s | %d | %d | %d | %s |\n",
			n, t.Games, t.Wins, pct(t.WinRate), pct(t.CILow), pct(t.CIHigh), pct(t.Null),
			t.Draws, t.Stalled, t.Turns.P50, yesNo(t.Beats()))
	}

	b.WriteString("\n### Funnel\n\n")
	b.WriteString("| policy | windows | A | B | C | escalated | calls | timeouts | fallbacks | in tok | out tok | prompt B p50 |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---|---:|---:|---:|\n")
	for _, n := range names {
		t := s.PerPolicy[n]
		windows := t.Funnel.Windows
		if windows == 0 {
			// A heuristic or random seat has no funnel; the Layer A
			// meter is the only record of its windows.
			windows = t.Meter.Windows
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d | %d | %d | %s | %d | %d | %d |\n",
			n, windows, layer(t, model.LayerA), layer(t, model.LayerB), t.Funnel.ByLayer[model.LayerC],
			t.Funnel.Escalated, t.Funnel.ModelCalls, t.Funnel.ModelTimeouts, counts(t.Funnel.ByFallback),
			t.Funnel.Usage.InputTokens, t.Funnel.Usage.OutputTokens, t.PromptBytesP50)
	}

	b.WriteString("\n### Latency\n\n")
	b.WriteString("| policy | decisions | decision p50 | p99 | p999 | max | model p50 | model p99 | model max |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, n := range names {
		t := s.PerPolicy[n]
		fmt.Fprintf(&b, "| %s | %d | %s | %s | %s | %s | %s | %s | %s |\n",
			n, t.Decision.Count, dur(t.Decision.P50), dur(t.Decision.P99), dur(t.Decision.P999), dur(t.Decision.Max),
			dur(t.ModelCall.P50), dur(t.ModelCall.P99), dur(t.ModelCall.Max))
	}

	b.WriteString("\n### Games\n\n")
	b.WriteString("| seed | turns | state | winner | wall | stalled |\n")
	b.WriteString("|---:|---:|---|---|---:|:--:|\n")
	for _, g := range s.Games {
		fmt.Fprintf(&b, "| %d | %d | %s | %s | %s | %s |\n",
			g.Seed, g.Turns, g.State, orDash(g.Winner), g.Elapsed.Round(time.Millisecond), yesNo(g.Stalled))
	}
	return b.String()
}

// layer reads a layer's share from whichever counter has it.
//
// A model seat records the whole funnel in model.Stats. A heuristic
// seat has no funnel at all — it is a rules.Filter wrapping the
// heuristic — so its only record is the Layer A meter, where the
// windows Layer A did not absorb are by definition the ones Layer B
// answered. Without this the report would say a heuristic seat played
// two thousand windows on no layer whatsoever.
func layer(t *PolicyTotals, name string) int64 {
	if v := t.Funnel.ByLayer[name]; v > 0 {
		return v
	}
	if t.Funnel.Windows > 0 {
		return 0
	}
	switch name {
	case model.LayerA:
		return t.Meter.Absorbed
	case model.LayerB:
		return t.Meter.Windows - t.Meter.Absorbed
	}
	return 0
}

// droppedLine says whether the corpus is complete, and when it is
// not, which way it leaked.
func droppedLine(t DecisionLogTotals) string {
	if t.Dropped == 0 {
		return "nothing dropped"
	}
	parts := make([]string, 0, 3)
	if t.DroppedQueue > 0 {
		parts = append(parts, fmt.Sprintf("%d queue", t.DroppedQueue))
	}
	if t.DroppedCap > 0 {
		parts = append(parts, fmt.Sprintf("%d cap", t.DroppedCap))
	}
	if t.DroppedError > 0 {
		parts = append(parts, fmt.Sprintf("%d error", t.DroppedError))
	}
	return fmt.Sprintf("**%d DROPPED** (%s) — this corpus has holes", t.Dropped, strings.Join(parts, ", "))
}

func seatLine(seats []SeatSpec) string {
	parts := make([]string, 0, len(seats))
	for _, s := range seats {
		if s.Deck != "" {
			parts = append(parts, fmt.Sprintf("%s(%s)", s.Label(), s.Deck))
			continue
		}
		parts = append(parts, s.Label())
	}
	return strings.Join(parts, ", ")
}

func counts(m map[string]int64) string {
	if len(m) == 0 {
		return "—"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if m[keys[i]] != m[keys[j]] {
			return m[keys[i]] > m[keys[j]]
		}
		return keys[i] < keys[j]
	})
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

func pct(v float64) string { return fmt.Sprintf("%.1f%%", v*100) }

// dur prints a latency at a precision that keeps it a latency. A
// heuristic decision is a few hundred nanoseconds and a model call is
// ten seconds, and rounding both to the millisecond turns half the
// table into "0s".
func dur(d time.Duration) string {
	switch {
	case d == 0:
		return "—"
	case d < time.Microsecond:
		return d.String()
	case d < time.Millisecond:
		return d.Round(10 * time.Nanosecond).String()
	case d < time.Second:
		return d.Round(time.Microsecond).String()
	default:
		return d.Round(time.Millisecond).String()
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
