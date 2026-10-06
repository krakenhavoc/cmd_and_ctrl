# ADR 0116 — A card filter on the revealed-hand pick

**Status:** Accepted · 2026-10-03 · S58 — Deck requests, October batch (tracker [#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077)). Amended 2026-10-05 for the pick's variants ([#2115](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2115)); see [the amendment](#amendment-2026-10-05--variants-of-the-pick-2115).
**Issues:** [#2078](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2078) (this change). Requested cards: Unmask and Pelakka Predation // Pelakka Caverns ([#2063](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2063)).
**Owner decision behind it (2026-10-03):** fix it now, faithfully, with this ADR first. Thoughtseize ships stronger than printed today, which AGENTS.md §7 forbids. See [Owner decisions](#owner-decisions-2026-10-03).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-03. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 43 remote heads: `origin/develop`, `origin/main`, the `feat/s58-*` card and seam branches, and the other chore, docs, feat, fix, repro and wip branches. The highest number anywhere is 0115, so this ADR takes **0116**.
**Builds on:** [ADR 0010](0010-card-effect-catalog.md) §10 (the `PendingChoices` queue and Thoughtseize's flow), [ADR 0018](0018-triggers-on-the-stack.md) §6 (prompts that block the table), [ADR 0033](0033-ai-bot-seat.md) §1 and §3 (the legal-move enumerator and the bot's filtered view), [ADR 0041](0041-game-persistence.md) (restore points) and [ADR 0060](0060-leaving-the-game.md) (CR 800.4g reassignment).

This ADR was written plan-first. No engine, client or card code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

"Target player reveals their hand. You choose a nonland card from it. That player discards that card." is one of Magic's most printed shapes: 67 Commander-legal cards in the local Scryfall dump choose a card of some kind out of a revealed hand and make its owner discard it. The engine has the pick but cannot restrict it, so the one catalog card that uses it, Thoughtseize, lets you take a land. Its caveat admits this. Unmask and Pelakka Predation, both requested in #2063, wait on the same gap, and so does the rest of the shape.

Every claim below was checked in the code on `origin/develop` at `da257aff`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`). Rulings are Scryfall's, read on 2026-10-03.

### The rules

- **CR 701.9b:** "By default, effects that cause a player to discard a card allow the affected player to choose which card to discard. Some effects, however, … allow another player to choose which card is discarded." This shape is the second case.
- **CR 701.20a:** "To reveal a card, show that card to all players". Unmask's ruling (2014-02-01) says it outright: if you target yourself, "you must reveal your entire hand to the other players just as any other player would." Looking at a card shows it only to the specified player (CR 701.20e), which is Gitaxian Probe's verb, not this one.
- **CR 402.3:** a player can't look at another player's hand but may count it. The reveal is what lets everyone see it.
- **CR 609.3:** "If an effect attempts to do something impossible, it does only as much as possible." When no card in the revealed hand matches, the player has still revealed, nothing is chosen, and nothing is discarded. The rest of the effect still happens: Thoughtseize's ruling (2020-08-07) is "You lose 2 life even if the target player has no nonland cards in their hand to discard."
- **The pick is not optional.** "You choose a nonland card from it" has no "may", so when a matching card exists one must be chosen. The cards printed with "You may choose" (Nightsnare, Binding Negotiation) are a different shape (see [Out of scope](#out-of-scope)).
- **Characteristics in a hand:** a double-faced card has only its front face's (CR 712.8a), so Pelakka Predation is a nonland sorcery with mana value 3 in a hand. A split card's mana value is its combined cost (CR 709.4b, 202.3d), X is 0 off the stack (CR 202.3e), and an adventurer card has only its normal characteristics (CR 715.4). Pelakka's 2020-09-25 rulings repeat the first and third.
- **CR 608.2c:** the instructions run in the order written. Thoughtseize's life loss is printed after the discard.
- **CR 118.9:** Unmask's "exile a black card from your hand rather than pay this spell's mana cost" is an alternative cost, paid as the spell is cast (CR 601.2h).

### What exists

- **The entry point.** `Game.QueueDiscardFromRevealedHand(chooser, fromPlayer, source, count, reason)` (`game/pending_choice.go`) marks the chooser, and only the chooser, a knower of every card in the hand, caps `count` at the hand size, and queues a `PendingChoiceDiscardFromHand`. It emits no reveal event. Thoughtseize is its only caller.
- **The resolve.** `ResolvePendingChoice` checks that `len(picks) == Count` and that each pick is in `FromPlayer`'s hand, then dequeues and calls `discardCardsLocked` with `DiscardCauseEffect`, the one discard path, so madness and discard triggers already work. It does not check that the picks are distinct.
- **The view.** `viewOfPendingChoices` (`protocol/view.go`) inlines `FromPlayer`'s whole hand as `options[]`. `redactChoiceCards` keeps unknown cards as answerable backs for the chooser because the pool is a hand, and drops them for every other viewer.
- **The bot.** `legal.EnumerateFor`'s `choiceMoves` (`legal/choices.go`) offers every `Count`-sized combination of the whole hand. The heuristic (`aiseat/heuristic/choices.go`) ranks those moves by card value and never reads a filter of its own.
- **The client.** `ChoicePromptModal.svelte` renders this kind in its generic card grid: every option is a button, and a button is disabled only once `count` cards are already selected.
- **Restore points.** The prompt has no resume frame, so a table waiting on Thoughtseize's pick is restorable today. `closure_fields.txt` lists no route through it.
- **The reveal primitive.** `Game.RevealForEffect(RevealSpec)` (`game/reveal.go`) makes every seated player a knower of each named card and emits one grouped `EventRevealCards` run, which the client shows as a reveal banner and a log line.
- **The predicates.** `cards/effects/targets.go` has `Nonland()`, `Creature()`, `Artifact()`, `OfColor()`, `Noncreature()`, `ManaValueLE()` and the `And`/`Or`/`Not` combinators, and `modes.go` has `ManaValueGE()`. They read `ManaValueForEffect`, `IsLand` and the rest of the card's current characteristics, which for a card in a hand are the ones the rules above describe: `Card.SettleImported` materialises face 0 and a split card's combined cost.
- **The precedent for a filtered card prompt.** `PendingChoice.SacrificeOptions`, `SearchCards` and `ChooseCards` each carry the candidates as a list of instance IDs worked out when the prompt is queued. The snapshot carries each list, and the resolver checks each pick against it.

### The gaps

1. Nothing restricts the pick, on the server, on the wire, or in the bot.
2. A hand with no legal card still raises a prompt that must take a card: with no filter, every card is legal.
3. The reveal is shown to the chooser only. The table never sees the hand, and a reassigned prompt (CR 800.4g, ADR 0060) hands its new chooser backs.

---

## Decision

### 1. One entry point, with a filter

`QueueDiscardFromRevealedHand` takes a struct:

```go
type RevealedHandDiscard struct {
    Chooser, FromPlayer, Source uuid.UUID
    Count  int
    Reason string            // the prompt header, e.g. "Thoughtseize"
    Filter func(Card) bool   // nil: any card
    Label  string            // what may be chosen, e.g. "nonland card"
}
```

The card side never writes the closure by hand. A shared helper in `cards/effects` takes an `effects.CardPredicate`, binds it to the game and the resolving controller, and passes the result to the engine, so the filter is always built from the same predicates targeting and searching use: `Nonland()` for Thoughtseize and Unmask, `ManaValueGE(3)` for Pelakka Predation, and `And(Noncreature(), Nonland())` for Duress. No new predicate is written for this seam.

### 2. The reveal is a real reveal

The entry point reveals the whole hand with `RevealForEffect` (CR 701.20a): every seated player becomes a knower of every card, and the table gets the reveal banner and a log line. This replaces the chooser-only `AddKnower` loop. It is also what keeps a reassigned prompt answerable, because its new chooser is already a knower. An empty hand reveals nothing and queues nothing.

### 3. The filter is applied once, at the reveal

The filter runs over the hand once, in hand order, and the matching instance IDs are stored on the prompt as `PendingChoice.DiscardOptions []uuid.UUID`, with `Label` as `PendingChoice.DiscardLabel`. `Count` becomes `min(Count, len(DiscardOptions))`.

A list of IDs rather than the predicate, for the same reason as `SacrificeOptions`: a func cannot be written to a restore point, and a prompt with no closure on it stays a restore point. The list cannot go stale while it is open. The prompt blocks the table (`choiceGateDecisions`), so no player can act until it is answered, and the resolver checks the hand again anyway.

Fixing the set at the reveal is also the reading CR 608.2c gives. The card is chosen from the hand that was revealed, so a card that reaches the hand afterwards cannot be chosen.

### 4. No legal card: reveal, then nothing

When `DiscardOptions` is empty, no prompt is queued and the entry point returns `uuid.Nil`. The hand has already been revealed. The resolving effect carries on, so Thoughtseize still costs 2 life, as its ruling says (CR 609.3).

### 5. The server enforces it

`ResolvePendingChoice` refuses an answer, with `ErrInvalidParam`, when:

- it has a pick that is not in `DiscardOptions`; or
- it names the same card twice, a check that is missing today.

A pick that has left the hand is still `ErrCardNotFound`. The client and the bot are both offered only legal picks, but the server never relies on that: a hand-built `resolve_choice` naming a land in answer to Thoughtseize is refused and the prompt stays open.

A prompt with `DiscardOptions == nil` means the whole hand. Only a restore point written before this change has one, and it keeps today's behaviour. Every prompt this change queues has a non-empty list.

### 6. Ordering: no continuation

The pick stays frame-free. Anything printed after the discard (Thoughtseize's life loss, a pool card's "draw a card" or "scry 1") runs on the next line of the resolving effect, before the pick is answered. That cannot be observed:

- State-based actions and the trigger drain wait for the answer (#1289, `resolution_pause.go`), so a caster at 2 life who casts Thoughtseize loses only after the discard, as under the printed order (CR 704.3).
- The candidates were fixed at the reveal (§3), so a card the trailing clause draws or moves cannot become a pick.

A pool card whose later text reads the discarded card or whether a card was chosen is not this shape. It needs a continuation and stays on the registry (see [Out of scope](#out-of-scope)).

### 7. Wire and client

`PendingChoiceView` gains two fields for this kind:

- `eligible: string[]`: the instance IDs in `options[]` that may be chosen, from `DiscardOptions`. It is filtered by the same knower rule as `options[]`. Every seat is a knower after §2, so in practice every viewer gets the whole list. It is absent on a legacy prompt, which means every option.
- `eligible_label: string`: from `DiscardLabel`, e.g. `"nonland card"`.

`options[]` stays the whole revealed hand, because it was revealed and the chooser is entitled to see all of it. This is unlike Mox Diamond's `entry_discard_from_hand`, whose options are only the legal cards out of the chooser's own hand.

In `ChoicePromptModal`'s generic grid, a card that is not eligible renders dimmed with its button disabled, and `toggle` ignores it. The hint reads "Choose a nonland card from Ana's revealed hand. Ana will discard it." No `aria-label` or dialog name changes (the AGENTS.md labels contract).

### 8. The bot

`choiceMoves` draws its combinations from `DiscardOptions` when it is set, and from the whole hand otherwise. The heuristic ranks the moves it is given and needs no change. The bot can therefore only ever be offered, and only ever send, a legal pick. A seat-level whole-game test is not needed. The enumerator's own test is the gate, together with a policy test that asks the heuristic and checks that its answer is in the eligible set.

### 9. Snapshot impact

Additive within schema v7. The changes:

- **Two new fields:** `pendingChoices[].discardOptions` (array of uuid) and `pendingChoices[].discardLabel` (string), both `omitempty`.
- **The shape guard:** recorded with `-update-shape`.
- **The drift test:** both fields go into `snapshot_drift_test.go`'s PendingChoice table as `carried`.
- **The corpus:** one new board, `revealed_hand_discard_filtered.json` (Thoughtseize's pick open over a hand of a land and two spells), written with `-write-corpus`, which adds a board without touching the existing files.

The closure ratchet does not move: the filter never reaches `Game`.

A rollback binary does not check pending-choice keys. It drops the two fields and restores the prompt unfiltered, which is today's behaviour. That is the rollback case and needs no bump.

### 10. Cards

All three requested cards go to **Full**, each verified against its oracle text and rulings:

- **Thoughtseize, Caveats to Full.** `Nonland()`, then 2 life. Its caveat is deleted.
- **Unmask: Full.** `Pitch("Exile a black card from your hand", 0, CardInYourHand(…, OfColor("B")), …)`, Snapback's zero-life form of Force of Will's helper, then the same pick with `Nonland()`. A card cannot pitch itself (CR 601.2a moves it to the stack first), which the helper already enforces.
- **Pelakka Predation, the front face: Full.** It targets an opponent and filters with `ManaValueGE(3)`. The back, Pelakka Caverns, is already registered in `mdfc_lands.go` under `<oracle>#1`, so the face gets its own file under the bare oracle ID, in `agadeems_awakening.go`'s shape.

No other catalog card uses this pick. A grep for `QueueDiscardFromRevealedHand`, `PendingChoiceDiscardFromHand` and "reveals their hand" in `cards/effects` finds only Thoughtseize, and no other caveat names this gap.

**The pool.** Per [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) owner decision 6, as restated by [ADR 0113](0113-small-seams-for-the-s58-deck-requests.md) owner decision 2, the seam lands every catalog-pool card it alone unblocks. The 65 other cards of the discard shape split in two:

- **Plain shape (24 cards).** These print the plain sentence, alone or followed by an independent clause. They are Coercion, Dark Inquiry, Despise, Distress, Divest, Duress, Encroach, Inquisition of Kozilek, Lay Bare the Heart, Ostracize, Pilfer, Psychic Spear, Shattered Dreams, Brainbite, Harsh Scrutiny, Thought Erasure, Gix's Caress, Render Speechless, Humiliate, Toll of the Invasion, The Torment of Gollum, Diplomacy of the Wastes, Ego Drain and Memory Theft.
- **Everything else (41 cards).** These are modal, triggered or activated, or carry another mechanic: Mardu Charm, Grief, Pilfering Imp, Collective Brutality, the sagas, Venarian Glimmer's X, Nightmare Void's dredge and Down for Repairs' Attractions among them. Each lands only if its other text already works.

Every card in both groups is verified against its full oracle text in a pool PR. A card that needs anything else goes on the matching `Waiting` list with the reason.

### 11. Tests

One per rule interaction. In `internal/game`:

- **The filter offers only matching cards.** A hand of Forest, Pelakka Predation and Lightning Bolt gives Thoughtseize's prompt the two spells. The MDFC counts as nonland (CR 712.8a). Pelakka's filter over Fire // Ice (combined mana value 4, CR 709.4b) and an `{X}{R}` card (mana value 1, CR 202.3e) offers only Fire // Ice.
- **The server refuses a non-matching pick sent anyway.** Answering with the Forest is `ErrInvalidParam`, the prompt stays open, and the hand is unchanged. A duplicated pick is refused the same way.
- **A hand with no match reveals and discards nothing.** A hand of three lands: no prompt, every seat is a knower of every card, one reveal run is emitted, the hand is unchanged, and Thoughtseize's caster still loses 2 life.
- **The reveal reaches the whole table.** In a four-player game every seat is a knower after the reveal, not only the caster.
- **Count is capped** at the number of matching cards.
- **A legacy prompt** with `DiscardOptions == nil` still accepts any card in the hand.
- **Snapshot restore mid-prompt.** Capture with the prompt open, `RestoreStrict`, then refuse the land and accept the spell on the restored game. The new corpus board covers the same.
- **Reassignment.** The chooser concedes with the prompt open. The new chooser gets the same eligible set and can answer it.

In `internal/legal` and `internal/aiseat`:

- **The bot picks only matching cards.** `EnumerateFor` offers no move naming the land, and the heuristic's chosen answer is in the eligible set.

In `internal/protocol`:

- `eligible` and `eligible_label` reach the chooser and the other seats after the reveal, and are absent on a legacy prompt.

In the client, a vitest:

- An ineligible card's button is disabled, clicking it selects nothing, and the hint names the label.

The card tests: Unmask cast through its pitch (and refused when the only black card is itself), and the Pelakka front-face round trip with its back still a land.

### 12. Registry

- **`revealed-hand-restricted-pick` closes.** It goes to **Implemented**. `Missing` and `Waiting` are emptied, `Rules` becomes 701.9b, 701.20a and 609.3, and `Examples` gains Thoughtseize, Unmask and Pelakka Predation. It gets its fragment under `docs/engine-seams/closed/`.
- **A new row for the variants.** The variants of the pick that this seam does not build get a new row, `revealed-hand-pick-variants`, with the cards from [Out of scope](#out-of-scope) on its `Waiting` list.
- **Regeneration:** `go test ./internal/roadmap/ -update`.

---

## Delivery

1. **This ADR.** Docs only.
2. **The seam PR** (`feat/s58-seam-2078-revealed-hand-filter`). It carries the engine, wire, client, bot and snapshot changes; Thoughtseize, Unmask and Pelakka Predation; the registry; `docs/protocol.md` for the two fields; and a recipe in `docs/adding-cards.md`. Its `## Issues` section says `Closes #2078`.
3. **Pool PRs** of 10 to 20 cards each, branched from develop after the seam PR merges, plain-shape cards first.

## Consequences

- Thoughtseize stops being stronger than printed, and the shape becomes a few lines per card.
- Every player now sees a revealed hand, as the rules say. Until now only the caster did.
- The prompt stays a restore point.

## Out of scope

These go on the new `revealed-hand-pick-variants` row:

- **Exiling instead of discarding.** Appetite for Brains, Aggressive Negotiations and Agonizing Remorse exile the chosen card, which is not a discard: no madness, no discard triggers. It needs a destination on the pick.
- **"You may choose … If you do / If you don't".** Nightsnare, Binding Negotiation, Reckoner Shakedown, Traumatic Revelation and Revealing Eye need an optional pick and a continuation that learns whether a card was chosen.
- **Clauses that read the chosen card.** Talara's Bane gains life equal to its toughness.
- **Choices made before the pick.** Addle chooses a color first.
- **Counts that are not fixed.** Last Rites chooses one card for each card discarded this way.

## Owner decisions (2026-10-03)

The owner approved the faithful filter on 2026-10-03: fix Thoughtseize now, and build the filter that its printed text and the rules require. The rules above settle every remaining question. A card with no legal pick reveals and discards nothing (CR 609.3 and Thoughtseize's ruling), a reveal is shown to every player (CR 701.20a and Unmask's ruling), and the pick is mandatory when a legal card exists. No question is left for the owner.

---

## Amendment 2026-10-05 — variants of the pick (#2115)

**Issue:** [#2115](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2115), the `revealed-hand-pick-variants` registry row. **Sprint:** S58.

[Out of scope](#out-of-scope) listed five variants of the pick. Each was checked in the code on `origin/develop` at `de685935a` before it was called missing, and each rule below against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`). The rulings are Scryfall's, read on 2026-10-05.

### The rules

- **CR 701.9a:** "To discard a card, move it from its owner's hand to that player's graveyard." Exiling a card from a hand is not a discard. Madness (CR 702.35a: "If a player would discard this card …") and every "whenever a player discards" therefore never see Appetite for Brains, Aggressive Negotiations or Agonizing Remorse.
- **"You may choose"** is optional even with a legal card in the hand (Extract the Truth's 2022-04-29 ruling). "If you don't" happens when nothing was chosen, which includes a hand with nothing to choose. Nightsnare's 2015-06-22 ruling says the player then discards two cards of their own choice.
- **CR 608.2c:** the instructions run in the order written. Talara's Bane's life gain is printed before the discard, and its 2008-08-01 ruling reads the toughness while the card is in the hand. Reckoner Shakedown's counters are chosen after the pick (its 2022-02-18 ruling).
- **CR 608.2h:** information an effect reads off an object that has left its zone is last-known information.
- **CR 113.6a, 208.2a, 604.3:** a characteristic-defining ability works in every zone. Talara's Bane's ruling says the same of a `*` toughness in a hand.
- **CR 608.2n:** a resolved instant or sorcery is put into its owner's graveyard as the final part of its resolution. In this engine that happens before a prompt the resolution raised is answered, which matters only to a clause that reads the graveyards (§6 below).
- **CR 701.27e:** "When this creature transforms into Revealing Eye" triggers only on a transform that leaves Revealing Eye's face up.
- **Agonizing Remorse** (its 2020-01-24 rulings): a graveyard card may be a land; a card must be chosen from either place if any exists; nothing can happen between the choice and the exile; and with nothing to choose the caster still loses 1 life.

### What existed

- `RevealedHandDiscard` (§1) always discarded and always took exactly `Count`, with no continuation.
- The resolution-time colour prompt (`QueueColorChoiceThenForEffect`, Wash Out) and the prompted discard run (`PlayerDiscardsThenForEffect`, #1027) each hand a continuation the answer, with the resolution still open (#1289).
- The keyed body registry of ADR 0041 phase 3 (`effect_bodies.go`) names a delayed trigger's body by an on-disk key, with a ledger (`testdata/effect_keys.txt`) and a restore-time refusal of an unknown key (`ErrUnknownEffectKey`).
- `TransformPermanentForEffect` emits `EventTransform`, but no card watched it.
- Nothing applied a characteristic-defining ability off the battlefield.

### Decision

1. **A kind of its own.** A pick with any variant is `PendingChoiceRevealedHandPick` (`"revealed_hand_pick"`), raised by `Game.RevealedHandPickForEffect(RevealedHandDiscard)`. `QueueDiscardFromRevealedHand` hands a variant to it, and a plain pick is unchanged. A new kind rather than more fields on `discard_from_hand`, for the rollback case (ADR 0041 Decision 5). A binary from before this change does not know the kind, so it refuses the restore point (ADR 0115 §8) and the file is kept. Fields on the old kind would have been dropped by that binary, restoring an exile as a discard or an optional pick as a mandatory one.
2. **A destination.** `RevealedHandDiscard.Destination` is `PickDiscard` (the zero value) or `PickExile`. An exile goes through the one exile route (`ExileCardsThenForEffect`), never the discard path, so no `EventDiscardCard` is emitted (CR 701.9a). The vocabulary is closed: restore refuses any other value.
3. **An optional floor.** `Optional` makes the empty answer legal. Any other answer must still name exactly `Count` cards from the candidates, none twice (§5).
4. **The graveyard.** `FromGraveyard` adds every card in `FromPlayer`'s graveyard to the candidates, after the hand's matches and unfiltered. A pick is accepted while the card is still in the hand or that graveyard.
5. **A keyed continuation.** `Then` is a `RevealedPickThen` registered once at init with `RegisterRevealedPickThen(key, fn)`, or `RegisterRevealedPickFirst(key, fn)` for a clause printed before the move. The key shares the effect-key namespace, sits in the ledger as `pick <key>`, and is refused at restore if this binary lacks it. The prompt carries only the key (`PendingChoice.PickThen`), so it stays a restore point. The function is handed a `RevealedPick`: the chooser, the revealing player, the source, and the chosen cards as value copies read before they moved (their last-known information, CR 608.2h). A `Then` runs after the move has finished; the move is itself a continuation, so a move that pauses on a replacement's prompt holds it. A `First` runs before the move and moves the cards by calling `RevealedPick.Done`. Keys are on-disk identities, never renamed or reused, and they have no alias.
6. **A number read as the pick goes up.** `Measure func(Card) int` reads each candidate when the pick is raised, while the spell is still on the stack. The readings are stored as `PendingChoice.PickMeasures` and handed to the continuation for the chosen cards. Talara's Bane needs this. Read at the answer, a Tarmogoyf's toughness would count Talara's Bane, a sorcery, already in its owner's graveyard (CR 608.2n as this engine orders it), which is one more than printed. Nothing else can change in between, because the prompt stops the table. Like `Filter`, the func is never stored. `Game.ToughnessAnywhereForEffect` is the toughness reader: off the battlefield it applies the card's own layer-7a abilities (CR 113.6a), so a `*` creature card is read correctly.
7. **Nothing to choose still runs the rest.** With no candidate no prompt goes up (§4), and the continuation runs at once, told that nothing was chosen (CR 609.3). "If you don't" happens, and Tourach's Canticle still discards at random.
8. **Opponent protection is checked at the discard.** #2178's "can't cause you to discard" stops the discard, not the reveal or the choice, so a variant pick still goes up and Talara's Bane still gains the life. The check is made when the cards would move, with the chooser as the cause. The chooser is the controller of the effect that asked, and asking the prompt makes the answer independent of whatever resolved last.
9. **Choices and counts from earlier text compose.** Addle's colour is the existing resolution-time colour prompt, whose answer raises the plain pick filtered by `OfColor`. Last Rites' count is the prompted discard run's `PromptedDiscards.Count()`. Both prompts are queued while the resolution is open, so the pick is stamped `midResolution` and state-based actions still wait (#1289). The earlier prompt holds a closure, as it always has, so the table is not a restore point until it is answered. The pick it raises is a restore point.
10. **Wire and client.** The view sends the new kind with the same `options[]` (the revealed hand, then the graveyard when it is open), `eligible` and `eligible_label`, plus `choose_min` / `choose_max`, `pick_destination` and `pick_from_graveyard`. Every seat gets all of it, because the hand was revealed (CR 701.20a) and a graveyard is public, so nothing new is hidden or leaked. `ChoicePromptModal` greys the ineligible cards as before. With nothing selected on an optional pick, its primary button reads "Choose nothing". It says when the card will be exiled, and it captions a graveyard card. No registered label changes.
11. **The bot.** `legal.EnumerateFor` offers every `Count`-sized set of candidates still in place, using `Game.RevealedPickCandidateLocked`, the resolver's own test. An optional pick also gets the empty answer, marked `AlwaysLegal`. The heuristic prices the kind as it prices the plain pick: it takes an opponent's best card, and from its own hand it chooses nothing when it may. Layer A escalates choices to the heuristic as before.
12. **Snapshot impact.** Additive within schema v7. Five `omitempty` fields on `pendingChoices[]` are recorded with `-update-shape` and classified `carried` in the drift table: `pickDestination`, `pickOptional`, `pickFromGraveyard`, `pickThen` and `pickMeasures`. `checkEffectKeys` refuses an unknown `pickThen` or `pickDestination`. Three corpus boards are new files, written with `-write-corpus`: `revealed_hand_pick_optional_then`, `revealed_hand_pick_exile_graveyard` and `revealed_hand_pick_measures`. The closure ratchet does not move: no closure reaches `Game`. The departure table reassigns the kind (CR 800.4g) like `discard_from_hand`.

### Cards

All Full, each checked against its oracle text and rulings: Appetite for Brains, Aggressive Negotiations, Agonizing Remorse, Nightsnare, Binding Negotiation, Reckoner Shakedown, Extract the Truth, Talara's Bane, Tourach's Canticle, Concealing Curtains // Revealing Eye, Addle and Last Rites. Revealing Eye is the first card to watch `EventTransform`.

Still waiting, on the same registry row, now `partial`:

- **Distended Mindbender:** "you choose from it a nonland card with mana value 3 or less and a card with mana value 4 or greater" is one pick of two cards under two different filters, which this shape does not express. Its cast also needs emerge (CR 702.119), which the engine lacks.
- **Traumatic Revelation:** the pick is this shape, but its "if you don't" branch is incubate (CR 701.53), which the engine lacks.

So #2115 stays open for the two-filter pick.
