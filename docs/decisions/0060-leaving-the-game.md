# ADR 0060 — Leaving the game (CR 800.4): a departed player's objects leave with them

**Status:** Accepted · 2026-09-18 · S36 — Tables that wedge: the engine's dead ends
**Issue:** [#769](https://github.com/krakenhavoc/cmd_and_ctrl/issues/769)
**Numbering:** on 2026-09-18, after `git fetch --all --prune`, every remote
branch was listed with `git ls-tree docs/decisions/`. The highest number
present anywhere is 0059 (`origin/docs/adr-0059-turn-machinery`, since
merged to `develop`); no branch and no commit reachable from any ref has a
`0060-*` file (`git log --all --diff-filter=A -- 'docs/decisions/006*'` is
empty). 0005, 0024, 0029 and 0030 stay permanently unused per AGENTS.md §4.
**Related:** [ADR 0059](0059-turn-machinery.md) (the rotation seam — what
happens to the TURN when the active player leaves, CR 800.4j),
[ADR 0012](0012-layer-system.md) (layer 2 is where control lives),
[ADR 0007](0007-stack-foundation.md) (the S13.1 stack cleanup this extends)

## Context

Conceding, or losing to a state-based action, set `Player.Eliminated`, swept
the leaver's stack items, pending triggers and prompts, and moved play on.
Nothing touched their **objects**. Their permanents stayed on the
battlefield under their control; their commander stayed in the command zone;
their hand, library and graveyard stayed full. Their creatures kept blocking
and kept dealing combat damage, their anthems kept pumping, their taxes kept
taxing, and a creature they had stolen with a Mind Control stayed stolen —
by a player who was not at the table.

That is not a cosmetic gap. A dead player's static abilities are still
changing what the living players may legally do, which is the difference
between a sandbox that cannot yet do something and a sandbox that is
actively wrong. The owner's direction on this was explicit: **we follow the
rules.** Clearing a departed player's board by hand is what the manual
affordances are a fallback *for*, never a reason to leave a rule
unenforced.

**The rule** (CR 800.4a, Comprehensive Rules effective August 7 2026). When
a player leaves the game, in this order:

1. all objects **owned** by that player leave the game;
2. any effects giving that player **control** of any objects or players end;
3. any objects they controlled on the stack that are **not represented by
   cards** cease to exist;
4. any objects **still controlled** by that player are exiled.

And it is explicitly **not a state-based action** — it happens the moment
the player leaves, not at the next priority boundary.

Step 3 has been in the tree since S13.1
(`cleanupStackForEliminatedLocked`). Steps 1, 2 and 4 are this ADR.

## Decisions

### 1. Leaving the game is a real removal, not a "left the game" bucket

A card owned by the departed player is dropped out of its zone's slice. It
is not moved to exile, and not moved to a hidden per-player "gone" zone kept
for replay.

The alternative — a bucket — was considered for replay and history and
rejected. The engine has no zone for "outside the game" and adding one would
mean a `ZoneKind`, a wire shape, a view filter, a snapshot field and a
`Clone` leg, all to hold objects that no rule may ever look at again. The
history that matters is already kept: `Game.Events` is append-only and
records every card these objects passed through on their way to the
battlefield, plus the `EventPlayerEliminated` that ends them. A replay
rebuilt from the log still shows the departed player's whole game; it simply
stops showing their board afterwards, which is what happened.

**The log reads as one line.** `EventPlayerEliminated` is emitted **before**
the objects go, so anything watching a player lose the game — a listener, a
future "whenever a player loses the game" trigger — sees the board they lost
with. Nothing is emitted per object: see Decision 3.

### 2. Ownership is the test, and every zone is swept

`removeObjectsOwnedByLocked` walks the battlefield, the stack, exile, and
**every** player's library, hand, graveyard and command zone, and drops
every card whose `Owner` is the departed player. Not just their own four
zones: a card they own can be on the shared battlefield under somebody
else's control, in shared exile, on the stack, or — by a sandbox move or an
effect that puts a card into another player's zone — in another player's
hand or graveyard. CR 800.4a names ownership, so the code tests ownership.

The converse holds and is pinned by a test: another player's card sitting in
the departed player's graveyard does **not** leave.

This is also what folds in the loose end S13.1 left. `cleanupStackForEliminatedLocked`
moved the departing player's spell *cards* to exile as the closest available
analogue to "cease to exist". The sweep runs immediately after it, so a card
they own does not stop in exile — it leaves the game like everything else
they own. A spell card on the stack that somebody **else** owns still exiles,
which is step 4 of the rule, correctly, for that object.

This also answers the question [ADR 0057](0057-win-and-lose-by-effect.md)
Decision 3 left to #769 — *"where [a resolving card] goes when its owner has
left"*. Wherever it routes to: the sweep runs after the departure has
finished with the stack, so the card leaves the game from whatever zone it
reached. There is no special case for it, because ownership is the only
test.

### 3. Nothing announces: no zone change, no LTB, no dies trigger

The removal writes the zone slices directly instead of going through
`MoveCard` and the event emitters. That is the point rather than a shortcut.
Leaving the game is not a zone change, so there is no `EventZoneMove`, no
`EventLTB`, no `EventETB`, no `EventSacrifice`, and no CR 603.10
last-known-information snapshot for an event that never happened. A
departing player's Blood Artist does not drain the table on its way out, and
a survivor's "whenever a creature dies" does not see six deaths when a seat
concedes.

The one thing that **does** announce is step 4's exile (Decision 5 below):
that is a real zone change off the battlefield, so it goes through
`executeBattlefieldLeaveLocked` and its triggers fire.

### 4. The rule's order is the implementation's order

`leaveGameObjectsLocked` performs 1, then 2, then 4, and the order is
load-bearing rather than pedantic:

- **Mind Control** is an Aura owned by the departing player. Step 1 takes
  the Aura out of the game, which ends its layer-2 continuous effect, and
  the recompute hands the creature back to the player it entered under
  (`BaseController`, CR 613.1b). The creature is **not** exiled, because by
  the time step 4 looks, it is not controlled by the departed player any
  more. That is the printed behaviour.
- A permanent whose **baseline** is the departed player — another player's
  card put onto the battlefield under their control, reanimated out of an
  opponent's graveyard — has no effect to end. Step 4 finds it still
  controlled by a player who is not there, and exiles it.

Step 2 costs one manual `layerVersion` bump and one unconditional
recompute, because step 1 emitted no event for the listener to notice.
Every control-changing effect the engine has is a layer-2 continuous effect
sourced from a permanent (ADR 0012, the S24 Mind Control work), so removing
the source **is** ending the effect; there is no separate registry of
control grants to walk.

### 5. A departure that ENDS the game does not strip the board

When the departing player is the second-to-last, the game is over. That
departure skips the object sweep: the final board — both players' — is what
the winner, the post-game screen and the replay look at.

This is a presentation decision about a game that has already ended, not a
rules concession. No rule reads the table again, nothing can observe the
objects leaving, and no state check ever runs on an ended game. It is the
same judgement `endGameIfDecidedLocked` already makes about the turn cursor,
and for the same stated reason: *"a game that has ended keeps its cursor
where it was … ending the last turn would sweep marked damage and pull
attackers out of combat on the board the game ended with."* Pinned by
`TestGameOverKeepsTheFinalBoard` (which predates this change) and
`TestGameEndingDepartureKeepsTheFinalBoard`.

The consequence to be aware of: in a two-player game, CR 800.4a is never
observable, because every departure ends the game. Every test of this
behaviour therefore needs three or more seats.

### 6. The commander leaves with its owner, and is not offered the command zone

A commander is an object its owner owns, so it leaves the game with
everything else they own — from the battlefield, the command zone or
anywhere else. No CR 903.9 "put it in the command zone instead?" prompt is
offered, because that replacement applies to a **zone change**, and this is
not one (Decision 3); and because the command zone it would move to is
leaving the game in the same breath.

`Player.CommanderDamage` is left alone. It is keyed by commander instance ID
and is a record of damage already dealt to a living player; a departed
player's commander leaving does not un-deal the 21 it dealt to somebody
else.

### 7. CR 800.4c is performed on the next state-based-action pass

CR 800.4c is the case CR 800.4a cannot see. B leaves while D's creature is
under **A's** Control Magic: the creature is not "still controlled by B", so
step 4 correctly leaves it alone. Later A's effect ends, and the creature
would go back to its baseline controller — B, who is not there. Nobody can
control it, so it is exiled.

The rules perform that as part of the effect ending. The engine performs it
in `stateBasedActionsLocked`, which is the first moment after an effect ends
that anything looks at the whole board — the same place and for the same
reason the legend rule's choice is queued. A one-battlefield-walk check per
SBA pass, skipped entirely on an ended game (Decision 5).

**Why not in the layer recompute**, which is where the effect actually ends:
the recompute runs inside `ReadSnapshot` on behalf of the view builder, and
a snapshot read must not move cards between zones or emit events.

### 8. Nothing else in CR 800.4 is in scope, and here is what is left

Done here: **800.4a** (this ADR), **800.4b/d** to the extent the engine can
produce them — a delayed trigger scheduled by a departed player and any
triggered ability of theirs left on the APNAP queue are dropped, because
nothing goes on the stack under the control of a player who is not in the
game. Already done elsewhere: **800.4e** (a departed player is not a legal
attack target, `attack_target.go`), **800.4j** (ADR 0059 Decision 6's
declared simplification: the departed active player's turn ends at once
rather than continuing without an active player), **800.4k** (their turns
are skipped).

Still open, and deliberately not folded in here:

- **800.4g/h — a choice owed by a departed player on somebody ELSE's
  object.** `cleanupStackForEliminatedLocked` drops every prompt whose
  chooser has left. That is right for their own objects and wrong for an
  opponent-choice prompt on a living player's spell, where the rule says the
  choice is **reassigned**, not dropped. Reassignment needs an APNAP policy
  for who inherits it and a path through the bot enumerator and the client;
  it is a separate slice with its own decisions and it is not made safer by
  being bolted onto this one. The current drop is at least not a wedge — it
  is what stops the table blocking behind an unanswerable prompt.
- **800.4i — last known information about a departed player.** Unverified.
- **800.4m — "until that player's next turn" durations.** Ties to #755.

## Consequences

### Good

- A conceded or dead seat stops affecting the game. No phantom blockers, no
  phantom anthems, no phantom taxes, and no combat damage from a player who
  is not playing.
- Concede and every state-based loss (0 life, 21 commander damage, 10
  poison, empty-library draw) go through one door, `leaveGameLocked`, so
  they cannot drift apart.
- A stolen permanent behaves the way the card reads: the Mind Control leaves
  with its owner and the creature goes home; a permanent with no effect
  holding it up is exiled.
- Leaving the game fires no triggers, so a conceding player cannot use their
  own departure as a board wipe with upside.

### Tradeoffs

- The removal is not reversible except through undo. There is no "the cards
  are over here" bucket to recover from if a future rule wants one; the
  event log is the record.
- An eliminated seat's panel goes empty on the client — every zone count
  drops to zero. That is already a state the client can render (an empty
  hand, an empty graveyard and an empty command zone are all ordinary), and
  the seat itself still appears, greyed, with its life total and its
  "eliminated" tag.
- `exileGhostControlledLocked` walks the battlefield on every SBA pass, on
  every table, to catch a case that needs two control effects and a
  departure. That is one slice walk against a pass that already walks the
  same slice three times.
- In a two-player game none of this is observable (Decision 5). Every test
  of CR 800.4a needs a four-seat table, which is a trap worth knowing about
  before writing the next one.
