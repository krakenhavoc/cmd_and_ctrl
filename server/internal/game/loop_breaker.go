package game

import "github.com/google/uuid"

// loop_breaker.go — the CR 732 shortcut, engine side (#628).
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
// CR 732 is the paper answer. Players take a shortcut: the loop's
// controller says how many more times it happens and the table skips
// straight there, and a mandatory loop nobody can stop is a draw
// (CR 732.4). Both halves need a player to SAY something, so the
// engine's job is to hand priority back to the humans with the
// loop's trigger still on the stack — not to stop the game, not to
// refuse the pass.
//
// So this file notices, says so, and — since #804 — asks the one
// question CR 732 says the controller gets to answer. The shortcut
// half is at the bottom of the file; the detection half is here.
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
//   - notePlayerDecisionLocked clears both, and withdraws any
//     standing shortcut prompt. A decision is a cast, an activation,
//     an answered prompt, or an attack / block declaration; a bare
//     pass is not one, which is what keeps the notice up while the
//     loop is being stepped through by hand.
//
// Bots count as automatic, deliberately. A table of four bots
// spinning on a real loop is the same runaway as four browsers on
// autopass, and the same answer applies: the table stops with the
// notice in the game state saying which ability and how many times.
// A bot that has a real (non-pass) move may still play it, and doing
// so clears the notice like any other decision — so the breaker
// never wedges a table that still has something to do. A bot that
// controls the loop also answers the shortcut prompt: ten more
// iterations the first time it is asked in a turn, stop the second
// (internal/legal, docs/bot.md), so a bot-only table is bounded
// rather than merely stopped.
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
//
// #2450 found the exception: one event can trigger an ability dozens
// of times (CR 603.2c), and a wrath across a token board under Syr
// Konrad and Exquisite Blood got there in about nine deaths. That is
// handled by telling a draining batch from a loop (noteLoopWorkLocked),
// not by raising this number; the owner kept it at 25.
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
	// the player CR 732 would have name the number of iterations.
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
	// #2450: a batch of triggers that is draining is not a loop.
	// Measured before this resolution is counted, so a key that
	// resolves at a new low starts its run again at one.
	g.noteLoopWorkLocked(ev)
	if g.TurnTally.LoopRun == nil {
		g.TurnTally.LoopRun = map[string]int{}
	}
	g.TurnTally.LoopRun[key]++
	if !g.loopSuspectedLocked(key) {
		return
	}
	// #804, CR 732. The loop's controller may already have said how
	// many more iterations they want, and while that number is being
	// counted down there is nothing to tell the table.
	spent, exhausted := g.spendLoopAllowanceLocked(key)
	switch {
	case spent && !exhausted:
		return
	case !spent && g.loopShortcutRunningLocked():
		// The other half of a loop running under a shortcut its
		// controller agreed to. A two-permanent loop is two keys and
		// one conversation: raising a notice for the half nobody was
		// asked about would stop the table short of the number they
		// named.
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
	// `exhausted` here means this resolution was the last one the
	// controller asked for, so the prompt they answered is the one
	// being offered again. The flag rides the prompt so the bot
	// policy can tell a first ask from a second (docs/bot.md).
	g.queueLoopShortcutLocked(key, exhausted)
}

// spendLoopAllowanceLocked charges one resolution of `key` against the
// CR 732 shortcut its controller agreed to.
//
//	spent     — an allowance existed and this resolution used one of it.
//	exhausted — that was the last one: the table has now had exactly
//	            the K resolutions it was promised, and the question
//	            goes back to the controller.
//
// Exhausting one shortcut ends every shortcut. At most one is ever
// live in practice (the answer clears the tally's other runs), and a
// table that has just been stopped should not still be counting down
// somebody's leftover number.
//
// Caller must hold g.mu.
func (g *Game) spendLoopAllowanceLocked(key string) (spent, exhausted bool) {
	n, ok := g.TurnTally.LoopAllowance[key]
	if !ok || n <= 0 {
		return false, false
	}
	if n--; n > 0 {
		g.TurnTally.LoopAllowance[key] = n
		return true, false
	}
	g.TurnTally.LoopAllowance = nil
	return true, true
}

// loopShortcutRunningLocked reports whether any CR 732 shortcut still
// has iterations left to run. Caller must hold g.mu.
func (g *Game) loopShortcutRunningLocked() bool {
	return len(g.TurnTally.LoopAllowance) > 0
}

// grantLoopShortcutLocked records the shortcut the loop's controller
// proposed (CR 732): `iterations` more resolutions of `key`, and then
// the same question again.
//
// It re-arms LoopRun as well as writing the allowance, and that is the
// whole subtlety of #804. Answering a prompt is a player decision, and
// a decision restarts every run (notePlayerDecisionLocked) — right for
// every other prompt and exactly wrong for this one: with the run back
// at zero nothing would consult the allowance until the ability had
// resolved a further LoopThreshold times, so "resolve it 3 more times"
// would mean 28. Putting this key's run back where the notice found it
// makes the next resolution spend allowance, and K more resolutions is
// then exactly what happens. Every OTHER key stays cleared: the
// decision was real, and the shortcut is about one ability.
//
// `iterations` of zero re-arms the run and grants nothing, which is
// what "stop here" has to do: the table goes back to exactly where the
// breaker left it — the notice standing at its count, and the count
// still climbing if the players step the loop on by hand.
//
// Caller must hold g.mu.
func (g *Game) grantLoopShortcutLocked(key string, at, iterations int) {
	if key == "" {
		return
	}
	if threshold := g.loopThresholdLocked(); at < threshold {
		at = threshold
	}
	if g.TurnTally.LoopRun == nil {
		g.TurnTally.LoopRun = map[string]int{}
	}
	g.TurnTally.LoopRun[key] = at
	if iterations <= 0 {
		return
	}
	if g.TurnTally.LoopAllowance == nil {
		g.TurnTally.LoopAllowance = map[string]int{}
	}
	g.TurnTally.LoopAllowance[key] = iterations
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
	// #2450: the low is "since the last decision", and so is the
	// activation whose run a draining batch leaves alone.
	g.TurnTally.LoopLow, g.TurnTally.LoopLowSet = 0, false
	g.TurnTally.LoopActivated = ""
	g.TurnTally.LoopProgressed = false
	// #804: a real decision ends any CR 732 shortcut too. The
	// controller agreed to K more iterations of a loop that nobody was
	// doing anything about; somebody has now done something, so the
	// board the number was named against is gone. ResolveLoopShortcut
	// grants the allowance AFTER this runs, which is what lets the one
	// decision that IS the shortcut set it.
	g.TurnTally.LoopAllowance = nil
	g.LoopNotice = nil
	g.dropLoopShortcutPromptsLocked()
}

// notePlayerActivationLocked is notePlayerDecisionLocked for the one
// decision that can be the loop (#810).
//
// Activating an ability is a player decision and stays one: it
// restarts every run and clears the notice, exactly as casting a spell
// does. The exception is the ability being activated. ADR 0055 §3
// counted an activation as a decision without qualification, and for a
// TRIGGER loop that is right — a trigger loop runs itself, so anything
// its controller does instead is progress. An ACTIVATION loop is the
// other shape: the decision IS the loop. Soothsaying's "{X}: Look at
// the top X cards" at X=0 costs nothing, does nothing, and can be
// taken again the instant it resolves, and the run it should be
// counted against was being cleared by the very activation that added
// to it — so `LoopRun` never got past 1 and the breaker never fired
// while a bot table spun for five minutes (#810).
//
// So this key's run survives, and every other key's is cleared. The
// resolution that follows bumps it (turnTallyListener, EventResolve),
// and a repeated free activation reaches the threshold in exactly the
// number of iterations the threshold names.
//
// The CR 732 allowance for this key survives with it, for the same
// reason grantLoopShortcutLocked re-arms the run rather than letting
// the answer reset it: a shortcut its controller agreed to is about
// this ability, and the activations that spend it must not be what
// cancels it.
//
// `key` is TallyKey(source, label) for the ability being activated; an
// empty key falls back to the ordinary decision.
//
// Caller must hold g.mu.
func (g *Game) notePlayerActivationLocked(key string) {
	if key == "" {
		g.notePlayerDecisionLocked()
		return
	}
	run := g.TurnTally.LoopRun[key]
	allowance, granted := g.TurnTally.LoopAllowance[key]
	g.notePlayerDecisionLocked()
	if run > 0 {
		g.TurnTally.LoopRun = map[string]int{key: run}
	}
	if granted {
		g.TurnTally.LoopAllowance = map[string]int{key: allowance}
	}
	g.TurnTally.LoopActivated = key
}

// ---------------------------------------------------------------
// A batch is not a loop (#2450, ADR 0055's 2026-10-07 amendment, A)
// ---------------------------------------------------------------
//
// CR 603.2c: "An ability triggers only once each time its trigger
// event occurs. However, it can trigger repeatedly if one event
// contains multiple occurrences." A wrath across a token board with
// Syr Konrad out puts thirty Konrad triggers on the stack at once,
// and with Exquisite Blood beside it each of those triggers three
// more. Nothing in that is a set of actions that "could be repeated
// indefinitely" (CR 732.1b), but it is twenty-five resolutions of one
// ability with no decision in between, which is all
// loopSuspectedLocked asks. Three bot-only tables stopped on it.
//
// A loop refills the stack and a batch empties it. So as each
// TRIGGERED ability begins to resolve, the engine counts the work
// still waiting (noteLoopWorkLocked) and keeps the lowest count since
// the last decision. A count strictly below that low means the table
// is working through something finite, and every run restarts
// (noteLoopProgressLocked). A real loop replaces what it resolves, so
// it never makes a new low and trips at the threshold as before. The
// lows are whole numbers that only fall between decisions, so the
// restarts are finite too: an endless trigger sequence still trips.
//
// Only triggered resolutions are measured. A spell that resolves
// leaves an empty stack behind it (the wrath itself), which would set
// a low no batch it causes could ever beat. An activated ability is
// #810's shape, fed one activation at a time, and its run is kept
// across a draining batch (TurnTally.LoopActivated).

// loopWorkLocked counts the stack items still to resolve and the
// triggers waiting to join them. Caller must hold g.mu.
func (g *Game) loopWorkLocked() int {
	n := len(g.PendingTriggers)
	for _, item := range g.StackMeta {
		if item != nil {
			n++
		}
	}
	return n
}

// noteLoopWorkLocked measures the work left as a triggered ability
// begins to resolve, and restarts the runs when it is a new low.
// `ev` is the EventResolve; the resolving item has already left
// StackMeta. Caller must hold g.mu.
func (g *Game) noteLoopWorkLocked(ev Event) {
	if g.resolving == nil || g.resolving.item == nil {
		return
	}
	item := g.resolving.item
	if item.Kind != StackItemTriggered || item.ID != ev.ResolvedStackItemID {
		return
	}
	work := g.loopWorkLocked()
	t := &g.TurnTally
	if t.LoopProgressed {
		// Option B: a life total or a poison count moved toward a
		// loss since the last triggered resolution.
		t.LoopProgressed = false
		if !t.LoopLowSet || work < t.LoopLow {
			t.LoopLow, t.LoopLowSet = work, true
		}
		g.noteLoopProgressLocked()
		return
	}
	if !t.LoopLowSet {
		t.LoopLow, t.LoopLowSet = work, true
		return
	}
	if work >= t.LoopLow {
		return
	}
	t.LoopLow = work
	g.noteLoopProgressLocked()
}

// ---------------------------------------------------------------
// A loop that is ending the game keeps stepping (#2450, option B)
// ---------------------------------------------------------------
//
// CR 704.5a: "If a player has 0 or less life, that player loses the
// game." CR 104.2a: "A player still in the game wins the game if that
// player's opponents have all left the game." Exquisite Blood with
// Marauding Blight-Priest is a mandatory loop, and every iteration takes
// life from each opponent until they are dead: it ends, so it is not
// CR 104.4b's loop "with no way to stop", and CR 732.4's draw is not
// its answer. Option A does not see it, because the loop adds more
// triggers than it resolves.
//
// So a run also restarts on progress that can only happen finitely
// often (owner decision 3): a player leaving the game, a player
// reaching a new lowest life total this turn, or a new highest poison
// count this turn. A life total makes only so many new lows before
// CR 704.5a takes its player; poison stops at ten (CR 122.1f). A player
// who can't lose to that cause (ADR 0057's gate, owner decision 4)
// makes no progress that way, or a drain into a Platinum Angel would
// run for ever. A player's first loss of a turn counts as a new low:
// one per player per turn, which keeps the bound.
//
// A new low or a new high is marked (TurnTally.LoopProgressed) and
// applied as the next triggered ability begins to resolve, where
// option A measures too: a loss in the middle of a resolution is part
// of that iteration, and the runs restart before the next one is
// counted. A player leaving restarts them at once (leaveGameLocked).

// noteLoopLifeLowLocked records player's life total after a loss, and
// restarts the loop runs when it is the lowest this turn. Caller must
// hold g.mu.
func (g *Game) noteLoopLifeLowLocked(player uuid.UUID) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated || !g.canLoseLocked(p, LossLife) {
		return
	}
	progress := false
	g.bumpPlayerTally(player, func(t *PlayerTurnTally) {
		if !t.LifeLowSet || p.Life < t.LifeLow {
			t.LifeLow, t.LifeLowSet, progress = p.Life, true, true
		}
	})
	if progress {
		g.TurnTally.LoopProgressed = true
	}
}

// noteLoopPoisonHighLocked records player's poison count after it
// rose, and restarts the loop runs when it is the highest this turn.
// Caller must hold g.mu.
func (g *Game) noteLoopPoisonHighLocked(player uuid.UUID, poison int) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated || !g.canLoseLocked(p, LossPoison) {
		return
	}
	progress := false
	g.bumpPlayerTally(player, func(t *PlayerTurnTally) {
		if poison > t.PoisonHigh {
			t.PoisonHigh, progress = poison, true
		}
	})
	if progress {
		g.TurnTally.LoopProgressed = true
	}
}

// noteLoopProgressLocked restarts the runs because the table showed
// it is not looping. Owner decision 5 of the 2026-10-07 amendment: it
// clears the notice and withdraws a standing shortcut prompt as a
// decision does, and counts as a decision for nothing else. The low,
// the activation key and the run of that activation are kept: a
// batch draining between two activations is not a reason to stop
// counting the activations. A notice that names that activation
// stays up for the same reason.
//
// Caller must hold g.mu.
func (g *Game) noteLoopProgressLocked() {
	t := &g.TurnTally
	keep := t.LoopActivated
	run := t.LoopRun[keep]
	allowance, granted := t.LoopAllowance[keep]
	t.LoopRun, t.LoopAllowance = nil, nil
	if keep != "" && run > 0 {
		t.LoopRun = map[string]int{keep: run}
	}
	if keep != "" && granted {
		t.LoopAllowance = map[string]int{keep: allowance}
	}
	if n := g.LoopNotice; n != nil && (keep == "" || TallyKey(n.Source, n.Label) != keep) {
		g.LoopNotice = nil
		g.dropLoopShortcutPromptsLocked()
	}
}

// dropLoopShortcutPromptsLocked withdraws any outstanding CR 732
// shortcut prompt (#804). Called from notePlayerDecisionLocked,
// because the question it asks — "this loop is going nowhere, how many
// more times?" — is moot the moment somebody casts, activates,
// declares or answers something else: the notice it belongs to has
// just been cleared for the same reason.
//
// Moot is not the whole of it. The prompt BLOCKS the table
// (choice_gate.go), so one left behind by a cleared notice is not a
// stale question, it is a wedged game.
//
// dropChoiceLocked, not dequeueChoiceLocked: the engine is withdrawing
// a prompt nobody answered, and counting that as a decision is what
// ADR 0055 §3's prune rule exists to prevent.
//
// Caller must hold g.mu.
func (g *Game) dropLoopShortcutPromptsLocked() {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == PendingChoiceLoopShortcut {
			g.dropChoiceLocked(i)
		}
	}
}

// AutoPassSuspended reports whether the CR 732 breaker has paused
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

// ---------------------------------------------------------------
// The CR 732 shortcut prompt (#804)
// ---------------------------------------------------------------
//
// ADR 0055 §6 deferred this half: the breaker stops the loop and hands
// priority back, but the table's only ways on were to click through
// every further iteration by hand or to re-enable autopass and be
// paused again a threshold later. CR 732 is the conversation paper
// has instead — the loop's controller says how many more times it
// happens and the table skips there — and this is that conversation,
// as one prompt with a number on it.
//
// Three things make it work, and each is deliberately small:
//
//   - The prompt goes to the CONTROLLER of the ability the notice
//     names. CR 732 is their proposal to make; nobody else at the
//     table knows what the loop is for.
//   - The answer is an allowance on the tally
//     (TurnTally.LoopAllowance), not a second detector. The detector
//     is still loopSuspectedLocked; the allowance only decides whether
//     its answer is worth saying out loud. Every resolution spends
//     one, and the last one spent raises the notice again and re-asks.
//   - The prompt BLOCKS the table, like almost every other prompt
//     (choice_gate.go). That is a real change from ADR 0055 §4, which
//     was careful to refuse no passes — and it is right here where it
//     would be wrong for a Rhystic tax: the shortcut is proposed while
//     the loop's trigger is still on the stack, and the whole point of
//     the answer is how many times that trigger resolves next. A table
//     that passed through the question would answer it by doing.

// PendingChoiceLoopShortcut is the CR 732 shortcut: "<card> —
// <ability> has resolved N times this turn. Resolve it K more times,
// then stop?" Answered with an iteration count, 0 meaning "stop here"
// — which leaves the notice standing and automatic passing paused,
// exactly where the breaker put it.
//
// Queued to the loop's controller when the notice is raised, and only
// then: it is not something a player can ask for out of the blue.
const PendingChoiceLoopShortcut PendingChoiceKind = "loop_shortcut"

// MaxLoopShortcutIterations caps the answer. It is a sanity bound on a
// number a player types, not a rules limit: CR 732 lets a shortcut
// name any finite number, and a table that wants more than this can
// answer the prompt again when it comes back. Large enough that
// nobody legitimately hits it, small enough that a fat-fingered
// 1000000 does not hand the server an afternoon's work with no way to
// interrupt it.
const MaxLoopShortcutIterations = 1000

// DefaultLoopShortcutIterations is the K a bot seat takes the first
// time it is asked in a turn — the first answer `internal/legal`
// offers. See docs/bot.md.
const DefaultLoopShortcutIterations = 10

// queueLoopShortcutLocked offers the standing notice's controller the
// CR 732 shortcut. `repeat` says this is the second ask of the turn,
// which is what lets a bot stop rather than shortcut the same loop
// forever.
//
// Queues nothing when there is nobody to ask — no controller on the
// notice, or a controller who has left or been eliminated. The notice
// alone is then the whole answer, which is ADR 0055's behaviour and
// the only safe one: a blocking prompt addressed to an absent seat is
// a wedged table, which is the class of bug this sprint exists to
// close.
//
// Caller must hold g.mu.
func (g *Game) queueLoopShortcutLocked(key string, repeat bool) {
	n := g.LoopNotice
	if n == nil || key == "" {
		return
	}
	p := g.playerByIDLocked(n.Controller)
	if p == nil || p.Eliminated {
		return
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceLoopShortcut {
			// One conversation at a time.
			return
		}
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:               PendingChoiceLoopShortcut,
		Chooser:            n.Controller,
		Count:              1,
		Source:             n.Source,
		Reason:             n.Label,
		LoopShortcutKey:    key,
		LoopShortcutCount:  n.Count,
		LoopShortcutRepeat: repeat,
	})
}

// ResolveLoopShortcut answers a PendingChoiceLoopShortcut: run the
// loop `iterations` more times, then ask again. Zero is "stop here" —
// the prompt clears and the table stays exactly as the breaker left
// it, automatic passing paused and priority in the players' hands.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveLoopShortcut(choiceID, chooserID uuid.UUID, iterations int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceLoopShortcut {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if iterations < 0 || iterations > MaxLoopShortcutIterations {
		return ErrInvalidParam
	}
	key, at := choice.LoopShortcutKey, choice.LoopShortcutCount
	// Captured before the dequeue, which clears it: answering is a
	// player decision like any other (ADR 0055 §3), and for K = 0 the
	// notice has to go straight back up.
	notice := cloneLoopNotice(g.LoopNotice)
	g.dequeueChoiceLocked(idx)
	g.grantLoopShortcutLocked(key, at, iterations)
	if iterations == 0 {
		g.LoopNotice = notice
	}
	return nil
}

// LoopAllowanceFor reports how many more resolutions of the ability
// keyed by `key` the table's CR 732 shortcut still covers. Zero when
// no shortcut is running.
func (g *Game) LoopAllowanceFor(key string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.TurnTally.LoopAllowance[key]
}
