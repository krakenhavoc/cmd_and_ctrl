# ADR 0178 — Life instructions after a replaced move

Date: 2026-10-14
Status: Accepted
Sprint: S58 — Deck requests, October batch follow-up
Issues: #2950, #2910

## Context

The Darkness Crystal replaces a death with exile and a gain of 2 life. Its
callback currently starts the life event before the exile physically occurs.
When life replacements ask for CR 616 ordering, the parent move continues and
runs the original effect's continuation while that life question remains open.
A real destroy-then-draw regression reproduces the early continuation.

The repository-pinned 2026-09-25 rules require printed instruction order
(608.2c), allow a modified event to contain several instructions (614.6), and
consider replacements of contained events after the first replacement is
chosen (616.1g). An impossible move instruction does not erase a separate
life instruction. A stale original event that never occurs is different
(614.7). Existing move and life replacement pipelines already handle all
individual operations; their structured continuations need one link.

## Decision

1. Add a narrow value-only life instruction list to the transient
   ReplacementEvent for a replaced move. Each entry stores source ID, affected
   player ID and signed life delta. A locked-context helper validates the
   parent kind and appends it. The Crystal redirects the move and stamps its
   existing exile link, then declares its gain instead of executing it there.
2. The shared battlefield-leave finisher lands the modified move first. It
   then runs the declared life instructions through the ordinary life
   replacement pipeline, holding the outer route continuation until they all
   settle. This covers both inline moves and resumed move ordering.
3. A typed private continuation on lifeTail stores the remaining instruction
   values, the outer zoneRoute and its simultaneous-exit context. It captures
   no mutable parent event or route in a Go closure. The ordinary life tail
   clears itself before resuming this sequence; the sequence consumes each
   instruction once, then runs the outer route tail once. Publish the carried
   simultaneous-exit context while resuming that tail, as ordinary exits do.
4. Declared life instructions survive a later destination redirect or
   cancellation of the move subevent, per 614.6. A stale parent move abandoned
   before it occurs discards its instructions but still settles the existing
   route continuation. A prohibited or replaced-away gain settles its own
   life tail with zero and advances normally. Player/source identities are
   snapshots of the replacement's application, so later source departure
   does not rewrite who was to gain life.
5. Cost moves pass their existing mustSettleNow posture into the contained
   life events. An indivisible cost never pauses on an inner ordering prompt.
   Effect-time moves retain their ordinary choice window.
6. Clone replacement resumes with fresh instruction slices and a fresh typed
   continuation/route, including simultaneous context, so answering a life
   prompt cannot consume an undo snapshot's outer continuation. Additive
   private frame data adds no persisted card, event or snapshot field and
   does not change schema v8 or frozen fixtures. Existing continuation census
   and closure inventory must continue to count the entire paused resolution.

## Validation

Pin exile-before-life-before-original-continuation for inline and nested
ordering paths. Test a parent move ordering prompt followed by a life prompt,
multiple sequential life instructions, prohibited/cancelled life, redirected
and cancelled moves, stale parent withdrawal, affected-player departure,
source departure, must-settle-now costs, simultaneous-exit context and
independent clone/undo replay. Keep the Crystal's discount, nontoken/opponent
filter, linked targeted reanimation, tapped entry and additional counters.
Run affected game/effects/legal/protocol/coverage/roadmap suites and the
snapshot/census/source guards. Do not close the issue on its existing Full
catalog declaration; close after verified PR merge.

## Number reservation

Fetched all remote refs and checked 13 branches on 2026-10-14: origin,
origin/cost-ledger, origin/develop, origin/docs/3337-away-seats,
origin/docs/issue-audit, origin/feat/750-conditional-block-restrictions,
origin/feat/playmats, origin/fix/2948-liliana-simultaneous-sacrifice,
origin/fix/3159-3160-attachments, origin/fix/3338-tablet-card-reading,
origin/fix/caddy-reload-admin-off, origin/main, origin/wip/836-one-click-default.
0175–0177 were occupied; 0178 was the first free number.

## Implementation (2026-10-14, #2950)

Finished from Ian Robinson's design commit, unchanged above. The names
and the places it landed:

- `Game.ChangeLifeAfterReplacedMoveForEffect(ev, source, player, delta)`
  is the locked-context helper (decision 1). It refuses any event but
  `RepEventMove` and appends a `lifeInstruction` value to the
  unexported `ReplacementEvent.lifeAfterMove`. The Darkness Crystal's
  `Replace` calls it instead of `ChangePlayerLifeForEffect`.
- `runMoveTailLocked(ev)` replaces `runRouteTailLocked(ev.zoneRoute)`
  at each terminal outcome of a zone change: the battlefield-leave
  finisher, the routed-move finisher, the cancel branches of both entry
  points and of the CR 616 resume, and the sandbox move (decisions 2
  and 4). With instructions present it takes the route's continuation
  off the route, lands any paired simultaneous entry with the exit,
  then runs the instructions in order.
- The continuation (decision 3) is `moveLifeSequence{remaining, route,
  mustSettleNow}`, a value holding the instructions not yet run and a
  COPY of the route. It rides each life change's existing
  `lifeTail.then`, captured by value. No mutable parent event or route
  pointer is captured, and each run copies the route again before
  running its tail. So an undo snapshot that shares the tail
  (`cloneReplacementResume` copies the tail pointer) replays the same
  continuation instead of finding it consumed. A typed field on
  `lifeTail` was tried first. It added two new routes from Game to a
  closure, and the ADR 0041 P3 ratchet (`closure_fields.txt`) does not
  let the `ChoiceResumeFrames` class grow. The last instruction runs
  the route's tail inside the route's simultaneous exit.
- `cloneReplacementResume` copies `lifeAfterMove` (decision 6), and its
  route copy moved into `cloneZoneRoute`.
- An abandoned move (its prompt taken away) still goes through
  `abandonZoneRouteLocked`, which discards the instructions and settles
  the route (decision 4).

Found while testing, and out of scope here: catalog replacement IDs are
packed from the battlefield index (`encodeCatalogReplacementID`). If a
permanent earlier on the battlefield leaves while a CR 616 prompt is
open, the paused event's once-per-event marks and the prompt's IDs then
name the wrong card, so an applied replacement can apply again. This
happens with any paused life event, not only this one, and is filed as
its own bug (#3482).

## Amendment (2026-10-29, #4517): a reflexive trigger after a replaced move

Head of the Hunt: "If a creature an opponent controls would die, exile it
instead. When you do, create a 2/2 green Wolf creature token." The second
sentence is a reflexive triggered ability (CR 603.12), but its creator is a
replacement, not a resolving spell or ability, so
`QueueReflexiveTriggerForEffect`, which needs the resolving parent item,
cannot create it. Triggering on the exile event instead was rejected in
#4454: it would also fire when Rest in Peace, or any other effect, exiled
the creature, which is stronger than printed.

1. The replacement declares the trigger on the move it modifies, as it
   declares a life change (this ADR) or an extra turn (ADR 0188 Decision 2):
   `Game.QueueReflexiveTriggerAfterReplacedMoveForEffect(ev, src, rt)`,
   called in `Replace` after the destination is rewritten. The instruction
   (`reflexiveMoveInstruction`) is plain data: the replacement's source and
   source object, its controller, the zone the replacement sent the card to,
   and the trigger's label, `ReflexiveBody` key and params. The controller
   is fixed when the replacement applies (CR 603.12 gives the trigger to the
   creator's controller), so the source leaving afterwards changes nothing.
   `Optional` and `Payload` are refused for now; no card needs them.
2. The move's finisher (`runMoveTailLocked`) creates each declared trigger
   only if the move was not cancelled and the card is now in the declared
   zone ("when you do" means "when you exile it this way"). A later
   replacement that sends the card elsewhere or cancels the move means this
   replacement's action did not happen, so there is no trigger. A
   controller who has left the game gets none (CR 800.4a). The trigger goes
   through the same dispatch a resolving effect's reflexive trigger takes
   (`dispatchReflexiveTriggerLocked`, factored out of
   `QueueReflexiveTriggerForEffect`), onto `PendingTriggers`, so it goes on
   the stack the next time a player would receive priority (CR 603.3).
   Queuing it cannot pause, so it needs none of the life sequence's
   continuation machinery.
3. Each replaced move declares its own instruction, so a wipe that exiles
   three creatures creates three triggers (CR 603.12a). When the affected
   object's controller applies Rest in Peace first (CR 616.1), the creature
   no longer "would die", Head of the Hunt never applies, and nothing is
   declared.
4. Snapshot: once created the trigger is an ordinary keyed triggered item.
   A move event paused on a later CR 616 question with an instruction
   declared writes it as `replacementResume.event.reflexiveAfterMove`
   (additive in v9) under a resume key of its own,
   `replacement-resume/move-reflexive`, which takes precedence over every
   other move key so that a binary that would drop the instruction refuses
   the file. Restore checks each instruction's body key. No new route from
   `Game` to a func: the closure ratchet does not move.

Not addressed: Kalitas, Traitor of Ghet's "instead exile that card and
create a 2/2 black Zombie creature token" makes the token as part of the
modified event, not as a trigger. That is a token-creation instruction
after a replaced move, which can pause on its own CR 616 question, and
needs the life sequence's continuation handling rather than this one.

## Amendment (2026-11-05, #4522): a token creation after a replaced move

Kalitas, Traitor of Ghet: "If a nontoken creature an opponent controls would
die, instead exile that card and create a 2/2 black Zombie creature token."
The token is the second instruction of the modified event (CR 608.2c, CR
614.6), not a "when you do" trigger, so nothing goes on the stack between
the exile and the Zombie. The creation is a CR 701.7b event with its own CR
614 window, and each token's entry has another, so either can pause on a CR
616 question. The move's continuation has to wait for both, as it waits for
a life change (decision 2).

1. The replacement declares the creation on the move, after rewriting the
   destination: `Game.CreateTokensAfterReplacedMoveForEffect(ev, spec)`.
   The instruction (`tokenMoveInstruction`) is the source, the controller
   and the token groups, fixed when the replacement applies. A creation
   attacking, blocking or entering together is refused; no card needs one.
2. The move's finisher (`runMoveTailLocked`) runs the declared creations
   after the extra turns and reflexive triggers and before the life
   instructions, then the route's continuation. Like a life change, a
   creation survives a later redirect or cancel of the move (CR 614.6: a
   commander moved to the command zone still leaves a Zombie), is
   discarded with a move abandoned before it happens (CR 614.7), and is
   skipped for a controller who has left the game (CR 800.4a). A move paid
   as a cost creates the tokens without asking (decision 5); a token's
   entry is not held to that, since no entry event carries the posture.
3. A creation's continuation is a keyed tail all the way through its
   entries (`tokenTail`, then `tokenEntryRun`), and a route cannot ride a
   keyed tail. So the rest of the move is parked on the game while the
   creation runs: `Game.movesAwaitingTokens`, a carrier `ReplacementEvent`
   per parked move holding a copy of the route and the instructions not yet
   run, under a fresh event ID. The creation's tail,
   `engine/move-tokens-created`, names that ID, takes the carrier back on
   every terminal outcome (tokens made, creation replaced away, prompt
   abandoned) and goes on from a fresh copy. This is the frozen-entry
   pattern of #4361 S17 (owner decision D4): copy on write, the carrier
   never written after it is parked, so an undo snapshot that shares the
   slice replays the same continuation.
4. Snapshot: additive in v9. A move paused on a later CR 616 question with a
   creation declared writes `replacementResume.event.tokensAfterMove`
   under its own resume key, `replacement-resume/move-tokens`, which takes
   precedence over every other move key, the reflexive one included. A
   parked move is written as `GameSnapshot.MovesAwaitingTokens` through the
   replacement-event mirror; one that is not data is counted
   (`ChoiceResumeFrames`), as is one with no prompt open. Restore checks the
   keys the carrier names. **Closure ratchet:** three new edges, none
   counted: `Game.movesAwaitingTokens` keyed (its tail names it),
   `ReplacementEvent.tokensAfterMove` and `tokenMoveInstruction.Groups`
   rebuilt (token templates, as `ReplacementEvent.TokenGroups`).

Still open: #4521, a wipe that destroys the replacement's source loses it
for the later legs, applies to Kalitas as it does to Head of the Hunt.
