package aiseat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// spend.go is #735: what a game of Magic against the model tiers
// actually costs, measured rather than estimated.
//
// # Why this is its own thing and not a line in Stats
//
// S31's exit criterion 4 has two halves. "Layer A absorbs more than
// 80% of priority windows" was measured at 91.3% and closed. "Per-game
// model spend is measured and recorded, not estimated" was never done,
// and ADR 0033 §5 still carries an order-of-magnitude GUESS — "~30–50
// real decisions per bot per game, ~20% escalating, that is cents per
// game" — whose escalation half has already been shown wrong by a
// factor of three. An argument nobody can check is not an argument.
//
// The numbers existed; they were unreachable. model.Policy has
// counted its own tokens since sub-PR 7 and its improvisation tokens
// since #686, but a model.Stats lives inside one policy object that
// only the seat's own runner holds, the improvisation half is not on
// the decision-log path at all (the improvise path is deliberately
// not observed), and nothing anywhere adds four seats together and
// says what the TABLE cost. This file is that addition, in one shape,
// read at one moment: when a game's bots are done.
//
// # Decision spend and improvisation spend are never added up here
//
// They are two different questions with two different caps, and ADR
// 0033's §8 amendment is explicit that improvisation is bounded
// separately (MaxImprovCalls, 8 per seat per game) precisely so that
// it can be reasoned about on its own. A decision call is an index
// out of a closed list against a cached prefix; an improvisation call
// is a paragraph written from oracle text, with MaxTokens an order of
// magnitude larger, a handful of times a game. Averaging them into
// one tokens-per-call number would make both unreadable, so the split
// survives all the way to the admin line, and Total() is there for
// the one question — what did the table cost — that wants them added.

// PurposeSpend is one seat's spend on one PURPOSE: deciding, or
// improvising.
type PurposeSpend struct {
	// Calls is how many model calls were attempted. Attempted rather
	// than succeeded, deliberately: a call that timed out or came
	// back malformed was still billed.
	Calls int64 `json:"calls"`
	// Usage totals the provider's reported token counts over those
	// calls. Zero from a provider that reports none — which is a
	// measurement that could not be taken, not a call that was free.
	Usage TokenUsage `json:"usage"`
	// Latency totals the wall clock spent inside them.
	Latency time.Duration `json:"latency_ns,omitempty"`
}

// Add folds o into p.
func (p PurposeSpend) Add(o PurposeSpend) PurposeSpend {
	p.Calls += o.Calls
	p.Usage = p.Usage.Add(o.Usage)
	p.Latency += o.Latency
	return p
}

// Empty reports whether anything was spent at all — true for every
// `random` and `heuristic` seat, and for a model seat on a server
// with no transport.
func (p PurposeSpend) Empty() bool { return p == PurposeSpend{} }

// Add folds o into u.
func (u TokenUsage) Add(o TokenUsage) TokenUsage {
	u.InputTokens += o.InputTokens
	u.OutputTokens += o.OutputTokens
	u.CacheReadTokens += o.CacheReadTokens
	u.CacheWriteTokens += o.CacheWriteTokens
	u.CachedPromptTokens += o.CachedPromptTokens
	return u
}

// Spend is one seat's model spend for the life of the seat, split by
// purpose. See the file comment for why the split is not collapsed.
type Spend struct {
	// Decision is the funnel's Layer C: one call per window that
	// survived Layer A and Layer B and had budget left.
	Decision PurposeSpend `json:"decision"`
	// Improvisation is ADR 0033 §8's bundle writing, hard-capped per
	// seat per game by model.Config.MaxImprovCalls.
	Improvisation PurposeSpend `json:"improvisation"`
}

// Total is the two purposes added — the number that answers "what did
// this seat cost".
func (s Spend) Total() PurposeSpend { return s.Decision.Add(s.Improvisation) }

// Add folds o into s.
func (s Spend) Add(o Spend) Spend {
	s.Decision = s.Decision.Add(o.Decision)
	s.Improvisation = s.Improvisation.Add(o.Improvisation)
	return s
}

// Empty reports whether this seat spent nothing.
func (s Spend) Empty() bool { return s == Spend{} }

// Spender is an optional Policy extension: a policy that can reach a
// model reports what it has spent so far.
//
// Optional for the reason every other extension here is: `random` and
// `heuristic` have no model and nothing to report, and a required
// method would make every future policy implement a stub. A seat whose
// policy is not a Spender contributes a zero Spend, which is the true
// answer.
//
// Read from another goroutine while the seat plays (Runner.Stats does
// exactly that), so an implementation must be safe for concurrent
// reads.
type Spender interface {
	Spend() Spend
}

// SeatSpend is one seat's line in a game's spend record.
type SeatSpend struct {
	Seat uuid.UUID `json:"seat"`
	// Policy is Runner.PolicyName() — the tier as it actually played,
	// not as it was labelled.
	Policy string `json:"policy"`
	Spend  Spend  `json:"spend"`
}

// GameSpend is ONE record per game: every bot seat at the table, what
// each spent, split by purpose. Built when the game's runners are
// done, logged as the admin summary line and written to the decision
// log.
//
// One per game rather than one per seat because the question S31 asks
// is "what does a game cost", and four files a reader has to add up
// is an answer nobody will compute twice.
type GameSpend struct {
	Game uuid.UUID `json:"game"`
	// Seats are in the order the runners were started, which is seat
	// order at the table.
	Seats []SeatSpend `json:"seats"`
}

// Total is the whole table's spend.
func (g GameSpend) Total() Spend {
	var out Spend
	for _, s := range g.Seats {
		out = out.Add(s.Spend)
	}
	return out
}

// Empty reports whether the table spent nothing — a four-heuristic
// game, or a model table on a server with no transport.
func (g GameSpend) Empty() bool { return g.Total().Empty() }

// Tiers is the tiers that played, most seats first, for the one-line
// summary: "assisted x2, heuristic".
func (g GameSpend) Tiers() string {
	by := map[string]int{}
	for _, s := range g.Seats {
		name := s.Policy
		if name == "" {
			name = "unknown"
		}
		by[name]++
	}
	names := make([]string, 0, len(by))
	for k := range by {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool {
		if by[names[i]] != by[names[j]] {
			return by[names[i]] > by[names[j]]
		}
		return names[i] < names[j]
	})
	parts := make([]string, 0, len(names))
	for _, n := range names {
		if by[n] > 1 {
			parts = append(parts, fmt.Sprintf("%s x%d", n, by[n]))
			continue
		}
		parts = append(parts, n)
	}
	return strings.Join(parts, ", ")
}

// PerSeat is the per-seat breakdown as one short string, for the
// admin line: "assisted 31c/8i 41200in/1830out".
//
// It is a string rather than structured attributes because the admin
// summary is ONE line and a four-seat table would otherwise carry
// forty key-value pairs into it. The structured form is the record
// written to the decision log, which is where a tool should read it.
func (g GameSpend) PerSeat() string {
	parts := make([]string, 0, len(g.Seats))
	for _, s := range g.Seats {
		t := s.Spend.Total()
		parts = append(parts, fmt.Sprintf("%s %dc/%di %din/%dout",
			s.Policy, s.Spend.Decision.Calls, s.Spend.Improvisation.Calls,
			t.Usage.InputTokens, t.Usage.OutputTokens))
	}
	return strings.Join(parts, "; ")
}

// LogAttrs is the admin summary line's payload: the whole game's
// spend as slog key-value pairs.
//
//	log.Info("bot model spend for the game", gs.LogAttrs()...)
//
// It lives here rather than in the Manager so that the arena, a test
// and the server all print the same line, and so that adding a field
// to Spend cannot leave one caller behind.
func (g GameSpend) LogAttrs() []any {
	total := g.Total()
	sum := total.Total()
	return []any{
		"game", g.Game.String(),
		"seats", len(g.Seats),
		"tiers", g.Tiers(),
		"calls", sum.Calls,
		"decision_calls", total.Decision.Calls,
		"improv_calls", total.Improvisation.Calls,
		"input_tokens", sum.Usage.InputTokens,
		"output_tokens", sum.Usage.OutputTokens,
		"cache_read_tokens", sum.Usage.CacheReadTokens,
		"cache_write_tokens", sum.Usage.CacheWriteTokens,
		"cached_prompt_tokens", sum.Usage.CachedPromptTokens,
		"model_time", sum.Latency.Round(time.Millisecond).String(),
		"per_seat", g.PerSeat(),
	}
}

// SpendOfRunners reads one record off a game's runners. Runners that
// have exited are still readable, which is the point: this is called
// after every seat is done.
func SpendOfRunners(gameID uuid.UUID, runners []*Runner) GameSpend {
	gs := GameSpend{Game: gameID, Seats: make([]SeatSpend, 0, len(runners))}
	for _, r := range runners {
		if r == nil {
			continue
		}
		gs.Seats = append(gs.Seats, SeatSpend{
			Seat:   r.Seat(),
			Policy: r.PolicyName(),
			Spend:  r.Stats().Spend,
		})
	}
	return gs
}

// SpendObserver is an optional GameDecisionLog extension: a log that
// implements it is handed the game's one spend record before it is
// closed.
//
// Optional, and asserted rather than added to GameDecisionLog, for
// the reason every other extension in this package is optional: the
// Manager holds an interface so that aiseat never imports
// aiseat/decisionlog (which reads aiseat's own types, so the
// dependency would be a cycle), and widening that interface would
// break every test double that implements it.
type SpendObserver interface {
	ObserveSpend(gs GameSpend)
}
