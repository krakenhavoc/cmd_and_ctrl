# ADR 0009 — Smart priority auto-pass (S13.6)

**Status:** Accepted · 2026-04-23 · Sprint S13.6
**Amended by:** S31 sub-PR 2 ([ADR 0033](0033-ai-bot-seat.md) §1, PR #429) — the
mechanism under decisions 2–4 is gone, the policy above it is not.
**Amended by:** S35 (#526) — **decision 6 is reversed on one point: the autopass
toggle no longer overrides a manual pin.** See "Amendment: a manual pin beats the
autopass toggle (#526)" below. The gate chain also moved out of `Game.svelte`
into `client/src/lib/autopassDecision.ts`, which is where the precedence list
now lives in code.

`hasAnyLegalResponse` no longer walks the viewer's cards running per-action
predicates. The server enumerates the seat's legal moves and ships them as
`GameView.legal_moves`, so the aggregate question is
`legal_moves.some(m => m.kind !== "pass")` — one scan of a field the engine
already computed. Three consequences for a reader of the sections below:

- **Decision 3's conservatism is now structural rather than deliberate.** The
  client no longer has to assume every permanent is "potentially activatable"
  because it cannot see ability lists; the engine tells it exactly which
  activations are live. The false-positive-over-false-negative preference
  survives, and is what makes an ABSENT `legal_moves` (a pre-S31 server, or a
  frame where the seat owes nothing) resolve to "stop" rather than "skip".
- **Decision 4 is obsolete: mana affordability is now in scope**, for free. The
  enumerator pays for what it offers, so a hand of uncastable 7-drops correctly
  reports no response and the window is skippable.
- **Decision 2's memo cache is gone.** It existed because the card walk was
  O(hand + battlefield + command) per call; an array scan does not need it.
  `_resetCacheForTests` survives as a no-op so test teardowns don't break.

The module split decision 2 argues for (`timing.ts` per-action, `priority.ts`
aggregate) still holds and is unchanged.

## Amendment: a manual pin beats the autopass toggle (#526)

Decision 6 below says the autopass toggle "ignores every gate (stops grid,
smartAutoPass predicate, manual pins, `settings.autoPassPriority`)". The pins
half of that was wrong, and a playtester reported it as a bug: *"When autopass
is turned on, and a manual stop is placed on a step, the game continues to
autopass right through without ever stopping."*

**The precedence is now:** a manual pin (decision 5) sits **above** the autopass
toggle. Everything else in decision 6 stands — the toggle still out-votes
`settings.autoPassPriority`, the stops grid, the smartAutoPass predicate and the
#323 own-stack rule.

**Why the original reading doesn't survive contact with play.** Decision 6's
argument was "the point of autopass is 'no more asking'". But the pin is a
*later and narrower* instruction than the toggle: the toggle is standing intent
("I'm tapped out, carry me"), the pin is a click made for one step of one turn,
and the only reason to make it is to interrupt automatic passing. Under the old
order the affordance was dead in exactly the state where a player needs it —
the PhaseDisplay icon lit up and changed nothing. Decision 5's own argument
("if smartAutoPass could skip a pinned step, the affordance would be a lie")
applies with more force to the toggle, not less.

**It does not strand the toggle.** The pin is consumed on the next step
transition, so the cursor holds once and autopass resumes on its own with no
second click. The player who wants out of autopass entirely still clicks the
toggle.

**Two things deliberately kept above the pin:**

- The guards that are not questions about what the viewer *may* do — an open
  pending choice, an owed declare-blockers decision (#328), the CR 726 loop
  breaker ([ADR 0055](0055-loop-breaker.md), #628), mulligans, game over,
  elimination. All of these hold anyway, so the pin changes nothing there.
- The safety belt (decision 7). Entering the viewer's own `precombat_main`
  *disarms the toggle* instead of passing, which holds the cursor too — so a
  pin on your own main phase loses nothing by sitting under it, and gains the
  disarm. Putting the pin first would leave a forgotten toggle armed for the
  rest of that turn.

**Where the code lives.** The chain was ~90 lines inline in one `$effect` in
`Game.svelte`, which is how the bug survived four sprints: no test could reach
it. It is now `autopassDecision(gates) -> "hold" | "pass" | "clear-toggle"` in
`client/src/lib/autopassDecision.ts`, a pure function over resolved booleans,
with `autopassDecision.test.ts` covering the precedence, #526's reported
configuration and #599's declare-attackers window. The component keeps the
reactive reads and the two side effects.

## Context

S13 shipped the per-step "stops" grid and a global `autoPassPriority`
toggle (ADR [0006](0006-priority-foundation.md)): auto-pass priority
through any step the viewer hasn't pinned, stop at every step they
have. S13.3 shipped `client/src/lib/timing.ts` — the legality
predicates (`canCastFromHand`, `canActivateAbility`, etc.) that
grey out hand cards the viewer can't play right now.

Four-player playtests exposed a rough edge the original design
didn't anticipate: **a stop means "don't skip this step," not "I
definitely have something to do here."** In practice the active
player gets stopped on their own upkeep with an empty hand, stopped
on combat steps with no creatures, and so on — a click per stop per
turn cycle, each of which adds up to a drumbeat of pointless
interactions in a 2–3 hour game.

## Decisions

### 1. The stops grid becomes an intent filter, not a gate

**Decision:** Add `gameplay.smartAutoPass: boolean` (schema v5;
default `true`). When on, the existing `autoPassPriority` effect in
`Game.svelte` folds in `hasAnyLegalResponse(snap, viewerID)`: if
the viewer holds priority on a stopped step but the legality
engine reports nothing legal, auto-pass anyway. The stops grid
still controls _where the player wants to consider stopping_; the
predicate decides whether there's actually anything to consider.

**Why a new setting instead of just changing behaviour:**

- Some players (especially sandbox testers) want every stop to
  hold open so they can think or fake plays that aren't in the
  catalog yet. Making this a toggle lets them opt out.
- Matches the pattern we used for the v2→v3 `autoPassPriority`
  re-default (ADR 0006 §3): ship the better default, give the
  holdouts a knob.

**Why default on:** the predicate is intentionally permissive
(see §3) so false-negative skips — the failure mode that would
eat a player's response — are rare. False-positive stops (the
flip side: stop when there was nothing to do) cost one extra
click and are self-explanatory. The ratio says on is the right
default.

### 2. The predicate lives in `client/src/lib/priority.ts`

**Decision:** `hasAnyLegalResponse(snap, viewerID, snapSeq?)` folds
the S13.3 timing helpers across the viewer's hand, command zone,
and viewer-controlled battlefield cards. Returns true as soon as
any predicate returns legal.

Memoised by `snap.seq` so repeat calls inside one snapshot window
(the autoPassPriority effect + any future UI consumer) don't
rescan the same state.

**Why a new module and not an extension of `timing.ts`:** scope
clarity. `timing.ts` is the per-action predicate layer
(card-in-hand vs. ability vs. pass). `priority.ts` is the
aggregate question "anything at all?" built on top. Keeping them
separate lets `timing.ts` stay focused on the grey-card UX and
`priority.ts` stay focused on the auto-pass UX without each
module's API growing tendrils into the other.

### 3. Predicate is conservative (false-positive-stop, not false-negative-skip)

**Decision:** Treat any viewer-controlled battlefield card as
"potentially has an activated ability" (via `canActivateAbility`,
which is the generic instant-speed gate). Treat any commander in
the command zone as castable via the same predicate as hand
cards.

**Why:**

- Skipping a priority window the player wanted is the bad
  outcome. The grey-card UX recovers from a false positive (you
  see the grey, you click pass); there's no recovery from a
  skipped window short of undo.
- We don't know per-card ability lists on the client — the
  server's effect catalog (S14) does, but the client doesn't
  carry that metadata on-wire. Treating every permanent as
  _potentially_ activatable keeps the predicate's false-negative
  rate near zero.

### 4. Mana affordability is out of scope (S13.6 ships without it)

**Decision:** `hasAnyLegalResponse` does **not** check whether the
viewer can afford the mana cost of any of their hand cards. A
player with a 7-drop and zero lands still counts as "has a legal
response" under this predicate.

**Why not:**

- Mana affordability requires the server's `/auto-tap-preview`
  endpoint (S15). Hitting it on every priority window for every
  card in hand, on every snapshot tick, is a perf cliff. The
  endpoint is designed for one-shot lookups, not hot-path
  scanning.
- The predicate's job is to skip _dead air_, not to divine
  intent. A player who can't afford their one sorcery usually
  still wants to see their upkeep land — if only to draw into
  something castable on the next priority window.

Future sprint (punted): a light affordance-precomputation on the
client that runs once per snapshot would let the predicate
tighten. Not shipping here.

### 5. Manual one-time stops (click-to-pin on phase icons)

**Decision:** `client/src/lib/priorityStops.ts` owns a Svelte store
of pinned `StepID`s. Clicking a priority-granting phase icon on the
PhaseDisplay track toggles that step in the set. When the auto-pass
effect sees a pinned step, it short-circuits before any other rule
(stops grid, smartAutoPass) — the cursor holds regardless. The pin
is consumed on the next snapshot step transition, so the next cycle
of that step is unpinned unless re-clicked.

**Precedence order (strongest first)** — amended by #526, which put the pin
above the autopass toggle of decision 6 as well:

1. Manual pin → hold (this decision).
2. `stepStops[step] === true` + `smartAutoPass=true` + predicate
   says "something to consider" → hold.
3. `stepStops[step] === true` + `smartAutoPass=false` → hold.
4. Everything else → auto-pass.

**Why manual stops override smartAutoPass:**
The whole point of a manual pin is "I want the cursor even though
the engine sees no reason for it." A player might want to think
about an upcoming line, fake a cast to bluff a counter, or
consider a play the catalog doesn't model yet (our coverage is
opt-in; lots of legal actions aren't predicate-visible). If
smartAutoPass could skip a pinned step, the affordance would be a
lie — so pins sit above the predicate in the hierarchy.

**Why priority-granting only:**
Untap and Cleanup don't grant priority (CR 502.4 / 514.3). Pinning
them would put a visible affordance on a step where the server
actively rejects `pass_priority`. `canManuallyStop(step)` gates
the toggle; the click handler no-ops on non-priority steps, and
the rendered icon shows a "no priority" tooltip so the missing
affordance is self-explanatory.

**Why one-time, not sticky:**
Sticky pins are the stops grid (persisted, persistent intent).
Manual pins are now-intent ("this cycle I want to see declare
attackers"). Sticky would duplicate the grid with worse ergonomics
(no Settings UI to unpin) and risk "I pinned this days ago and
forgot" stuck-cursor incidents. One-time consume on step
transition matches the feature as described and keeps state
disposable.

**Why not persisted across reloads:**
The store is process-scoped — a reload drops all pins. Consistent
with one-time semantics (a reload is a stronger signal than a step
transition), and anyone who reloads mid-game probably wants the
default stops behaviour to reassert.

### 6. Pared-down priority toolbar: just `next` + `autopass`

**Decision:** Replace the (pre-S13.6) six-button priority toolbar
(next step / pass priority / ⇥ pass step / → next stop / pass
until end of turn / pass turn) with two controls that live inside
the PhaseDisplay box itself, below the priority pills:

- **`next`** — fires a single `pass_priority`. Labelled "next"
  because the meaningful action is "move to the next priority
  window," not "pass [the thing I have]."
- **`autopass`** — a session toggle (not a one-shot). When on,
  the auto-pass `$effect` ignores every gate (stops grid,
  smartAutoPass predicate, ~~manual pins~~, `settings.autoPassPriority`)
  and fires `pass_priority` on every snapshot where the viewer
  holds priority. Persists until clicked off or reload.
  **Amended (#526): manual pins are no longer among the gates it
  ignores** — see the amendment near the top of this ADR.

`pass turn` (active-player-only whole-turn skip) and `undo` stay in
the Game.svelte toolbar — they're not priority passes, and they
belong with the other table-wide shortcuts.

**Why the toolbar got pared down:**
The intelligent auto-pass system (stops grid + smartAutoPass +
manual pins) makes the batch-pass buttons redundant — the cursor
advances to the next configured stop automatically. The only thing
a player needs a button for is the two extreme cases: "one more
step please" (`next`) and "I'm out, stop asking" (`autopass`).

**Why autopass lives inside PhaseDisplay (not the main toolbar):**
The priority pills + phase track + step label are the information
context; the two buttons are actions against that context. Keeping
them in the same visual box trims eye-travel and makes the widget
self-contained. Game.svelte's main toolbar stays focused on
one-shot game actions (draw, untap all, shuffle, life, pass turn,
undo) that aren't priority-specific.

**Why autopass must only fire when the viewer holds priority:**
The first iteration of autopass was a `passToEnd`-style loop that
bashed `pass_priority` until the step hit cleanup. That loop
didn't check `viewerHasPriority` between sends — during an
opponent's turn, priority rotates off the viewer, and every
subsequent send was rejected with `bad_request: you do not hold
priority`. The toggle form sidesteps this: the `$effect` fires
only when `viewerHasPriority` is true, so opponents' turns don't
generate rejection spam.

### 7. Autopass safety belt (`autopassPersistThroughTurns: false`)

**Decision:** The autopass toggle auto-clears the first time the
cursor enters the viewer's own `precombat_main` step — preventing
the nightmare case where a forgotten autopass skips your own turn.
A new gameplay setting `autopassPersistThroughTurns` (default
`false`, schema v6) is the opt-out: flip it on to keep autopass
engaged indefinitely. Labelled DANGER in the Settings UI with
explicit warning copy ("WARNING: ENABLING THIS SETTING MAY CAUSE
YOU TO SKIP YOUR OWN TURN").

**Why the safety is default on:**
The autopass toggle is tempting to leave on across turns ("I'll
turn it off when it matters"), but the failure mode — losing your
entire main phase to a one-click-forgot — is much worse than the
inconvenience of having to re-click autopass each rotation.
Default-on safety puts the sharp edge behind an explicit opt-in.

**Why precombat_main (not earlier):**
A paranoid reading would clear autopass on the viewer's untap
step, but that's premature: most players actively want autopass to
carry them through their own upkeep and draw steps when they're
tapped out and nothing is happening. The one step that almost
universally matters is precombat_main — that's where you cast
things. Disabling there is the minimum-surprise default.

**Why a danger label (not just a toggle):**
The opt-in doesn't merely change UX — it changes whether the
player can lose a turn to a forgotten UI state. Danger styling
(amber border, WARNING copy in caps) matches how other
minefield-adjacent settings are marked in similar projects. The
cost of over-warning a rare "I know what I'm doing" case is low;
the cost of a player flipping the toggle without noticing is a
missed turn.

## Consequences

### Good

- A four-player turn cycle where nobody has anything to respond
  with auto-walks from Untap to Untap with exactly the clicks
  players intentionally made — no drumbeat of pointless passes.
- The predicate surface is tiny (one function, one cache) and
  trivially extensible. When S13.5 visibility tightens, we can
  refine the opponent-legality questions; when S15-derived mana
  affordability is worth precomputing, we can fold it in.

### Tradeoffs

- False-positive-stop: the viewer controls _any_ permanent →
  `hasAnyLegalResponse` returns true. In Commander that's almost
  every mid-game priority window. Practically this means
  smart-auto-pass shines in empty-board and empty-hand scenarios
  (turn 1, board-wiped mid-game, etc.) and is a gentle no-op
  once players are established. That's fine — the frustration it
  addresses is precisely the empty-state drumbeat.
- Server-side authority: smart-auto-pass dispatches real
  `pass_priority` actions. If the client's predicate diverges
  from the server's legality model (e.g. an unshipped split-
  second effect that the server enforces but the client doesn't
  surface), the client might auto-pass through a priority window
  the server would have given the viewer anyway. Mitigation: the
  server's action dispatcher is authoritative — the worst case is
  the exact same "viewer had the chance, nothing happened" state
  you'd get by clicking pass by hand.

### Not done

- **Opponent "thinking" indicator.** Accurate telegraph of
  "opponent can respond" requires their full hand / board
  visibility, which S13.5 doesn't grant across seats. Future
  sprint.
- **Triggered-ability responses.** S19 will auto-fire catalog
  triggers; predicate doesn't consider them since they aren't
  the viewer's action to dispatch.
- **Mana affordability.** Out of scope per §4.
- **Per-player granularity on the smartAutoPass toggle.** Global
  for now; split if users ask.
