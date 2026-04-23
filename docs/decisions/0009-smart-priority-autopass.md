# ADR 0009 — Smart priority auto-pass (S13.6)

**Status:** Accepted · 2026-04-23 · Sprint S13.6

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
still controls *where the player wants to consider stopping*; the
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
  *potentially* activatable keeps the predicate's false-negative
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
- The predicate's job is to skip *dead air*, not to divine
  intent. A player who can't afford their one sorcery usually
  still wants to see their upkeep land — if only to draw into
  something castable on the next priority window.

Future sprint (punted): a light affordance-precomputation on the
client that runs once per snapshot would let the predicate
tighten. Not shipping here.

### 5. Accompanying `⇥ pass step` button

**Decision:** Add a new priority-toolbar button between "pass
priority" and "→ next stop." Fires `pass_priority` repeatedly
until the step advances, the active seat changes, or the stack
changes. Capped at 24 iterations (same safety belt as
`passToEnd`).

**Why not just reuse "next stop":** "next stop" respects the
stops grid — if the current step *is* a configured stop, pressing
it once from there stops immediately on the very next iteration.
The new button ignores stops entirely: "nothing here, move on to
the next step boundary, I'll decide again there." Covers the
"I've decided there's nothing to do at this one stop" case
without forcing the player to temporarily un-configure the stop.

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
- False-positive-stop: the viewer controls *any* permanent →
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
