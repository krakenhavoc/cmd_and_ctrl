# ADR 0055 — CR 726 loop breaker: the engine detects trigger loops by tally and suspends autopass

**Status:** Accepted · 2026-09-17 · Post-S30 — Rolling deck-driven catalog growth
**Issue:** [#628](https://github.com/krakenhavoc/cmd_and_ctrl/issues/628)
**Related:** [ADR 0009](0009-smart-priority-autopass.md) (autopass lives in the
client), [ADR 0018](0018-triggers-on-the-stack.md) (triggers use the stack),
[ADR 0033](0033-ai-bot-seat.md) (bot seats)

## Context

Replacement loops are capped — `ErrReplacementIterationExceeded`. Trigger
loops were not.

A triggered ability goes on the stack and resolves once every player has
passed priority in succession, so the server does one bounded unit of work
per pass and never blocks. Two "whenever a creature enters, create a token"
permanents therefore do not hang the server: they loop at exactly the speed
the table passes priority. With autopass on for every seat that speed is the
network, and the loop becomes a tight
`pass → resolve → broadcast → autopass → pass` spin. The game never ends,
the event log grows without limit (the sibling issue, #629), and the only
way out is for somebody to find the autopass toggle mid-flight.

The paper answer is CR 726: players take a shortcut. The loop's controller
says how many more times it happens and the table skips there, and a
mandatory loop nobody can stop is a draw (CR 726.4). Both halves need a
player to *say* something. So the engine's job is not to stop the game — it
is to hand priority back to the humans with the loop's trigger still on the
stack, and say why.

## Decisions

### 1. Detection is one function over the tally the engine already keeps

`TurnTally.Resolved` (#586) already counts every stack item that resolved
this turn, keyed by `TallyKey(source, label)` — one printed ability of one
permanent. `TurnTally.LoopRun` is the same count restarted at every player
decision, bumped in the same `turnTallyListener` case, and
`loopSuspectedLocked(key)` is the whole rule:

```go
func (g *Game) loopSuspectedLocked(key string) bool {
	return g.TurnTally.LoopRun[key] >= g.loopThresholdLocked()
}
```

One place, one function, no per-card special cases, and no second copy of a
count the engine was already keeping. The alternative considered and
rejected in the issue — a hard cap on stack size or on triggers per turn —
would break real, finite storm and aristocrats turns long before it stopped
any loop.

**Why the key and not a total.** The count is per `(source, label)` in one
turn. Four upkeep triggers from four players are four different keys and
each sits at 1. The false positive this design has to avoid is "an ordinary
busy turn looks like a loop", and keying by ability is what makes that
structurally impossible rather than merely unlikely.

**Why a consecutive run and not the raw total.** A loop between two
permanents alternates, so "consecutive resolutions with nothing in between"
would never exceed one. `LoopRun` counts resolutions of each key since the
last decision, whatever else resolved in between.

### 2. The threshold is 25, a named constant, overridable per game

`DefaultLoopThreshold = 25`, the issue's number. `Game.LoopThreshold`
overrides it, so a test can trip the breaker in three resolutions without
every other test in the tree sharing the setting. A field on `Game` rather
than a package global for exactly that reason: parallel tests, one process.

25 is chosen to be unreachable by accident rather than to be tight. A loop
gets there in seconds; the longest real turn does not put one ability on the
stack twenty-five times without its controller casting or activating
something.

### 3. A "decision" is anything but a pass — and a bare pass is not one

`notePlayerDecisionLocked` restarts every run and clears the notice. It is
notched from four places, chosen so that each one is the only place that
fact is knowable:

| Decision | Where it notches |
| --- | --- |
| Cast a spell | `turnTallyListener`, on `EventCast` |
| Declare an attacker / blocker | `turnTallyListener`, on `EventAttack` / `EventBlock` |
| Activate an ability | `ActivateCatalogAbility` — the announce emits `EventTrigger`, the same kind a *triggered* ability announces with, so the listener cannot tell them apart |
| Answer a prompt | `dequeueChoiceLocked`, the tail of every `Resolve*` path |

The engine's own prune paths (a sacrifice prompt whose card has left, a
stale zone-change prompt) call the new `dropChoiceLocked` instead: a prompt
the engine withdrew is not a decision anybody made, and counting it as one
would let a loop that queues and prunes a prompt each iteration run forever.

**`pass_priority` is deliberately not a decision.** If it were, the first
manual "next" click would clear the notice and four autopassing clients
would spin the loop straight back up. Leaving the notice standing lets the
table step the loop through by hand for as long as it likes, getting
priority back every iteration — which is the CR 726 conversation happening
at human speed.

### 4. The effect is suspending automatic passing, not stopping the game

`Game.LoopNotice` (source, label, controller, count) plus one
`EventLoopSuspected` breadcrumb, emitted once per run. The engine does not
refuse a pass, does not halt the stack, and does not change whose priority
it is. Everything that changes is who passes *automatically*:

- **Clients**: `GameView.loop_notice` rides the wire, `autopassSuspended()`
  reads it, and the autopass `$effect` in `Game.svelte` returns early —
  above the toggle, so the toggle cannot out-vote it. The toggle renders as
  `autopass ⏸` with a banner under it naming the ability and the count.
  Suspended, not switched off: the player's intent is untouched.
- **Bot seats**: the runner skips a `KindPass` move while
  `game.AutoPassSuspended()`. A bot with a real (non-pass) move still plays
  it, and playing one clears the notice like any other decision.

### 5. Bots count as automatic, and a bot-only table stops

This is the one place where the simple rule has a visible cost, so it is
stated plainly: on a table of four bots with a real loop, every seat holds
and the room's commit sequence stops. The table is *stopped*, with the
notice in the game state naming the ability and the count, rather than
spinning until someone kills the process.

The alternative — a bot answers a CR 726 shortcut prompt with a fixed K and
then stops — needs the shortcut prompt (§6) and a `PendingChoiceKind` case
in `internal/legal`, and it buys a bot-only table a few more iterations of a
loop that has no ending. Stopping is the honest answer and it is the one a
human at the table gets too.

### 6. The CR 726 shortcut prompt is deferred

"This loop has resolved N times. Resolve it K more times and stop?" — and
CR 726.4's draw — are not in this change. A new `PendingChoiceKind` needs a
case in `internal/legal/choices.go` or bot seats owing the prompt get an
empty move list (the #499 / #618 class, flagged on #628 by the S31 audit),
plus a client modal, plus a bot answer. The breaker is useful without it:
the loop stops running on its own, priority comes back, and a player can
break it or concede. The prompt is a follow-up.

**Amendment (2026-09-18, #804): the shortcut prompt is built; §6 is
closed.** CR 726.4's draw is still not, and stays out of scope — a
mandatory loop nobody can stop is a different question from "how many more
times", and nothing in the catalog makes one.

*The kind.* `PendingChoiceLoopShortcut` (`"loop_shortcut"`), declared in
`loop_breaker.go` beside the notice that raises it and queued from
`queueLoopShortcutLocked` at the moment the notice goes up — to the
**controller** of the ability the notice names, because CR 726 is their
proposal to make. It carries the tally key the answer attaches to, the
count the notice is quoting, and a `LoopShortcutRepeat` flag (below). The
prompt is queued only when there is a live, unelminated seat to ask; a loop
with no controller raises the notice and no question, which is this ADR's
original behaviour and the only safe one, since the prompt blocks.

*Allowance semantics.* The answer is an integer K, 0 ≤ K ≤
`MaxLoopShortcutIterations` (1000), written to
`TurnTally.LoopAllowance[key]`. There is no second detector:
`loopSuspectedLocked` is still the whole rule, and the allowance only
decides whether its answer is worth saying out loud. Each further
resolution of that key spends one (`spendLoopAllowanceLocked`), silently;
the K-th raises the notice again and re-offers the prompt. While a shortcut
is live the notice is down, so the client's autopass and the bot runner
both run — they read `loop_notice` and `AutoPassSuspended()`, and neither
needed a line of change. A shortcut running on one key also suppresses a
notice for the loop's *other* key, because a two-permanent loop is two keys
and one conversation.

K = 0 is "stop here": the prompt clears and the table goes back exactly
where the breaker left it — notice standing at its count, automatic passing
suspended, and the count still climbing if the players step the loop on by
hand.

*The counter subtlety, stated because it is the whole of the change.*
Answering a prompt is a player decision (§3), and a decision restarts every
run — right for every other prompt and exactly wrong for this one. With
`LoopRun` back at zero nothing would consult the allowance until the
ability had resolved a further `LoopThreshold` times, so "resolve it 3 more
times" would mean 28. `grantLoopShortcutLocked` therefore re-arms *this
key's* run to where the notice found it, after the dequeue has cleared
everything. Every other key stays cleared: the decision was real. And
because the prompt blocks the table, `notePlayerDecisionLocked` now also
withdraws any outstanding shortcut prompt (`dropChoiceLocked`, not
dequeue — the engine is withdrawing it, nobody answered): a prompt left
behind by a cleared notice would not be a stale question, it would be a
wedged game.

*Blocking.* This prompt **blocks the table** (`ChoiceBlocksTable`, ADR 0018
§6's 2026-09-18 amendment), which is a real departure from §4's "the engine
refuses no passes". It is right here and wrong for a Rhystic tax for the
same reason: the shortcut is proposed while the loop's trigger is still on
the stack, and what the answer decides is how many times that trigger
resolves next. A table that could pass through the question would be
answering it by doing. §4's rule is otherwise untouched — once the question
is answered, every pass works and priority still rotates.

*Bot policy.* `internal/legal` enumerates the kind (it has to: a kind with
no case there is an empty move list, which is the #499 / #618 wedge this
ADR's §6 named). The first ask of a turn offers 10 — first in the list,
which is what a bot takes — then 100, then stop. The **second** ask for the
same loop in the same turn offers *only* stop. That termination guarantee
is in the enumerator rather than in a policy on purpose: a policy that can
rank "100 more" top can rank it top every time, and a random policy
eventually will. With one answer on the list, every policy at the table
stops. A bot-only table on a real loop therefore runs threshold + 10
iterations and comes to rest, which is §5's outcome with one shortcut's
worth of progress in front of it. See [docs/bot.md](../bot.md).

*Undo and restore.* The prompt's three fields and `TurnTally.LoopAllowance`
are `carried` by both copies. Writing the undo test turned up that
`RestoreFrom` had never copied `LoopNotice` or `LoopThreshold` at all —
#628 added them to `cloneLocked` and to the snapshot but not to the undo
path, so an undo across the moment the breaker fired left the live notice
untouched. Fixed in the same change, because a restored prompt without its
notice is a question about nothing.

## Consequences

### Good

- A trigger loop now stops itself after 25 iterations instead of running
  until a person finds the autopass toggle, on every kind of table —
  browsers, bots, and mixed.
- Nothing about it is per-card. Any future pair of permanents that loops is
  covered the day it is added to the catalog.
- The detection reuses `TurnTally`, so it costs one map bump per resolution
  and no new scan of the event log.

### Tradeoffs

- The notice names the ability that tripped *first*. In a two-permanent loop
  the other half is equally to blame; naming both would mean a second notice
  and a second breadcrumb per resolution for no extra information.
- A loop that manages a player decision every iteration (a cast, an answered
  prompt) is not detected. That is deliberate — those are not the runaway
  this exists to stop, and they are not driven by autopass.
- A bot-only table that hits a real loop stops rather than finishing, and
  the whole-game bot tests define "stalled" as "the commit sequence stopped
  moving". No catalog pair loops today and 25 resolutions of one ability in
  one turn does not happen in those games, so nothing goes red — but if a
  looping pair is ever added, those tests are where it will surface, which
  is the right place for it to.

## Amendment (2026-09-18, #810): the breaker counts repeated activations

§3's table said "activate an ability → `ActivateCatalogAbility`", with no
qualification, and §1's tradeoff list said in as many words that a loop
managing a player decision every iteration is not detected and that this is
deliberate. Both were written about a TRIGGER loop, which runs itself — and
for that shape they are right. An **activation** loop is the other shape,
and it is a runaway too: nothing repeats on its own, a seat simply keeps
taking the same free, repeatable ability, because the ability is back on its
move list the instant it resolves. `internal/legal` offered Soothsaying's
"{X}: Look at the top X cards of your library" at X=0 to a table of bots,
which took it 79,519 times in five minutes and never finished turn 18.

*The change is one exception, in one function.* `notePlayerActivationLocked`
is `notePlayerDecisionLocked` with the run of the ability **being
activated** kept, and every other run cleared as before. An activation is
still a player decision about everything else — alternating two abilities
never trips the breaker, and neither does a turn with a cast in the middle
of it — but it is not a decision about the loop it *is*. Before this, the
activation cleared the very run its own resolution was about to add to, so
`LoopRun` never got past 1 and `loopSuspectedLocked` never saw a thing. The
detector, the threshold, the notice and the tally are untouched: the
resolution still notches through `turnTallyListener` on `EventResolve`,
keyed by `TallyKey(source, label)` like everything else. The CR 726
allowance for that key survives the activation too, for the same reason
`grantLoopShortcutLocked` re-arms the run: a shortcut is about one ability,
and the activations that spend it must not be what cancels it.

*Two bot-side rules make the table actually stop*, which is what §5 asks of
a bot-only table and what the notice alone could not deliver here. A trigger
loop is fed by passes and the breaker starves it by suspending automatic
passing; an activation loop is fed by activations, which are not passes, so
the runner would have gone on feeding it with the notice up.

- `internal/legal` offers **only "stop here"** on the *first* ask for a loop
  whose repeating ability is an activated ability of the chooser's own
  permanent — derived from the prompt's source and label against the board,
  no new field. "Resolve it ten more times" is an answer about a loop that
  runs itself; for one somebody has to crank it buys ten more turns of the
  crank.
- The **runner holds** on an activation of the permanent the notice names,
  exactly as it holds on a pass (§4). It matches on the permanent rather
  than the ability, which is the conservative direction — the notice means
  "stop feeding this card" — and, like the pass rule, it parks the seat
  rather than choosing a different move.

*Tradeoff.* A human stepping their own activation loop by hand is asked the
CR 726 question again each time round, where a trigger loop's notice simply
stays up. That is the honest consequence of an activation being a real
decision: it clears the notice, and the next resolution raises it again. A
player who wants to keep going answers with a number and gets it.

*Where the fix actually belongs.* This amendment is the belt. The braces are
in the enumerator, which since #810 does not offer an {X} move at X=0 when X
is the whole of what it does — so the Soothsaying table has no loop to
break. See [ADR 0033](0033-ai-bot-seat.md)'s amendment of the same date.
