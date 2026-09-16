# ADR 0006 — Priority foundation (S13)

**Status:** Accepted · 2026-04-20 · Sprint S13

## Context

S13 is the kickoff of the **B→C rules graft track** (see
[PLAN.md §6](../../PLAN.md#6-phased-roadmap) and
[docs/sprints.md S13](../sprints.md#s13--priority-foundation-rules-graft-kickoff)).
Up through S12 the engine was an Option-B sandbox: every action — draw a card,
untap permanents, pass priority — was a manual click. Real Commander games
exposed two problems that the sandbox couldn't paper over:

1. **Cleanup is busywork.** Untap and Cleanup don't grant priority in real
   MTG (CR 502.4 / 514.3). Forcing every player to click `pass_priority`
   through them was at minimum 12 extra clicks per turn cycle, the active
   seat actually had to *also* click `untap_all` and `draw_card` to do the
   things the rules say happen automatically.
2. **A 4-player turn cycle is too many clicks.** Even on priority-granting
   steps, the active player almost never has anything to do on their own
   upkeep / draw / end / opponents' main phases. The pre-S13 client had a
   global "auto-pass priority on opponents' turns" toggle but no per-step
   control. Real engines (Magic Online, Arena, Cockatrice) all expose a
   stops grid.

S13's scope is the **priority semantics + turn-based-action automation**
slice. The stack and SBAs land in S13.1; counter SBAs in S13.2; greyed
illegal actions in S13.3; interactive cleanup discard in S13.4. This ADR
captures the architectural decisions that this slice locks in and that the
later sprints inherit.

## Decisions

### 1. `PriorityHolder == -1` sentinel for no-priority steps

**Decision:** Reuse the existing `Turn.PriorityHolder int` field; introduce
`game.NoPriority = -1` as the sentinel meaning "no player holds priority
right now." `Turn.advance` lands on Untap and Cleanup with PriorityHolder set
to `NoPriority`; every other step lands on `ActiveSeat`.

**Why a sentinel and not a separate enum / bool:**
- Adding a new `Turn.HoldsPriority bool` field would have meant a protocol
  change for every existing client + every snapshot in the replay store.
  The `int` field is already on the wire; flipping its semantics from
  "always a seat index" to "seat index OR -1" is a single value-shape
  expansion that clients can decode opportunistically.
- The `actions.requirePriorityHolder` guard at
  [actions.go:128](../../server/internal/actions/actions.go) already
  defensively bypasses on `PriorityHolder < 0`. The sentinel formalises
  what was already a code-level invariant.
- TypeScript clients decode `priority_holder: number` and lose nothing; UI
  derives `priorityHeld = (priority_holder >= 0)` cheaply.

### 2. Auto turn-based actions live in `runStepEntryHooksLocked`

**Decision:** The S07 hook function that previously refreshed undo budget
and resolved combat damage now also auto-untaps on `StepUntap`, auto-draws
on `StepDraw`, and auto-advances on `StepCleanup`. The hook recurses through
`Turn.advance` for the no-priority steps so a single advance from `StepEnd`
walks all the way through Cleanup → next-Untap → Upkeep inside one write
lock.

**Why the step-entry hook (vs. a per-tick loop or a dedicated
"turn-based-action processor"):**
- Step transitions already funnel through one function. Every existing
  call site (AdvanceStep, PassPriority's wrap branch, KeepHand-on-mulligan-
  close, advancePastEliminatedLocked, PassTurn) gets the auto-fire for
  free.
- Synchronous recursion under the same write lock is the simplest mental
  model: "advancing the cursor is one mutation, even if multiple steps
  collapse." Snapshots / undo / WS broadcast all see the post-collapse
  state, never the intermediate.
- A dedicated processor or tick loop would add a goroutine, lock-handoff
  semantics, and a "what runs first when the stack arrives" question we
  don't have an answer to until S13.1. Keep the model boring.

### 3. `Game.StartingSeat int` for the turn-1 skip-draw rule

**Decision:** Add a new `Game.StartingSeat int` field, set in `Start()` to
the active seat at game start (always `0` today). The `StepDraw` hook
checks `Turn.Number == 1 && Turn.ActiveSeat == StartingSeat` and skips the
auto-draw per CR 103.8a. The field is carried in `GameView` so spectators
and reconnects see the same skip-draw decision.

**Amendment (2026-09-16, #692):** the skip is gated on the game having
exactly two players. CR 103.8a covers only a two-player game and CR 103.8b
only Two-Headed Giant (a format the engine does not implement); CR 103.8c
says that in all other multiplayer games no player skips the draw step of
their first turn. The hook therefore also requires
`startingPlayerCountLocked() == 2`, which counts seats and *not* who is
still alive — the rule keys off the game's player count at the start, so a
turn-1 concession does not turn a three-player game into a two-player one.

**Why a separate field vs. inferring from turn state:**
- "First player" can't be inferred from the current cursor — by turn 2 the
  cursor has moved on, but if a future feature ever lets a player rejoin
  mid-game we still need to know who skipped. A sentinel field is the
  durable record.
- Pre-S13 replays predate this field; protobuf-style implicit defaults of
  `0` happen to match the only seat games started on before this field
  existed, so old replays decode correctly without a migration.
- It avoids a "is this turn 1?" check sprinkled across multiple code
  paths — there's exactly one consumer.

### 4. Per-step stops are client-only (no protocol bump)

**Decision:** The per-step stops grid lives entirely in
`localStorage.cmdctrl.settings.v1.gameplay.stepStops`, persisted via the
S11.5 settings system. The server doesn't know which steps any seat wants
to stop on. The client's auto-pass effect simply doesn't fire
`pass_priority` on a stop step.

**Why client-only:**
- Stops are a UX optimisation, not a rule. Two players at the same table
  can have completely different stop preferences and the game is still
  consistent — neither's behaviour affects the other's board state.
- Putting stops on the wire would mean every snapshot carries 10
  booleans × 4 seats of preferences metadata that no other client cares
  about. Wasteful and protocol-coupled.
- The S11.5 settings system already handles persistence + migration; we
  reuse it. New `__version: 2` schema seeds defaults during the v1 → v2
  bump.

### 5. Manual `untap_all` / `draw_card` actions stay dispatchable

**Decision:** The pre-S13 manual actions remain registered in the
dispatcher. Inside their mutations, they no-op when called during the auto-
fire window for the active seat (`Turn.Step == StepUntap` and active player
for `UntapAll`; `Turn.Step == StepDraw` and active player for `DrawCard`).
Outside the window, they behave exactly as before.

**Why keep the manual actions:**
- **Sandbox flexibility.** Casual play sometimes wants "untap that opponent's
  permanent because their card said to" — with the action gone, players
  would have no escape hatch.
- **Replay backwards compatibility.** Every existing replay in the undo
  stack contains `untap_all` and `draw_card` actions. Re-applying them
  on restore must remain a no-op (instead of a hard error) so the undo
  path doesn't bit-rot.
- **Migration cost is zero.** Gating to no-op during the auto-window is
  one branch each.

### 6. KeepHand fires the first step-entry hook

**Decision:** The very first auto-untap doesn't fire at `Start()` time
(mulligans are open then — the active seat hasn't even kept their opening
hand). Instead, the moment the last seated, non-eliminated player calls
`KeepHand` and `MulligansOpen` flips false, `KeepHand` invokes
`runStepEntryHooksLocked` itself. That fires seat 0's auto-untap and walks
the cursor on to Upkeep.

**Why not at Start():**
- Auto-untapping a seat that hasn't decided whether to keep their opening
  hand would be confusing — the keep dialog would briefly show "your turn
  is starting" while the player is still making the keep/mulligan call.
- The MulligansOpen guard is a one-line check inside the StepUntap branch
  of the hook (`if g.MulligansOpen { return }`), and KeepHand calls the
  hook unconditionally on close — both paths cost essentially nothing.

## Consequences

- **AdvanceStep is no longer a pure cursor move.** Externally, the
  observable step sequence is now Upkeep → Draw → … → End → next seat's
  Upkeep — Untap and Cleanup are invisible to AdvanceStep callers. The
  `TestFullFourPlayerTurnCycle` exit-criteria test in
  [game_test.go](../../server/internal/game/game_test.go) was rewritten to
  assert the new shape; future tests should use the same convention.
- **PassPriority can fail with `ErrNoPriority`.** Callers that previously
  treated PassPriority as infallible during normal play need to handle
  this error (or check `priority_holder >= 0` first). The client's
  passToNextStop helper does the latter — it skips `pass_priority` on
  steps that don't grant priority and just waits for the next snapshot.
- **Snapshot wire format gains `starting_seat`.** Pre-S13 snapshots decode
  with `starting_seat: 0` (Go's int zero), which matches every game ever
  started before S13. No migration needed.
- **Per-step stops settings schema bumped to v2.** The migration in
  [client/src/lib/settings.ts](../../client/src/lib/settings.ts) seeds
  defaults for any user with an empty stepStops map and strips entries
  for no-priority steps. User-configured stops survive untouched.
- **Step-entry recursion is bounded.** Cleanup → Untap → Upkeep is at
  most 3 step transitions per AdvanceStep call. The recursive
  `runStepEntryHooksLocked` calls are cheap (no I/O, all under the same
  write lock) and the recursion depth is provably small (the cursor
  always lands on a priority-granting step within 2 hops).
- **S13.4 inherits the cleanup gap.** The current cleanup hook auto-
  advances unconditionally; S13.4 will replace that with the interactive
  discard pause + per-player MaxHandSize. Until S13.4 lands, hand sizes
  exceed 7 with no enforcement (matches pre-S13 behaviour).
- **S13.1 inherits the priority shape.** The stack work in S13.1 will
  hook `cast_spell` and `auto-resolve` into the same priority-rotation
  mutation; the `NoPriority` sentinel + step-entry hook architecture
  scales to "stack-aware priority resets" without further restructuring.

## Alternatives considered

- **Drop manual untap_all / draw_card.** Rejected — replays + sandbox
  flexibility (see decision 5).
- **Server-pushed per-step stops.** Rejected — protocol coupling for
  client-only UX (see decision 4).
- **Run auto turn-based actions on a goroutine tick loop.** Rejected —
  goroutine handoff + lock semantics for no actual benefit. Synchronous
  recursion under one write lock is simpler and easier to test.
- **`SkipNextDrawForActive bool` instead of `StartingSeat int`.** Considered
  in the original [s13-priority-foundation.md](/home/node/.claude/plans/s13-priority-foundation.md)
  research synthesis. Rejected during the 2026-04-20 re-plan: a one-shot
  flag works for the common case but doesn't survive a Concede-then-
  Pass-Turn that would land back on the starting seat at turn 1; a
  durable `StartingSeat` field is the simpler invariant.
