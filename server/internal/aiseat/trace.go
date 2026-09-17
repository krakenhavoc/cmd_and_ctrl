package aiseat

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// trace.go is the observation surface: what a policy did on one
// window, and what the runner then did with it.
//
// It exists because the funnel's working is otherwise unreachable.
// The prompt and the reply are locals inside model.Policy.Decide, the
// heuristic's ranking is thrown away by Decide, and the runner's own
// fallbacks are log lines. None of that can be measured, replayed or
// reviewed — and ADR 0033 §5's cost and quality claims are all
// measurements nobody has taken yet. A Trace is the record that makes
// them takeable.
//
// Everything here is PLAIN DATA with JSON tags. It names no policy
// package (a Trace describes the funnel without importing any layer
// of it), holds no engine types, and can be marshalled, written to a
// file and read back by a tool that has never linked the policies.
// aiseat/decisionlog is the first such reader.
//
// # Hidden information
//
// A Trace, like the Input it describes, is built from the seat's OWN
// filtered view. It leaks nothing that seat could not see. The FILE a
// decision log aggregates them into is a different matter — it holds
// every bot seat's view of one game — which is why that file is
// operator-only and off by default. See aiseat/decisionlog.

// Candidate is one move with the price Layer B put on it. It mirrors
// heuristic.Candidate, deliberately: a Trace may not import the
// heuristic (the ban runs the other way, but the dependency would
// still be wrong — a trace is data about a decision, not a decision).
type Candidate struct {
	Index  int     `json:"index"`
	Value  float64 `json:"value"`
	Reason string  `json:"reason,omitempty"`
}

// Prompt is exactly what a model was shown: the system blocks in
// order, and the per-decision user delta.
type Prompt struct {
	System []string `json:"system,omitempty"`
	User   string   `json:"user,omitempty"`
}

// TokenUsage is what one model call cost, as the provider reported
// it. It mirrors model.Usage for the same reason Candidate mirrors
// heuristic.Candidate.
type TokenUsage struct {
	InputTokens      int `json:"input_tokens,omitempty"`
	OutputTokens     int `json:"output_tokens,omitempty"`
	CacheReadTokens  int `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens int `json:"cache_write_tokens,omitempty"`
	// CachedPromptTokens is a local server's PREFIX-cache hit count
	// (Ollama reports it as usage.prompt_tokens_details.cached_tokens).
	// It is not CacheReadTokens: that one is Anthropic's explicit
	// cache-breakpoint read, and conflating them would corrupt the
	// one measurement ADR 0033 §5's cost argument rests on.
	CachedPromptTokens int `json:"cached_prompt_tokens,omitempty"`
}

// Trace layer names. They match model.LayerA/B/C, plus one the funnel
// has no name for because it never produced it.
const (
	// TraceLayerRandom is the random policy: no layers at all. The
	// runner writes it, because a policy that is not a Tracer cannot.
	TraceLayerRandom = "random"
)

// Trace is one policy's working on one window.
//
// The zero value is a valid trace of a policy that told us nothing.
// HeuristicIndex is -1 when Layer B was never consulted (a Layer A
// verdict, or a policy with no heuristic underneath), so a reader can
// tell "the heuristic wanted move 0" from "nobody asked it".
type Trace struct {
	// Layer is which layer answered: "A", "B", "C", or "random".
	Layer string `json:"layer,omitempty"`
	// Rule is the Layer A rule that fired, when Layer is "A".
	Rule string `json:"rule,omitempty"`
	// HeuristicIndex is the move Layer B would have taken, or -1.
	HeuristicIndex int `json:"heuristic_index"`
	// Candidates is Layer B's ranking, best first, when it exposed
	// one. Empty for the windows the scalar scorer does not price.
	Candidates []Candidate `json:"candidates,omitempty"`
	// Escalations are the triggers that sent this window to the
	// frontier model, sorted.
	Escalations []string `json:"escalations,omitempty"`
	// Model is the model id that was asked, empty when none was.
	Model string `json:"model,omitempty"`
	// Prompt is what it was shown. Nil when no call was assembled.
	Prompt *Prompt `json:"prompt,omitempty"`
	// Reply is the raw reply text.
	Reply string `json:"reply,omitempty"`
	// ParsedIndex is the index parsed out of Reply, nil when the
	// reply was not an index.
	ParsedIndex *int `json:"parsed_index,omitempty"`
	// Fallback names why Layer C's answer was not used — the funnel's
	// own classification, not the runner's.
	Fallback string `json:"fallback,omitempty"`
	// TimedOut refines Fallback: the call did not fail, it did not
	// finish.
	TimedOut bool `json:"timed_out,omitempty"`
	// ModelLatency is the model call alone.
	ModelLatency time.Duration `json:"model_latency_ns,omitempty"`
	// Usage is what the call cost.
	Usage TokenUsage `json:"usage"`
}

// Tracer is an optional Policy extension, the same shape as Conceder:
// a policy that implements it can answer a window AND say how.
//
// DecideTraced must return exactly what Decide would have returned
// for the same Input — it is the same decision with the working
// shown, not a second opinion. Callers use one or the other, never
// both on the same window: running both would double every side
// effect the funnel has (a Layer A meter tick, a model call, a
// per-decision record).
type Tracer interface {
	DecideTraced(ctx context.Context, in Input) (Decision, Trace, error)
}

// Runner fallback causes: why the move the runner DISPATCHED was not
// the move the policy asked for. They are the runner's own
// classification and are deliberately separate from Trace.Fallback,
// which is the funnel's: a window can be a model failure inside the
// policy (Trace.Fallback) and still have the policy's Layer B answer
// dispatched untouched (no runner fallback at all).
const (
	// FallbackTimeout — Decide did not return inside MaxThink.
	FallbackTimeout = "timeout"
	// FallbackPolicyError — Decide returned an error.
	FallbackPolicyError = "policy-error"
	// FallbackOutOfRange — Decide returned an index that is not a move.
	FallbackOutOfRange = "out-of-range"
	// FallbackDeclinePass — the policy declined while holding
	// priority, so the decline became the pass it stood in for.
	FallbackDeclinePass = "decline-pass"
	// FallbackForcedPass — MaxConsecutiveRejects reached; the runner
	// yielded priority.
	FallbackForcedPass = "forced-pass"
	// FallbackForcedAlwaysLegal — the same, with no pass on offer, so
	// the enumerator's unconditional answer was taken (#544).
	FallbackForcedAlwaysLegal = "forced-always-legal"
)

// DecisionEvent is one decision window, after the dispatcher has had
// its say. The runner emits exactly one per window it decided in.
//
// It carries the whole Input on purpose. An event with the view in it
// can be REPLAYED — rules.Resolve and heuristic.Decide are pure
// functions of an Input — which is what turns a log of these into a
// regression suite rather than a pile of prose.
type DecisionEvent struct {
	// Game and Seat identify the window.
	Game uuid.UUID `json:"game"`
	Seat uuid.UUID `json:"seat"`
	// Policy is Runner.PolicyName().
	Policy string `json:"policy"`
	// Seq is the room sequence the move committed at, 0 when nothing
	// was committed (a decline, a hold, a rejection).
	Seq uint64 `json:"seq,omitempty"`
	// Input is what the policy was given.
	Input Input `json:"input"`
	// Traced says whether Trace came from the policy (a Tracer) or
	// was synthesised by the runner.
	Traced bool  `json:"traced"`
	Trace  Trace `json:"trace"`
	// Decision and DecisionErr are what the policy returned, before
	// the runner applied any of its own rules to it.
	Decision    Decision `json:"decision"`
	DecisionErr error    `json:"-"`
	// Fallback is the runner's own fallback cause, empty when the
	// policy's answer was dispatched as given.
	Fallback string `json:"fallback,omitempty"`
	// Index, Label and Reason are the move actually dispatched.
	// Index is Decline when nothing was.
	Index  int    `json:"index"`
	Label  string `json:"label,omitempty"`
	Reason string `json:"reason,omitempty"`
	// Latency is the whole Decide call, deadline included.
	Latency time.Duration `json:"latency_ns"`
	// Applied is whether the dispatcher took the move; RejectErr is
	// why it did not.
	Applied   bool  `json:"applied"`
	RejectErr error `json:"-"`
}

// DecisionObserver receives one DecisionEvent per window. Observe is
// called from the runner's own goroutine, inline, so an
// implementation that blocks holds up a bot seat; implementations
// that do real work (a decision log writes to disk) buffer.
//
// Several seats share one observer at a table, so Observe must be
// safe from several goroutines.
type DecisionObserver interface {
	Observe(ev DecisionEvent)
}

// ObserverFunc adapts a function to DecisionObserver.
type ObserverFunc func(ev DecisionEvent)

// Observe calls f.
func (f ObserverFunc) Observe(ev DecisionEvent) { f(ev) }

// GameDecisionLog is one game's decision log: an observer with a
// file behind it.
//
// It is an interface here, rather than the concrete writer, so that
// Manager never imports aiseat/decisionlog — which would be a cycle,
// since decisionlog reads aiseat's own types.
type GameDecisionLog interface {
	DecisionObserver
	io.Closer
}

// DecisionLogger opens one log per game. main.go builds one from
// CMDCTRL_BOT_DECISION_LOG and hands it to Manager.SetDecisionLogger.
type DecisionLogger interface {
	OpenGame(gameID uuid.UUID) (GameDecisionLog, error)
}

// DecisionLoggerFunc adapts a function to DecisionLogger.
type DecisionLoggerFunc func(gameID uuid.UUID) (GameDecisionLog, error)

// OpenGame calls f.
func (f DecisionLoggerFunc) OpenGame(gameID uuid.UUID) (GameDecisionLog, error) { return f(gameID) }
