package suite

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// run.go asks a policy every labelled question in the suite and
// counts the answers.
//
// The classification is deliberately the funnel's own, not a second
// opinion invented here: when the policy is an aiseat.Tracer, a
// malformed reply, an index that is not a move and a call that did
// not come back are read off Trace.Fallback — the same field the
// decision log records and the same one Stats.ByFallback counts. A
// suite that classified failures differently from the live seat would
// report numbers nobody could act on.

// Outcome is how one position came out.
type Outcome string

const (
	// OutcomeAgree — the policy chose one of the accepted moves (or
	// declined on a position where declining is defensible).
	OutcomeAgree Outcome = "agree"
	// OutcomeRejectHit — the policy chose a move the label names as
	// specifically wrong. Worse than a disagreement: somebody wrote
	// down that this exact move is a blunder.
	OutcomeRejectHit Outcome = "reject-hit"
	// OutcomeDisagree — a legal move that is neither accepted nor
	// rejected.
	OutcomeDisagree Outcome = "disagree"
	// OutcomeMalformed — the model answered something that is not an
	// index (Trace.Fallback == model.FallbackMalformed).
	OutcomeMalformed Outcome = "malformed"
	// OutcomeOutOfRange — the answer was a number that is not a move.
	OutcomeOutOfRange Outcome = "out-of-range"
	// OutcomeTimeout — the decision did not come back inside MaxThink,
	// or the model call inside it did not.
	OutcomeTimeout Outcome = "timeout"
	// OutcomeError — the policy returned an error.
	OutcomeError Outcome = "error"
	// OutcomeDecline — the policy declined on a position where the
	// label does not allow it.
	OutcomeDecline Outcome = "decline"
	// OutcomeSkipped — the position carries no label yet.
	OutcomeSkipped Outcome = "skipped"
)

// Miss reports whether an outcome is anything other than agreement on
// a labelled position. A skipped position is not a miss: nobody has
// said what the right answer is.
func (o Outcome) Miss() bool { return o != OutcomeAgree && o != OutcomeSkipped }

// Result is one position's answer.
type Result struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags,omitempty"`
	Gate    []string `json:"gate,omitempty"`
	Note    string   `json:"note,omitempty"`
	Outcome Outcome  `json:"outcome"`
	// Index and Label are what the policy chose; Index is
	// aiseat.Decline when it declined.
	Index int    `json:"index"`
	Label string `json:"label,omitempty"`
	// Reason is the policy's own explanation.
	Reason string `json:"reason,omitempty"`
	// Want is the labels of the accepted moves, for the report.
	Want []string `json:"want,omitempty"`
	// Err is the policy's error, when there was one.
	Err string `json:"err,omitempty"`
	// Layer, Fallback and HeuristicIndex come off the trace: which
	// layer answered, why the model's answer was not used, and what
	// Layer B would have played.
	Layer          string `json:"layer,omitempty"`
	Fallback       string `json:"fallback,omitempty"`
	HeuristicIndex int    `json:"heuristic_index"`
	// Latency is the whole decision. PromptBytes is the model prompt
	// this window assembled, zero when no call was made.
	Latency     time.Duration     `json:"latency_ns"`
	PromptBytes int               `json:"prompt_bytes,omitempty"`
	Usage       aiseat.TokenUsage `json:"usage"`
}

// TagStats is one tag's slice of the report.
type TagStats struct {
	Positions  int `json:"positions"`
	Labelled   int `json:"labelled"`
	Agree      int `json:"agree"`
	RejectHits int `json:"reject_hits"`
}

// AgreeRate is agreement over labelled positions, 0 when none.
func (t TagStats) AgreeRate() float64 {
	if t.Labelled == 0 {
		return 0
	}
	return float64(t.Agree) / float64(t.Labelled)
}

// Report is the whole run.
type Report struct {
	Policy    string `json:"policy"`
	Positions int    `json:"positions"`
	// Labelled is how many positions had an answer to grade against;
	// everything else in this report is computed over those.
	Labelled   int `json:"labelled"`
	Agree      int `json:"agree"`
	RejectHits int `json:"reject_hits"`
	Disagree   int `json:"disagree"`
	Malformed  int `json:"malformed"`
	OutOfRange int `json:"out_of_range"`
	Timeouts   int `json:"timeouts"`
	Errors     int `json:"errors"`
	// Declines is the labelled positions the policy declined where
	// the label does not allow it. It is its own bucket because a
	// declining policy is a different operational problem from one
	// that picks the wrong move.
	Declines int                 `json:"declines"`
	ByTag    map[string]TagStats `json:"by_tag,omitempty"`
	Latency  aiseat.Percentiles  `json:"latency"`
	// Tokens is the SUM over the run, not a per-window figure.
	Tokens         aiseat.TokenUsage `json:"tokens"`
	PromptBytesP50 int               `json:"prompt_bytes_p50,omitempty"`
	Results        []Result          `json:"results"`
}

// AgreeRate is agreement over labelled positions, 0 when none.
func (r Report) AgreeRate() float64 {
	if r.Labelled == 0 {
		return 0
	}
	return float64(r.Agree) / float64(r.Labelled)
}

// GateFailures returns one line per gated position the named policy
// missed. An empty slice is a passing gate.
//
// The policy name is a parameter rather than r.Policy because the
// gate is a property of the POSITION: a position gated on `heuristic`
// pins a heuristic invariant, and running the assisted tier over the
// same suite must not fail the build on it — the assisted tier is
// allowed to play differently, and often better.
func (r Report) GateFailures(policyName string) []string {
	var out []string
	for _, res := range r.Results {
		if !gatedBy(res.Gate, policyName) || !res.Outcome.Miss() {
			continue
		}
		line := fmt.Sprintf("%s: %s — chose %d %q; expected one of %v",
			res.ID, res.Outcome, res.Index, res.Label, res.Want)
		if res.Note != "" {
			line += " (" + res.Note + ")"
		}
		if res.Err != "" {
			line += " [" + res.Err + "]"
		}
		out = append(out, line)
	}
	return out
}

func gatedBy(gate []string, policy string) bool {
	for _, g := range gate {
		if strings.EqualFold(strings.TrimSpace(g), policy) {
			return true
		}
	}
	return false
}

// RunOptions pace a run.
type RunOptions struct {
	// MaxThink is the hard deadline on each position, the same
	// deadline the runner would impose at a table. Zero takes
	// DefaultMaxThink.
	MaxThink time.Duration
	// Parallel is how many positions run at once. Zero or one runs
	// them in order, which is what a heuristic run wants (it is
	// microseconds) and what a single-GPU model endpoint wants (more
	// in flight is slower, not faster).
	Parallel int
	// Log receives progress. Nil discards it.
	Log *slog.Logger
}

// DefaultMaxThink is the per-position deadline when RunOptions does
// not set one. It is deliberately generous: the suite is not a
// latency test of the table's pacing, it is a question of what the
// policy WOULD play, and a model on somebody's own GPU takes tens of
// seconds. The reported latency says what it cost.
const DefaultMaxThink = 60 * time.Second

// Run asks policy every position and grades the answers.
//
// Unlabelled positions are skipped without being asked: running them
// would cost a model call for a question nobody has answered.
func Run(ctx context.Context, positions []Position, policy aiseat.Policy, opt RunOptions) Report {
	if opt.MaxThink <= 0 {
		opt.MaxThink = DefaultMaxThink
	}
	if opt.Parallel <= 0 {
		opt.Parallel = 1
	}
	results := make([]Result, len(positions))
	tracer, _ := policy.(aiseat.Tracer)

	var (
		wg   sync.WaitGroup
		work = make(chan int)
	)
	for w := 0; w < opt.Parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range work {
				results[i] = decide(ctx, positions[i], policy, tracer, opt)
				if opt.Log != nil {
					opt.Log.Info("suite position",
						"id", positions[i].ID, "outcome", results[i].Outcome,
						"label", results[i].Label, "latency", results[i].Latency)
				}
			}
		}()
	}
	for i := range positions {
		work <- i
	}
	close(work)
	wg.Wait()

	return summarise(policy.Name(), positions, results)
}

// decide runs one position under its own deadline and classifies the
// answer.
func decide(ctx context.Context, p Position, policy aiseat.Policy, tracer aiseat.Tracer, opt RunOptions) Result {
	res := Result{
		ID: p.ID, Tags: p.Tags, Gate: p.Gate, Note: p.Note,
		Index: aiseat.Decline, HeuristicIndex: aiseat.Decline,
		Want: p.AcceptLabels(),
	}
	if !p.Labelled() {
		res.Outcome = OutcomeSkipped
		return res
	}

	dctx, cancel := context.WithTimeout(ctx, opt.MaxThink)
	defer cancel()
	started := time.Now()
	var (
		d   aiseat.Decision
		tr  aiseat.Trace
		err error
	)
	if tracer != nil {
		d, tr, err = tracer.DecideTraced(dctx, p.Input)
	} else {
		d, err = policy.Decide(dctx, p.Input)
		tr = aiseat.Trace{HeuristicIndex: aiseat.Decline}
	}
	res.Latency = time.Since(started)
	res.Layer, res.Fallback, res.HeuristicIndex = tr.Layer, tr.Fallback, tr.HeuristicIndex
	res.Usage = tr.Usage
	if tr.Prompt != nil {
		res.PromptBytes = len(tr.Prompt.User)
		for _, b := range tr.Prompt.System {
			res.PromptBytes += len(b)
		}
	}
	res.Index, res.Reason = d.Index, d.Reason
	if d.Index >= 0 && d.Index < len(p.Input.Moves) {
		res.Label = p.Input.Moves[d.Index].Label
	}

	// The policy failed outright. This mirrors runner.decide: a
	// deadline miss and a policy that threw take the same fallback
	// and are counted apart, because the fixes are different.
	if err != nil {
		res.Err = err.Error()
		res.Outcome = OutcomeError
		if errors.Is(err, context.DeadlineExceeded) || dctx.Err() != nil {
			res.Outcome = OutcomeTimeout
		}
		return res
	}

	// The funnel's own classification of a model failure, ahead of
	// the grading. A window where the model produced nothing usable
	// is a model failure even though Layer B's move underneath may
	// well be the accepted one — and counting it as agreement would
	// hide exactly the transport bug this harness exists to find.
	switch {
	case tr.TimedOut:
		res.Outcome = OutcomeTimeout
		return res
	case tr.Fallback == model.FallbackMalformed:
		res.Outcome = OutcomeMalformed
		return res
	case tr.Fallback == model.FallbackOutOfRange:
		res.Outcome = OutcomeOutOfRange
		return res
	case tr.Fallback == model.FallbackError || tr.Fallback == model.FallbackPolicyError:
		res.Outcome = OutcomeError
		return res
	}

	switch {
	case d.Index == aiseat.Decline:
		res.Outcome = OutcomeDecline
		if p.Expected.DeclineOK {
			res.Outcome = OutcomeAgree
		}
	case d.Index < 0 || d.Index >= len(p.Input.Moves):
		res.Outcome = OutcomeOutOfRange
	case p.Rejects(d.Index):
		res.Outcome = OutcomeRejectHit
	case p.Accepts(d.Index):
		res.Outcome = OutcomeAgree
	default:
		res.Outcome = OutcomeDisagree
	}
	return res
}

func summarise(policyName string, positions []Position, results []Result) Report {
	rep := Report{
		Policy:    policyName,
		Positions: len(positions),
		ByTag:     map[string]TagStats{},
		Results:   results,
	}
	var (
		lat    []time.Duration
		prompt []int
	)
	for i, res := range results {
		labelled := res.Outcome != OutcomeSkipped
		if labelled {
			rep.Labelled++
			lat = append(lat, res.Latency)
		}
		if res.PromptBytes > 0 {
			prompt = append(prompt, res.PromptBytes)
		}
		rep.Tokens.InputTokens += res.Usage.InputTokens
		rep.Tokens.OutputTokens += res.Usage.OutputTokens
		rep.Tokens.CacheReadTokens += res.Usage.CacheReadTokens
		rep.Tokens.CacheWriteTokens += res.Usage.CacheWriteTokens
		rep.Tokens.CachedPromptTokens += res.Usage.CachedPromptTokens

		switch res.Outcome {
		case OutcomeAgree:
			rep.Agree++
		case OutcomeRejectHit:
			rep.RejectHits++
		case OutcomeDisagree:
			rep.Disagree++
		case OutcomeMalformed:
			rep.Malformed++
		case OutcomeOutOfRange:
			rep.OutOfRange++
		case OutcomeTimeout:
			rep.Timeouts++
		case OutcomeError:
			rep.Errors++
		case OutcomeDecline:
			rep.Declines++
		case OutcomeSkipped:
		}
		for _, tag := range positions[i].Tags {
			t := rep.ByTag[tag]
			t.Positions++
			if labelled {
				t.Labelled++
			}
			switch res.Outcome {
			case OutcomeAgree:
				t.Agree++
			case OutcomeRejectHit:
				t.RejectHits++
			}
			rep.ByTag[tag] = t
		}
	}
	rep.Latency = aiseat.PercentilesOf(lat)
	if len(prompt) > 0 {
		sort.Ints(prompt)
		rep.PromptBytesP50 = prompt[(len(prompt)-1)/2]
	}
	return rep
}

// Markdown renders the report as the block that goes into a PR
// description or the ADR. It is the same shape the arena prints, so
// the two read as one report when both are pasted.
func (r Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "### Position suite — `%s`\n\n", r.Policy)
	fmt.Fprintf(&b, "%d positions, %d labelled, **%.0f%% agreement**", r.Positions, r.Labelled, r.AgreeRate()*100)
	if skipped := r.Positions - r.Labelled; skipped > 0 {
		fmt.Fprintf(&b, " (%d unlabelled, skipped)", skipped)
	}
	b.WriteString("\n\n")

	b.WriteString("| agree | reject-hit | disagree | decline | malformed | out-of-range | timeout | error |\n")
	b.WriteString("|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	fmt.Fprintf(&b, "| %d | %d | %d | %d | %d | %d | %d | %d |\n\n",
		r.Agree, r.RejectHits, r.Disagree, r.Declines, r.Malformed, r.OutOfRange, r.Timeouts, r.Errors)

	if len(r.ByTag) > 0 {
		tags := make([]string, 0, len(r.ByTag))
		for t := range r.ByTag {
			tags = append(tags, t)
		}
		sort.Strings(tags)
		b.WriteString("| tag | labelled | agree | agree % | reject-hits |\n")
		b.WriteString("|---|---:|---:|---:|---:|\n")
		for _, t := range tags {
			s := r.ByTag[t]
			fmt.Fprintf(&b, "| %s | %d | %d | %.0f%% | %d |\n", t, s.Labelled, s.Agree, s.AgreeRate()*100, s.RejectHits)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "latency p50 %v · p99 %v · max %v",
		r.Latency.P50.Round(time.Microsecond), r.Latency.P99.Round(time.Microsecond), r.Latency.Max.Round(time.Microsecond))
	if r.PromptBytesP50 > 0 {
		fmt.Fprintf(&b, " · prompt bytes p50 %d", r.PromptBytesP50)
	}
	if r.Tokens.InputTokens > 0 || r.Tokens.OutputTokens > 0 {
		fmt.Fprintf(&b, " · tokens in %d / out %d", r.Tokens.InputTokens, r.Tokens.OutputTokens)
	}
	b.WriteString("\n")

	var misses []Result
	for _, res := range r.Results {
		if res.Outcome.Miss() {
			misses = append(misses, res)
		}
	}
	if len(misses) > 0 {
		b.WriteString("\nMisses:\n\n")
		for _, m := range misses {
			fmt.Fprintf(&b, "- `%s` **%s** — chose %d %q, wanted %v", m.ID, m.Outcome, m.Index, m.Label, m.Want)
			if m.Note != "" {
				fmt.Fprintf(&b, " — %s", m.Note)
			}
			if len(m.Gate) > 0 {
				fmt.Fprintf(&b, " (gated: %s)", strings.Join(m.Gate, ", "))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
