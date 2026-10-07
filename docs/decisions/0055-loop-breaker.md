# ADR 0055 — CR 732 loop breaker: the engine detects trigger loops by tally and suspends autopass

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

The paper answer is CR 732: players take a shortcut. The loop's controller
says how many more times it happens and the table skips there, and a
mandatory loop nobody can stop is a draw (CR 732.4). Both halves need a
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
priority back every iteration — which is the CR 732 conversation happening
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

The alternative — a bot answers a CR 732 shortcut prompt with a fixed K and
then stops — needs the shortcut prompt (§6) and a `PendingChoiceKind` case
in `internal/legal`, and it buys a bot-only table a few more iterations of a
loop that has no ending. Stopping is the honest answer and it is the one a
human at the table gets too.

### 6. The CR 732 shortcut prompt is deferred

"This loop has resolved N times. Resolve it K more times and stop?" — and
CR 732.4's draw — are not in this change. A new `PendingChoiceKind` needs a
case in `internal/legal/choices.go` or bot seats owing the prompt get an
empty move list (the #499 / #618 class, flagged on #628 by the S31 audit),
plus a client modal, plus a bot answer. The breaker is useful without it:
the loop stops running on its own, priority comes back, and a player can
break it or concede. The prompt is a follow-up.

**Amendment (2026-09-18, #804): the shortcut prompt is built; §6 is
closed.** CR 732.4's draw is still not, and stays out of scope — a
mandatory loop nobody can stop is a different question from "how many more
times", and nothing in the catalog makes one.

*The kind.* `PendingChoiceLoopShortcut` (`"loop_shortcut"`), declared in
`loop_breaker.go` beside the notice that raises it and queued from
`queueLoopShortcutLocked` at the moment the notice goes up — to the
**controller** of the ability the notice names, because CR 732 is their
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
keyed by `TallyKey(source, label)` like everything else. The CR 732
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
CR 732 question again each time round, where a trigger loop's notice simply
stays up. That is the honest consequence of an activation being a real
decision: it clears the notice, and the next resolution raises it again. A
player who wants to keep going answers with a number and gets it.

*Where the fix actually belongs.* This amendment is the belt. The braces are
in the enumerator, which since #810 does not offer an {X} move at X=0 when X
is the whole of what it does — so the Soothsaying table has no loop to
break. See [ADR 0033](0033-ai-bot-seat.md)'s amendment of the same date.


## Amendment (2026-09-18, #936): the breaker keeps the CARD key; the once-each-turn gates get the OBJECT

`TurnTally.LoopRun` shares its key with `TurnTally.Resolved` —
`TallyKey(source, label)` — and that sharing is why a real bug sat
open for a sprint. `Resolved` and `Triggered` are also the catalog's
"only once each turn" gates, and CR 400.7 says a permanent that leaves
the battlefield and comes back is a NEW OBJECT whose clause may fire
again. The obvious fix, dropping the entries at the one battlefield
exit alongside the loyalty and combat registries (#630, #935), would
have reset `LoopRun` on every iteration of a loop that blinks its own
source: the detector would never have reached `DefaultLoopThreshold`,
and the escape hatch would have been created by exactly the loops this
ADR exists for.

The key is split instead of cleared, and the detector keeps the half
it had:

- **`LoopRun` and `LoopAllowance` stay keyed by `TallyKey` — the
  CARD.** A loop is a loop whichever object is running it, so the
  count survives the permanent leaving and re-entering, and the
  threshold means what §2 says it means. Nothing in `loop_breaker.go`
  changed: `loopSuspectedLocked` is still the one detector over the
  one count, and `notePlayerActivationLocked` and
  `grantLoopShortcutLocked` still receive `TallyKey(source, label)`
  from their callers.
- **`Resolved` and `Triggered` are keyed by
  `ObjectTallyKey(source, Card.ObjectEpoch, label)` — the OBJECT.**
  The returning permanent reads a key nothing has written; the old
  object's entries are left in the map, unreachable, until the turn
  boundary flushes the whole tally.

One (source, label) pair, two projections, and the reader's question
picks the projection — `TurnTally`'s field comments say which is
which, and ADR 0049's amendment of the same date carries the table.
Pinned by `TestLoopBreakerStillTripsWhenTheSourceLeavesAndReturns`
(turn_tally_object_test.go), which is the test the naive fix would
have failed.

## Amendment (2026-10-07, #2450): a batch of triggers is not a loop, and a loop that ends the game is not a runaway

**Status of this amendment: Proposed.** Nothing below is built. The
owner's answers to the questions at the end decide what is.

### The problem

Three bot-only tables in ADR 0126's measurement runs stopped at this
breaker. Every seat logged `bot holding: automatic passing is suspended
by the loop breaker (CR 732)`, and the commit sequence stopped:

| Run | Seed | Where | Life totals when it stopped |
| --- | --- | --- | --- |
| run 1 (#2445, ADR 0126 PR 4) | 15 | turn 11, precombat main | aristocrats 82; 7, 2, 5 |
| run 1 (ADR 0126 PR 7 and PR 8) | 21 | turn 10, precombat main | aristocrats 75; 32, 2, one seat out |
| run 2 (ADR 0126 PR 9, #2463) | 141 | turn 13, declare attackers | simic 57; esper 9; two seats out at -48 and -50 |

#2450 named the Sanguine Bond and Exquisite Blood drain loop. The arena
records say otherwise. **Sanguine Bond was cast in none of the three
games.** In seeds 15 and 21 the aristocrats seat had cast Syr Konrad,
the Grim and Exquisite Blood. In seed 141 no seat cast either
enchantment.

A scratch test (not committed) rebuilt seeds 15 and 21's shape on four
seats. The board was Syr Konrad and Exquisite Blood, with 30 opposing
creatures destroyed at once:

- One event kills 30 creatures, so Konrad triggers 30 times (CR 603.2c).
  Each Konrad trigger deals 1 damage to each of three opponents, which
  triggers Exquisite Blood three times.
- Exquisite Blood's key reaches 25 resolutions after nine Konrad
  triggers, with no player decision in between, and the notice goes up.
- The bot answers the shortcut prompt with 10. The 35th Blood resolution
  asks again, and the second ask offers only "stop here" (§6 amendment).
  Every seat then holds with Konrad triggers still waiting.
- The aristocrats seat ends at 40 + 35 = 75 life, which is seed 21's
  number exactly.

Konrad alone, with 30 deaths, trips the breaker too. There the
shortcut's 10 more covers the last five triggers, so that table finishes.
Seed 141 was not rebuilt. Its board fits the same shape: no drain
enchantment was cast, and simic had Avenger of Zendikar's Plants on the
battlefield. It is listed here as unconfirmed.

Two more findings:

- **Sanguine Bond and Exquisite Blood never trip the breaker in this
  engine.** Each Bond trigger asks its controller for a target
  (`pick_target`), and answering a prompt is a decision (§3). So every
  run restarts each iteration. A scratch run of that pair drained 372
  times in one turn with no notice.
- **The notice is not cleared when a player leaves the game.** Only a
  decision, the turn boundary (`resetTurnTallyLocked`) or a running
  shortcut takes it down. That much of the triage was right. It is not
  what stopped these tables, though: in seed 15 no seat had left.

So the observed stalls are a **false positive**. They are a finite batch
of triggers, all caused by one event, and the stack is draining. §2
argued that 25 resolutions of one key "is unreachable by accident". A
wrath across a token board with one death trigger and one per-opponent
payoff gets there in about nine deaths.

There is also a second case, latent so far: **a real mandatory loop that
ends the game.** Exquisite Blood with Marauding Blight-Priest or
Cliffhaven Vampire ("Whenever you gain life, each opponent loses 1
life") is such a loop. All three cards are catalogued, and nothing in
the loop targets. Each iteration adds more triggers than it resolves,
until every opponent is dead. Any detector should call it a loop. On a
bot-only table it stops after threshold + 10, exactly as the drain was
expected to. No curated deck holds the pair, but the catalog soak can
deal it. This is the shape #2450 described.

### What the rules say

From the Comprehensive Rules effective September 25, 2026:

- **CR 732.1b:** "Occasionally the game gets into a state in which a set
  of actions could be repeated indefinitely (thus creating a 'loop'). In
  that case, the shortcut rules can be used to determine how many times
  those actions are repeated without having to actually perform them,
  and how the loop is broken."
- **CR 732.2a:** the player with priority may suggest a shortcut, and
  "This sequence may be a non-repetitive series of choices, a loop that
  repeats a specified number of times, multiple loops, or nested loops,
  and may even cross multiple turns."
- **CR 732.4:** "If a loop contains only mandatory actions, the game is a
  draw. (See rules 104.4b and 104.4f.)" **CR 104.4b** says when that
  applies: the game "somehow enters a 'loop' of mandatory actions,
  repeating a sequence of events with no way to stop".
- **CR 603.2c:** "An ability triggers only once each time its trigger
  event occurs. However, it can trigger repeatedly if one event contains
  multiple occurrences."
- **CR 704.5a:** "If a player has 0 or less life, that player loses the
  game." **CR 104.2a:** "A player still in the game wins the game if that
  player's opponents have all left the game."

Three readings follow:

1. Thirty triggers from one wrath are thirty triggers (CR 603.2c). They
   are not a set of actions that "could be repeated indefinitely" (CR
   732.1b), so CR 732 has nothing to say about them. The game resolves
   them.
2. A mandatory loop that drives every opponent to 0 life ends. Its
   opponents lose to CR 704.5a and its controller wins under CR 104.2a.
   It is not 104.4b's loop "with no way to stop", so CR 732.4's draw
   does not apply. CR 732.2a lets it be shortcut ("a loop that repeats a
   specified number of times"), and the number is however many
   iterations it takes.
3. The draw is for a mandatory loop that never ends. This engine still
   does not build it (§6 amendment), and nothing below changes that.

### Options

**A. A batch that drains is not a loop.** A loop refills the stack and a
batch empties it. While the stack plus `PendingTriggers` keeps reaching
new lows, the table is working through something finite.

- *Rule.* After each resolution, once the triggers it caused are
  waiting, the engine counts the stack plus `PendingTriggers`.
  `TurnTally` keeps the lowest such count since the last decision; the
  first resolution after a decision sets it. When a resolution leaves
  the count strictly below that low, every run of a **triggered** key
  restarts.
- *Effect on the cases.* The wrath resolves and leaves 30 Konrad
  triggers waiting: the low is 30. One Konrad trigger and its three
  Blood triggers make a cycle (30, 32, 31, 30, 29), so every cycle
  makes a new low and no run gets past three. A real trigger loop replaces what it resolves, so it
  never makes a new low and trips at 25 as it does today. This covers
  #628's token engines and Blight-Priest with Blood.
- *#810 is untouched.* An activation loop empties the stack every
  iteration by construction, so the rule never restarts an activated
  key's run. That run stays with `notePlayerActivationLocked`.
- *Cost.* One integer on `TurnTally`, carried by undo and the snapshot (an
  additive field within schema v7, recorded with `-update-shape`), and one
  comparison in `noteResolutionForLoopLocked`. It fixes seeds 15 and 21
  as rebuilt, and seed 141 if that is a batch too.

**B. A loop that is ending the game keeps stepping.** A run restarts on
progress that can only happen finitely often:

- a player leaves the game, or
- a player who can lose the game reaches a new lowest life total for the
  turn, or
- a player reaches a new highest poison count for the turn (CR 122.1f
  ends it at ten).

Each player's life can make only so many new lows before CR 704.5a takes
them, and players can leave only so many times. So a loop that keeps
restarting this way must end. Blight-Priest with Blood plays out to the
win. A loop that gains life forever, or makes tokens forever, still
trips the breaker.

B alone would also release the observed batches, because every Konrad
trigger lowers three life totals. A batch that touches no life total
would still trip it, though: thirty "whenever a creature dies, scry 1"
triggers, say. B costs one low-water mark per player on
`PlayerTurnTally` (additive), set from the same life-change events the
tally already reads.

**C. Clear the notice when a player leaves the game.** This is the triage
proposal. It is narrow: seed 15 stopped with nobody gone. B includes it
as one of its progress events.

**D. No hold at a table with no humans.** The runner passes through the
notice when every seat is a bot. That plays out every batch and every
ending loop, but a real infinite loop on a bot table then spins until the
wall clock, which is the runaway §5 exists to stop. The event log grows
without limit too (#629). It is also the least faithful option: it makes
bot tables play by a different rule, where the CR gives such a loop a
draw. Not recommended.

### Recommendation

**A and B together.** Ship A first, because it is the observed stall,
then B as the second half of the same amendment. A tells a batch from a
loop. B tells a loop that is ending the game from one that is not, which
is the distinction CR 104.4b draws. With both, the breaker fires only on
a loop that neither drains nor progresses. That is a loop CR 732.4 would
call a draw, and §5's stop is still the honest answer for it until the
draw is built. Both rules apply to every seat, human and bot, because
ADR 0055 is one rule for the table. A human can still turn autopass off
and step through either case by hand. C is subsumed by B, and D is
rejected.

### Questions for the owner

1. Adopt A, so that a run of triggered abilities restarts whenever the
   stack plus pending triggers reaches a new low since the last
   decision? *Recommended: yes.*
2. Adopt B as well, as a second PR under this amendment? *Recommended:
   yes.*
3. Which events count as progress for B? The candidates are (a) a player
   leaves the game, (b) a new lowest life total this turn, (c) a new
   highest poison count this turn and (d) a new smallest library.
   *Recommended: a, b and c. Leave out d: milling a player to an empty
   library ends nothing until they draw.*
4. Should a player who can't lose the game, or can't lose life (ADR
   0057), be left out of B's life and poison checks? Otherwise a drain
   into a Platinum Angel would count as progress for ever.
   *Recommended: yes, using ADR 0057's own predicate.*
5. When A or B restarts the runs while a notice is standing, should it
   also clear the notice and withdraw a pending shortcut prompt, as a
   decision does? *Recommended: yes. A new low or new progress means the
   notice's claim no longer holds. A stale shortcut prompt blocks the
   table, so leaving one up would wedge it.*
6. Engine-wide, so that human autopass also plays through a batch or a
   loop that is ending the game? Or bot seats only? *Recommended:
   engine-wide, the same rule for every seat. A human can still turn
   autopass off to step by hand.*
7. Keep `DefaultLoopThreshold` at 25? *Recommended: yes. A removes the
   reason to raise it.*
8. Build CR 732.4's draw now, for a mandatory loop that neither drains
   nor progresses? *Recommended: no. File it separately and keep §5's
   stop until then.*
9. Retitle #2450 to the observed shape (a batch of Syr Konrad and
   Exquisite Blood triggers), and correct docs/bot.md's paragraph naming
   Sanguine Bond and Exquisite Blood, in the PR that implements A?
   *Recommended: yes.*
