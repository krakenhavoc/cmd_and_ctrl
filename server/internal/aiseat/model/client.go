// Package model is ADR 0033 §5's Layer C: the model-backed policy,
// and the funnel that keeps most windows away from it.
//
// # The shape of a decision
//
//	Decide
//	 ├─ Layer A (aiseat/rules)  forced / trivial  → answer, no cost
//	 ├─ Layer B (aiseat/heuristic)                → always computed,
//	 │                                              in microseconds
//	 ├─ escalation triggers fire?                 → frontier model
//	 ├─ otherwise                                 → cheap model
//	 └─ model absent / slow / wrong               → Layer B's answer
//
// Layer B is computed BEFORE the model is called, not after it fails.
// That is what makes every failure mode — outage, timeout, malformed
// output, an index that is not a move — cost nothing but the latency
// already spent: there is always an answer in hand. ADR 0033 §10 is
// explicit that the table never waits on a bot, and the only way to
// honour a hard deadline with a network call inside it is to have
// already decided what to do when the call does not land.
//
// # The model selects, it does not act
//
// The model is shown a numbered list of `legal.Move` labels and asked
// for one number. It cannot emit an action, so it cannot invent one;
// an index outside the list is simply discarded. This is the
// structural defence ADR 0033 §1 builds the enumerator for, and it is
// why a model is safe here at all.
//
// # The type gate
//
// Like every policy package under aiseat/, this one may not import
// internal/game (ADR 0033 §3). Prompt assembly reads aiseat.Input —
// the seat's own filtered protocol.GameView and the closed move list
// — and the static half of the prompt comes from a Config the seat
// was constructed with. Nothing in this package can reach an
// opponent's hand, so nothing it sends to a model can leak one.
// TestPolicyPackagesDoNotImportGame in aiseat/heuristic enforces it.
package model

import (
	"context"
	"errors"
)

// Block is one system-prompt block. Cache places a prompt-cache
// breakpoint at the end of it.
type Block struct {
	Text string
	// Cache marks this block as the end of the stable prefix. The
	// caching contract is a PREFIX match: every byte before the
	// breakpoint must be identical call to call or nothing is
	// reused, which is why the board state travels in User and never
	// in a System block.
	Cache bool
}

// Request is one model call. It is deliberately small — this is a
// select-an-index call, not a conversation: no history, no tools, no
// streaming, one turn.
type Request struct {
	// Model is the provider's model id.
	Model string
	// System is the static, prompt-cached half: rules primer,
	// decklist with oracle text, archetype plan.
	System []Block
	// User is the per-decision delta: board state and the move list.
	User string
	// MaxTokens caps the reply. A reply is a JSON object with an
	// integer in it, so this is small on purpose.
	MaxTokens int
	// Effort maps to output_config.effort where the model supports
	// it. Empty omits the field — which is required for models that
	// reject it.
	Effort string
	// Thinking is "", "adaptive" or "disabled". Empty omits the
	// field and takes the model's default.
	Thinking string
}

// Usage is what one call cost, as the provider reported it.
type Usage struct {
	InputTokens      int
	OutputTokens     int
	CacheReadTokens  int
	CacheWriteTokens int
}

// Response is one model reply.
type Response struct {
	// Text is the concatenated text content of the reply.
	Text string
	// Model is the model that actually answered.
	Model string
	// StopReason is the provider's stop reason, for the log.
	StopReason string
	Usage      Usage
}

// Client is the model transport, and the whole of Layer C's contact
// with the outside world. It is an interface for two reasons: the
// obvious one is testing, and the load-bearing one is that a policy
// that cannot reach a model must still be a complete policy. Every
// test in this package that is not about HTTP runs against a fake,
// and the outage drill runs against a Client that only ever fails.
//
// Implementations must respect ctx. The caller gives them a budget
// strictly smaller than the runner's MaxThink; one that overruns it
// has taken the table's time, which ADR 0033 §10 does not allow.
type Client interface {
	// Name identifies the transport in logs.
	Name() string
	// Complete makes one call.
	Complete(ctx context.Context, req Request) (Response, error)
}

// ErrNoClient is returned when a policy configured for a model tier
// has no Client. It is not fatal: the policy runs as Layer A + B.
var ErrNoClient = errors.New("aiseat/model: no model client configured")

// ErrOutage is what a Client returns when the endpoint is
// unreachable. Nothing in this package treats it specially — every
// error takes the same path to Layer B — but it makes the outage
// drill readable.
var ErrOutage = errors.New("aiseat/model: model endpoint unreachable")

// ErrNoBudget means the runner's deadline left too little time to
// attempt a call. The decision is Layer B's and no call is made.
var ErrNoBudget = errors.New("aiseat/model: no time left to call a model")
