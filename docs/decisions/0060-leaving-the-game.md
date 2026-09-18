# ADR 0060 — Leaving the game (CR 800.4): a departed player's objects leave with them

**Status:** Accepted · 2026-09-18 · S36 — Tables that wedge: the engine's dead ends
**Amended:** 2026-09-18 — the CR 800.4 remainder
([#902](https://github.com/krakenhavoc/cmd_and_ctrl/issues/902)): reassigning
a departed player's choice (800.4g/h), last known information (800.4i), and
"until that player's next turn" (800.4m). Amended again 2026-09-18 — CR
800.4f's consequence and the departure table's drop-action column
([#961](https://github.com/krakenhavoc/cmd_and_ctrl/issues/961)). See the
amendments at the end of this file; Decision 8's "Still open" list is now
empty.
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

Still open, and deliberately not folded in here — **all three closed on
2026-09-18 by the amendment at the end of this file (#902); the list
below is what this ADR shipped with:**

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

---

## Amendment — 2026-09-18: the CR 800.4 remainder (#902)

**Status:** Accepted · 2026-09-18 · S36 — Tables that wedge: the engine's
dead ends
**Issue:** [#902](https://github.com/krakenhavoc/cmd_and_ctrl/issues/902)
(split out of [#769](https://github.com/krakenhavoc/cmd_and_ctrl/issues/769))
**Related:** [#864](https://github.com/krakenhavoc/cmd_and_ctrl/issues/864) /
PR #868 (`EventPendingChoiceDropped` — the wedge this replaces for one
case and keeps for the rest),
[ADR 0063](0063-durations-and-control.md) (`Player.TurnsBegun`, which is
what makes 800.4m already true)

This amendment closes the three items Decision 8 left open. Two of them
were already true and needed a test and a sentence; one is new code.

### Amendment 1 — CR 800.4g/h: a departed player's choice on somebody else's object is reassigned

**The rules, quoted, because the gap is in the detail:**

> **800.4f** If an object requires a player who has left the game to pay
> a cost or choose whether to pay a cost, that cost is not paid.
>
> **800.4g** If an object requires a player who has left the game to make
> a choice other than whether to pay a cost, the controller of the object
> chooses another player to make that choice. If the original choice was
> to be made by an opponent of the controller of the object, that player
> chooses another opponent if possible.
>
> **800.4h** If a rule requires a player who has left the game to make a
> choice, the next player in turn order makes that choice.

**The inheritor policy.** One function, `choiceInheritorLocked`
(`leave_game.go`), and one walk: the seats in turn order starting after
the seat that left.

1. **The first surviving opponent of the object's controller inherits.**
   That is 800.4g's *second* sentence, and it is the sentence that always
   fires. A choice the controller themselves owed cannot reach the policy
   at all: CR 800.4a would have taken the object out of the game along
   with them, and the policy refuses a prompt whose object is gone.
2. **The controller is the "if possible" fallback**, taken only when no
   other opponent is still in the game — 800.4g's first sentence, read
   plainly: the controller chooses another player, and the only other
   player left is themselves.
3. **Nobody left ⇒ dropped**, with `EventPendingChoiceDropped`, exactly
   as before this amendment.

The engine does **not** prompt the controller to nominate somebody,
which is what 800.4g literally describes. Turn order from the departed
seat is the deterministic stand-in, for the reason every other "choose a
player" default here is APNAP: four bots have to reach the same answer
four humans would, and a prompt to decide who answers a prompt is a
second place the table can wedge.

**CR 800.4h reassigns nothing today**, and that is a statement about the
engine rather than about the rule. Every rule-required prompt it has —
the legend rule, a cleanup discard, a mulligan — is a choice about the
asked player's own permanents or hand, so all of them are already in the
never-reassign table below for the *material* reason.

**Three gates, in order** (`reassignDepartedChoiceLocked`):

1. **The kind** — the table below.
2. **The object** — the prompt's `Source` must still be findable in a
   zone and controlled by a player still in the game. 800.4g reassigns a
   choice an *object* requires, so there has to still be one.
3. **The material** — a prompt whose `FromPlayer` is the departed chooser
   themselves is about their own pool, which CR 800.4a took out of the
   game a moment ago. Every answer is a no-op, so it is dropped.
   `FromPlayer` is already the engine's name for whose material a prompt
   is about; protocol's `redactChoiceCards` reads the same field for the
   same meaning.

**Torment of Hailfire is the worked example of gate 3.** "Each opponent
loses 3 life unless that player sacrifices a nonland permanent of their
choice or discards a card" is an `option_pick` addressed to each
opponent, and read literally 800.4g would hand a departed opponent's
copy to another opponent. Every branch of it acts on the player who
left — their life, their permanents, their hand — and all three are gone
under 800.4a, so the reassignment would be a question with no answer
that does anything. It is dropped. The *kind* stays reassignable,
because the second half of a Fact or Fiction pile split is the same kind
over a living player's cards; the discriminator is the material, not the
card.

#### The never-reassign table

Every `PendingChoiceKind`, with the reason. Enforced by one predicate
(`choiceReassignDecisions` + `reassignDepartedChoiceLocked`) and by
`TestEveryChoiceKindHasAReassignmentDecision`, which fails until a new
kind has a row — the same mechanism `choiceGateDecisions` uses, with the
opposite default: **an unclassified kind is dropped**, which is the
pre-#902 behaviour and is safe.

| Kind | Reassigned? | Why |
|---|---|---|
| `pick_target` | **yes** | CR 800.4g — a CR 603.3d target pick over the whole board. |
| `trigger_prompt` | **yes** | CR 800.4g — a CR 603.5 "you may" pointed at another seat (`TriggerOptionalPrompt.Chooser`; Edric, Spymaster of Trest). |
| `choose_cards` | **yes** | CR 800.4g — a pile split's first half, over cards the splitter does not own. |
| `discard_from_hand` | **yes** | CR 800.4g — Thoughtseize-shaped, chooser ≠ the hand. |
| `option_pick` | **yes** | CR 800.4g — a pile split's second half. Torment's own-material ask is stopped by gate 3. |
| `pay_unless` | no | CR 800.4f — that cost is not paid. Its DROP runs the "unless" branch; see the 2026-09-18 #961 amendment. |
| `entry_pay_life` | no | CR 800.4f. |
| `mana_pick` | no | The mana would enter a pool that left the game. |
| `sacrifice_choice` | no | Their permanents left with them (800.4a). |
| `legend_rule` | no | Their permanents. |
| `scry` / `surveil` / `look_at_top` / `search_library` | no | Their library left with them. |
| `may_cast` | no | The offered card, and the free-cast permission, are theirs. |
| `copy_target` | no | The copy is theirs. |
| `choose_creature_type` / `choose_color` | no | CR 614.12-style choices about their own permanent or resolving effect. |
| `damage_assignment` | no | The attacker was theirs. |
| `trigger_order` | no | Their triggers are dropped (CR 800.4d). |
| `mode_pick` | no | CR 800.4d — the trigger is not put on the stack at all, so there is no mode left to choose. |
| `choose_protector` | no | CR 310.9, chosen as their battle enters. |
| `coin_call` | no | The flip belongs to its flipper, whom the untouched frame names. |
| `loop_shortcut` | no | CR 726 — the allowance is on their loop's tally key. |
| `confirm` | **yes** | Since #961: `ConfirmPrompt.FromPlayer` records whose material it is about, so gate 3 can tell a self-question from a cross-table one. Every confirm queued today defaults the field to the chooser and is still dropped. |
| `replacement_order` / `optional_replacement` | no | The CR 616 pair settles its own drop through `finishDroppedReplacementLocked` (#808). |

#### What the reassignment does, and what it does not

`reassignChoiceLocked` (`pending_choice.go`) is the only code that
rewrites a queued `PendingChoice.Chooser`. It:

- **prunes the candidates leaving with the old chooser** — a card they
  own, or one already out of every zone, and the departed player as a
  target. If that empties the question the prompt is **not** moved and
  is dropped instead: handing a seat a prompt with no answers is the
  #544 wedge. `ChooseMin` / `ChooseMax` / `Count` are clamped with the
  list so every answer the enumerator offers is one the resolver takes;
- **rewrites `Chooser`, which IS the re-redaction.** Redaction is
  computed per viewer at VIEW time from `PendingChoiceView.Chooser` and
  `.FromPlayer` (#918's `redactChoiceCards`), so the next snapshot
  already shows the new chooser exactly what that seat may see of a pool
  that is not theirs, and shows the departed seat nothing. There is no
  stored redaction to re-run;
- **emits `EventPendingChoiceReassigned`** — `Actor` the departed
  chooser, `Target` the inheritor, `Source` the object, `Label` the kind
  — the twin of #868's `EventPendingChoiceDropped`, so a stall dump says
  where a prompt went rather than going quiet;
- **leaves the continuation untouched.** Every resume frame is the rest
  of the *card*, and the card did not change because the player
  answering it did.

**The enumerator and the client needed nothing.** `legal.choiceMoves`
already keys on `PendingChoice.Chooser`, and so does the client's
picker, so one field moving re-addresses the prompt for both. Pinned by
`legal.TestReassignedPromptIsEnumeratedForItsNewChooser` (the inheritor
is offered the prompt's answers and nothing else, and every one of them
dispatches) and by
`aiseat.TestConcedeMidPromptLeavesTheBotTablePlaying` (a bot table with
a concede mid-prompt clears the inherited prompt and plays on).

**Undo and the snapshot owe nothing new.** `Chooser`, the candidate
lists and the bounds were already cloned and snapshotted, so an undo
puts the prompt back in front of the player who left and a replayed
departure moves it again.

### Amendment 2 — CR 800.4i is true for a player's own information, and deviates for their board

> **800.4i** If an effect requires information about a specific player,
> the effect uses the current information about that player if they are
> still in the game; otherwise, the effect uses the last known
> information about that player before they left the game. …

**Verified, no code.** A departed seat stays in `g.Seats` with
`Eliminated` set and every scalar as it was — life total, poison,
commander damage, energy — and nothing zeroes a conceding player's life.
So every effect that reads a player through `PlayerByID` reads last
known information by construction. The event log is append-only and
`SpellsCastThisTurn` is keyed by player, so 800.4i's second sentence
("the effect can find actions that were taken by a player who has left
the game") holds too. Both are now pinned by tests rather than left to
luck.

**The deviation, stated so it is a decision.** The engine keeps a
departed player's SCALARS and does **not** keep a pre-departure copy of
their BOARD, because CR 800.4a removed it. So "the number of creatures
that player controls", asked after they left, reads zero rather than
what they had. Closing that means a per-player LKI snapshot taken at the
moment of departure — a new field, a clone leg, a snapshot leg and a
census row — and no card in the catalog asks the question. Out of scope,
pinned by `TestDepartedPlayersObjectsAreGoneNotLastKnown` so a future
change has to come here and say so.

800.4i's third sentence — "actions taken during their *last turn*" can
be found "only until that player's next turn after leaving the game
would have begun" — has nothing to be true or false about: the engine
has no effect that reads another player's last turn.

### Amendment 3 — CR 800.4m is complete, by #921

> **800.4m** When a player leaves the game, any continuous effects with
> durations that last until that player's next turn or until a specific
> point in that turn will last until that turn would have begun. They
> neither expire immediately nor last indefinitely.

**No code.** [ADR 0063](0063-durations-and-control.md) Decision 3 already
made this true: `beginNextTurnLocked` bumps `Player.TurnsBegun` for a
seat the rotation *steps over* because that player has left, and
`UntilYourNextTurn` expires on `TurnsBegun >= ExpiresAtTurnsBegun`. So
the effect ends when the departed player's turn *would* have begun —
neither at the moment they leave nor never.

Proven by `TestUntilYourNextTurnEndsWhenADepartedPlayersTurnWouldHaveBegun`
(#921, four seats, the leaver at seat 2) and, added here for the boundary
#921 did not cover, `TestUntilThatPlayersNextTurnWhenTheyLeaveOnTheirOwnTurn` — the effect is created during the departing player's *own* turn,
that turn then ends early because its active player left (ADR 0059
Decision 6), and the duration still waits a full round for the
never-taken turn.

The engine has no "until a specific point in that turn" duration, so
that half of the rule has nothing to implement.

### Consequences of this amendment

**Good**

- A prompt a living player's card pointed at a seat that then conceded
  is answered by somebody instead of vanishing. The rule is enforced
  rather than approximated, and the table still cannot wedge: every path
  out of the departure sweep either moves the prompt to a seat that can
  answer it or drops it with the event #868 added.
- One reassignment function, one inheritor policy, one never-reassign
  predicate, all rules-named. No card is special-cased and nothing reads
  a card name.
- The bots and the client inherited the behaviour for free, which is the
  #794 property holding: the enumerator and the engine read the same
  field for the same question.

**Tradeoffs**

- **`confirm` is never reassigned**, including the cross-table kind
  (Combustible Gearhulk's question to its target). `QueueConfirmForEffect`
  sets `FromPlayer` to the chooser unconditionally, so gate 3 cannot tell
  a self-question from one asked across the table, and the safe answer is
  the pre-#902 drop. Lifting it is one field — a `From` on
  `ConfirmPrompt`, carried into the `PendingChoice` the way
  `ChooseCardsPrompt.FromPlayer` already is — and deliberately not done
  in the PR that decides the policy. **Resolved by #961**, which added
  exactly that field; see the amendment below.
- **`pick_target`, `trigger_prompt` and `discard_from_hand` are
  classified reassignable but are not reachable from the harvester
  today**, because `queuePickTargetLocked` and
  `QueueDiscardFromRevealedHand` address the prompt to the object's own
  controller and a controller who leaves takes the object with them. The
  rows are decisions about the *rule*, kept so that a card addressing one
  of those prompts across the table (#918 already allows it) does not
  land on a silent wrong default. `trigger_prompt` *is* reachable, via
  `TriggerOptionalPrompt.Chooser`, which is why it is the headline test.
- **CR 800.4f's consequence is not run when a `pay_unless` is dropped.**
  The rule says the cost is not paid, which should run the prompt's
  "unless" branch — Rhystic Study still draws. Today the branch is
  discarded with the prompt. It is a real gap, it is one rule away from
  this one, and it is left out of scope on purpose: the fix is to run a
  continuation from inside the elimination sweep, which is the class of
  change #808 had to be careful about. See
  `pending_choice.go`'s `payUnlessFrame`. **Resolved by #961**; see the
  amendment below.

### Decision 8's "Still open" list, updated

- **800.4g/h** — done, Amendment 1.
- **800.4i** — verified and tested, Amendment 2; one documented deviation
  (a departed player's board is not kept as last known information).
- **800.4m** — complete by #921 / ADR 0063 Decision 3, Amendment 3.
- **800.4b** (an object that would change to a departed player's control
  doesn't; no token is created), **800.4n** (the ante zone) and
  **800.4p** (Planechase) remain unimplemented and unfiled. 800.4n and
  800.4p are variants this sandbox does not play. 800.4b is partly true
  by consequence — `QueueChoiceForEffect` refuses a departed chooser and
  the trigger queue drops a departed controller's abilities — and has no
  test of its own.

## Amendment — 2026-09-18: CR 800.4f's consequence, and the drop-action column (#961)

**Status:** Accepted · 2026-09-18 · S36 — Tables that wedge: the engine's
dead ends
**Issue:** [#961](https://github.com/krakenhavoc/cmd_and_ctrl/issues/961)
(noted by the #902 agent in PR #959)
**Related:** [#808](https://github.com/krakenhavoc/cmd_and_ctrl/issues/808)
(a continuation run from inside the elimination sweep — the pattern this
follows), [#864](https://github.com/krakenhavoc/cmd_and_ctrl/issues/864)
(`QueueChoiceForEffect` refuses a departed chooser, which is what makes
running one safe)

The amendment above closed 800.4g/h and left 800.4f half-implemented: a
departed player's `pay_unless` was never reassigned — correct — and was
then discarded whole, which is not. The rule says the **cost is not
paid**; "unless that player pays {1}" is a sentence about what happens
when it is not, so the Study's controller draws and the Tithe's
controller makes a Treasure. This amendment runs that branch.

### The table gains a second column

`choiceReassignDecisions` is now `choiceDepartureDecisions`
(`leave_game.go`), one row per `PendingChoiceKind` and two columns:

| column | question | default |
|---|---|---|
| `reassign` | CR 800.4g — does somebody else make this choice? | no |
| `onDrop` | when it is dropped, does the rule that ends it say what happens instead? | `dropDiscard` — nothing |

Deny by default in both, for the reason the first column already had:
an unclassified kind behaves exactly as it did before #902, which is
safe and cannot wedge. `TestEveryChoiceKindHasAReassignmentDecision`
still fails until a new kind has a row.

**One kind declares an action today.** `pay_unless` is
`{onDrop: dropDecline}`: the drop runs the frame's decline continuation
(`declineDepartedChoiceLocked`, `pending_choice.go`) — the same
continuation `ResolvePayUnless` runs for an answered "no", reached from
the elimination sweep instead of from an answer. `entry_pay_life` is
CR 800.4f too and keeps the default, because its "unless" branch (the
permanent enters tapped) is about the departed player's own permanent,
which 800.4a takes in the same breath.

**The action is declared in the table, not in the sweep.** The
departure sweep (`dropChoicesForPlayerLocked`, `mutations.go`) is one
loop reading one table: it drops, and then runs whatever action the row
names. A future kind with a default action of its own adds a row, not a
branch.

### What makes running a continuation inside the sweep safe

Three things, and none of them is a new check:

1. **It runs after the queue is rewritten**, exactly like #808's
   replacement frames. A continuation may queue the next prompt, and it
   must not land in a slice the drop loop is still writing over.
2. **It cannot re-queue a prompt to the departed seat.**
   `QueueChoiceForEffect` refuses an eliminated chooser (#864) and emits
   `EventPendingChoiceDropped` instead. A prompt queued to a **survivor**
   is kept: that seat really does owe it.
3. **It only runs for a prompt whose object a survivor still controls**
   — `departedChoiceObjectLocked`, gates 2 and 3 of the reassignment
   predicate, shared rather than re-written. CR 800.4f's subject is "an
   OBJECT requires a player who has left the game to pay a cost", which
   is 800.4g's opening clause one rule earlier.

Gate 3 is what makes **cumulative upkeep** come out right. "Sacrifice
this permanent unless you pay its upkeep cost" is asked of the
permanent's own controller, so when they concede the permanent leaves
the game with them (800.4a) instead of dying. The difference is
observable: a sacrifice is a death, and the rest of the table's triggers
watch for one.

**Undo owes nothing new.** The prompt, its frame and the board are all
already cloned, so a rewind past the concede takes the drawn card back
and a replayed departure draws it again.

### `ConfirmPrompt.FromPlayer`, and the row it unlocks

The amendment above listed `confirm` as never reassigned, for a reason
that was about the engine rather than the rule: `QueueConfirmForEffect`
stamped `FromPlayer` with the chooser unconditionally, so gate 3 could
not tell a self-question from one asked across the table. That is now
the one field it named — `ConfirmPrompt.FromPlayer`, defaulting to
`Chooser`, carried the way `ChooseCardsPrompt.FromPlayer` already is —
and the row is `{reassign: true}`, `trigger_prompt`'s resolution-time
twin.

Nothing changes for any confirm the engine queues today: all of them
leave the field defaulted, so gate 3 still drops every one of them. The
row is a decision about the rule, kept so that the first card to ask a
cross-table yes/no (Combustible Gearhulk's question to its target) does
not land on a silent wrong default.

### Consequences

**Good**

- CR 800.4f is enforced in both halves: nobody inherits a departed
  player's tax, and the tax still pays out. A player conceding in
  response to a Rhystic trigger no longer eats the card.
- The sweep stayed one loop over one table. The second column is where
  the next kind's default action goes, and the elimination sweep does
  not have to learn about it.

**Tradeoffs**

- **A dropped `pay_unless` runs its decline branch, not its "pay"
  branch**, even for a may-pay frame whose whole payload hangs off the
  payment (`QueueMayPayForEffect`, Hashaton). Nothing happens there,
  which is the right reading of "if you do" for a player who no longer
  can — but it does mean the departure is silent for that shape rather
  than announced.
- **`pay_unless` is still the only kind with a drop action.** Every
  other CR 800.4f prompt the engine has is about the departed player's
  own material, so the honest row is the default; the column exists
  because the next one may not be.
