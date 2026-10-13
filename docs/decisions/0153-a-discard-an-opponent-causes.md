# ADR 0153 — A discard an opponent causes: who controls the cause, and a discard that enters the battlefield (CR 614.1a, CR 701.9a)

**Status:** Proposed · 2026-10-13 · S53 — Faces, zones, and shared payment (tracker [#1910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1910))
**Issues:** [#1861](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1861) (seam: "if an opponent causes you to discard this card", Nullhide Ferox; found landing ADR 0106 PR 6, #1793, tracker #1784). Registry row: `discarded-by-an-opponent`.
**Owner decisions:** needed. See [Questions for the owner](#questions-for-the-owner). Nothing in the engine changes until they are answered.
**Numbering:** assigned **0153** for S53. After `git fetch --all --prune` I read the `docs/decisions/` file names on every remote head (`origin/develop`, `origin/main` and the seven other branches). The highest number anywhere is 0149, and 0150–0155 are being written for S53 in parallel.
**Builds on:** [ADR 0013](0013-replacement-effects.md) (the CR 614 pipeline; §5g, a discard counted wherever the card lands; §10a, the three discard causes), [ADR 0061](0061-token-creation-and-discard-are-replaceable-events.md) (`RepEventDiscard`, #650), #1320 (`MoveCause` on zone-change events), #478 / #1322 (the resumable battlefield entry and the entry batch), [ADR 0098](0098-discard-as-a-permanent-would-enter.md) Decision 4 (an entry the window sends elsewhere), [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) owner decision 6 (a PR lands every card its seam unblocks), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic prices declared card facts).

---

## Context

### The rules

- **CR 701.9a:** "To discard a card, move it from its owner's hand to that player's graveyard." The engine reads a discard by its source zone (ADR 0013 §5g, `discardedThisWayLocked`), so a card that a replacement sends somewhere else was still discarded.
- **CR 614.1a:** an effect that uses "instead" is a replacement effect. **CR 614.6:** the modified event happens in place of the original, and it may trigger abilities.
- **CR 614.5 / 616.1:** a replacement applies to an event at most once, and when several apply, the affected player (here, the discarding player) chooses the order. Madness (CR 702.35a), CR 903.9 and Rest in Peace already compete for a discard this way.
- **CR 614.12:** a replacement effect that changes how a permanent enters the battlefield is checked against the permanent as it would exist there. That applies to a card a discard replacement puts onto the battlefield, just as it applies to any other entry.
- **CR 113.8 / 602.2 / 603.3a:** the controller of a spell or ability on the stack is the player who cast, activated, or controlled its source when it triggered. That player need not control the card it came from (an ability of a stolen permanent, a copy), so "a spell or ability an opponent controls" is a question about the **stack item**, not about the card that asked.
- **CR 118.12:** "do X unless you discard" makes the discard a cost paid while a spell or ability resolves. The engine is not consistent about this today. The generic pay-unless payment (`payActionLocked`, `pay_unless_action.go`: echo and cumulative upkeep on your own permanents) files it under `DiscardCauseCost`. Painful Quandary, the one opponent-forced "unless you discard" in the catalog, queues an ordinary discard prompt, so it is `DiscardCauseEffect`. See question 4.
- **CR 514.1:** the cleanup step's hand-size discard is a turn-based action. Nobody's spell or ability causes it.
- **CR 110.2:** a permanent enters under the control of the player who put it onto the battlefield. A discard replacement is controlled by the card's owner (CR 616.1, the affected player), and the card is in its owner's hand, so it enters under its owner's control.

The rulings on the family (Gatherer, Obstinate Baloth, Loxodon Smiter, Wilt-Leaf Liege) say three things the design has to keep: a discard you cause yourself, a discard paid as a cost, and the cleanup discard all leave the card in the graveyard. The card was still discarded, so "whenever a player discards a card" abilities trigger. Madness and these replacements are chosen between by the discarding player.

### What the engine has today (origin/develop at `d66aec828`)

- **One discard path.** Every discard goes through `discardCardsLocked` → `discardBatchLocked` (`game/discard.go`) → `routeCardToZoneLocked` (`game/zone_route.go`) with `zoneRoute.Discard`, `DiscardCause` (effect, cost or cleanup) and `Source`.
- **The route knows who caused it.** Before the window opens, `routeCardToZoneLocked` fills `zoneRoute.Cause` from `resolutionCauseLocked()` (`move_cause.go`, #1320). For an effect, that is `MoveCause{Kind: MoveCauseEffect, Controller: <resolving item's controller>, Item: <item ID>}`. A cost names the payer, and the cleanup discard is `MoveCauseRule`. The completed move's `EventDiscardCard` carries this as `Event.Cause` / `CauseController` / `CauseItem` (`stampCause`). **So the TRIGGERED half of the family already has its data.**
- **The replacement event does not.** The `RepEventDiscard` that `routeCardToZoneLocked` builds copies `DiscardPlayer` and `DiscardCause` and nothing else about the cause. `ReplacementEvent.Source` is **not set** on a discard either. The comment on `ReplacementEvent.DiscardCause` and the registry row's engine notes both say it is, which is wrong. `zoneRoute` is unexported, so a catalog `AppliesTo` in `cards/effects` cannot reach it. Today a replacement cannot tell Mind Rot from your own Careful Study.
- **A discard cannot land on the battlefield.** `DiscardBecomes` (`effects/discard_replacements.go`) rewrites `ev.NewZone`, and `executeZoneRouteLocked` moves the card to `routeDestinationLocked(ev.NewZone)`. That handles graveyard, hand, library and command, and returns `ErrZoneNotFound` for `ZoneBattlefield` (pinned for the stack exit by `resolution_self_move_test.go`). Even if it resolved a zone, a bare `MoveCard` onto the battlefield would skip everything a battlefield entry owes: the entry's own CR 614 window (enters tapped, enters with counters, an "exile it instead"), the controller stamp, summoning sickness, `EventETB`, the AsEnters hook and the entry ordinal.
- **The entry pipeline exists, and it can pause.** `PutFromHandOntoBattlefieldThenForEffect` → `startEntryBatchLocked` (`battlefield_put.go`, `entry_batch.go`) runs the entry window, can queue a prompt and resume, and runs a continuation exactly once at a terminal outcome. ADR 0098 Decision 4 (`moveRedirectedEntryLocked`) is the reverse case: an entry the window sends elsewhere is turned into a routed exit.
- **Both finishers meet in one place.** An inline discard and a discard resumed after a CR 616 ordering prompt (`applyResolvedReplacementEventLocked`, the `RepEventMove, RepEventDiscard` arm) both settle in `executeZoneRouteLocked`. That is the one place a battlefield destination has to be handled.

### The cards

The Scryfall dump the repo uses (`data/scryfall/default-cards.json`) has **17** cards with "a spell or ability an opponent controls causes you to discard". **16** are Commander-legal. Gorilla Tactics is not, and is left out. None is in the catalog. Only Nullhide Ferox is on the registry row's `Waiting` list.

| Shape | Cards | What they need |
|---|---|---|
| Replacement: put it onto the battlefield instead (5) | Obstinate Baloth, Loxodon Smiter, Nullhide Ferox, Wilt-Leaf Liege, Dodecapod (with two +1/+1 counters) | §1 (the cause controller on the event) and §2 (the battlefield handoff) |
| Replacement: put it on top of your library instead (1) | Nephalia Academy (a land; "may reveal", any card you discard) | §1 only. `DiscardBecomes{Dst: ZoneLibrary, Optional: true}` already does the move |
| Trigger on the discard (10) | Ajani's Last Stand, Guerrilla Tactics, Mangara's Blessing, Metrognome, Orvar, the All-Form, Psychic Purge, Quagnoth, Sand Golem, Spiritual Focus, Pure Intentions | A shared predicate over data `EventDiscardCard` already carries (§4) |

So the issue is right about the battlefield redirect. It is wrong that the cause controller is missing everywhere: it is on the event, just not on the replacement event. It also undercounts the family. The seam closes for 6 cards, and the predicate in §4 lets the other 10 land in the same sprint.

## Decisions (proposed)

### 1. The replacement event carries the cause

`routeCardToZoneLocked` already has `r.Cause` resolved when it builds the `RepEventDiscard`. It copies two more facts onto the event:

- `ReplacementEvent.Source = r.Source`. This fixes the documented-but-missing field for every discard.
- A new `ReplacementEvent.DiscardCauseController uuid.UUID`: `r.Cause.Controller` when `r.Cause.Kind == MoveCauseEffect`, and `uuid.Nil` otherwise (a cost, the cleanup step, a manual move).

The test is the **stack item's** controller (CR 113.8), never `Card.Controller` of `Source`, because the two can differ (an activated ability of a stolen permanent, a copy of a spell). `uuid.Nil` for a cost is deliberate. Every cost discard in the engine today is a cost the discarding player chose to pay for their own spell or ability (an additional cost, an activation, echo, cumulative upkeep), so none of them is caused by an opponent. Question 4 covers the one opponent-forced "unless you discard" in the catalog.

Card side, `DiscardBecomes` gains `ByOpponent bool`. It applies only when `ev.DiscardCauseController` is a player other than `ev.DiscardPlayer`. Any other player at a Commander table is an opponent (CR 102.3). A teammate rule would go here if team play arrives. A test helper and an exported predicate, `game.DiscardCausedByOpponent(ev *ReplacementEvent) bool`, keep the reading in one place.

The new field rides on the event the way `DiscardCause` does: plain data, copied by the undo clone with the rest of the event. A paused replacement prompt (`replacementResume`) is never written to a restore point, so there is **no snapshot schema change**.

### 2. A discard that settles on the battlefield enters through the entry pipeline

`DiscardBecomes` gets a battlefield destination: `Dst: game.ZoneBattlefield`, with an optional `EntersWithCounters map[string]int` for Dodecapod. Its `AppliesTo` additionally requires `ev.NewZone == ZoneGraveyard`. The printed text is "instead of putting it into your graveyard", and CR 614.5 means that once madness or Rest in Peace has sent the card elsewhere, these no longer apply. Madness reads the same way (`MadnessReplacement`).

The engine side is one branch at the top of `executeZoneRouteLocked`, before its deferred route tail is registered:

```go
if ev.Kind == RepEventDiscard && ev.NewZone == ZoneBattlefield {
    return g.discardOntoBattlefieldLocked(ev)
}
```

`discardOntoBattlefieldLocked` (new, `game/discard_entry.go`):

1. Re-checks that the card is still in `ev.DiscardPlayer`'s hand. If not, it runs the route tail and returns, the same as a stale exit.
2. Calls `PutFromHandOntoBattlefieldThenForEffect(card, HandEntryOptions{Controller: ev.DiscardPlayer}, then)`, seeding the entry event's `EntersWithCounters` from what the discard replacement asked for. This opens the **entry's own** CR 614 window (decision 3).
3. In `then`, which runs once at the entry's terminal outcome (entered, cancelled, redirected by the entry window, or a dropped prompt): it emits `EventDiscardCard` with `NewZone` set to where the card really is and the route's `Cause` stamped, then runs `runRouteTailLocked(ev.zoneRoute)`. That is the rest of the discard batch, and through it the discard effect's own "then".

Why this shape:

- **One entry primitive.** The card enters exactly as a Sneak Attack'd card does, with the controller stamp, the CR 614.12 window, summoning sickness, `EventETB`, the AsEnters hook and the entry ordinal. No second copy of the entry code.
- **It can pause safely.** If the entry window asks a question, `discardOntoBattlefieldLocked` returns with the card still in hand and the discard batch waiting. The batch continues from the entry's `then`, which is the contract a paused CR 903.9 exit already gives it (#853, #1027).
- **"Discarded this way" stays right.** `discardedThisWayLocked` reads "not in the hand any more", so Syphon Mind's count, the discard prompt's "then", and every payoff count a Baloth that went to the battlefield, as the rulings require. A card the entry window cancelled outright is still in hand and is correctly not counted.
- **Inline and resumed paths cannot drift.** Both reach `executeZoneRouteLocked`, so the branch covers a Baloth whose owner was first asked to order it against madness.
- **A redirect inside the entry window is ADR 0098's case.** If the entry window sends the card elsewhere (an "if a nontoken creature would enter and it wasn't cast, exile it instead" effect), `moveRedirectedEntryLocked` moves it as a plain exit. The card has left the hand, so the `then` above still emits `EventDiscardCard` with `NewZone: ZoneExile`. The discard happened. Only where the card ended up changed.

`EventDiscardCard` is emitted **after** `EventETB` in the log. Both are one event under CR 614.6, and the triggers they raise are put on the stack together at the next opportunity (CR 603.3), so log order changes nothing a card reads. Emitting after the terminal outcome is what lets `NewZone` be true. Question 3 is the alternative.

### 3. The entry gets a second, chained CR 614 window

The discard window and the entry window are two windows in sequence, not one. Decision 2 does it this way because:

- The replacements that watch the two are disjoint. Discard replacements watch `RepEventDiscard`. Enters-tapped, enters-with-counters and entry-exile replacements watch a battlefield `RepEventMove`. CR 903.9 only offers the command zone for graveyard and exile destinations (and for hand and library under its other clauses), not for the battlefield. No replacement can apply twice, so CR 614.5 holds.
- CR 614.12 says to check entry replacements against the permanent as it will exist on the battlefield. The entry window does that check, and the discard window cannot.
- ADR 0098 Decision 4 decided the opposite for entry → exit (no second window). It could, because that exit's replacements had already seen the event as a move to the graveyard. Here, nothing in the discard window has seen a battlefield entry.

### 4. The triggered half: one predicate, no new seam

`game.DiscardCausedByOpponentOf(ev Event, player uuid.UUID) bool` goes in `move_cause.go`, next to its twin `ExiledBySpellOrAbilityOf`. It is true when `ev.Kind == EventDiscardCard && ev.Actor == player && ev.Cause == MoveCauseEffect && ev.CauseController != uuid.Nil && ev.CauseController != player`. Card side: `effects.DiscardedByAnOpponent(self bool)`.

"When … causes you to discard **this card**" is a trigger on the card's own discard. It is written like cycling's "when you cycle this card": `TriggeredAbility.Zones = []ZoneKind{ZoneGraveyard}`, matching `ev.CardID == source.InstanceID`. A card that a replacement sent to exile instead does not trigger from exile. That is weaker than printed in a rare corner, the same posture cycling takes, and the card's `Caveats` says so. Spiritual Focus and Pure Intentions are battlefield and this-turn readers of the same predicate. Psychic Purge's "that player" is `CauseController`.

### 5. The bot

Two facts go on the wire, declared by the catalog constructor and never inferred from a card name (ADR 0126):

- `CardView.discard_by_opponent` on a hand card: `"battlefield"` (the five creatures), `"library"` (any card under Nephalia Academy, set on the land's controller's hand cards) or `"trigger"` (the ten). The value comes from the spec's declaration, not from parsing text.
- The discard prompt already carries `PendingChoiceView.Source`. The heuristic reads that source's controller and stack item from the view it already has to tell "an opponent is making me discard" apart from its own loot.

`handKeepValue` uses them only when an opponent caused the prompt. A `"battlefield"` card's discard price is `-(cardValue)`: the bot wants to pitch it, because it gets the permanent for free. A `"trigger"` card's price is reduced by what the trigger declares, using the same units `discardPayoff` (ADR 0126, 2026-10-06 amendment) uses. The heuristic does not try to bait discards (it does not hold a Baloth in hand on purpose). That is a strategy question for a later bot PR. Arena measurement is not expected to move. These are rare cards, so the PR carries a suite position rather than an arena claim.

### 6. Client and protocol

No new prompt and no new event kind. `EventDiscardCard` already has `NewZone`, so the log line becomes "Alice discarded Obstinate Baloth (onto the battlefield)" using the existing NewZone suffix the client prints for madness and Library of Leng. If that suffix does not exist yet for the battlefield, it is a one-line `labels.ts` addition. `discard_by_opponent` is additive and omitted when empty, so the protocol schema stays compatible. Nephalia Academy's "reveal" is the existing `EventReveal` emitted by its `Replace`.

### 7. Testing

Engine (`game/discard_entry_test.go`), with a test-registered replacement so the engine is pinned without a card:

- An opponent's resolving spell makes the owner discard the card: it enters under the **owner's** control, summoning sick. `EventETB` and then `EventDiscardCard{NewZone: ZoneBattlefield, Cause: effect, CauseController: opponent}` are emitted. The discard effect's `then` counts it as discarded.
- The owner's own spell, a cost discard (including the owner's own echo or cumulative-upkeep discard) and the cleanup discard all leave the card in the graveyard. An opponent's Painful Quandary follows question 4's answer.
- An ability whose source card the owner controls but whose stack item an opponent controls (and the reverse) is decided by the item.
- Madness and the battlefield replacement on the same card: the CR 616 ordering prompt is offered. Either order settles correctly, and the second replacement no longer applies once the first has moved the destination off the graveyard.
- The entry window pauses (two enters-tapped replacements to order). The card stays in hand, the rest of the batch waits, and both finish after the answer. Undo across the pause restores the hand.
- The entry window redirects to exile: the card ends in exile, `EventDiscardCard{NewZone: ZoneExile}` is emitted, and the card counts as discarded.
- Dodecapod's two counters arrive, and they stack with Hardened Scales-style entry modifiers.
- The card left the hand before the handoff ran: nothing moves, and the batch continues.

Card tests (one file per card per convention) for all 16 cards, each verified against its full oracle text. Roadmap: the `discarded-by-an-opponent` row moves to Implemented and its engine notes are rewritten. The wrong `DiscardCause` comment in `replacements.go` is corrected. Heuristic: a suite position where the bot is made to discard with a Baloth and a land in hand and pitches the Baloth.

### 8. PR plan

1. **Engine:** §1, §2 and §3, with the §7 engine tests and the comment fix. No cards.
2. **Cards:** all 16 cards (ADR 0106 owner decision 6), §4's predicate, the §5 declarations and heuristic, and the roadmap row.

## Alternatives considered

- **Read the cause controller in the catalog from the `Source` card's controller.** This is simpler and needs no new event field, but it is wrong whenever the stack item's controller is not the source card's controller (CR 113.8): a stolen Liliana's +1 or a copied Mind Rot. Rejected.
- **Expose `zoneRoute` (or its `Cause`) to `cards/effects`.** This leaks route plumbing (continuations, batch state) into the catalog to answer one question. Rejected in favour of one plain field.
- **Cancel the discard and queue a separate "put it onto the battlefield" effect.** The card would never have been discarded (no `EventDiscardCard`, so Syphon Mind and the discard payoffs miss it, against the rulings), and the entry would happen at a different time from the discard. Rejected.
- **Approximate the replacement with a trigger that returns the card from the graveyard.** It gives the opponent a window to exile the card in response, and it fires "dies"/"put into a graveyard" payoffs that should not. Rejected. This is the approximation Sand Golem prints on purpose, and these cards print the opposite.
- **Let `routeDestinationLocked` resolve `ZoneBattlefield` and `MoveCard` it there.** This skips the entire entry contract (see Context). Rejected.
- **One combined window for the discard and the entry.** That would need the replacement loop to change event kind mid-loop and re-gather appliers, which no other event does. Decision 3 explains why two chained windows are rules-equivalent here.

## Questions for the owner

1. **Where the battlefield redirect is handled.**
   - (a) In the engine: the one branch in `executeZoneRouteLocked` that hands a discard settled on the battlefield to the entry pipeline (§2).
   - (b) In the catalog: the card's `Replace` cancels the discard and calls `PutFromHandOntoBattlefieldForEffect` itself.
   - (c) Do the cause-controller half only (Nephalia Academy and the ten triggers), and leave the five creatures waiting.

   **Recommendation: (a).** (b) cancels the discard, so the card is not "discarded this way" and the discard event is lost (see Alternatives). (c) leaves the issue's own card, Nullhide Ferox, waiting.

2. **The entry's replacement window.**
   - (a) A second, chained CR 614 window for the entry (§3), so enters-tapped, enters-with-counters and entry-exile effects apply.
   - (b) No entry window: the card lands with only what the discard replacement said.

   **Recommendation: (a).** CR 614.12 makes entry replacements apply to this entry. (b) would let a Baloth dodge an opponent's "creatures enter tapped" effect.

3. **Log order of the two events.**
   - (a) `EventETB` first, then `EventDiscardCard` carrying the true landing zone, emitted at the entry's terminal outcome (§2).
   - (b) `EventDiscardCard` first, at the handoff, with `NewZone: ZoneBattlefield` assumed before the entry window has run.

   **Recommendation: (a).** Triggers are gathered together either way (CR 603.3). (b) can name the wrong zone when the entry window redirects or cancels the card.

4. **An "unless you discard" forced by an opponent's spell** (Painful Quandary: "that player loses 5 life unless they discard a card"). Under CR 118.12 the discard is a cost the player chooses to pay while the opponent's spell resolves. The engine files this one as an effect discard today, because `painfulQuandaryDiscard` queues an ordinary discard prompt. So under §1 it would put a Baloth onto the battlefield.
   - (a) Leave it as it is: the opponent's resolving spell is what makes you discard, so the card's replacement applies. No change.
   - (b) Make it a cost: move Painful Quandary onto the pay-unless discard (`PayActionDiscard`, `DiscardCauseCost`), so the Baloth goes to the graveyard. This matches the reading that costs are not effects (Library of Leng's ruling).

   **Recommendation: (a)** for now. The printed trigger is "a spell or ability an opponent controls **causes** you to discard", and the opponent's spell is what puts the choice in front of you. Library of Leng's ruling is about a different phrase ("an effect causes"). I could not find a Gatherer ruling that settles this combination, so it is worth a check. Either answer is one small change in Painful Quandary's file, not in the engine.

5. **Scope of the card PR.**
   - (a) All 16 Commander-legal cards (the 6 replacements and the 10 triggers) in PR 2, per ADR 0106 owner decision 6.
   - (b) Only the 6 replacements, and file the 10 triggered cards as a separate tier-2 issue.

   **Recommendation: (a).** The triggers need only §4's predicate, and splitting them out leaves them uncatalogued for no engine reason.

6. **The triggered "this card" when the discard was sent to exile instead** (madness or Rest in Peace ordered first).
   - (a) Trigger from the graveyard only, like cycling's "when you cycle this card", with a caveat.
   - (b) Watch from the graveyard and exile both.

   **Recommendation: (a)** to match the cycling precedent. (b) is a one-field change if the owner wants the corner covered.

## Consequences

- A replacement can tell who controls the spell or ability behind a discard, and `ReplacementEvent.Source` is finally set for discards, as its comment claims.
- A discard can land on the battlefield through the same entry every other effect uses. No new prompt kind, event kind, snapshot field or continuation slot is added.
- 16 cards become catalogable. Nullhide Ferox leaves the row's `Waiting` list, and the row moves to Implemented.
- Out of scope: discard caused by a static ability (no printed card needs it), team play's notion of an opponent, and any bot strategy that holds these cards to bait a discard.
