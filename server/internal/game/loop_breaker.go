package game

import "github.com/google/uuid"

// loop_breaker.go — the CR 726 shortcut, engine side (#628).
//
// Replacement loops are capped (ErrReplacementIterationExceeded).
// Trigger loops were not. A triggered ability goes on the stack and
// resolves once every player has passed in succession, so the server
// does one bounded unit of work per pass and never blocks — but two
// "whenever a creature enters, create a token" permanents trigger
// each other forever, at whatever speed the table passes. With
// autopass on for every seat that is a tight client-driven loop:
// pass → resolve → broadcast → autopass → pass, until somebody
// notices and turns autopass off. The game never ends.
//
// CR 726 is the paper answer. Players take a shortcut: the loop's
// controller says how many more times it happens and the table skips
// straight there, and a mandatory loop nobody can stop is a draw
// (CR 726.4). Both halves need a player to SAY something, so the
// engine's job is to hand priority back to the humans with the
// loop's trigger still on the stack — not to stop the game, not to
// refuse the pass.
//
// So this file does exactly one thing: it notices, and it says so.
//
//   - Detection is the count already on TurnTally. Resolved counts
//     every stack item that resolved this turn keyed by
//     TallyKey(source, label); LoopRun is the same count restarted
//     at each player decision. loopSuspectedLocked is the whole
//     rule, and it is the only place the threshold is read.
//   - The notice is Game.LoopNotice plus one EventLoopSuspected.
//     While it is set, AUTOMATIC passing is suspended: the client's
//     autopass toggle stops firing (client/src/lib/priority.ts) and
//     a bot runner holds rather than dispatch another pass
//     (internal/aiseat/runner.go). Every seat still holds priority
//     and every pass_priority the engine is handed still works — a
//     player who wants to watch the loop run clicks "next".
//   - notePlayerDecisionLocked clears both. A decision is a cast, an
//     activation, an answered prompt, or an attack / block
//     declaration; a bare pass is not one, which is what keeps the
//     notice up while the loop is being stepped through by hand.
//
// Bots count as automatic, deliberately. A table of four bots
// spinning on a real loop is the same runaway as four browsers on
// autopass, and the same answer applies: the table stops with the
// notice in the game state saying which ability and how many times.
// A bot that has a real (non-pass) move may still play it, and doing
// so clears the notice like any other decision — so the breaker
// never wedges a table that still has something to do.
//
// See ADR 0055.

// DefaultLoopThreshold is how many times one triggered ability may
// resolve in a single turn, with no player decision in between,
// before the engine calls it a loop.
//
// 25 is the issue's number (#628) and it is chosen to be
// unreachable by accident rather than to be tight. It counts
// resolutions of ONE TallyKey — one ability of one permanent — in
// ONE turn: four upkeep triggers from four players are four
// different keys and never approach it, and the longest real
// aristocrats or storm turn recorded in the bot soaks does not put
// the same ability on the stack twenty-five times without its
// controller casting or activating something in between. A loop, by
// contrast, passes it on the second or third second of wall clock.
const DefaultLoopThreshold = 25

// LoopNotice is the engine's "this looks like a loop" flag: which
// ability, whose, and how many times it has resolved since the last
// player decision. Nil when nothing is suspected.
//
// Pure data, and it has to stay that way — it rides the undo clone
// and the persisted snapshot by value.
type LoopNotice struct {
	// Source is the permanent whose triggered ability is repeating.
	Source uuid.UUID `json:"source"`
	// Label is that ability's stack label, which by catalog
	// convention reads "<card> — <what happens>" and is what the
	// client puts in the banner.
	Label string `json:"label"`
	// Controller is the seat that controls the repeating ability —
	// the player CR 726 would have name the number of iterations.
	Controller uuid.UUID `json:"controller,omitempty"`
	// Count is how many times it has resolved in the current run. It
	// keeps rising if the table steps the loop on by hand; the
	// EventLoopSuspected breadcrumb is emitted once, when the notice
	// is first raised.
	Count int `json:"count"`
}

// loopThresholdLocked is the count this game trips at: the per-game
// override when set, DefaultLoopThreshold otherwise. Caller must
// hold g.mu.
func (g *Game) loopThresholdLocked() int {
	if g.LoopThreshold > 0 {
		return g.LoopThreshold
	}
	return DefaultLoopThreshold
}

// loopSuspectedLocked is the detector, and the only reader of the
// threshold: has the ability keyed by `key` resolved enough times
// in this turn, with no player decision in between, to be a loop?
//
// One function on purpose. Everything else in this file is
// bookkeeping around this question, and no other package asks it.
//
// Caller must hold g.mu.
func (g *Game) loopSuspectedLocked(key string) bool {
	return g.TurnTally.LoopRun[key] >= g.loopThresholdLocked()
}

// noteResolutionForLoopLocked records one resolution of a stack item
// against the current run and raises the notice when that run
// crosses the threshold. Called from turnTallyListener on
// EventResolve, next to the Resolved bump it shares a key with.
//
// Caller must hold g.mu.
func (g *Game) noteResolutionForLoopLocked(ev Event, key string) {
	if g.TurnTally.LoopRun == nil {
		g.TurnTally.LoopRun = map[string]int{}
	}
	g.TurnTally.LoopRun[key]++
	if !g.loopSuspectedLocked(key) {
		return
	}
	if g.LoopNotice != nil {
		// A notice already stands, so the table has already been told
		// to stop passing automatically and there is nothing new to
		// say. Keep the count live for the banner when it is the same
		// ability, and leave the notice naming the ability that
		// tripped first when it is the loop's other half — a loop
		// between two permanents would otherwise raise a fresh notice
		// (and a fresh breadcrumb) on every single resolution.
		if g.LoopNotice.Source == ev.Source && g.LoopNotice.Label == ev.Label {
			g.LoopNotice.Count = g.TurnTally.LoopRun[key]
		}
		return
	}
	g.LoopNotice = &LoopNotice{
		Source:     ev.Source,
		Label:      ev.Label,
		Controller: ev.Actor,
		Count:      g.TurnTally.LoopRun[key],
	}
	g.EmitEvent(Event{
		Kind:   EventLoopSuspected,
		Actor:  ev.Actor,
		Source: ev.Source,
		Label:  ev.Label,
		Amount: g.LoopNotice.Count,
	})
}

// notePlayerDecisionLocked records that a player decided something
// rather than passing: a spell cast, an ability activated, a prompt
// answered, an attacker or blocker declared. It restarts every run
// and clears any standing notice, which is what "automatic passing
// is suspended until a human acts" means in code.
//
// A bare pass_priority is deliberately NOT a decision. If it were,
// the first manual "next" click would clear the notice and four
// autopassing clients would immediately spin the loop up again.
// Leaving the notice up lets the table step the loop by hand for as
// long as it wants and still hands priority back every iteration.
//
// Caller must hold g.mu.
func (g *Game) notePlayerDecisionLocked() {
	g.TurnTally.LoopRun = nil
	g.LoopNotice = nil
}

// AutoPassSuspended reports whether the CR 726 breaker has paused
// automatic passing. Read by the bot runner before it dispatches a
// pass; the client reads the same fact off GameView.loop_notice.
func (g *Game) AutoPassSuspended() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.LoopNotice != nil
}

// CurrentLoopNotice returns a copy of the standing loop notice, or
// nil. A copy, so a caller outside the lock cannot read a Count that
// is being bumped underneath it.
func (g *Game) CurrentLoopNotice() *LoopNotice {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return cloneLoopNotice(g.LoopNotice)
}

// cloneLoopNotice deep-copies the notice (it is all value fields, so
// this is a shallow copy behind a fresh pointer).
func cloneLoopNotice(n *LoopNotice) *LoopNotice {
	if n == nil {
		return nil
	}
	out := *n
	return &out
}
