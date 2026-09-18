# ADR 0061 — Token creation and discard are replaceable events (CR 701.7b, CR 701.8, CR 614)

**Status:** Accepted · 2026-09-18 · S39 — The CR 614 replacement surface and its paused continuations
**Issues:** [#762](https://github.com/krakenhavoc/cmd_and_ctrl/issues/762) (token creation),
[#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650) (discard with a cause)
**Numbering:** on 2026-09-18, after `git fetch origin`, every remote branch head was
listed with `git ls-tree -r --name-only <sha> docs/decisions/`, and every commit
reachable from any ref was swept with `git log --all --name-only -- docs/decisions/`.
The highest number present anywhere is 0060 (`docs/decisions/0060-leaving-the-game.md`,
merged to `develop`). Nothing holds a `0061-*` file. 0005, 0024, 0029 and 0030 stay
permanently unused per AGENTS.md §4.
**Related:** [ADR 0013](0013-replacement-effects.md) (the CR 614/616 engine these two
events join, and §10a, which this ADR closes),
[ADR 0018](0018-triggers-on-the-stack.md) §6 (what a post-event listener sees, and why
a replacement is not one),
[ADR 0010](0010-card-effect-catalog.md) (the `Spec` slots these cards declare through)

## Context

Two of the engine's mutations still happened *behind* the CR 614 window, and both
were named on the seam registry.

**Token creation.** `CreateTokenForEffect`, `CreateTokensForEffect` and
`CreateTokensAttackingForEffect` each minted their tokens and pushed them onto the
battlefield with no replacement window at either end. Nothing could double a
creation and nothing could change what kind of token it made, so Doubling Season and
Primal Vigor shipped with only their counter halves and Parallel Lives, Anointed
Procession, Academy Manufactor and Mondrak could not be written at all. The entry
half was missing too: a created token never ran the battlefield-entry pipeline, so
no enters-tapped effect (Urabrask the Hidden, Manglehorn, Kinjalli's Sunwing,
Kismet, Thalia), no enters-with-counters effect (Renata, Arwen, Dragonstorm Globe)
and no `fireETBHookLocked` ever saw one. Nine catalog cards carried a caveat about
it. The `TokenEntryOptions.Counters` comment stated the gap in so many words.

**Discard.** #853 put every discard through the shared exit primitive, so the
CR 614 window does open on a discard today and a discarded commander is offered the
command zone (CR 903.9). What the window sees, though, is an ordinary
`RepEventMove` hand → graveyard with nothing on it to say it is a discard or why it
is happening. Library of Leng ("if an **effect** causes you to discard a card"),
madness (CR 702.35a, [#657](https://github.com/krakenhavoc/cmd_and_ctrl/issues/657))
and the Obstinate Baloth family ("a spell or ability an **opponent** controls causes
you to discard") all key on the discard itself, and two of the three also key on the
cause. ADR 0013 §10a established what the cause has to be — effect, cost, or
turn-based action, not "voluntary" — and left the event shape to this ADR.

Both are the same shape of gap, which is why they are one ADR: a mutation that a
card wants to replace, reaching the board without an event that names it.

## Decisions

### 1. `RepEventCreateTokens` is opened once per creation INSTRUCTION

CR 701.7b: "To create one or more tokens … is to put the specified numbers and
kinds of those tokens onto the battlefield." One instruction is one event, so a
doubler modifies the instruction rather than each token — "create two Treasures"
under Parallel Lives is **one** event that becomes four Treasures, not two events of
two. That distinction is what makes a second doubler multiplicative rather than
additive.

The event carries:

```go
RepEventCreateTokens:
    TokenController uuid.UUID     // "under your control"
    TokenGroups     []TokenGroup  // {Template, Count, Entry} per KIND
    TokenAttacking  uuid.UUID     // CR 506.3c, uuid.Nil for the ordinary case
    Source          uuid.UUID     // the card whose effect is creating them
```

**Groups, not a bare count**, and that is a decision rather than a detail. Parallel
Lives, Anointed Procession, Doubling Season, Primal Vigor and Mondrak change *how
many*; Academy Manufactor changes *which kind* — "if you would create a Clue, Food,
or Treasure token, instead create one of each" turns one group into three, keeping
the count. A count alone can express the first family and not the second, and the
audit that filed #762 said so.

Card authors never touch the slice. The event exposes `TokenCount()`,
`MultiplyTokens(n)`, `ReplaceTokenKinds(templates…)`,
`ReplaceTokenKindsWhere(pred, templates…)` and `TokenTemplatesMatch(pred)`, and the
`effects` package wraps the first of those in `TokensDoubled(label)` and
`AnyPlayersTokensDoubled(label)` so a doubler's card file is its printed sentence
and nothing else.

**Ordering follows CR 616.1 and the engine's existing skips.** Two Anointed
Processions are two objects contributing ONE declared effect, so
[#792](https://github.com/krakenhavoc/cmd_and_ctrl/issues/792)'s identical-window
skip applies and nobody is asked to order them — ×2 then ×2 is ×4 either way. A
Doubling Season beside an Academy Manufactor is two *different* declared effects and
the creation's controller really is asked, because the engine cannot know the
orderings agree and CR 616 gives them the choice.

### 2. Every created token then takes the ordinary battlefield-entry pipeline

After the creation window settles, each token the instruction produced enters
through `enterBattlefieldThroughPipelineLocked` — the one effect-side entry
primitive [#478](https://github.com/krakenhavoc/cmd_and_ctrl/issues/478) built for
the library search, the exile return and the reanimation, landed in PR #919 while
this work was in flight. **A created token is a fourth caller of it and nothing
more.** There is no token entry path: the same `executeEntryToBattlefieldLocked`
finishes it, on the inline path and on the resumed one, so a token gets
`EntersTapped`, `EntersWithCounters` (through the CR 614 counter pipeline, so a
Doubling Season doubles them and All Will Be One sees them placed), CR 614.12
self-replacement, `applyEntersAsCopyLocked`, the ETB event and
`fireETBHookLocked` — the same list, in the same order, as everything else that
enters.

**`OldZone` is empty**, because a token comes from no zone at all (CR 111.1 — it is
created on the battlefield). Every entry replacement in the catalog keys on
`NewZone`, so none of them notices, and the empty old zone is what the shared
finisher reads twice: once to push the staged token instead of moving a card out of
a zone, and once to decide that the arrival announces `EventTokenCreated` rather
than `EventZoneMove`. One event per arrival, never two: a token that emitted both
would be counted twice by everything that watches permanents arrive.

**The token is staged in `Game.enteringTokens` while its window is open.** A card
entering the battlefield sits in the zone it is leaving while its entry
replacements are consulted; a token has no such zone, and an entry replacement
still has to be able to read it — Urabrask the Hidden asks whether the entering
permanent is a creature an opponent controls, and it asks through
`LookupCardForEffect`. So that function looks in the staging slice when no zone
holds the card. A token whose entry is cancelled or redirected is dropped from the
slice and ceases to exist (CR 111.8: a token put anywhere but the battlefield
ceases to exist).

### 3. Both halves can pause, and the batch is a value carried forward

A creation whose window queues a CR 616 ordering prompt returns with **nothing
created**; the resume mints the tokens when the prompt is answered
(`applyResolvedReplacementEventLocked` gains a `RepEventCreateTokens` case, which
calls exactly the function the unpaused path calls). An individual token's ENTRY can
pause the same way — two distinct enters-tapped effects on one opponent's token —
and it resumes through **#478's entry frame**, the one every other effect-side entry
already uses. Token entries are flagged `entryResumable`, and nothing is skipped by
resuming one: there is no shuffle owed and no new object identity to mint.

The tokens *behind* a paused one are that entry's continuation rather than the next
line of a loop, so they land on the far side of the prompt — and that continuation
is #478's `entryTail`, not a token-shaped copy of it. `tokenTail` exists only for
the CREATION event in Decision 1, which has no entry to hang a tail on yet; it is
what `CreateTokensThenForEffect(spec, then)` hands over, and `then` receives the IDs
of the tokens that actually landed, running from the landing on every terminal
outcome — created, cancelled, or abandoned because the prompt was taken away.

One necessary widening in #478's code: a paused entry is pruned when its card has
left the zone the window opened over (`pausedZoneChangeStaleLocked`), and a staged
token is in no zone at all. A token whose entry is paused is not stale — it is
exactly where the paused entry left it.

`CreateTokensForEffect` keeps its `([]uuid.UUID, error)` signature for the ~200
callers that create a token and stop there, and its returned slice is **empty when
the creation paused**. A caller whose sentence continues past the tokens must use
the `Then` form. That is the contract `MillToZoneThenForEffect` and
`discardCardsLocked` already carry, and it is stated on the function.

### 4. `RepEventDiscard` replaces the plain move the discard route opens

`zoneRoute.Discard` already told the exit primitive that a move was a discard.
It now also decides the event kind: a discard opens `RepEventDiscard` rather than
`RepEventMove`, carrying

```go
RepEventDiscard:
    CardID        uuid.UUID     // the discarded card
    DiscardPlayer uuid.UUID     // its owner, who is discarding it
    DiscardCause  DiscardCause  // "effect" | "cost" | "cleanup"
    Source        uuid.UUID     // the card whose effect or cost asked
    OldZone / NewZone / NewZoneOwner, zoneRoute, mustSettleNow  // the move payload
```

It is a distinct kind rather than a flag because a discard replacement watches for
the **discard**, not for the zone move: `Watches: []EventKind{EventDiscardCard}`
is what Library of Leng, madness and "exile it instead" declare, and
`eventKindMatches` maps the new kind to `EventDiscardCard`. It still carries the
whole move payload, because a discard IS a move out of the hand and its
replacements rewrite the destination, so everything in the engine that finishes,
abandons or prunes a routed exit handles the two kinds together through one
predicate (`isExitMove`).

**The CR 903.9 built-in watches both.** A discarded commander is still offered the
command zone: `commanderZoneReplacement` now declares `EventZoneMove` and
`EventDiscardCard` and accepts either kind. Nothing else in the catalog keys on a
hand → graveyard `RepEventMove`, which was checked card by card before the switch —
every catalog `RepEventMove` predicate is either an entry (`NewZone ==
ZoneBattlefield`) or a battlefield exit (`OldZone == ZoneBattlefield`).

**`EventDiscardCard` still fires whatever the window did with the destination**
(CR 701.8a: the discard is the move out of the hand). Library of Leng putting the
card on top of the library, madness exiling it and CR 903.9 sending a commander to
the command zone are all still discards, and Megrim, Containment Construct and the
rest of the family still see them. What the event now carries in addition is
`Source` and `DiscardCause`, so a log line can say who caused it and a future payoff
can read it.

**A cost discard still settles now.** `discardCauseCost` keeps setting
`zoneRoute.MustSettleNow` (CR 601.2h / 602.2b: costs are paid as one indivisible
step), which means the window still runs and a mandatory discard replacement still
applies, but anything that would ask a question is skipped un-applied. A commander
pitched to a cost goes to the graveyard, and Library of Leng does not offer its
"may" — which is also the right answer for the other reason: costs are not effects
(Gatherer ruling, 2004-10-04), so Leng's `AppliesTo` refuses a cost discard anyway.
The two guards agree, and the card-side one is the one that expresses the rule.

### 5. The cause taxonomy is the one ADR 0013 §10a named

`DiscardCause` is a string enum — `"effect"`, `"cost"`, `"cleanup"` — so it reads
in the event log and on the wire without a translation table:

| value | what it means | CR |
|---|---|---|
| `DiscardCauseEffect` | an effect instructed it: Mind Rot, looting, a random discard, a revealed-hand pick | 701.8a, 608.2c |
| `DiscardCauseCost` | a cost paid it: an additional cost to cast, an ability's cost, and the CR 118.12 "unless you discard" branch of a resolving spell | 601.2h, 602.2b, 118.12 |
| `DiscardCauseCleanup` | the hand-size discard, a turn-based action nobody's effect caused | 514.1, 703.1 |

Library of Leng reads `Effect`. Madness will read nothing (CR 702.35a replaces every
discard, "independently of why you're discarding"). The Obstinate Baloth family will
read `Effect` plus the controller of `Source`, which is on the event and needs no
further field.

**There is no `IsVoluntary`, and there will not be.** ADR 0013 §10a withdrew that
framing; this is the shape that replaces it.

### 6. What the two events carry afterwards, and `Event.Source`

`EventTokenCreated` now carries `Source` — the card whose effect created the token —
and is emitted from the entry body rather than from the creation loop, so it is
emitted once per token, after the token is on the battlefield and before its
`EventETB`, which is the order it had before. `EventDiscardCard` gains `Source`
(already added by #853) and `DiscardCause`. Both keep every field they had.

This is the first `ReplacementEvent.Source` that a creation sets, and it is
threaded from the `CreateToken` primitive's `ctx.Source()` for the catalog's
creations; a creation with no source card (a test fixture, an admin verb) leaves it
`uuid.Nil`.

### 7. Scope: what this ADR does NOT decide

- **Madness** ([#657](https://github.com/krakenhavoc/cmd_and_ctrl/issues/657)) is
  not implemented here. What it needs from #650 is the event, and the event is here.
- **Search-to-graveyard and surveil** still bypass the CR 614 window. Rest in Peace
  needs those two as well as the discard, so it stays on
  [#383](https://github.com/krakenhavoc/cmd_and_ctrl/issues/383)'s skip list. #650
  asked whether they join this work; the answer is no — they are a graveyard-arrival
  seam, not a discard one, and folding them in would make one PR out of two.
  *(Closed 2026-09-18 by [#931](https://github.com/krakenhavoc/cmd_and_ctrl/issues/931),
  which routed both and shipped Rest in Peace and Leyline of the Void off that skip
  list — see [ADR 0013 §5q](0013-replacement-effects.md).)*
- **A "discard these card IDs" primitive** (Kroxa, the Obstinate Baloth redirect)
  is not added. Nothing in the catalog needs it yet.
- **Choices inside a token-creation replacement** — Jinnie Fay's two-way pick,
  Esix's creature pick — are out. The event supports them (an `Optional`
  replacement on a creation pauses like any other), but the prompt shapes they
  need are their own work.
- **CR 111.13**: a copy of a permanent spell becoming a token
  ([#666](https://github.com/krakenhavoc/cmd_and_ctrl/issues/666)) is not
  "created" and must not reach `RepEventCreateTokens`. Nothing routes it there
  today, because that path does not exist yet; the note is here so the one who
  writes it does not reach for `CreateTokensThenForEffect`.
- **No new prompt kind.** Both events reuse the CR 616 ordering prompt and the
  CR 614.10 optional prompt, so the bots and the client need nothing new.

## Consequences

### Good

- Token doubling, token-kind rewriting and the whole entry pipeline for tokens are
  one mechanism each, opened in one place each. Nine catalog caveats go away and
  five cards become writable.
- Library of Leng, madness and the Obstinate Baloth family have an event to key on,
  with the cause the rules actually distinguish.
- A created token is now indistinguishable from any other permanent from the entry
  pipeline's point of view, so the next entry replacement anybody writes covers
  tokens for free. Thalia, Heretic Cathar's card file already claimed a token
  entered tapped; the claim is now true.

### Tradeoffs

- `CreateTokensForEffect`'s returned IDs are empty on a paused creation. Every
  in-catalog caller that reads them creates tokens and stops, so none is affected
  today; the failure mode for a future one is "the rest of the sentence does not
  happen", not a wrong board, and the `Then` form is the fix.
- `Game.enteringTokens` is one more piece of transient state. It is dropped by the
  persisted snapshot and copied by `Clone`, on exactly the reasoning
  `replacementsAppliedThisEvent` already carries: it is non-empty between actions
  only while a token entry is paused on a prompt, and that prompt's resume frame is
  already counted by `ContinuationCensus`.
- A token's entry now runs `fireETBHookLocked`, which it never did. For a plain
  token that is a no-op (no oracle ID), and for a token COPY it means the copied
  card's `Spec.AsEnters` fires — closing a gap `token_copy.go`, Saw in Half and
  Hashaton, Scarab's Fist all documented, and a behaviour change for anybody who had
  learned to expect the old one. A token copy of Adaptive Automaton now names a
  creature type as it enters.
- Sequencing. The tokens of one instruction enter one at a time, so a paused one
  puts the rest on the far side of a prompt. They were already sequential; what is
  new is that the sequence can now be interrupted. CR 701.7b creates them
  simultaneously, and holding a whole simultaneous creation open across a prompt
  needs the continuation frame [#478](https://github.com/krakenhavoc/cmd_and_ctrl/issues/478)
  is about — the same call `zone_route.go` makes for a simultaneous exit.
