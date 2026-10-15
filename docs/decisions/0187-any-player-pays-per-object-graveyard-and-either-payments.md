# ADR 0187 — Unless any player pays: one offer per land, graveyard exile, and an either payment

**Status:** Accepted (owner decisions 2026-10-14) · 2026-10-14 · Sprint: none (Backlog, parked on a dependency or a single card)
**Issues:** [#3205](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3205) (Cleansing: destroy each land unless any player pays 1 life) and [#3206](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3206) (any-player offers paid by graveyard exile or an either payment: Carrion Rats, Carrion Wurm, Worms of the Earth). Registry row `unless-any-player-pays` (Partial; tracked under #1903, S50 tracker #1784, found landing #1860, S53 tracker #1910).
**Builds on:** [ADR 0150](0150-other-players-pay-any-player-offers-and-assist.md) (the any-player offer, owner decisions 1–9, and its follow-ups for Cleansing and the two payments), [ADR 0108](0108-turn-scoped-effects-object-history-and-damage-shields.md) §2 and §5 (rules-gate scoped effects; non-mana pay-unless payments), [ADR 0175](0175-players-choose-before-simultaneous-sacrifices.md) (public choices one seat at a time, `EarlierChoices`), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers), [ADR 0054](0054-dice-rolls-and-coin-flips.md) decision 4 (undo never lets a seat see the future twice).
**Depends on (not this ADR):** [#1985](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1985), "Lands can't enter the battlefield", for Worms of the Earth only.
**Numbering:** 0187 was reserved for this seam. After `git fetch --all --prune`, `git ls-tree --name-only <ref> docs/decisions/` on every remote branch (`develop`, `main`, `cost-ledger`, `docs/issue-audit`, `feat/2417-incubate`, `feat/3343-autopass-reason-undo`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2950-darkness-replacement-continuation`, `fix/3159-3160-attachments`, `fix/caddy-reload-admin-off`, `wip/836-one-click-default`, and, at the re-check before pushing, `docs/0189-face-down-remainder`, `docs/2704-crew-and-saddle-adr`, `feat/rf-f-cast-during-resolution`, `feat/sideboard-outside-the-game`) showed 0185, 0186, 0189 and 0191 in use and 0187 free.

**Owner decisions:** the owner answered all six questions on 2026-10-14. Question 1 went against the recommendation: Cleansing follows CR 101.4 literally. The answers are under [Owner decisions](#owner-decisions-answered-2026-10-14) and are binding.

This ADR is plan-first: no code changes with it. Every claim about the code was checked on `origin/develop` at `8dd2c2e2c`. Every rule was checked against the Comprehensive Rules effective 2026-09-25 (`MagicCompRules 20260925.txt`).

---

## Context

ADR 0150 shipped the any-player offer: `Game.QueueAnyPlayerPaysForEffect` (`server/internal/game/any_player_pays.go`) asks every player still in the game, in APNAP order from the active player, one ordinary `PendingChoicePayUnless` at a time. The first payer ends it (`OnPaid`). When everyone has declined, `OnNobody` runs. Each link blocks its step, takes no standing answer, and carries `PendingChoice.AnyPlayerOffer` for the view and the bot. PR 2 added `PayActionLife`, `PayActionDamage` and `PayActionNothing` (`pay_unless_action.go`), so 21 cards ship. ADR 0150 left three gaps, and the registry row's `Missing` names all three: "Exiling cards from your graveyard as the payment, a choice between two payments, and one offer for each land (Cleansing) aren't supported yet."

### The waiting cards

The row's `Waiting` list is exactly four cards. A search of `data/scryfall/default-cards.json` (one row per oracle ID, `Card` and `Card // Card` type lines skipped) for `for each … unless (any|a) player`, `any player may (exile|sacrifice … or)` and `unless any player … or` finds the same four and no others. All four are Commander-legal and none is catalogued.

| Card | Oracle text | Gap |
|---|---|---|
| **Cleansing** {W}{W}{W} Sorcery | "For each land, destroy that land unless any player pays 1 life." | One offer per land (§2) |
| **Carrion Rats** {B} Creature — Rat 2/1 | "Whenever this creature attacks or blocks, any player may exile a card from their graveyard. If a player does, this creature assigns no combat damage this turn." | Graveyard exile payment (§3); "assigns no combat damage this turn" (§4) |
| **Carrion Wurm** {3}{B}{B} Creature — Zombie Wurm 6/5 | "Whenever this creature attacks or blocks, any player may exile three cards from their graveyard. If a player does, this creature assigns no combat damage this turn." | Same as Carrion Rats |
| **Worms of the Earth** {2}{B}{B}{B} Enchantment | "Players can't play lands. / Lands can't enter the battlefield. / At the beginning of each upkeep, any player may sacrifice two lands of their choice or have this enchantment deal 5 damage to that player. If a player does either, destroy this enchantment." | Either payment (§5); "Lands can't enter" is #1985, not this ADR |

Scryfall has no rulings for Cleansing, Carrion Rats or Carrion Wurm. Worms of the Earth's rulings (2004-10-04, 2006-05-01, 2008-08-01) are all about "Lands can't enter" and Two-Headed Giant, so none affects this ADR.

### The rules

- **CR 118.12 / 118.12a.** "[Do something] unless [a player does something else]" means "[A player may do something else]. If [that player doesn't], [do something]." The action is a cost, paid when the spell or ability resolves, and the "If … does" clause "checks whether the player chose to pay an optional cost … regardless of what events actually occurred". So Worms of the Earth is destroyed once a player has chosen and paid either payment, even if the 5 damage is then prevented (ADR 0150 §3 already treats `PayActionDamage` that way).
- **CR 118.3.** A player can't pay a cost without the resources to pay it fully: no "exile three cards" from a graveyard of two, and no "sacrifice two lands" with one land.
- **CR 119.4.** A payment of more than 0 life is allowed only if the life total is at least the amount. **CR 119.8.** A player who can't lose life can't pay life. Both go through `CanPayLifeLocked`, which every life cost already uses.
- **CR 101.4** (APNAP): the active player chooses first, then each nonactive player in turn order, and "Then the actions happen simultaneously." **CR 101.4b**: each player knows the earlier players' choices. **CR 608.2e**: an effect's choices that involve several players are made in APNAP order, then that action is processed simultaneously. ADR 0150 decision 1 records the one simplification the engine makes: it pays each answer as it is given and asks nobody after the first payer, because a second payment buys nothing.
- **CR 608.2h.** Information an effect needs from the game is determined once, when the effect is applied. Cleansing's "each land" is the lands on the battlefield as it resolves.
- **CR 701.8a.** Destroy moves a permanent to its owner's graveyard (so regeneration and indestructible apply). **CR 701.21a.** A player can sacrifice only permanents they control. **CR 406.2.** Exile is a move to the exile zone.
- **CR 510.1 / 510.1a.** Each attacking and blocking creature assigns combat damage equal to its power, and one that would assign 0 or less "doesn't assign combat damage at all". "Assigns no combat damage" is therefore not prevention (CR 615). An effect that says damage can't be prevented does not undo it, and nothing is assigned, so nothing triggers on assignment.
- **CR 800.4f.** If an object requires a player who has left the game to pay a cost or choose whether to pay one, that cost is not paid.

### What the engine has (develop at `8dd2c2e2c`)

- **The offer.** `AnyPlayerPays` / `AnyPlayerOffer` and `offerAnyPlayerFromLocked` (`game/any_player_pays.go`). Each link is `queuePayUnlessLocked(..., payUnlessExtras{onPay, action, offer, acceptLabel, declineLabel})` (`game/pending_choice.go`). `resolvePayUnless` validates the answer before the prompt is dequeued, and `answerPayUnlessLocked` pays (`payActionLocked`) and runs exactly one branch. Card side: `effects.AnyPlayerPays` (`cards/effects/any_player_pays.go`) with `Pay UpkeepPayment`, `PerPayer`, `Opponents`, `GuardsSource` and the labels.
- **Non-mana payments.** `PayAction{Kind, Count, Of PermanentQuery}` is plain data on the frame. `payActionOptionsLocked` lists what may be named (hand for a discard, matching permanents for a sacrifice), `validatePayActionLocked` checks exactly `Count` distinct live options, `payActionLocked` pays, and `NamesNothing` marks the yes/no kinds (energy, life, damage, nothing). Wire: `PendingChoiceView.PayCards{Action, Count, Options}` and `PayLife` / `PayDamage` / `PayNothing`. Answer: `{apply, card_ids}` routed through `actions.go` to `ResolvePayUnlessWithCards`.
- **Exiling cards as a cost.** `payExileCardsCostLocked` (`game/exile_cost.go`) moves each card through `routeCardToZoneLocked` with `MoveCause{Kind: MoveCauseCost}` and `MustSettleNow`, with CR 903.9 answers collected before the payment (#1397). A pay-unless payment happens while something resolves, not during a cast, so it may pause the way `payActionLocked`'s discard batch already does.
- **A card-set pick, one seat at a time.** `PendingChoiceChooseCards` with `chooseCardsFrame` (`game/chained_choice.go`), including `earlierPermanentChoices` (ADR 0175, `ordered_permanent_choices.go`). The view turns those into `PendingChoiceView.EarlierChoices`, and the client already draws them (`ChoicePromptModal.svelte`, `lilianaEarlierChoices.render.test.ts`). `PendingChoice.ChooseSuggested` (with `withSuggestedFirst` in `legal/choices.go`) puts a suggested set first in the move list. The kind blocks the table (`choiceGateDecisions`). Its departure decision is `{reassign: true, onDrop: dropDefault}` (`leave_game.go`), and reassignment needs `FromPlayer != Chooser` (`departedChoiceObjectLocked`).
- **A rules gate on one object for the turn.** `ModCantBeRegenerated` (`game/scoped_effects.go`, reader `readerRule`) is a scoped effect pinned to an object until cleanup (`CantBeRegeneratedThisTurnForEffect`) and read at one gate (`permanentCantBeRegeneratedLocked`, `game/regeneration.go`), with a view chip (`CardView.CantBeRegenerated`). Combat damage amounts come from one place: `assignAndDealCombatDamageLocked` (`game/mutations.go`) snapshots `c.combatDamagePower()` for every combatant before it marks anything. Nothing in the engine says "assigns no combat damage", and 26 Commander-legal cards print it (Ophidian, Laccolith Rig, Goblin Vandal and others, beyond the two Carrion cards).
- **Simultaneous destruction.** `DestroyPermanentsThenForEffect(ids, then)` (`game/simultaneous.go`).
- **Triggers.** "Whenever this creature attacks or blocks" is `OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, … attackDeclared(ev, source) || selfBlocksOnce(ev, source) …)` (Goblin Psychopath).

## Decisions

### 1. Who is asked, and in what order: unchanged from ADR 0150

Every new shape below goes through the same chain: the players still in the game, in APNAP order from the **active** player (`votersFromLocked(activePlayerIDLocked())`), the controller included, with `Opponents` narrowing it. A player who has left is skipped, and one who leaves while asked declines (CR 800.4f). ADR 0150 owner decisions 2 and 7 bind here too: each link blocks its step, no standing answer applies, and a bot doesn't free-ride. Decision 1's order (APNAP) holds everywhere. Its "the first payer ends the offer" holds for Carrion and Worms, but not for Cleansing, where every player chooses and every choice is paid (§2, owner decision 1).

**When nobody pays, or nobody can be asked:** Cleansing destroys every land it could not save, Carrion Rats and Carrion Wurm deal their combat damage as usual, and Worms of the Earth stays. `OnNobody` runs once, at once if there is nobody to ask.

### 2. One offer per object: Cleansing (#3205)

"For each land" is one offer per land. ADR 0150 owner decision 4 rejected asking it one land at a time (up to 160 prompts at a four-seat table). Under [owner decision 1](#owner-decisions-answered-2026-10-14), Cleansing follows **CR 101.4 literally**: every player, in APNAP order, chooses which lands they will pay for, knowing what the earlier players chose (CR 101.4b). Then all the payments happen. Two players may pay for the same land, and each of those payments is still made. A land is saved if anyone paid for it. Unlike ADR 0150 decision 1, this shape does **not** stop at the first payer, and it doesn't hide lands someone has already chosen.

**A new door, `Game.QueueAnyPlayerPaysForEachForEffect(AnyPlayerPaysForEach)`**, in a new file `game/any_player_pays_each.go`:

```go
type AnyPlayerPaysForEach struct {
    Source, Controller uuid.UUID
    Objects  []uuid.UUID // the battlefield objects, frozen as the effect is applied (CR 608.2h)
    EachLife int         // life per object paid for; the only payment this door takes (1 for Cleansing)
    Noun     string      // "land" / "lands", for the words
    Question string
    // OnUnsaved gets the objects nobody paid for, still the same objects
    // (epoch-checked), and runs exactly once, after every payment.
    OnUnsaved func(g *Game, unsaved []uuid.UUID) error
}
```

**Phase 1: choosing, one player at a time, in APNAP order.** The door builds one leg per player still in the game, in `votersFromLocked(activePlayerIDLocked())` order, and runs them through the ADR 0175 collector, `ChoosePermanentsInOrderThenForEffect` (`game/ordered_permanent_choices.go`). That collector already does four things this needs. It asks one seat at a time. It skips a player who has left. It binds each pick to the object by `ObjectEpoch`. It carries the earlier players' picks on each later prompt (`earlierPermanentChoices`), which the view already sends as `EarlierChoices`. **Nothing is paid during this phase.** Each leg is a `PendingChoiceChooseCards` (the existing kind, so the gate, the enumerator, the dock, the departure table and the restore census all treat it as they already do), with:

- `Cards`: **every** object in `Objects` still on the battlefield as the same object, including ones an earlier player has chosen.
- `Min: 0`, `Max: min(len(Cards), affordable)`. Here `affordable` is how many `EachLife` payments the player can make one after another: `life / EachLife` when `CanPayLifeLocked(p, EachLife)`, else 0 (CR 119.4 applies to each payment, and CR 119.8). It's worked out when the door builds the legs. No life changes before the payment phase, so it still holds when each player answers. A player who can afford none gets no leg. They are logged as declining (`EventAnyPlayerDeclined`) when the door is called, because a prompt whose only answer is "none" is a wasted click.
- `Zone: ZoneBattlefield`, `FromPlayer: Chooser`. Setting `FromPlayer` to the chooser stops the departure table from reassigning the leg to another seat, because `departedChoiceObjectLocked` needs `FromPlayer != Chooser`. A leg whose chooser leaves is dropped with `dropDefault`: its continuation runs with nothing picked, so that player pays for nothing (CR 800.4f), and the next player is asked.
- `AnyPlayerOffer`: the same plain data ADR 0150 put on a pay-unless link (`Controller`, `Order`, `Declined`), plus one additive field, **`EachLife int`** (`json:"eachLife,omitempty"`). On this shape, `Declined` lists the earlier players who chose nothing.
- `ChooseSuggested`: the chooser's own objects that **no earlier player has chosen**, in engine order, cut to `Max`. It's worked out when this leg is queued, from the earlier picks. A player's own lands that someone already pays for are left out, so the suggestion never pays twice for nothing ([owner decision 6](#owner-decisions-answered-2026-10-14)).

The collector's `ChooseCardsPrompt` legs gain three unexported fields to carry this: `anyPlayerOffer *AnyPlayerOffer`, `suggestUnchosenOwn bool`, and the `Declined` update. `collectPermanentChoicesLocked` copies them onto each queued `PendingChoice` and fills `ChooseSuggested` and `Declined` from the receipts so far. Leg-level `Then` stays refused, as the collector requires.

**What the earlier choices look like.** Every leg's `EarlierChoices` is one entry per earlier player, `{player_id, card_ids}`, in APNAP order. It holds the lands that player has chosen to pay for, filtered to the ones still on the battlefield as the same object. An earlier player who chose nothing appears with an empty `card_ids` (they're also in `AnyPlayerOffer.Declined`). The client (§8) draws this three ways:

- a badge on each land with the names of everyone already paying for it ("Alice pays");
- a summary line ("So far: Alice pays for 3 lands, Carol for none");
- a warning on a land the viewer selects that someone already pays for ("Alice already pays for this: you'd both pay").

Every seat sees the same entries. The battlefield and these choices are public (CR 101.4b).

**Phase 2: paying, after the last player has chosen.** The collector's final `then` gets `PromptedPicks`, one `SeatCards` per player in APNAP order, already filtered to the objects still there. For each player in that order, it pays `EachLife` life **once per object they chose**, as separate payments: one `payLifeAsCostLocked` per object ([owner decision 2](#owner-decisions-answered-2026-10-14)), so a "whenever you lose life" trigger sees each one. Two players who chose the same land both pay for it. `CanPayLifeLocked` is checked before each payment. A payment that fails is a land that player didn't pay for. Nothing inside the resolution can make it fail today, but the check keeps CR 118.3 true. Each payer gets one `EventAnyPlayerPaid` with the label "N life (N lands)". The payments of different players happen inside one locked step, with no priority between them, so they are simultaneous in effect (CR 101.4 "Then the actions happen simultaneously"), and any life-loss triggers wait for the next time a player would get priority.

**Then the outcome.** A land is saved if at least one player paid for it. `OnUnsaved` gets every object in `Objects` that nobody paid for and that is still the same object. Cleansing destroys them in one `DestroyPermanentsThenForEffect` (CR 701.8a), so regeneration, indestructible and "dies" triggers apply as they do for Armageddon. With nobody to ask, or nobody able to pay, `OnUnsaved` runs at once on every object.

**Lands that change during the offer.** Nothing resolves while the offer is open, so the set can only shrink, and only through a departure (CR 800.4a takes that player's lands). The epoch check drops those lands from later legs, from the picks and from `OnUnsaved`.

**Door scope.** `EachLife` is the only payment. A mana or card payment per object has no card. A zero `EachLife` is a programming error: the door reports it with an `EventEffectError` and runs `OnUnsaved` on every object, so the card still does something, as `QueueAnyPlayerPaysForEffect` does with an unparseable cost.

Card side: `effects.AnyPlayerPaysForEach{Objects func(ctx) []uuid.UUID, EachLife, Noun, Question, OnUnsaved func(ctx, []uuid.UUID) error}`, bound to the resolving item as `AnyPlayerPays` is.

**As built (slice 4, #3572, 2026-10-15).** Three details the text above left open:

- `AnyPlayerOffer.Declined` on a leg lists, in `Order`, the earlier players who chose nothing **and** the players who were given no leg because they can't pay. Both said no. `EarlierChoices` lists only players who were asked, so a player with no leg has no entry there.
- A leg's `EarlierChoices` and `ChooseSuggested` are fixed when the leg is queued, like every ADR 0175 prompt. A player who leaves afterwards takes their lands off the table, so the candidate list and the payments shrink at once, but a leg already open keeps the earlier entry it was built with.
- `ChooseSuggested` is cleared on every seat's view but the chooser's, with `Options` and the bounds that `choose_cards` already withholds from them. The offer, `Declined` and `EarlierChoices` stay public (CR 101.4b).

- `SnapshotSchemaVersion` is 9 now, not 8 (§9 was written against 8). The additive `anyPlayerOffer.eachLife` line went into `testdata/snapshot_shape/v9.txt`, and `v8.txt` stays frozen. The rest of §9 holds without a bump.

### 3. Exiling cards from your own graveyard as the payment (#3206)

A new `PayActionKind`, **`PayActionExileFromGraveyard`** (`"exile_graveyard"`), in `pay_unless_action.go`: exile `Count` cards from the payer's own graveyard.

- `payActionOptionsLocked`: the payer's graveyard, in `Zone.Cards` order. It is a card pick (`NamesNothing` stays false), so the existing `{apply: true, card_ids}` answer, `validatePayActionLocked` (exactly `Count` distinct live options) and the "fewer options than Count makes Pay a decline" rule in `resolvePayUnless` all apply unchanged (CR 118.3).
- `payActionLocked`: one batch through the exit primitive, `routeCardToZoneLocked` to `ZoneExile` with `MoveCause{Kind: MoveCauseCost, Controller: payer}` and `Source`, chained card by card with the route's `then`. `MustSettleNow` is **not** set: this payment is made while something resolves, so a commander exiled this way gets its CR 903.9a answer the way the discard batch's commanders do, and the `then` that runs `onPay` runs once the last card has landed. A graveyard exile does not use `payExileCardsCostLocked`, because that function assumes the answers were collected before a cast.
- Wire: `PayCards{Action: "exile_graveyard", Count, Options}`. Graveyards are public, so `Options` goes to every seat, unlike a discard's.
- Card side: `effects.ExileFromGraveyardPayment(n)` returning an `UpkeepPayment` with the words "exile a card from your graveyard" / "exile three cards from your graveyard".

### 4. "Assigns no combat damage this turn"

Carrion Rats and Carrion Wurm need it, and nothing in the engine says it. It's a rule about how damage is assigned (CR 510.1a), not prevention, so a prevention shield would be wrong whenever damage can't be prevented. It gets the `ModCantBeRegenerated` shape:

- `ModAssignsNoCombatDamage ModKind = "assignsNoCombatDamage"` (reader `readerRule`), pinned to the object, until cleanup. The constructor is `Game.AssignsNoCombatDamageThisTurnForEffect(sourceID, cardID uuid.UUID, label string) bool`, which registers nothing for a card that isn't on the battlefield. The kind is a new on-disk identity, never renamed or reused.
- One reader, `permanentAssignsNoCombatDamageLocked(c *Card) bool`, consulted in `assignAndDealCombatDamageLocked` where the per-card `power` map is built: such a combatant's amount is 0, ahead of `combatDamagePower()`'s `DamageByToughness` and `DamageAsPositivePower`. Everything downstream (the damage-assignment prompt, trample, lifelink, deathtouch, commander damage) already reads that map. A creature with 0 there assigns nothing (CR 510.1a).
- View: `CardView.AssignsNoCombatDamage bool` (`json:"assigns_no_combat_damage,omitempty"`), stamped like `CantBeRegenerated`. The client shows a status chip ("Assigns no combat damage this turn").
- Bot: a `combatPower(c *protocol.CardView) int` helper in `aiseat/heuristic` answers 0 for such a creature. Where the attack and block scorers (`combat.go`, `blockset.go`, `race.go`) read a combatant's power as the damage it deals, they read it through this helper instead.

The other 24 cards that print "assigns no combat damage" are not in scope. Under owner decision 4, slice 1 adds the registry row `assigns-no-combat-damage` with them on its Waiting list, tracked by the follow-up seam issue [#3531](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3531).

### 5. An either payment: Worms of the Earth (#3206)

"Any player may sacrifice two lands of their choice or have this enchantment deal 5 damage to that player. If a player does either, destroy this enchantment." It's one offer with two ways to pay, not two offers: "does either" is one condition.

- **`PayActionEither`** (`"either"`), and a new plain-data field on `PayAction`: **`Either []PayAction`** (two or more payments, none of them `PayActionEither`). It's a `PayAction` on the frame like any other, so it adds no closure and no snapshot field (`payUnlessResume` is already a census frame).
- **Payable** when any branch is: `payActionPayableLocked` for a branch that names nothing, and `len(payActionOptionsLocked) >= Count` for a card branch. Worms' damage branch is always payable, so Worms' offer always is.
- **The answer names the branch.** A new answer field, **`pay_option`** (a `PayOption *int` field, `json:"pay_option"`, on the `resolve_choice` params struct in `actions/actions.go`, beside `OptionalApply`), is an index into `Either`, sent with `apply: true` and, for a card branch, `card_ids`. `resolvePayUnless` validates it before dequeueing: the index must be in range, the branch is validated as that `PayAction` on its own (card pick or names-nothing), and a `pay_option` on a prompt that isn't `PayActionEither` is refused (`ErrInvalidParam`), as a stray `card_ids` is. `apply: true` on an either payment without `pay_option` is refused too, so no branch is ever guessed. A new entry point, `ResolvePayUnlessWithOption(choiceID, chooser, apply, option, cardIDs)`, is routed from `actions.go`.
- **Paying** runs `payActionLocked` with the chosen branch. `onPay` runs once that branch is paid. For damage, that is once the damage event has settled, however much landed (CR 118.12).
- **Wire:** `PendingChoiceView.PayEither []PayOptionView`, where each entry carries `{action, count, words, options}` (options only for a card branch, filtered as `PayCards` is: hand cards to the chooser only). `PayCost` carries the card's words. At most one of `PayCards` / `PayLife` / `PayDamage` / `PayNothing` / `PayEither` is present.
- Card side: `effects.EitherPayment(a, b UpkeepPayment) UpkeepPayment`.

**As built (slice 7, #3575, 2026-10-15).** Details the text above left open:

- `PayOptionView` also carries `payable` (always sent): whether the chooser can make that branch right now. The client greys a branch it can't pay with the reason, and no rule is duplicated in it. A card branch's `options` is omitted when empty, unlike `PayCards.options`.
- `PayAction` gains `Words`, the verb phrase naming a branch ("sacrifice two lands", "take 5 damage"). `EitherPayment` fills it from the two `UpkeepPayment`s; `PayActionPhrase` falls back to one built from the kind and count. It is plain data on the frame, so §6's census claim holds.
- `ResolvePayUnlessWithOption(choiceID, chooser, apply, option *int, cardIDs)` takes the index as a pointer, because 0 is the first branch. `resolvePayUnless` resolves the chosen branch first and validates it as its own payment, and `answerPayUnlessLocked` takes that branch as its payment. A branch the payer can't make, named with no cards, is a decline (CR 118.3), as every other unpayable "yes" is. An automatic answer names no branch and is never payable, so it declines (an either offer is step-owed and takes no standing answer in any case).
- An `Either` with fewer than two branches, or with an `Either` branch, is a programming error: `queuePayUnlessLocked` reports it as an `EventEffectError` and runs the decline, so the card still does something. `EitherPayment` panics on a mana or nested branch.
- The log line reads "Alice paid for Worms of the Earth (sacrifice two lands or take 5 damage)": the `any_player_paid` event is emitted by a closure that doesn't know the branch, so it carries the whole offer's words, and the sacrifice or damage event that follows names the branch taken.
- Bot (§7): damage at or above `phyrexianLifeFloor` scores 0.9; a sacrifice branch scores between 0.55 and 0.85 by what the lands are worth, and 0.25 when it would leave fewer than three lands (`eitherLandFloor`); both sit on either side of the decline's 0.5 as that rule needs. The MCP seat gets the branches from the enumerator's moves and from `boardtext`'s "(pay one of: …)" line.

### 6. How it pauses and resumes

Nothing new. Every link is a prompt with a closure frame the restore census already counts: `payUnlessResume` ("payUnlessResume") for §3 and §5, and `chooseCardsResume` for §2. A table with an offer open is not a restore point until it is answered, as for every ADR 0150 link. The rest of the resolution lives in the frame's continuation (`decline` → next link → `OnNobody`, or `pay` → `OnPaid`, and for §2 each leg's `then` → the next leg → the payment phase → `OnUnsaved`), so the chain resumes on each answer and on each departure. No new field holds a func, so `closure_fields.txt` gains no route and the `census:` and `transient` counts don't move. The slice that adds `AnyPlayerPaysForEach` checks this with `TestClosureFieldsReachableFromGame`: its closures are captured by `chooseCardsFrame.then`, which is already classified.

### 7. Legal moves, the heuristic bot and the MCP seat

**Enumerator (`legal/choices.go`).**
- Graveyard exile is a `PayAction` card pick, so the existing `e.combos(pool, Count, Count, …)` branch offers it. The pool is ordered `cheapestFuelFirst` (the seat's `CostFuelPricer`, as delve and escape use), so the payments that survive `MaxExpansionPerSource` are the cheapest ones. The label is "pay: exile — Card A, Card B" through `anyPlayerVerbs`.
- Either: one move per payable branch, plus the decline. A card branch adds combos as above (the sacrifice pool in `SacrificePaymentOrderForEffect` order) with `pay_option`. A names-nothing branch adds one move with `pay_option` and the branch's `MoveCost` (life, or the damage as a life price, as PR 2's `pay_damage` move does). Labels name the branch: "Worms of the Earth: take 5 damage", "Worms of the Earth: sacrifice two lands — Forest Island".
- Cleansing's `choose_cards` (carrying `AnyPlayerOffer`): `cardSetPickPool` orders the pool in three groups:
  1. the chooser's own lands that no earlier player chose, most valuable first;
  2. their own lands someone already chose;
  3. everyone else's.

  The moves are "pay for none" (`AlwaysLegal`), then the suggested set (group 1, cut to `Max`) through `withSuggestedFirst`, then the existing `filteredCombos` within budget (owner decision 6). Labels read "Cleansing: pay 3 life for Forest, Island, Plains". When a land in the set is already chosen by an earlier player, the label adds "(Alice also pays for Forest)".

**Heuristic (`aiseat/heuristic`).** ADR 0150 §6's rules hold: never pay against your own effect, pay an opponent's when it's worth it, and no free-riding. The new parts:
- **Graveyard exile** (Carrion): the bot pays only if the creature is in combat with it, meaning it attacks the bot or the bot's planeswalker or battle (`AttackingTarget`), or it blocks or is blocked by a creature the bot controls (owner decision 3). It scores the exiled cards with `fuelValue`, as dredge and devour do.
- **Either** (Worms): pay. Take the damage when it leaves the bot at or above `phyrexianLifeFloor`. Otherwise sacrifice the two least valuable lands (the existing sacrifice score). With neither acceptable, decline.
- **Cleansing:** the `choose_cards` branch reads `AnyPlayerOffer.EachLife` and `EarlierChoices`. It pays only for the bot's own lands that **no earlier player has chosen** (owner decision 6), most valuable first, as many as keep it at or above `phyrexianLifeFloor` with `EachLife` per land. It never pays for a land an earlier player already pays for, so it doesn't pay twice for nothing. It never pays for an opponent's land, whoever controls Cleansing. A later player may still choose the bot's land too. The bot can't see that coming, and paying twice is then that player's choice.
- **Combat:** `combatPower` (§4).

**MCP seat (`internal/mcpseat`).** The MCP seat answers from the numbered legal-move list and reads the board with `get_state`, so it gets every new shape through the enumerator's moves and labels, and nothing in `mcpseat` changes. Its tests gain one render case per shape (`energy_render_test.go` is the pattern). One limit is recorded: on Cleansing an agent can choose only among the enumerated sets (nothing, its own lands, and the budgeted combinations), not every subset of 40 lands.

### 8. Wire and client

- **Protocol additions** (all additive, `omitempty`): `AnyPlayerOfferView.EachLife` (`each_life`); `PayCardsView.Action` gains the value `"exile_graveyard"`; `PendingChoiceView.PayEither []PayOptionView` (`pay_either`); `CardView.AssignsNoCombatDamage` (`assigns_no_combat_damage`); the `resolve_choice` answer field `pay_option`. `docs/protocol.md` is updated in the slice that adds each one.
- **Client (`client/src`).**
  - The `pay_unless` picker (`ChoicePromptModal.svelte`, `payCardsOptions`) resolves `exile_graveyard` options from the chooser's graveyard as well as from the battlefield and hand.
  - An either payment draws one button per branch ("Sacrifice two lands" opens the existing card picker, "Take 5 damage" answers at once) plus "Don't". The offeree header (`anyPlayerOfferLine`, `payPhrase`) reads "sacrifice two lands or take 5 damage".
  - Cleansing: the `choose_cards` modal, when `any_player_offer` is set, shows:
    - The same whose-effect and who-is-next lines as a pay-unless link: "Bob's Cleansing. Each player in turn chooses lands to pay 1 life for; a land anyone pays for is saved. Then Carol."
    - **The earlier choices** from `earlier_choices`:
      - a badge on each land with every earlier payer's name ("Alice pays");
      - a summary line ("So far: Alice pays for 3 lands, Dan for none");
      - on a selected land that already has a payer, a warning ("Alice already pays for this: you'd both pay 1 life").
    - `choose_suggested` preselected.
    - The running price ("Pay 3 life for 3 lands").

    `anyPlayerOfferRequest` (`choiceDock.ts`) and `dockHint.ts` give every other seat the read-only line for a `choose_cards` leg too, with the same summary: "Carol is choosing lands to pay 1 life for (Bob's Cleansing). So far: Alice 3, Dan none." Nothing is paid until the last player has chosen, and the hint says so.
  - The status chip for `assigns_no_combat_damage`.
  - New player-facing names go through `L.<key>` in `client/src/lib/labels.ts` (AGENTS.md §5).

**As built (slice 5, #3573, 2026-10-15).**

- **The leg is answered in the sheet, not on the board.** A `choose_cards` over battlefield permanents is board-answered by default (`answeredOnBoard`, #2394). A per-land leg opts out: the badge, the warning and the running price have nowhere to sit on the board. "Show as list" is not needed for it.
- **Where the words live.** `choiceDock.ts` holds `isAnyPlayerEachLeg`, `anyPlayerEachHeader`, `anyPlayerEachSummary`, `anyPlayerEachPrice`, `anyPlayerEachBadge` and `anyPlayerEachWarning`. `anyPlayerOfferLine`, `anyPlayerWaiting` and `anyPlayerOfferRequest` switch to the per-land wording when the leg carries `each_life`, so the other seats' read-only line and `passHint` both include "So far: Alice 3, Dan none."
- **The noun.** The wire carries no noun, so the copy says "lands" when every option is a land, and otherwise "permanents". A seat watching has no options (they are the chooser's), so it reads the earlier picks off the board; before anyone has chosen it says "permanents".
- **No new contract labels.** The new elements are not tutorial anchors or e2e selectors, so `labels.ts` is unchanged.
- **The bot.** `aiseat/heuristic/any_player_each.go` scores each enumerated answer. An answer naming an opponent's land or a land an earlier player chose scores 0.25, below "pay for none" (0.5). One that would take the bot below `phyrexianLifeFloor` scores `phyrexianLifeDeclined`. Any other answer scores above 0.5, rising with the number and value of lands, so the bot pays for as many of its own unchosen lands as the floor allows, most valuable first.

### 9. Snapshot, restore and undo

- **Additive v8 only.** No `SnapshotSchemaVersion` bump. New persisted shape: `pendingChoices[].anyPlayerOffer.eachLife` (int), and the new mod kind `assignsNoCombatDamage` in `ScopedEffects` (a new vocabulary value: an older binary refuses a restore point naming it with `ErrUnknownEffectKey` and keeps the file, docs/snapshots.md). `PayAction.Either` and `PayActionExileFromGraveyard` live on the `payUnlessResume` frame, which is never serialised. `testdata/snapshot_shape/v8.txt` gains the one line, and no frozen fixture changes.
- **Restore:** as §6. A table with an offer open writes no restore point, as today.
- **Undo (ADR 0054 decision 4):** each answer is one commit, like an ADR 0150 link. No answer reveals hidden information (graveyards and the battlefield are public, and the life and damage payments are public), so no new undo barrier is needed.
- **Closure ratchet:** no new route (§6).

### 10. Roadmap registry

Each card slice removes its card from the `unless-any-player-pays` row's `Waiting` list and rewrites its `Missing` and `EngineNotes` to match what shipped. Slice 1 adds the `assigns-no-combat-damage` row (owner decision 4, #3531). After slice 6 the row's `Missing` names only Worms of the Earth (blocked on #1985), and the row turns Implemented with Worms' slice (8). Rhystic Cave and Leonin Arbiter stay recorded as out of scope, as ADR 0150 left them.

## Alternatives considered

1. **Cleansing as one pay-unless per land around the table** (ADR 0150 Q4 option c). It's faithful to the words, but it's up to 160 prompts at a four-seat table with 40 lands. Rejected, as ADR 0150 did.
2. **Cleansing batched with ADR 0150 decision 1's shortcut:** each player picks only among lands not yet saved and pays as they answer, so nobody pays twice for one land. It saves a little life. The owner chose the literal CR 101.4 reading instead (owner decision 1), so a second payment is made when two players choose the same land. Rejected.
3. **Cleansing with a `pay_amount` stepper per player.** A number can't say which lands are saved. Rejected.
4. **A new prompt kind for Cleansing.** `choose_cards`, through the ADR 0175 collector, already blocks the table, is answered by every client, has a departure decision and shows earlier public choices. A new kind would need all of that again. Rejected.
5. **The either payment as two chained prompts:** a `confirm` "which payment?", then a pay-unless. That's two prompts per player, and the chain's decline would have to cross both. Rejected for one prompt with `pay_option`.
6. **The either payment inferred from the answer's shape** (`card_ids` present means sacrifice, absent means damage). It's ambiguous once a branch is a names-nothing payment beside another names-nothing payment, and a malformed answer would silently pick a branch. Rejected for an explicit index.
7. **"Assigns no combat damage" as a prevention shield** (Gossamer Chains' `preventFromSource`). It's wrong under "damage can't be prevented" and visible to prevention triggers (CR 615). Rejected.
8. **Graveyard exile through `payExileCardsCostLocked` with `MustSettleNow`.** That needs CR 903.9 answers collected before the payment. That's right for a cast, but a resolution may pause, and the discard batch already does. Rejected for the pausable batch.

## Consequences

- Four cards become expressible. Cleansing, Carrion Rats and Carrion Wurm ship in this ADR's slices, and Worms of the Earth ships once #1985 does.
- `ModAssignsNoCombatDamage` is a general primitive. 24 more cards wait on it under the new `assigns-no-combat-damage` registry row and [#3531](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3531).
- On Cleansing, two players can both pay for one land. That's what owner decision 1 asked for. The client warns before it happens, and the bot avoids it.
- **The thing most likely to be got wrong later:** a per-object offer built from bare `choose_cards` prompts, outside the door and the ADR 0175 collector. It would lose the departure rule (`FromPlayer == Chooser`, so no reassignment), the epoch check, the earlier choices, and the offer data the bot and view read. Cleansing-like cards go through `QueueAnyPlayerPaysForEachForEffect`.
- The MCP seat can't name an arbitrary subset of lands on Cleansing (§7, owner decision 6).

## Delivery plan

Each slice is one PR into `develop`, independently shippable, in this order unless "Depends on" says otherwise. Every slice runs `scripts/go-docker.sh test ./...` and `scripts/go-docker.sh lint`. A card slice also generates only its own oracle fixtures (`-update-oracle -oracle-ids=…`) and declares `Completeness` truthfully (docs/adding-cards.md).

### Slice 1 — `feat(server): assigns no combat damage this turn` (tier:2-standard)

**Scope:**
- `game/scoped_effects.go`: the `ModAssignsNoCombatDamage` constant, its `readerRule` row and its doc comment.
- A new `game/assigns_no_combat_damage.go`: `AssignsNoCombatDamageThisTurnForEffect`, `permanentAssignsNoCombatDamageLocked` and `PermanentAssignsNoCombatDamageForEffect`.
- The reader in `assignAndDealCombatDamageLocked` (`game/mutations.go`).
- `protocol/view.go`: `CardView.AssignsNoCombatDamage`.
- `aiseat/heuristic`: `combatPower`, and its use in `combat.go`, `blockset.go` and `race.go`.
- The client chip and its label.
- A new registry row `assigns-no-combat-damage` in `roadmap/registry.go` (owner decision 4): `KindSeam`, `StatusPartial`, `Issue: 3531`, `Rules: 510.1, 510.1a`, `ADR` this file, and `Waiting` the 24 cards listed on #3531.

No card.
**Acceptance (tests):**
- An attacker with the mod, unblocked, deals no damage to the defending player, while one without it does (back-out: removing the reader fails it).
- Blocked by one creature, it marks no damage. Multi-blocked, no `damage_assignment` prompt is queued for it.
- With trample, lifelink or deathtouch it causes nothing. With `DamageByToughness` it still assigns 0.
- A blocker with the mod deals no damage to the attacker.
- Under a "damage can't be prevented" effect, it still assigns 0 (it is not prevention).
- The mod ends at cleanup. A flickered creature returns without it (pinned, CR 400.7).
- `CardView.AssignsNoCombatDamage` round-trips. A heuristic test shows the bot blocks a creature with the mod as a 0-power attacker.
- A snapshot written with the mod restores, and the shape drift test passes.
- The roadmap tests accept the new row, and `docs/engine-seams.md` is regenerated the way the registry's other rows are.
**Depends on:** nothing.

### Slice 2 — `feat(server): any-player offers paid by exiling from your graveyard` (tier:2-standard)

**Scope:**
- `game/pay_unless_action.go`: `PayActionExileFromGraveyard`, with its options, the payable count, and the pausable exile batch in `payActionLocked`.
- `protocol/view.go`: the new `PayCards.Action` value, with the options public.
- `legal/choices.go`: the `cheapestFuelFirst` pool for this kind.
- `cards/effects/any_player_pays.go`: `ExileFromGraveyardPayment`.
- The client's `payCardsOptions` graveyard lookup.
- `docs/protocol.md`.

No card.
**Acceptance (tests):**
- `QueueAnyPlayerPaysForEffect` with the payment: a player with fewer than `Count` cards in their graveyard gets no "pay" move, and a "yes" without picks is a decline that moves the offer on.
- Exactly `Count` distinct cards from the payer's own graveyard are accepted. Another player's card, a duplicate, or a wrong count is refused with the prompt kept.
- The cards go to exile and `OnPaid` runs after the last card lands.
- A commander among them is offered CR 903.9a and `OnPaid` still runs once.
- The options are visible to every seat in the view.
- The enumerator's first payment is the policy's cheapest fuel.
**Depends on:** nothing.

### Slice 3 — `feat(server): carrion rats and carrion wurm` (tier:2-standard)

**Scope:**
- `cards/effects/carrion_rats.go` and `cards/effects/carrion_wurm.go`: an attacks-or-blocks trigger whose `AnyPlayerPays{Pay: ExileFromGraveyardPayment(n), OnPaid: …}` applies `AssignsNoCombatDamageThisTurnForEffect` to the source if `sourceIsStillThisPermanent`.
- The heuristic's combat rule for a graveyard-exile offer (§7, owner decision 3).
- Card tests, and oracle fixtures for the two oracle IDs.
- The `unless-any-player-pays` row update (`Waiting`, `Missing`, `EngineNotes`), and the two cards named as the first users in the `assigns-no-combat-damage` row's notes.

**Acceptance (tests):**
- Rats attack, an opponent exiles a card, and the Rats deal no combat damage this turn. With nobody paying, they deal 2.
- The trigger fires on blocking too.
- If the Rats left and came back before the payment, the new object is unaffected.
- The Wurm needs three cards.
- The bot that controls the Rats never pays. A bot that is being attacked by them and has a card in its graveyard pays. A bot not in that combat declines.
- Both cards are `CompletenessFull`.
**Depends on:** slices 1 and 2.

### Slice 4 — `feat(server): one any-player offer for each object, paid in life` (tier:2-standard)

**Scope:**
- New `game/any_player_pays_each.go`: `AnyPlayerPaysForEach`, `QueueAnyPlayerPaysForEachForEffect`, the legs, the payment phase and `OnUnsaved`.
- `game/ordered_permanent_choices.go` and `chained_choice.go`: the three unexported leg fields of §2 (the offer, the suggested-unchosen-own flag, and the `Declined` update) copied onto each queued prompt.
- `AnyPlayerOffer.EachLife` (`any_player_pays.go`, `cloneAnyPlayerOffer`), and `testdata/snapshot_shape/v8.txt`.
- `protocol/view.go`:
  - `AnyPlayerOfferView.EachLife`, and `AnyPlayerOffer` populated on a `choose_cards` too.
  - `ChooseSuggested` populated for a `choose_cards` that carries `AnyPlayerOffer`. Today the view does this only for proliferate and `entry_choose_permanents`.
  - `EarlierChoices`, which is already populated from the frame.
- `legal/choices.go`: `cardSetPickPool` for an offer leg (§7), `withSuggestedFirst` for it as for proliferate, and the labels.
- `cards/effects`: `AnyPlayerPaysForEach`.
- `docs/protocol.md`.

No card, client or heuristic change beyond what keeps the build green.
**Acceptance (tests):**
- With four players, the legs run in APNAP order from the active player.
- Each leg offers every land still there, **including** ones earlier players chose. `Max` is `min(lands, life/EachLife)`.
- A player at 0 life, or one who can't lose life (CR 119.8), gets no leg and is logged as declining.
- No life is paid until the last leg is answered.
- Then each player pays once per land they chose, as separate 1-life payments: k payments make k life-loss events.
- Two players who choose the same land both pay, and the land is saved once.
- `EarlierChoices` on a later leg lists each earlier player's choices in APNAP order, an empty list included. `AnyPlayerOffer.Declined` lists those who chose nothing.
- `ChooseSuggested` is the chooser's own lands that no earlier player chose, capped at `Max`.
- `OnUnsaved` gets exactly the lands nobody paid for that are still the same objects, and runs once.
- A chooser who leaves mid-prompt is not replaced (no reassignment, since `FromPlayer == Chooser`): they pay nothing, and the next player is asked.
- A land whose controller leaves during the picks is gone from later legs and from the payments.
- The leg blocks the table, and a standing answer never answers it. The game is not restorable while a leg is open.
- `TestClosureFieldsReachableFromGame` is unchanged. The snapshot drift test includes `eachLife`.
**Depends on:** nothing.

### Slice 5 — `feat(client): show and answer a per-land any-player offer` (tier:2-standard)

**Scope:**
- `client/src/lib/components/board/ChoicePromptModal.svelte`, the `choose_cards` modal when `any_player_offer` is set:
  - the whose-effect and who-is-next header;
  - per-land badges from `earlier_choices` naming every earlier player paying for that land;
  - the summary line of earlier choices;
  - the "you'd both pay" warning on a badged land the viewer selects;
  - `choose_suggested` preselected;
  - the running price ("Pay 3 life for 3 lands").
- `client/src/lib/choiceDock.ts`: `anyPlayerOfferRequest` and `anyPlayerWaiting` for a `choose_cards` leg, so the other seats' read-only line includes the choices so far.
- `dockHint.ts`, the `protocol.ts` types and `labels.ts`.
- `aiseat/heuristic`: the `choose_cards` branch (§7) that pays for the bot's own lands no earlier player chose, down to `phyrexianLifeFloor`.

**Acceptance (tests):**
- `choiceDock.test.ts` / `dockHint.test.ts`: the chooser's header, and every other seat's read-only line with the choices so far.
- A render test (the pattern is `lilianaEarlierChoices.render.test.ts`):
  - a land chosen by Alice shows "Alice pays";
  - selecting it shows the double-payment warning;
  - the suggested lands are preselected;
  - the price updates as lands are toggled.
- The heuristic:
  - pays for its own most valuable lands that no earlier player chose;
  - never pays for a land an earlier player already pays for;
  - never pays for an opponent's land;
  - stops at the life floor.
**Depends on:** slice 4.

### Slice 6 — `feat(server): cleansing` (tier:1-mechanical)

**Scope:**
- `cards/effects/cleansing.go`: an `AnyPlayerPaysForEach` over every land on the battlefield, with `EachLife: 1`, and `OnUnsaved` destroying the unsaved lands with `DestroyPermanentsThenForEffect`.
- Card tests and the oracle fixture.
- The `unless-any-player-pays` registry row update.

**Acceptance (tests):**
- With no one paying, every land is destroyed at once (one dies batch).
- A land any player paid for survives. Each payer loses 1 life per land they chose.
- When two players chose the same land, both lose the life and the land survives.
- An indestructible land survives anyway.
- The card is `CompletenessFull`.
**Depends on:** slices 4 and 5.

### Slice 7 — `feat(server): either payment for any-player offers` (tier:2-standard)

**Scope:**
- `game/pay_unless_action.go`: `PayActionEither`, `PayAction.Either`, and payability and validation per branch.
- `game/pending_choice.go`: `resolvePayUnless` with an option, `ResolvePayUnlessWithOption`, and `answerPayUnlessLocked` paying the chosen branch.
- `actions/actions.go`: `pay_option` routing.
- `protocol`: `PendingChoiceView.PayEither`.
- `legal/choices.go`: one move per branch.
- `aiseat/heuristic`: take the damage above the floor, else sacrifice the cheapest two lands, else decline.
- The client's either buttons.
- `cards/effects`: `EitherPayment`.
- `docs/protocol.md`.

Engine tests use a test-only offer, since no catalog card uses it until slice 8.
**Acceptance (tests):**
- A `pay_option` out of range, a `pay_option` on a non-either prompt, and `apply: true` without `pay_option` are each refused with the prompt kept.
- The sacrifice branch with fewer than two lands is not offered, and the damage branch always is.
- Choosing damage that is then fully prevented still counts as paid (`OnPaid` runs, CR 118.12).
- Each branch's move is in the move list, and the bot picks damage above its floor and lands below it.
- The client sends `{apply: true, pay_option: 1}` for damage, and `{apply: true, pay_option: 0, card_ids: […]}` for the sacrifice.
**Depends on:** nothing (`PayActionDamage` already shipped in ADR 0150 PR 2).

### Slice 8 — `feat(server): worms of the earth` (tier:1-mechanical)

**Scope:**
- `cards/effects/worms_of_the_earth.go`:
  - the land-play prohibition (the existing static in `land_play_restriction.go`);
  - the #1985 "lands can't enter" static;
  - the upkeep trigger `AnyPlayerPays{Pay: EitherPayment(SacrificePayment(2, "land", "lands", QueryType("land")), DamagePayment(5)), OnPaid: destroy this enchantment}`.
- Card tests and the oracle fixture.
- The registry row's final update: the row turns Implemented, apart from the recorded out-of-scope cards.

**Acceptance (tests):**
- At each player's upkeep the offer goes round in APNAP order.
- Either payment destroys Worms, and with no payer Worms stays.
- Lands can't be played or put onto the battlefield.
- The card is `CompletenessFull`.
**Depends on:** slice 7 and #1985.

## Owner decisions (answered 2026-10-14)

The owner answered all six questions on 2026-10-14. The answers are binding and are written into the decisions above. The options not chosen are kept here.

1. **How Cleansing asks (§2).** **Decision: (b), literal CR 101.4.** Every player, in APNAP order, picks which lands they'll pay for, knowing the earlier picks. Then all the payments happen. Two players may pay for the same land, and each such payment is still made. A land is saved if anyone paid for it. Not chosen:
   - (a) The batched pick, where lands already saved aren't offered again and each player pays as they answer. This was the recommendation.
   - (c) One pay-unless per land around the table.
2. **Saving k lands (§2).** **Decision: (a), k separate 1-life payments**, as recommended. Not chosen: (b) one payment of k life.
3. **When a bot pays against Carrion Rats or Carrion Wurm (§7).** **Decision: (a)**, as recommended. It pays only when the creature is fighting it: attacking it, its planeswalker or battle, or blocking or blocked by its creature. Not chosen: (b) paying like any tax.
4. **"Assigns no combat damage" beyond the Carrion cards (§4).** **Decision: (a)**, as recommended. The general `ModAssignsNoCombatDamage` ships in slice 1, together with a registry row (`assigns-no-combat-damage`). The follow-up seam issue [#3531](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3531) lists the 24 cards. Not chosen:
   - (b) no row;
   - (c) adding those cards to this ADR now.
5. **Building the either payment before Worms can ship (§5).** **Decision: (a)**, as recommended. It's built now as slice 7, and Worms (slice 8) waits on #1985. Not chosen: (b) waiting for #1985.
6. **Cleansing's move list for bots and the MCP seat (§7).** **Decision: (a)**, as recommended: "pay for none", then the suggested set, then a capped set of other combinations. Under decision 1, the suggested set and the bot's pick leave out lands an earlier player already chose to pay for, so a bot doesn't pay twice for nothing. Not chosen: (b) every prefix of the seat's own lands as extra moves.
