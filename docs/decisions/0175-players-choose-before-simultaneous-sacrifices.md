# ADR 0175 — Players choose before simultaneous sacrifices

Date: 2026-10-14
Status: Accepted
Sprint: S58 — Deck requests, October batch follow-up
Issues: #2948, #2910

## Context

Liliana, Dreadhorde General's −4 currently queues two sacrifice prompts per
player and moves each selected creature as its answer arrives. A Blood Artist
chosen by the first player observes its own death, but misses the later seven
selected deaths. CR 101.4 and 608.2e require choices in APNAP order followed by
one simultaneous action; 101.4b requires later players to know earlier public
choices. Queuing every choice at once also permits answers out of that order.
The −9 already gathers picks before one exit, but shares that choice-order gap.

Existing choose-cards prompts and the simultaneous exit primitive provide the
necessary operations. A new sacrifice event or a second exit system is not
needed. Existing forced-sacrifice protection must remain authoritative.

## Decision

1. Add `ChoosePermanentsInOrderThenForEffect` over ordinary
   `ChooseCardsPrompt` legs. Only the first live leg is queued. Each answer
   advances to the next leg, retaining immutable receipts; no permanent moves
   while players choose. Empty legs and departed players are skipped. A leg
   with no actual choice can be collected immediately. Leg-level `Then` is
   rejected: this API collects choices, and its one final continuation owns
   the action. The caller supplies APNAP leg order.
2. Each receipt keeps the chosen permanent's object epoch. The final
   `PromptedPicks` excludes an object that left and returned before the action.
   The continuation receives choices in leg order and runs once. A dropped
   leg continues with no picks while its source survives, so an affected
   player's concession or a stale choice cannot strand the resolution. A
   departed caster cancels its own resolution through existing departure rules. Closures capture only fresh immutable values and use the
   live Game supplied to them, preserving clone/undo behavior.
3. Carry earlier public choices in the existing private choose-cards frame.
   A computed protocol field `earlier_choices` contains player/card IDs;
   it adds no persisted choice or event field. The human picker and bot/MCP
   text show those earlier choices. Names come from the viewer's filtered
   board, so a face-down permanent's hidden identity is never disclosed.
4. Expose the existing effect-time forced-sacrifice gate and candidate walk
   through thin locked-context queries. Use those same checks for ordinary
   sacrifice prompts and Liliana's collected choices; costs retain their
   existing independent rules.
5. Liliana's −4 asks each eligible player for exactly two legal creatures,
   or all available creatures when fewer exist, then calls
   `SacrificeAllThenForEffect` once with the collected set. The −9 uses the
   same ordered choice collector and protection gate before its existing
   one-of-each-type validation and simultaneous exit. Its retention rule
   remains unchanged.

## Compatibility

The public choice kind and answer payload remain `choose_cards`. Earlier
choices are transient frame data, projected for a decision, and the frame
already prevents durable restoration until resolution completes. Clone/undo
retains independent receipt slices and drop continuations. Snapshot schema v8
and frozen fixtures stay unchanged. The existing general choose-cards fan-out
API remains available for callers that do not require this ordered collection.

## Validation

Reproduce the existing early sacrifice through a real loyalty activation.
Then test no movement before the final choice, all eight Blood Artist
triggers, exact counts with zero/one/two/more candidates, APNAP ordering and
public earlier-choice visibility, protected players, tokens, concessions,
object changes, undo/replay, and paused sacrifice replacements. Preserve the
−9 multitype retention, dies-draw and Zombie ability tests. Verify the client,
legal moves and bot/MCP representation of the ordinary choice kind.

## Number reservation

After `git fetch --all --prune`, checked all 13 remote refs on 2026-10-14:
origin, origin/cost-ledger, origin/develop, origin/docs/issue-audit,
origin/feat/3429-movable-stack, origin/feat/750-conditional-block-restrictions,
origin/feat/playmats, origin/fix/3159-3160-attachments,
origin/fix/3434-stats-privacy, origin/fix/caddy-reload-admin-off,
origin/main, origin/results-feed, origin/wip/836-one-click-default.
0171–0174 were occupied; 0175 was the first unused number.

## Amendment 2026-10-26 — the collector is plain data, so a table waiting on a seat's choice is a restore point (#4290)

The collector above asked one seat at a time with every leg's continuation a closure over the legs still to come, the picks so far, the objects' epochs and the card's `then`. A table waiting on a seat was therefore not a restore point (`ContinuationCensus.ChoiceResumeFrames` counted the leg's frame). Every one of those closures is data, and the prompt already carries what the engine needs to ask the next seat.

**Decision.** `game/ordered_picks_keyed.go` is the collector with its state on the prompt (`PendingChoice.ordered`): the legs (a seat, its candidates as `{id, epoch}` refs taken when the collection began, and the bounds), the picks so far, the picks the chooser is shown (CR 101.4b), and the registered continuation (`RegisterPromptRunThen`) with its `RunData`. A seat's turn is a keyed `choose_cards` pick under `choose-cards/ordered-pick-leg`, or under a key registered with the leg's own set rule (`RegisterOrderedPickLegRule`, for Liliana, Dreadhorde General's −9). The leg's continuation files the answer and asks the next seat; after the last the picks go to the continuation, or, for a sacrifice, to one `SacrificeAllThenForEffect` first, so nothing moves while seats are choosing and the continuation reads what really left (`RunDone.Sacrifices()`). The seats, the "forced valid set needs no prompt" shortcut, the departed-seat skip and the epoch binding are the closure form's, unchanged: a chooser who leaves with the prompt open chooses nothing (`defaultDroppedChoiceLocked` runs the keyed pick with nobody's choice) and the collection goes on.

**Callers.** `EachPlayerSacrificesThenKeyedForEffect` and `PlayersSacrificeThenKeyedForEffect` (Rise of the Witch-king; `effects.EachPlayerSacrifices.Keyed` for Extus, Oriq Overlord, Fatal Grudge, Garruk, Veiled Butcher's −2, Momentum Breaker, Syphon Flesh, Throne of the Grim Captain and Witch-king of Angmar; `Context.PlayersSacrificeThen` for Cornered by Black Mages, Will of the Abzan, Shadowgrange Archfiend, Priest of Forgotten Gods, Reactor Raid, Yuna's Decision and Capital Punishment), the fire-and-forget `EachPlayerSacrificesForEffect` and the new `PlayersSacrificeForEffect` (Ember Swallower, Papalymo Totolymo, Portal to Phyrexia), and `ChoosePermanentsInOrderSacrificeForEffect` / `ChoosePermanentsInOrderKeyedForEffect` for Liliana's −4 and −9. The −9's rule depends on the permanent types the opponent's board has, so it is registered once for each of the 63 non-empty sets of the six types (`choose-cards/liliana-dreadhorde-general-keep-<types>`), and a leg names the one for its board. `ChoosePermanentsInOrderThenForEffect` stays for the closure form of the per-object any-player offer, whose keyed form (#4264) already carries its picks on the offer.

**Compatibility.** Additive within snapshot v9 (`pendingChoices[].ordered`), see ADR 0041's amendment of the same date. No wire change: the leg is the same `choose_cards` prompt, and `EarlierPermanentChoicesForEffect` reads the collection's picks.

**Tests.** `game/keyed_collections_test.go`: a four-seat collection across a strict restore with one seat answered (nothing sacrificed until the last, the second seat sees the first pick, the continuation handed one entry per seat asked with its carry and number); an unknown collection or leg key refused; picks bound to the objects they were; a seat conceding with its turn open; the fire-and-forget form; a leg's registered set rule enforced after a restore; a leg with a closure, a foreign leg key or no continuation refused. `effects/keyed_collections_restore_test.go`: Syphon Flesh and Rise of the Witch-king across a strict restore, and Syphon Flesh's unknown-key refusal.

**Back-outs**, each failing the named tests:

- Not writing `ordered` on a prompt fails `TestAKeyedEdictCollectionIsARestorePointWithASeatAnswered`.
- Not checking the collection's continuation key at restore fails `TestAKeyedEdictCollectionWithAnUnknownKeyIsRefused`.
- Dropping the epoch check from `livePickRefs` fails `TestAKeyedEdictCollectionBindsPicksToTheObjectsTheyWere`.
