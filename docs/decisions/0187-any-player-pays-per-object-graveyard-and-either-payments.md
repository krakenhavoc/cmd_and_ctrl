# ADR 0187 — Unless any player pays: one offer per land, graveyard exile, and an either payment

**Status:** Proposed · 2026-10-14 · Sprint: none (Backlog, parked on a dependency or a single card)
**Issues:** [#3205](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3205) (Cleansing: destroy each land unless any player pays 1 life) and [#3206](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3206) (any-player offers paid by graveyard exile or an either payment: Carrion Rats, Carrion Wurm, Worms of the Earth). Registry row `unless-any-player-pays` (Partial; tracked under #1903, S50 tracker #1784, found landing #1860, S53 tracker #1910).
**Builds on:** [ADR 0150](0150-other-players-pay-any-player-offers-and-assist.md) (the any-player offer, owner decisions 1–9, and its follow-ups for Cleansing and the two payments), [ADR 0108](0108-turn-scoped-effects-object-history-and-damage-shields.md) §2 and §5 (rules-gate scoped effects; non-mana pay-unless payments), [ADR 0175](0175-players-choose-before-simultaneous-sacrifices.md) (public choices one seat at a time, `EarlierChoices`), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers), [ADR 0054](0054-dice-rolls-and-coin-flips.md) decision 4 (undo never lets a seat see the future twice).
**Depends on (not this ADR):** [#1985](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1985), "Lands can't enter the battlefield", for Worms of the Earth only.
**Numbering:** 0187 was reserved for this seam. After `git fetch --all --prune`, `git ls-tree --name-only <ref> docs/decisions/` on every remote branch (`develop`, `main`, `cost-ledger`, `docs/issue-audit`, `feat/2417-incubate`, `feat/3343-autopass-reason-undo`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2950-darkness-replacement-continuation`, `fix/3159-3160-attachments`, `fix/caddy-reload-admin-off`, `wip/836-one-click-default`, and, at the re-check before pushing, `docs/0189-face-down-remainder`, `docs/2704-crew-and-saddle-adr`, `feat/rf-f-cast-during-resolution`, `feat/sideboard-outside-the-game`) showed 0185, 0186, 0189 and 0191 in use and 0187 free.

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

## Decisions (proposed)

### 1. Who is asked, and in what order: unchanged from ADR 0150

Every new shape below goes through the same chain: the players still in the game, in APNAP order from the **active** player (`votersFromLocked(activePlayerIDLocked())`), the controller included, with `Opponents` narrowing it. A player who has left is skipped, and one who leaves while asked declines (CR 800.4f). Owner decisions 1, 2 and 7 of ADR 0150 bind here too: it is APNAP, each link blocks its step, no standing answer applies, and a bot doesn't free-ride.

**When nobody pays, or nobody can be asked:** Cleansing destroys every land it could not save, Carrion Rats and Carrion Wurm deal their combat damage as usual, and Worms of the Earth stays. `OnNobody` runs once, at once if there is nobody to ask.

### 2. One offer per object: Cleansing (#3205)

"For each land" is one offer per land. ADR 0150 owner decision 4 rejected the literal shape (up to 160 prompts at a four-seat table) and kept the batched one on file. This ADR adopts it.

**A new door, `Game.QueueAnyPlayerPaysForEachForEffect(AnyPlayerPaysForEach)`**, in a new file `game/any_player_pays_each.go`:

```go
type AnyPlayerPaysForEach struct {
    Source, Controller uuid.UUID
    Objects  []uuid.UUID // the battlefield objects, frozen as the effect is applied (CR 608.2h)
    EachLife int         // life per object saved; the only payment this door takes (1 for Cleansing)
    Noun     string      // "land" / "lands", for the words
    Question string
    // OnUnsaved gets the objects nobody paid for, still the same objects
    // (epoch-checked), and runs exactly once after the last player.
    OnUnsaved func(g *Game, unsaved []uuid.UUID) error
}
```

- **One prompt per player, one at a time, in APNAP order.** Each player's link is a `PendingChoiceChooseCards` (the existing kind, so the gate, the enumerator, the dock, the departure table and the restore census all treat it as they already do), with:
  - `Cards`: the objects not yet saved, still on the battlefield as the same object (the epoch check `collectPermanentChoicesLocked` uses, `ObjectEpoch`).
  - `Min: 0`, `Max: min(len(Cards), affordable)`, where `affordable` is how many `EachLife` payments the player can make one after another: `life / EachLife` when `CanPayLifeLocked(p, EachLife)`, else 0 (CR 119.4 is per payment, CR 119.8). A player who can afford none is not asked. Their turn in the order is logged as a decline (`EventAnyPlayerDeclined`) and the offer moves on, because a prompt with only "none" to answer is a wasted click.
  - `Zone: ZoneBattlefield`, `FromPlayer: Chooser`. Setting `FromPlayer` to the chooser stops the departure table from reassigning the link to another seat (`departedChoiceObjectLocked` needs `FromPlayer != Chooser`). A link whose chooser leaves is dropped with `dropDefault`: its continuation runs with nothing picked, and the next player is asked (CR 800.4f).
  - `earlierPermanentChoices`: the saves so far, one `SeatCards` per earlier player, set the way the ADR 0175 collector sets it. The view's `EarlierChoices` and the existing client rendering show "Alice saved Forest, Island" with no new wire field.
  - `ChooseSuggested`: the chooser's own unsaved objects, in engine order, cut to `Max`. That is the answer most players want, and `withSuggestedFirst` puts it first in the move list (§7).
  - `AnyPlayerOffer`: the same plain data ADR 0150 put on a pay-unless link (`Controller`, `Order`, `Declined`), plus one additive field, **`EachLife int`** (`json:"eachLife,omitempty"`). Here `Declined` lists the players who saved nothing.
- **Paying.** The answer's `Then` pays `EachLife` life once per object named, as **separate payments** (one `payLifeAsCostLocked` per object), because the card prints a 1-life payment for each land and a "whenever you lose life" trigger should see each one ([question 2](#open-owner-questions)). If a payment fails (the life total changed under the prompt, which nothing in a resolution can do today), it stops and only the objects already paid for are saved. It logs one `EventAnyPlayerPaid` per payer with the label "N life (N lands)", then asks the next player.
- **Lands saved by an earlier player are not offered again.** That is ADR 0150 decision 1's recorded simplification, applied per object: CR 101.4 lets two players pay for the same land, and the second payment buys nothing.
- **After the last player**, `OnUnsaved` gets every object in `Objects` that nobody saved and that is still the same object. Cleansing destroys them in one `DestroyPermanentsThenForEffect` (CR 701.8a, simultaneous per CR 608.2e), so regeneration, indestructible and "dies" triggers apply as for Armageddon.
- **Lands that change during the offer.** Nothing resolves while the offer is open, so the set can only shrink through a departure (CR 800.4a takes that player's lands). The epoch check drops those objects from later links and from `OnUnsaved`.
- **Door scope.** `EachLife` is the only payment. A mana or card payment per object has no card, so the door refuses a zero `EachLife` with an `EventEffectError`, and `OnUnsaved` runs on every object (the card still does something, as `QueueAnyPlayerPaysForEffect` does with an unparseable cost).

Card side: `effects.AnyPlayerPaysForEach{Objects func(ctx) []uuid.UUID, EachLife, Noun, Question, OnUnsaved func(ctx, []uuid.UUID) error}`, bound to the resolving item as `AnyPlayerPays` is.

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

The other 24 cards that print "assigns no combat damage" are not in scope. The primitive is general, and [question 4](#open-owner-questions) asks whether they get their own registry row.

### 5. An either payment: Worms of the Earth (#3206)

"Any player may sacrifice two lands of their choice or have this enchantment deal 5 damage to that player. If a player does either, destroy this enchantment." It's one offer with two ways to pay, not two offers: "does either" is one condition.

- **`PayActionEither`** (`"either"`), and a new plain-data field on `PayAction`: **`Either []PayAction`** (two or more payments, none of them `PayActionEither`). It's a `PayAction` on the frame like any other, so it adds no closure and no snapshot field (`payUnlessResume` is already a census frame).
- **Payable** when any branch is: `payActionPayableLocked` for a branch that names nothing, and `len(payActionOptionsLocked) >= Count` for a card branch. Worms' damage branch is always payable, so Worms' offer always is.
- **The answer names the branch.** A new answer field, **`pay_option`** (a `PayOption *int` field, `json:"pay_option"`, on the `resolve_choice` params struct in `actions/actions.go`, beside `OptionalApply`), is an index into `Either`, sent with `apply: true` and, for a card branch, `card_ids`. `resolvePayUnless` validates it before dequeueing: the index must be in range, the branch is validated as that `PayAction` on its own (card pick or names-nothing), and a `pay_option` on a prompt that isn't `PayActionEither` is refused (`ErrInvalidParam`), as a stray `card_ids` is. `apply: true` on an either payment without `pay_option` is refused too, so no branch is ever guessed. A new entry point, `ResolvePayUnlessWithOption(choiceID, chooser, apply, option, cardIDs)`, is routed from `actions.go`.
- **Paying** runs `payActionLocked` with the chosen branch. `onPay` runs once that branch is paid. For damage, that is once the damage event has settled, however much landed (CR 118.12).
- **Wire:** `PendingChoiceView.PayEither []PayOptionView`, where each entry carries `{action, count, words, options}` (options only for a card branch, filtered as `PayCards` is: hand cards to the chooser only). `PayCost` carries the card's words. At most one of `PayCards` / `PayLife` / `PayDamage` / `PayNothing` / `PayEither` is present.
- Card side: `effects.EitherPayment(a, b UpkeepPayment) UpkeepPayment`.

### 6. How it pauses and resumes

Nothing new. Every link is a prompt with a closure frame the restore census already counts: `payUnlessResume` ("payUnlessResume") for §3 and §5, and `chooseCardsResume` for §2. A table with an offer open is not a restore point until it is answered, as for every ADR 0150 link. The rest of the resolution lives in the frame's continuation (`decline` → next link → `OnNobody`/`OnUnsaved`, or `pay` → `OnPaid`), so the chain resumes on each answer and on each departure. No new field holds a func, so `closure_fields.txt` gains no route and the `census:` and `transient` counts don't move. The slice that adds `AnyPlayerPaysForEach` checks this with `TestClosureFieldsReachableFromGame`: its closures are captured by `chooseCardsFrame.then`, which is already classified.

### 7. Legal moves, the heuristic bot and the MCP seat

**Enumerator (`legal/choices.go`).**
- Graveyard exile is a `PayAction` card pick, so the existing `e.combos(pool, Count, Count, …)` branch offers it. The pool is ordered `cheapestFuelFirst` (the seat's `CostFuelPricer`, as delve and escape use), so the payments that survive `MaxExpansionPerSource` are the cheapest ones. The label is "pay: exile — Card A, Card B" through `anyPlayerVerbs`.
- Either: one move per payable branch, plus the decline. A card branch adds combos as above (the sacrifice pool in `SacrificePaymentOrderForEffect` order) with `pay_option`. A names-nothing branch adds one move with `pay_option` and the branch's `MoveCost` (life, or the damage as a life price, as PR 2's `pay_damage` move does). Labels name the branch: "Worms of the Earth: take 5 damage", "Worms of the Earth: sacrifice two lands — Forest Island".
- Cleansing's `choose_cards` (carrying `AnyPlayerOffer`): `cardSetPickPool` puts the chooser's own objects first, most valuable first, then everyone else's. The moves are "save nothing" (`AlwaysLegal`), then the suggested set (the chooser's own, cut to `Max`) through `withSuggestedFirst`, then the existing `filteredCombos` within budget. Labels read "Cleansing: pay 3 life to save Forest, Island, Plains".

**Heuristic (`aiseat/heuristic`).** ADR 0150 §6's rules hold: never pay against your own effect, pay an opponent's when it's worth it, and no free-riding. The new parts:
- **Graveyard exile** (Carrion): the bot pays only if the creature is in combat with it, meaning it attacks the bot or the bot's planeswalker or battle (`AttackingTarget`), or it blocks or is blocked by a creature the bot controls ([question 3](#open-owner-questions)). It scores the exiled cards with `fuelValue`, as dredge and devour do.
- **Either** (Worms): pay. Take the damage when it leaves the bot at or above `phyrexianLifeFloor`. Otherwise sacrifice the two least valuable lands (the existing sacrifice score). With neither acceptable, decline.
- **Cleansing:** the `choose_cards` branch reads `AnyPlayerOffer.EachLife` and saves only the bot's own lands, most valuable first, as many as keep it at or above `phyrexianLifeFloor` with `EachLife` per land. It never pays for an opponent's land, whoever controls Cleansing.
- **Combat:** `combatPower` (§4).

**MCP seat (`internal/mcpseat`).** The MCP seat answers from the numbered legal-move list and reads the board with `get_state`, so it gets every new shape through the enumerator's moves and labels, and nothing in `mcpseat` changes. Its tests gain one render case per shape (`energy_render_test.go` is the pattern). One limit is recorded: on Cleansing an agent can choose only among the enumerated sets (nothing, its own lands, and the budgeted combinations), not every subset of 40 lands.

### 8. Wire and client

- **Protocol additions** (all additive, `omitempty`): `AnyPlayerOfferView.EachLife` (`each_life`); `PayCardsView.Action` gains the value `"exile_graveyard"`; `PendingChoiceView.PayEither []PayOptionView` (`pay_either`); `CardView.AssignsNoCombatDamage` (`assigns_no_combat_damage`); the `resolve_choice` answer field `pay_option`. `docs/protocol.md` is updated in the slice that adds each one.
- **Client (`client/src`).**
  - The `pay_unless` picker (`ChoicePromptModal.svelte`, `payCardsOptions`) resolves `exile_graveyard` options from the chooser's graveyard as well as from the battlefield and hand.
  - An either payment draws one button per branch ("Sacrifice two lands" opens the existing card picker, "Take 5 damage" answers at once) plus "Don't". The offeree header (`anyPlayerOfferLine`, `payPhrase`) reads "sacrifice two lands or take 5 damage".
  - Cleansing: the `choose_cards` modal, when `any_player_offer` is set, shows the same whose-effect and who-is-next lines as a pay-unless link ("Bob's Cleansing. Any player may pay 1 life for each land: Alice saved 2; then Carol."), preselects `choose_suggested`, and shows the running price ("Save 3 lands for 3 life"). `anyPlayerOfferRequest` (`choiceDock.ts`) and `dockHint.ts` give every other seat the read-only line for a `choose_cards` link too: "Carol is choosing lands to save for 1 life each (Bob's Cleansing)".
  - The status chip for `assigns_no_combat_damage`.
  - New player-facing names go through `L.<key>` in `client/src/lib/labels.ts` (AGENTS.md §5).

### 9. Snapshot, restore and undo

- **Additive v8 only.** No `SnapshotSchemaVersion` bump. New persisted shape: `pendingChoices[].anyPlayerOffer.eachLife` (int), and the new mod kind `assignsNoCombatDamage` in `ScopedEffects` (a new vocabulary value: an older binary refuses a restore point naming it with `ErrUnknownEffectKey` and keeps the file, docs/snapshots.md). `PayAction.Either` and `PayActionExileFromGraveyard` live on the `payUnlessResume` frame, which is never serialised. `testdata/snapshot_shape/v8.txt` gains the one line, and no frozen fixture changes.
- **Restore:** as §6. A table with an offer open writes no restore point, as today.
- **Undo (ADR 0054 decision 4):** each answer is one commit, like an ADR 0150 link. No answer reveals hidden information (graveyards and the battlefield are public, and the life and damage payments are public), so no new undo barrier is needed.
- **Closure ratchet:** no new route (§6).

### 10. Roadmap registry

Each card slice removes its card from the `unless-any-player-pays` row's `Waiting` list and rewrites its `Missing` and `EngineNotes` to match what shipped. After slice 6 the row's `Missing` names only Worms of the Earth (blocked on #1985), and the row turns Implemented with Worms' slice (8). Rhystic Cave and Leonin Arbiter stay recorded as out of scope, as ADR 0150 left them.

## Alternatives considered

1. **Cleansing as one pay-unless per land around the table** (ADR 0150 Q4 option c). It's faithful to the words, but it's up to 160 prompts at a four-seat table with 40 lands. Rejected, as ADR 0150 did.
2. **Cleansing as a literal CR 101.4 batch:** every player picks in APNAP order, then all pay at once, so two players may pay for the same land. It's closer to the letter, but the second payment buys nothing and costs the payer life. Rejected for the reason ADR 0150 decision 1 gives ([question 1](#open-owner-questions) keeps it on file).
3. **Cleansing with a `pay_amount` stepper per player.** A number can't say which lands are saved. Rejected.
4. **A new prompt kind for Cleansing.** `choose_cards` already blocks the table, is answered by every client, has a departure decision and shows earlier public choices (ADR 0175). A new kind would need all of that again. Rejected.
5. **The either payment as two chained prompts:** a `confirm` "which payment?", then a pay-unless. That's two prompts per player, and the chain's decline would have to cross both. Rejected for one prompt with `pay_option`.
6. **The either payment inferred from the answer's shape** (`card_ids` present means sacrifice, absent means damage). It's ambiguous once a branch is a names-nothing payment beside another names-nothing payment, and a malformed answer would silently pick a branch. Rejected for an explicit index.
7. **"Assigns no combat damage" as a prevention shield** (Gossamer Chains' `preventFromSource`). It's wrong under "damage can't be prevented" and visible to prevention triggers (CR 615). Rejected.
8. **Graveyard exile through `payExileCardsCostLocked` with `MustSettleNow`.** That needs CR 903.9 answers collected before the payment. That's right for a cast, but a resolution may pause, and the discard batch already does. Rejected for the pausable batch.

## Consequences

- Four cards become expressible. Cleansing, Carrion Rats and Carrion Wurm ship in this ADR's slices, and Worms of the Earth ships once #1985 does.
- `ModAssignsNoCombatDamage` is a general primitive that 24 more cards could use.
- **The thing most likely to be got wrong later:** a per-object offer built from bare `choose_cards` prompts outside the door would lose the departure rule (`FromPlayer == Chooser`, so no reassignment), the epoch check, and the offer data the bot and view read. Cleansing-like cards go through `QueueAnyPlayerPaysForEachForEffect`.
- The MCP seat can't name an arbitrary subset of lands on Cleansing (§7).

## Delivery plan

Each slice is one PR into `develop`, independently shippable, in this order unless "Depends on" says otherwise. Every slice runs `scripts/go-docker.sh test ./...` and `scripts/go-docker.sh lint`. A card slice also generates only its own oracle fixtures (`-update-oracle -oracle-ids=…`) and declares `Completeness` truthfully (docs/adding-cards.md).

### Slice 1 — `feat(server): assigns no combat damage this turn` (tier:2-standard)

**Scope:** `game/scoped_effects.go` (the `ModAssignsNoCombatDamage` constant, its `readerRule` row, and its doc comment), a new `game/assigns_no_combat_damage.go` (`AssignsNoCombatDamageThisTurnForEffect`, `permanentAssignsNoCombatDamageLocked`, `PermanentAssignsNoCombatDamageForEffect`), the reader in `assignAndDealCombatDamageLocked` (`game/mutations.go`), `protocol/view.go` (`CardView.AssignsNoCombatDamage`), `aiseat/heuristic` (`combatPower` and its use in `combat.go`, `blockset.go`, `race.go`), the client chip and its label. No card.
**Acceptance (tests):**
- An attacker with the mod, unblocked, deals no damage to the defending player, while one without it does (back-out: removing the reader fails it).
- Blocked by one creature, it marks no damage. Multi-blocked, no `damage_assignment` prompt is queued for it.
- With trample, lifelink or deathtouch it causes nothing. With `DamageByToughness` it still assigns 0.
- A blocker with the mod deals no damage to the attacker.
- Under a "damage can't be prevented" effect, it still assigns 0 (it is not prevention).
- The mod ends at cleanup. A flickered creature returns without it (pinned, CR 400.7).
- `CardView.AssignsNoCombatDamage` round-trips. A heuristic test shows the bot blocks a creature with the mod as a 0-power attacker.
- A snapshot written with the mod restores, and the shape drift test passes.
**Depends on:** nothing.

### Slice 2 — `feat(server): any-player offers paid by exiling from your graveyard` (tier:2-standard)

**Scope:** `game/pay_unless_action.go` (`PayActionExileFromGraveyard`: options, payable count, the pausable exile batch in `payActionLocked`), `protocol/view.go` (`PayCards.Action` value, public options), `legal/choices.go` (the `cheapestFuelFirst` pool for this kind), `cards/effects/any_player_pays.go` (`ExileFromGraveyardPayment`), the client `payCardsOptions` graveyard lookup, and `docs/protocol.md`. No card.
**Acceptance (tests):**
- `QueueAnyPlayerPaysForEffect` with the payment: a player with fewer than `Count` cards in their graveyard gets no "pay" move, and a "yes" without picks is a decline that moves the offer on.
- Exactly `Count` distinct cards from the payer's own graveyard are accepted. Another player's card, a duplicate, or a wrong count is refused with the prompt kept.
- The cards go to exile and `OnPaid` runs after the last card lands.
- A commander among them is offered CR 903.9a and `OnPaid` still runs once.
- The options are visible to every seat in the view.
- The enumerator's first payment is the policy's cheapest fuel.
**Depends on:** nothing.

### Slice 3 — `feat(server): carrion rats and carrion wurm` (tier:2-standard)

**Scope:** `cards/effects/carrion_rats.go`, `cards/effects/carrion_wurm.go` (an attacks-or-blocks trigger whose `AnyPlayerPays{Pay: ExileFromGraveyardPayment(n), OnPaid: …}` applies `AssignsNoCombatDamageThisTurnForEffect` to the source if `sourceIsStillThisPermanent`), the heuristic's combat-involvement rule for a graveyard-exile offer (§7), card tests, oracle fixtures for the two oracle IDs, and the registry row update (`Waiting`, `Missing`, `EngineNotes`).
**Acceptance (tests):**
- Rats attack, an opponent exiles a card, and the Rats deal no combat damage this turn. With nobody paying, they deal 2.
- The trigger fires on blocking too.
- If the Rats left and came back before the payment, the new object is unaffected.
- The Wurm needs three cards.
- The bot that controls the Rats never pays. A bot that is being attacked by them and has a card in its graveyard pays. A bot not in that combat declines.
- Both cards are `CompletenessFull`.
**Depends on:** slices 1 and 2.

### Slice 4 — `feat(server): one any-player offer for each object, paid in life` (tier:2-standard)

**Scope:** new `game/any_player_pays_each.go` (`AnyPlayerPaysForEach`, `QueueAnyPlayerPaysForEachForEffect`, the per-player link builder, the payment, `OnUnsaved`), `AnyPlayerOffer.EachLife` (`any_player_pays.go`, `cloneAnyPlayerOffer`), `testdata/snapshot_shape/v8.txt`, `protocol/view.go` (`AnyPlayerOfferView.EachLife`, populated on a `choose_cards` too; `ChooseSuggested` populated for a `choose_cards` that carries `AnyPlayerOffer`, which today the view does only for proliferate and `entry_choose_permanents`), `legal/choices.go` (`cardSetPickPool` for an offer link, `withSuggestedFirst` for it as for proliferate, labels), `cards/effects` (`AnyPlayerPaysForEach`), `docs/protocol.md`. No card, client or heuristic change beyond what keeps the build green.
**Acceptance (tests):**
- With four players, the order is APNAP from the active player.
- Each link's `Cards` excludes objects saved earlier, and `Max` is `min(unsaved, life/EachLife)`. A player at 0 life, or one who can't lose life (CR 119.8), is not asked and is logged as declining.
- Saving k objects makes k separate 1-life payments (k life-loss events).
- `EarlierChoices` on a later link lists the earlier saves.
- `OnUnsaved` gets exactly the unsaved objects that are still the same objects, and runs once.
- A chooser who leaves mid-prompt is not replaced (no reassignment, `FromPlayer == Chooser`): the next player is asked.
- The link blocks the table, and a standing answer never answers it.
- The game is not restorable while a link is open.
- `TestClosureFieldsReachableFromGame` is unchanged. The snapshot drift test includes `eachLife`.
**Depends on:** nothing.

### Slice 5 — `feat(client): show and answer a per-land any-player offer` (tier:2-standard)

**Scope:** `client/src/lib/components/board/ChoicePromptModal.svelte` (the `choose_cards` modal with `any_player_offer`: header lines, preselected `choose_suggested`, running life price), `client/src/lib/choiceDock.ts` (`anyPlayerOfferRequest` and `anyPlayerWaiting` for a `choose_cards` link), `dockHint.ts`, `protocol.ts` types, `labels.ts`, and `aiseat/heuristic` (the `choose_cards` branch that saves the bot's own lands down to `phyrexianLifeFloor`).
**Acceptance (tests):**
- `choiceDock.test.ts` / `dockHint.test.ts`: the chooser's header and every other seat's read-only line.
- A render test: the suggested lands are preselected, and the price updates as lands are toggled.
- The heuristic saves its own most valuable lands, never an opponent's land, and stops at the life floor.
**Depends on:** slice 4.

### Slice 6 — `feat(server): cleansing` (tier:1-mechanical)

**Scope:** `cards/effects/cleansing.go` (an `AnyPlayerPaysForEach` over every land on the battlefield, `EachLife: 1`, `OnUnsaved` destroying them with `DestroyPermanentsThenForEffect`), card tests, the oracle fixture, and the registry row update.
**Acceptance (tests):**
- With no one paying, every land is destroyed at once (one dies batch).
- A land saved by any player survives, and the payer loses 1 life per land.
- An indestructible land survives anyway.
- The card is `CompletenessFull`.
**Depends on:** slices 4 and 5.

### Slice 7 — `feat(server): either payment for any-player offers` (tier:2-standard)

**Scope:** `game/pay_unless_action.go` (`PayActionEither`, `PayAction.Either`, payability and validation per branch), `game/pending_choice.go` (`resolvePayUnless` with an option, `ResolvePayUnlessWithOption`, `answerPayUnlessLocked` paying the chosen branch), `actions/actions.go` (`pay_option` routing), `protocol` (`PendingChoiceView.PayEither`), `legal/choices.go` (one move per branch), `aiseat/heuristic` (damage above the floor, else the cheapest two lands, else decline), the client either buttons, `cards/effects` (`EitherPayment`), and `docs/protocol.md`. Engine tests use a test-only offer, since no catalog card uses it yet.
**Acceptance (tests):**
- A `pay_option` out of range, a `pay_option` on a non-either prompt, and `apply: true` without `pay_option` are each refused with the prompt kept.
- The sacrifice branch with fewer than two lands is not offered, and the damage branch always is.
- Choosing damage that is then fully prevented still counts as paid (`OnPaid` runs, CR 118.12).
- Each branch's move is in the move list, and the bot picks damage above its floor and lands below it.
- The client sends `{apply: true, pay_option: 1}` for damage, and `{apply: true, pay_option: 0, card_ids: […]}` for the sacrifice.
**Depends on:** nothing (`PayActionDamage` already shipped in ADR 0150 PR 2).

### Slice 8 — `feat(server): worms of the earth` (tier:1-mechanical)

**Scope:** `cards/effects/worms_of_the_earth.go`: the land-play prohibition (`land_play_restriction.go`'s existing static), the #1985 "lands can't enter" static, and the upkeep trigger `AnyPlayerPays{Pay: EitherPayment(SacrificePayment(2, "land", "lands", QueryType("land")), DamagePayment(5)), OnPaid: destroy this enchantment}`. Card tests, the oracle fixture, and the registry row's final update (the row turns Implemented apart from the recorded out-of-scope cards).
**Acceptance (tests):**
- At each player's upkeep the offer goes round in APNAP order.
- Either payment destroys Worms, and with no payer Worms stays.
- Lands can't be played or put onto the battlefield.
- The card is `CompletenessFull`.
**Depends on:** slice 7 and #1985.

## Open owner questions

1. **Cleansing's shape (§2).**
   - (a) **Recommended:** one `choose_cards` per player in APNAP order. Each player picks any of the lands not yet saved and pays 1 life each as they answer. Unsaved lands are destroyed together at the end. This is ADR 0150 Q4 option (b), and it keeps ADR 0150 decision 1's "no second payment" simplification.
   - (b) Literal CR 101.4: every player picks first, then everyone pays at once, so two players may pay for the same land.
   - (c) One pay-unless per land around the table (rejected by ADR 0150 owner decision 4 as unplayable).
2. **Saving k lands: k payments of 1 life, or one payment of k life (§2)?**
   - (a) **Recommended:** k separate payments. The card prints a 1-life payment per land, and "whenever you lose life" triggers see each one.
   - (b) One payment of k life. That's one event, which is simpler to log, but it under-triggers life-loss watchers.
3. **When does a bot pay against Carrion Rats or Carrion Wurm (§7)?**
   - (a) **Recommended:** only when the creature is in combat with it (attacking it, its planeswalker or battle, or blocking or blocked by its creature). Otherwise its graveyard buys it nothing.
   - (b) Like any tax (ADR 0150's default): any opponent asked first pays, even one the creature isn't fighting. It spends that bot's graveyard for another player's benefit.
4. **"Assigns no combat damage this turn" beyond the Carrion cards (§4)?**
   - (a) **Recommended:** ship it as the general `ModAssignsNoCombatDamage` with slice 1, and file a follow-up seam issue plus a registry row (`assigns-no-combat-damage`) listing the 24 other Commander-legal cards that print it (Ophidian, Laccolith Rig, Goblin Vandal …) for card work.
   - (b) Ship it, but add no row. The other cards are found when someone asks for them.
   - (c) Add the other cards to this ADR's delivery now. That's more scope for one seam.
5. **Build the either payment before Worms of the Earth can ship (§5)?**
   - (a) **Recommended:** yes, as slice 7, so that once #1985 lands Worms is a one-file card (slice 8). The slice is small and is tested against a test-only offer.
   - (b) No. Defer slice 7 until #1985 ships, so no code lands without a catalog consumer.
6. **Cleansing's move list for bots and the MCP seat (§7)?**
   - (a) **Recommended:** "save nothing", the suggested set (your own lands, best first, capped by life), then budgeted combinations. That keeps the list short. An MCP agent can't pick an arbitrary subset.
   - (b) Add every prefix of your own lands (k = 1…cap) as separate moves, so a seat can save "my best k" for any k. This adds up to 40 moves per prompt.
