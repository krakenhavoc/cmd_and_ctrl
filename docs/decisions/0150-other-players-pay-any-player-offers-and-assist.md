# ADR 0150 — Other players pay: "unless any player pays" and assist

**Status:** Accepted (owner decisions 2026-10-13) · 2026-10-13 · S53 — Faces, zones, and shared payment (ADR first), tracker [#1910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1910)
**Issues:** [#1903](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1903) (seam: unless any player pays — Rhystic Circle) and [#1857](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1857) (seam: assist — another player helps pay the generic mana — Fan Favorite). Registry rows: `unless-any-player-pays` and `assist`, both open; this ADR does not touch them (they close with the code).
**Owner decisions:** the owner answered all nine questions on 2026-10-13, each with the recommended option (a). The answers are recorded under [Questions for the owner (answered 2026-10-13)](#questions-for-the-owner-answered-2026-10-13) and are binding. The options not chosen are kept there.
**Numbering:** assigned by the S53 orchestrator. `origin/develop` at `9a9372749` ends at **0149**, so this is **0150**.
**Builds on:** [ADR 0018](0018-triggers-on-the-stack.md) §6 (the pay-unless prompt, and the latitude that lets a table walk past it), [ADR 0013](0013-replacement-effects.md) §5af and #1397 (a cost announcement parked on another player's answer), [ADR 0060](0060-leaving-the-game.md) (the departure table), [ADR 0108](0108-turn-scoped-effects-object-history-and-damage-shields.md) §5 (non-mana pay-unless payments), [ADR 0111](0111-action-dock.md) (the dock), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic and `Purpose`), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers), [ADR 0129](0129-energy-getting-and-paying-it.md) §3 (the `pay_amount` prompt), [ADR 0146](0146-voting.md) (one prompt per player in turn order, and a read-only view of it for everyone else).

This ADR was written plan-first. No code changed with it. The changes land in the PRs under [Delivery](#delivery). Every claim about the code below was checked on `origin/develop` at `9a9372749`. Every rule was checked against the Comprehensive Rules effective 2026-09-25, and the assist rulings against Scryfall's rulings for Fan Favorite (2018-06-08).

---

## Context

Two seams with one gap: **a payment made by a player other than the one the engine already asks.**

- **#1903.** "Unless any player pays {2}" and "Any player may pay {1}. If no one does, …" offer one payment to the whole table. The engine's pay-unless prompt asks one named player.
- **#1857.** Assist lets the caster pick another player, who may pay some of the spell's generic mana. The cast path has one payer.

They are one group because they need the same three things: a prompt addressed to a player who is not acting, a payment from that player's own mana (or life, or cards) that pays for an effect someone else controls, and a way for the rest of the table to see who is being asked. They are not one mechanism. One happens while an effect resolves and the other while a spell is cast, so the engine work is two pieces with a shared client and bot pattern.

### The rules

From the Comprehensive Rules effective 2026-09-25. Short quotes are exact; everything else is a summary.

- **CR 118.12** covers "[A player] may [do something]. If [that player] [does, doesn't, or can't], [effect]": "The action [do something] is a cost, paid when the spell or ability resolves." **CR 118.12a**: "[Do something] unless [a player does something else]" "means the same thing as '[A player may do something else]. If [that player doesn't], [do something].'" So an "unless" payment follows the cost rules: **CR 118.3** (a player can't pay a cost without the resources to pay it fully) and **CR 119.4** (a payment of more than 0 life is allowed only if the life total is at least the amount). **CR 119.8**: a player who can't lose life can't pay a cost that involves paying life.
- **CR 101.4** (APNAP): when several players would make choices at the same time, the active player chooses first, then each nonactive player in turn order, and "Then the actions happen simultaneously." Read literally, every player decides in turn and everyone who said yes then pays. Decision 1 stops at the first payer instead (see its note). Neither Scryfall nor this repo's rulings data has a ruling on the Rhystic cards or Browbeat that says otherwise.
- **"Any player" includes the controller.** "Any opponent" (Innocent Traveler, Distant Memories) does not.
- **CR 702.132a (assist):** "a static ability that modifies the rules of paying for the spell with assist (see rules 601.2g-h)." If the spell's total cost has a generic mana part, then before mana abilities are activated during the cast, the caster may choose another player. That player gets the chance to activate mana abilities, then the caster does. "Before you begin to pay the total cost of the spell, the player you chose may pay for any amount of the generic mana in the spell's total cost." The reminder text ("up to {3}") is the printed generic. The rule's cap is the generic in the **total** cost (CR 601.2f: mana or alternative cost, plus additional costs and increases, minus reductions), so commander tax and Thalia-style increases raise it and reductions lower it. Fan Favorite's rulings (2018-06-08) say the same, and add three things: targets are chosen before the helper is, only generic mana can be assisted, and any colored mana in the cost is paid by the caster.
- **CR 601.2g**: mana abilities are activated before costs are paid. **CR 601.2h**: the player pays the total cost, and "Partial payments are not allowed."
- **CR 605.3a**: a mana ability can be activated while a spell is being cast, or in the middle of resolving a spell or ability. **CR 605.3b**: it doesn't go on the stack and "resolves immediately after it is activated." That's why Rhystic Cave is out of scope (see [Out of scope](#out-of-scope)).
- **CR 800.4f**: "If an object requires a player who has left the game to pay a cost or choose whether to pay a cost, that cost is not paid." A departed player's offer is a decline, and a departed helper paid nothing.
- **CR 106.4**: a mana pool empties at the end of each step and phase.

### What the engine has

**The pay-unless prompt.** `PendingChoicePayUnless` (`server/internal/game/pending_choice.go`) is one yes/no question to one `Chooser`, carrying a server-only `payUnlessFrame` (`cost ParsedCost`, `onDecline`, `onPay`, the waterbend `tap` clause, and a non-mana `action *PayAction`). `queuePayUnlessLocked` is the one body behind every door: `QueuePayUnlessForEffect` (Rhystic Study, Smothering Tithe), `queueMayPayLocked` ("you may pay. If you do"), `QueueCounterUnlessPaidForEffect` (`counter_unless_paid.go`, which guards a stack object), `QueueUpkeepPayUnlessForEffect` (`upkeep_pay_unless.go`, owed in a step), and `QueueEnergyPaymentForEffect` (`energy_payment.go`). `resolvePayUnless` validates an answer (taps, cards, Phyrexian life) before the prompt leaves the queue. `answerPayUnlessLocked` pays from the chooser's pool, then their auto-tapper (`payCostWithLifeLocked`), or makes the `PayAction` (`payActionLocked`: discard, sacrifice, energy), then runs exactly one branch. On the card side, `effects.PayUnless` and `effects.MayPay` wrap it.

**The chooser already doesn't have to be the controller.** Rhystic Study asks the opponent who cast the spell, and the payment comes out of that opponent's pool. The inline dock prompt (`client/src/lib/choiceDock.ts`, `case "pay_unless"`), the legal-move enumerator (`server/internal/legal/choices.go`) and the heuristic (`aiseat/heuristic/choices.go`, and `tax.go`'s `taxAgainstPlan`) all answer it for whoever the chooser is. **What's missing is the fan-out:** nothing offers the same payment to the next player when the first one declines, and nothing tells the table who is being offered what.

**The vote.** ADR 0146 built exactly that fan-out for votes: `StartVoteForEffect` lists the players still in the game in turn order (`votersFromLocked`, `council_vote.go`), asks one ordinary option pick at a time, carries the vote so far as plain data on the open prompt (`PendingChoice.CouncilVote`), and every other seat sees it read-only in the dock (`councilVoteRequest`).

**A cast parked on another player.** #1397 / ADR 0013 §5af (`cost_commander_choice.go`): `castSpellLocked` validates the whole announcement, and if a commander's owner must answer CR 903.9 first, it **parks** the announcement on a prompt to that owner. Nothing has been paid and nothing has moved. The `costCommanderFrame.announce` closure re-runs `castSpellLocked` with the answer attached, and the re-run re-validates from scratch. It's the only place a cast waits on another player, and it's the shape assist needs.

**A chosen amount.** `PendingChoicePayAmount` (`energy_payment.go`, widened by `choose_number.go`) asks one player for a number between `Min` and `Max`, with `PayAmountPrompt.Resource` saying what each point costs (`energy`, `life`, `none`). The dock draws it as a stepper. Its continuation is `chooseValueResume`.

**Choosing a player at cast time.** Gift (ADR 0089) names an opponent on `cast_spell` (`CastSpellParams.GiftOpponent`, wire `gift_opponent`, offered by `OptionalCostView.ChoosesOpponent` / `OpponentOptions` in `protocol/view.go`, picked in `client/src/lib/targeting.ts`). Assist's "choose another player" is the same kind of pick.

### The cards

Counted from `data/scryfall/default-cards.json` (Commander-legal, one per oracle ID) on 2026-10-13.

**"Unless any player pays" / "any player may pay": 13 cards, matching #1903's count. None is catalogued.**

| Card | Payment | When | Fits |
|---|---|---|---|
| Rhystic Circle | {1} | activated ability resolves | PR 1 |
| Rhystic Shield | {2} | instant resolves | PR 1 |
| Rhystic Tutor | {2} | sorcery resolves ("Unless any player pays {2}, search …") | PR 1 |
| Wild Might | {2} | instant resolves | PR 1 |
| Soul Strings | {X} | sorcery resolves (X from the spell) | PR 1 |
| Nakaya Shade | {2} | activated ability resolves | PR 1 |
| Icy Prison | {3} | upkeep trigger ("sacrifice … unless any player pays {3}") | PR 1 |
| Aether Rift | 5 life | upkeep trigger | PR 2 |
| Dash Hopes | 5 life | cast trigger; if a player pays, counter Dash Hopes | PR 2 |
| Temporal Extortion | half their life, rounded up | cast trigger; if a player pays, counter it | PR 2 |
| Cleansing | 1 life, once per land | sorcery resolves | Q4 |
| Rhystic Cave | {1} | **mana ability** | out (CR 605.3a-b) |
| Leonin Arbiter | {2} | a static's special action, not an offer | out |

**The same offer with a non-mana payment ("any player may [do X]. If a player does / If no one does"): 14 more cards, none catalogued.** Book Burning, Breaking Point, Browbeat, Barbarian Bully (have this deal N damage to them); Argothian Wurm, Shivan Wumpus (sacrifice a land); Brain Gorgers (sacrifice a creature); Prowling Pangolin (sacrifice two creatures); Phantasmagorian (discard three cards); Carrion Rats, Carrion Wurm (exile cards from their graveyard); Innocent Traveler (any opponent sacrifices a creature); Worms of the Earth (sacrifice two lands *or* take 5 damage); Distant Memories (any opponent may let you have the card, which costs nothing). They use the same fan-out with a different payment. Which of them come in this ADR's PRs is [question 3](#questions-for-the-owner-answered-2026-10-13).

**Assist: 16 cards.** Fan Favorite is catalogued with a caveat ("Assist isn't implemented …", `cards/effects/fan_favorite.go`). The other 15 aren't catalogued: Bloodborn Scoundrels, Bring Down, Charging Binox, Dwarven Lightsmith, Game Plan, Gang Up, Huddle Up, Lava-Field Overlord, Magma Hellion, Out of Bounds, Play of the Game, Skystreamer, Spellweaver Duo, The Crowd Goes Wild, Vampire Charmseeker. Gang Up and The Crowd Goes Wild have "up to {X}" (X chosen by the caster), which the total-cost cap already covers. Most of these also need nothing else: their other text is a vanilla body, an ETB the catalog already writes, or a spell effect with an existing primitive. Support (The Crowd Goes Wild) and "your team" (Dwarven Lightsmith: with no teams, that is just you) are card work.

**In all: 43 Commander-legal cards** (13 + 14 + 16) wait on this ADR, 42 of them not catalogued (Fan Favorite is the one that is).

## Decisions (proposed)

### 1. "Any player pays" is a chain of ordinary pay-unless prompts

One new door, `Game.QueueAnyPlayerPaysForEffect(AnyPlayerPays)`, and one card primitive, `effects.AnyPlayerPays`:

```go
type AnyPlayerPays struct {
    Source     uuid.UUID
    Controller uuid.UUID     // whose effect it is; read by the view and the bot
    Cost       string        // mana, as ParseCost reads it; X resolved by the card
    Action     *PayAction    // or a non-mana payment (decision 3); exactly one of the two
    Question   string        // dialog header
    Opponents  bool          // "any opponent" (Innocent Traveler, Distant Memories)
    OnPaid     func(g *Game, payer uuid.UUID) error // "If a player does, …"
    OnNobody   func(g *Game) error                  // "unless …" / "If no one does, …"
    InThisStep bool          // decision 2
    Guards     uuid.UUID     // a stack object the outcome decides (Dash Hopes), or Nil
}
```

The engine lists the players still in the game in APNAP order from the active player (CR 101.4; a `votersFromLocked`-style helper that starts at `activePlayerIDLocked`, minus the controller when `Opponents` is set). It then queues **one `PendingChoicePayUnless` to the first** of them through `queuePayUnlessLocked`, with its frame's `onPay` set to `OnPaid(payer)` and its `onDecline` set to "queue the same offer to the next player still in the game, or run `OnNobody` when nobody is left".

- **One prompt at a time, each an ordinary pay-unless.** The prompt kind, its validator, the payment path, `legal/choices.go`'s moves, the inline dock prompt and its keys, the heuristic and the departure table's `dropDecline` all answer it unchanged. A player who leaves while asked declines (`declineDepartedChoiceLocked`), so the offer moves on (CR 800.4f). Players who left before their turn are skipped when the next prompt is built.
- **The first payer ends the offer.** Paying runs `OnPaid` with that player and nobody after them is asked. An unfundable "Pay" is a decline (unchanged) and moves the offer on.
- **Where this differs from a literal CR 101.4.** CR 101.4 has every player choose in APNAP order and then "the actions happen simultaneously". So a later player could also choose to pay after an earlier one already had, and both would pay. On every card in [the list](#the-cards), a second payment buys nothing: the effect is already stopped, or the spell already countered. It only costs the second payer something. The engine skips that extra payment and pays each answer as it is given, so the later players see the earlier payment, as they would at a paper table. It's a recorded simplification. If a card turns up where a second payment matters, it should be asked in full instead.
- **The offer is plain data on the open prompt:** `PendingChoice.AnyPlayerOffer *AnyPlayerOffer{Controller, Order []uuid.UUID, Declined []uuid.UUID, Opponents bool}`. That gives the view (decision 5) and the bot (decision 6) the whole picture without a new kind. The branches stay closures on `payUnlessFrame`, as every pay-unless's do, so a table waiting on an offer is not a restore point and the snapshot census already says so (`snapshot.go`, `"payUnlessResume"`). The new field is additive (`pendingChoices[].anyPlayerOffer` in the shape file).
- **Card helpers.** `effects.AnyPlayerPays{Cost, Question, OnNobody}` covers the "unless any player pays" wording and `{…, OnPaid}` covers "if a player does". Both are bound to the stack item the way `PayUnless` and `MayPay` are (`NewContext(g, item)` in the closures), so `OnNobody` reads the resolving item's targets and X (Soul Strings' {X}, Wild Might's target).

### 2. It blocks the table, in the step it was asked in

Rhystic Study's prompt is the one kind a table may walk past (ADR 0018 §6; `choice_gate.go`'s allowlist). That latitude only makes sense when nothing on the board waits on the answer. Here the effect is paused partway through resolving: Rhystic Shield's +0/+2 and Wild Might's +4/+4 must land before damage, and Rhystic Tutor's search is the rest of the spell. So every link of the chain is queued with `OwedInStep` set to the current step (the anchor `QueueMayPayInThisStepForEffect` uses), and, for a cast trigger whose outcome counters the spell (Dash Hopes, Temporal Extortion, Brain Gorgers, Phantasmagorian), with `GuardsStackItem` set to that spell (`counter_unless_paid.go`'s rule).

A consequence: `payUnlessCovered` (`auto_answer.go`) leaves a step-owed prompt uncovered, so **no standing answer (ADR 0127) applies** to an any-player offer. A remembered "Always pay Rhystic Study" never silently pays an opponent's Rhystic Tutor. This is [question 2](#questions-for-the-owner-answered-2026-10-13).

### 3. Payments other than mana

`PayAction` (`pay_unless_action.go`) already has discard, sacrifice and energy. Two kinds are added so the 13 cards and the life-paying ones work:

- **`PayActionLife`** (`Count` life): payable when `CanPayLifeLocked(p, Count)` (CR 119.4, and CR 119.8's "can't lose life" through the same gate). Paid with `payLifeAsCostLocked`. Temporal Extortion's "half their life, rounded up" is worked out **per payer** when their link is queued, since each player's half is different. That's why the frame's action is built per link, not once.
- **`PayActionDamage`** (`Count` damage from `Source` to the payer): always choosable. Being dealt damage isn't a cost the payer can lack the resources for, so "Pay" always counts as paying even if the damage is then prevented. It's dealt through the ordinary damage path, so prevention and replacement apply. Needed only for the Browbeat family ([question 3](#questions-for-the-owner-answered-2026-10-13)).

Exiling from your own graveyard (Carrion Rats, Carrion Wurm) and Worms of the Earth's "either" are left for a later slice. Each is one more `PayActionKind`.

### 4. Assist parks the cast on the helper

`CastSpellParams` gains `AssistPlayer uuid.UUID` (wire `assist_player` on `cast_spell`), allowed only on a spell with assist (a `Spec.Assist` flag, keyword `assist` added to `canonicalKeywords`), naming another player still in the game, and only when the total cost has a generic part. Anything else is refused at announce, the way a stray `gift_opponent` is.

`castSpellLocked` validates the whole announcement as it does today, works out the **total** cost (CR 601.2f: printed or alternative cost, X, additional costs, commander tax, increases, then reductions), and reads its generic part, `G`. Then, at the same point it asks `askCostCommanderLocked` and before the auto-tap and `applyCastCostLocked`:

1. **Park.** If `AssistPlayer` is set and no assist answer is attached yet, it queues a `PendingChoicePayAmount` to the helper and returns without moving or paying anything:
   - `PayAmount{Min: 0, Max: min(G, what the helper can make), Resource: PayResourceMana, Unit: PayAmountAssist}`, where "what the helper can make" is their floating pool plus their auto-tapper's plan for generic mana, read-only (`autoTapTopUpLocked` plans without tapping).
   - A new `assistResume` frame carries the `announce` closure (`castSpellLocked` again with `params.assistPaid = {helper, n}`), exactly as `costCommanderFrame` does. The caster's own taps and choices ride along by value.
   - The prompt also carries plain data for the view, `PendingChoice.Assist{Caster, Spell, SpellName, Generic G}`.
2. **The helper answers** with an amount `n`. An `n` above `Max`, or one the helper can no longer fund, is **refused** with the prompt left open, the way a malformed pay-unless answer is. Zero is always an answer.
3. **Re-run, pay both at once (CR 601.2h).** The re-run re-validates everything. In `applyCastCostLocked` it then **plans** the helper's `n` generic (their pool, then their auto-tapper) and the caster's remaining cost (total minus `n` generic) **before materialising either**. If both are fundable, it taps, spends and puts the spell on the stack with `PaidCost.AssistedBy = helper, AssistPaid = n` (for the log and a later "if this spell was assisted" reader). If either is short, nothing is tapped or spent, the card stays where it was, and the caster gets an `EventEffectError`-style log line ("Fan Favorite couldn't be cast: you no longer have {1}") instead of a refused action, since the action being answered is the helper's.

Why park and not pause with the spell on the stack: the engine casts a spell in one locked step (`castSpellLocked` → `applyCastCostLocked` → push), and every validator assumes that. The parked announcement is already how the engine waits on another player during a cast, and a re-run that re-checks everything can't pay for a board that changed under it. The one thing that gets reordered is that the caster has announced their own taps before the helper answers. That's harmless: the caster's leftover mana stays in their pool (CR 106.4 empties it at the end of the step), and the helper sees `G` and the caster's spell before choosing, as at a paper table.

**Departure.** If the helper leaves with the prompt open, the departure table answers `0` (`dropDefault`, which is how `pay_amount` already drops) and the cast re-runs unassisted. If the caster leaves, the prompt is withdrawn (`dropChoicesForPlayerLocked`) and nothing was ever paid.

**The mana a helper spends** is spent in the cast's spend context (`ManaSpendContext` for this spell), since it's spent to cast this spell. "Spend only on creature spells" mana in the helper's pool may pay for Fan Favorite. Whose mana it is doesn't change what it may be spent on.

### 5. The client

**The offeree's prompt is the one they already know.** An any-player link is a `pay_unless`, drawn inline in the dock with Pay / Don't pay. The header gains one line from `any_player_offer`: "Any player may pay {2}: Alice declined; then Carol." It also gains whose effect it is ("Bob's Rhystic Tutor"), because the payer is usually not the controller and needs to see who they're stopping.

**Everyone else sees the offer read-only**, as ADR 0146 decision 6 does for a vote: `anyPlayerOfferRequest` in `choiceDock.ts` builds a `step` request with no buttons, "Carol is deciding whether to pay {2} (Bob's Rhystic Tutor)", with the declines so far, and `dockHint.ts`'s `passHint` gets the same sentence for a seat with nothing to answer, as `commanderReturnWaiting` has. Spectators get the log.

**The log.** `EventAnyPlayerPaid` ("Carol paid {2}: Rhystic Tutor does nothing") and `EventAnyPlayerDeclined` (one line per decline) go in the public log, so a table reading back can see who paid.

**Assist, the caster.** The cast flow gains an "Assist" row exactly where gift's opponent picker sits (`OptionalCostView`-shaped, from a new per-viewer `CastSurfaceView.Assist{Max, Players []string}` in `protocol/view.go`, beside `optional_costs`): "Ask a player to help pay (up to {3})", with a player pick. The cast is sent with `assist_player`. While it's parked, the caster's dock shows "Waiting for Bob to help pay for Fan Favorite", with a **Cancel request** button ([question 5](#questions-for-the-owner-answered-2026-10-13)).

**Assist, the helper.** A `pay_amount` with `resource: "mana"` is drawn as the existing stepper (ADR 0129 §3) with a new header: "Alice asks you to help pay for Fan Favorite ({3}{B}). Pay up to {3} of its generic mana." The stepper runs 0 to `max` and shows the mana the answer will tap, from the auto-tapper's plan. The buttons are "Pay {n}" and "Don't help" (0). The helper pays from their pool and auto-tapper only, with no tap picker in the first PR ([question 6](#questions-for-the-owner-answered-2026-10-13)).

### 6. The bot

**As an offeree (any player pays).** The card declares what the offer is worth to its controller in the existing `Purpose` (Rhystic Tutor: `Tutors: 1`; Rhystic Shield, Wild Might: `Answers: AnswerPump`; Icy Prison: the other way round, since nobody paying sacrifices the Prison and returns the exiled creature, so the payment is what the controller wants; PR 1 marks that offer `HelpsController`). The heuristic reads `any_player_offer` and:

- **is the controller** (or, for "if a player does" cards, the player the payment hurts): never pays;
- **is an opponent of the controller:** pays when it can and the price is below what not paying gives the controller, scored against `giftsValue`. A mana price goes through `taxAgainstPlan`'s check (don't pay a tax that strands the turn plan's next spell). A life price uses the Phyrexian floor (`phyrexianLifeFloor`, ADR 0131 §5). A sacrifice or discard uses the existing least-valuable-cards score.

A bot opponent does not **free-ride**, that is, decline because a later opponent could still pay ([owner decision 7](#questions-for-the-owner-answered-2026-10-13)).

**As an assist helper.** It answers `0`. In free-for-all Commander, helping an opponent cast a spell is never worth it to the heuristic, and nothing in `Purpose` says otherwise. **As a caster**, the enumerator doesn't offer `assist_player` variants of a cast, so bots never ask. That keeps the move list the same size and adds nothing to tune.

Model tiers see the same move list. The labels say who is offering what ("pay {2} to stop Bob's Rhystic Tutor") so a model can price it.

### 7. Testing

- **Engine (`internal/game`).** Offer order is APNAP from the active player, whoever controls the source. The first payer ends the offer and `OnPaid` gets them. Everyone declining runs `OnNobody` exactly once. An unfundable "Pay" moves on. A departed offeree is skipped, and one who leaves while asked declines. `Opponents` skips the controller. Life: `CanPayLifeLocked` gates "Pay", and Temporal Extortion's half is per payer. The prompt blocks the step and guards a spell. No standing answer applies.
- **Assist.** The cap is the total cost's generic (commander tax raises it, a reduction lowers it). The park moves and pays nothing. The helper's over-cap or unfundable answer is refused with the prompt kept. Re-runs are all or nothing (helper fundable and caster short: nothing tapped on either side). The helper leaving gives 0, and the caster leaving withdraws the prompt. `assist_player` on a non-assist spell, on the caster, or on a departed player is refused. Each new shape gets a test that fails without it, the "back-out" discipline of ADR 0146.
- **Enumerator and heuristic.** `legal` offers each link to its chooser only. The bot controller never pays its own offer. The bot opponent pays a cheap tax and declines below the life floor. The helper answers 0.
- **Protocol.** `any_player_offer` and `assist` round-trip in `view_test.go`, and the snapshot shape-file drift test covers the two additive fields.
- **Client.** `choiceDock.test.ts` and render tests for the offeree header, the read-only offer request, the helper's stepper and the caster's waiting line. `targeting.ts` sends `assist_player`. One e2e path: a four-seat table with Rhystic Tutor, one bot opponent paying.
- **Cards.** One test per card in `cards/effects` (any_player_cards_test.go already holds Fan Favorite's pump).

## Out of scope

1. **Rhystic Cave.** "Add one mana … unless any player pays {1}" is a mana ability, which can be activated in the middle of paying a cost (CR 605.3a) and resolves immediately, off the stack (CR 605.3b). Pausing it for three other players' answers would pause a cast mid-payment, which the engine doesn't do (decision 4 parks *before* payment). It stays uncatalogued, and the registry row can say so.
2. **Leonin Arbiter.** "Any player may pay {2} for that player to ignore this effect until end of turn" is an action a player takes whenever they could cast an instant. It isn't an offer made during a resolution. It's its own seam, closer to ADR 0106's any-player activations.
3. **Cleansing**, deferred to a follow-up issue by [owner decision 4](#questions-for-the-owner-answered-2026-10-13).
4. **Carrion Rats, Carrion Wurm, Worms of the Earth**: two more `PayActionKind`s (exile from your graveyard, and an "either" payment), for a later slice.
5. **Two-Headed Giant teams.** Assist's natural home. C2 has no team format, and "another player" is any other player.

## Consequences

**Cost.** One door and one card primitive for any-player offers, two `PayActionKind`s, one `PendingChoice` field for each seam (`AnyPlayerOffer`, `Assist`), both additive in the snapshot shape. One `CastSpellParams` field, a `pay_amount` resource (`mana`) and unit (`assist`), two wire fields on the pending-choice view, one on the cast options, two log events, and the read-only dock request. No new prompt kind: every question either seam asks is a `pay_unless` or a `pay_amount`.

**Cards.** PR 1 and PR 2 ship the 10 cards the table marks PR 1 or PR 2. Under owner decision 3, PR 2 also ships 11 of the 14 "any player may [do X]" cards: everything except Carrion Rats, Carrion Wurm and Worms of the Earth. Fan Favorite loses its caveat in PR 3, and the other 15 assist cards become card work.

**The thing most likely to be got wrong later.** Treating an any-player link like Rhystic Study's prompt. It must block its step (decision 2), and a standing answer must not reach it. A new door that skips `QueueAnyPlayerPaysForEffect` and queues a bare `PayUnless` to "each player" would lose both, which is why there is one door.

## Delivery

- **PR 1: the any-player offer with mana.** `QueueAnyPlayerPaysForEffect`, `AnyPlayerOffer`, the gate and guard wiring, the view field, the offeree header and read-only dock request, the bot's offeree rule, the log events. Cards: Rhystic Circle, Rhystic Shield, Rhystic Tutor, Wild Might, Soul Strings, Nakaya Shade, Icy Prison. Closes the `unless-any-player-pays` row as Partial.
- **PR 2: life, damage, and the "any player may [do X]" cards.** `PayActionLife` and `PayActionDamage`, the existing sacrifice and discard payments offered through the chain, and a no-payment offer for Distant Memories. Cards: Aether Rift, Dash Hopes, Temporal Extortion, Book Burning, Breaking Point, Browbeat, Barbarian Bully, Argothian Wurm, Shivan Wumpus, Brain Gorgers, Prowling Pangolin, Phantasmagorian, Innocent Traveler, Distant Memories. The `unless-any-player-pays` row stays Partial for Cleansing, Rhystic Cave, Leonin Arbiter, Carrion Rats, Carrion Wurm and Worms of the Earth.
- **PR 3: assist.** `assist_player`, the park and re-run, `PayResourceMana`, `CastSurfaceView.Assist`, the caster's and helper's client, departure, the bot. Fan Favorite drops its caveat, and the `assist` row closes.
- **Follow-up issues:** the remaining 15 assist cards (one batch), Cleansing (owner decision 4), the graveyard-exile and "either" payments (Carrion Rats, Carrion Wurm, Worms of the Earth), and Rhystic Cave and Leonin Arbiter (documented as out).

## Questions for the owner (answered 2026-10-13)

Each question lists the recommended option first. On 2026-10-13 the owner chose the recommended option, (a), on all nine. The questions are kept with the options not chosen.

1. **Who is offered "any player may pay", and in what order (§1; CR 101.4)?**
   - (a) **Recommended:** every player still in the game, controller included, in APNAP order from the **active** player, stopping at the first who pays. It follows CR 101.4's order (decision 1 notes the one difference: a second payment is never asked for). On your own turn you are asked first and simply decline.
   - (b) The controller's opponents only, in turn order from the controller. Fewer clicks (the controller never wants to pay), but it isn't the card, and on a "you may pay" card the controller is a legal payer.
   - (c) Ask everyone at once and take the first "Pay" to arrive. Fastest at a live table, but it's a race, not APNAP, and bots would always win it.

   **Decision:** (a), as recommended (owner decision 1): every player still in the game, controller included, in APNAP order from the active player, stopping at the first who pays.

2. **Does an any-player offer block the table, and can a standing answer reach it (§2; ADR 0018 §6, ADR 0127)?**
   - (a) **Recommended:** it blocks the step it was asked in (and guards the spell for the "counter it" cards), and no standing answer applies. The effect is mid-resolution, so a table walking past it would be answering it by doing.
   - (b) It keeps Rhystic Study's latitude (doesn't block) and standing answers apply. Fewer waits, but Rhystic Shield's toughness can arrive after combat damage.
   - (c) It blocks, but a "Never" standing answer applies (so a player can mute all Rhystic offers). Convenient, but it needs a new key family in ADR 0127 for one prompt shape.

   **Decision:** (a), as recommended (owner decision 2): each link blocks the step it was asked in, guards the spell for the "counter it" cards, and takes no standing answer.

3. **Which payments ship with #1903 (§3)?**
   - (a) **Recommended:** mana and life (PR 1 and PR 2: the 10 cards), plus `PayActionDamage` and the existing sacrifice and discard payments in PR 2. That adds Book Burning, Breaking Point, Browbeat, Barbarian Bully, Argothian Wurm, Shivan Wumpus, Brain Gorgers, Prowling Pangolin, Phantasmagorian and Innocent Traveler for one new payment kind. Distant Memories (no payment) comes too.
   - (b) Mana and life only: the 13-card family as #1903 counted it, minus Rhystic Cave and Leonin Arbiter. The smallest change. The rest is filed as a follow-up.
   - (c) Everything, including the graveyard-exile and "either" payments. Most cards, but two more payment kinds for three cards.

   **Decision:** (a), as recommended (owner decision 3): mana and life, plus `PayActionDamage` and the existing sacrifice and discard payments, so 11 of the 14 "any player may [do X]" cards ship in PR 2.

4. **Cleansing ("for each land, destroy that land unless any player pays 1 life")?**
   - (a) **Recommended:** later, as a follow-up issue. One card, and it needs its own prompt shape.
   - (b) Batched: each player in APNAP order picks any number of the lands not yet saved and pays 1 life each, in one `choose_cards` prompt. Equivalent in outcome, with far fewer clicks than one offer per land.
   - (c) Literally one offer per land, around the table. Faithful but unplayable: 40 lands at a four-seat table is up to 160 prompts.

   **Decision:** (a), as recommended (owner decision 4): Cleansing is a follow-up issue.

5. **Can the assist caster withdraw a parked request (§4, §5)?**
   - (a) **Recommended:** yes. A "Cancel request" button withdraws the prompt, and nothing was moved or paid, so it's a clean undo of an announcement that never happened. Without it, a helper who walks away holds the caster's cast hostage.
   - (b) No. The caster waits for the answer (the helper can always answer 0), as a commander owner's CR 903.9 answer is waited on today.
   - (c) No button, but a timeout answers 0 after N seconds. Needs a timer the engine has nowhere else.

   **Decision:** (a), as recommended (owner decision 5): the caster can withdraw a parked request with a Cancel request button.

6. **How does the assist helper pay (§5)?**
   - (a) **Recommended:** the amount stepper, paid from their pool and then their auto-tapper. One number and one click, with the planned taps shown, and the same posture as a mana pay-unless.
   - (b) The stepper plus a tap picker, so the helper chooses which lands to tap. More control and more UI, for a prompt that comes up rarely.
   - (c) The helper floats mana by hand first and pays from the pool only. Most like paper and the slowest.

   **Decision:** (a), as recommended (owner decision 6): the helper pays with the amount stepper, from their pool and then their auto-tapper, with no tap picker.

7. **Does a bot opponent free-ride on an any-player offer (§6)?**
   - (a) **Recommended:** no. A bot pays when it is worth it to the bot, whoever is still to be asked. Predictable, needs no model of the other seats, and matches how it answers Rhystic Study today.
   - (b) Yes. A bot declines when a later opponent could still pay and would want to, and pays only as the last opponent asked. Sharper in theory, but it needs a guess about the others, and at a table of bots everyone declines until the last.
   - (c) Configurable, a `heuristic.Config` switch on in `DefaultConfig()`, off in `BaselineConfig()`. Measurable in the arena, at the cost of one more switch.

   **Decision:** (a), as recommended (owner decision 7): no free-riding. A bot pays when it's worth it to the bot.

8. **Can an assist caster ask a bot (§4, §6)?**
   - (a) **Recommended:** yes. Any other player still in the game can be chosen (CR 702.132a), and the picker marks bot seats with "(always declines)", since the heuristic answers 0.
   - (b) Only seats played by people. Saves a wasted click, but it hides a legal choice and an MCP seat (ADR 0122) might help.
   - (c) Yes, and the bot sometimes helps (for example, a symmetrical Game Plan or Huddle Up that names it). That needs a new `Purpose` reading for little gain.

   **Decision:** (a), as recommended (owner decision 8): any other player can be asked, and bot seats are marked "(always declines)".

9. **Delivery.**
   - (a) **Recommended:** three PRs as listed: mana offer, life (and the question 3 cards), assist. Each is reviewable alone, and assist doesn't wait on the offer.
   - (b) Two PRs: the whole any-player offer, then assist.
   - (c) One PR for both seams. Fewer reviews, but a large diff across engine, protocol, client and bot.

   **Decision:** (a), as recommended (owner decision 9): three PRs, as listed under [Delivery](#delivery).

## Amendment 2026-10-24 — an any-player offer's consequences can be a key, so a table waiting on any link is a restore point (#4264)

Each link of the offer is an ordinary pay-unless (decision 1) whose decline is a closure over the whole offer: who is next, the cost, the labels and the two consequences. A table waiting on any player's answer could be undone but not saved and restored.

**Decision.** `game.RegisterAnyPlayerPays(key, onPaid, onNobody, perPayer)` registers the two consequences and the optional per-payer payment (a function of the game, the player and the carried IDs) under one key in the effect-key namespace, ledgered `anyplayer <key>`, and `QueueAnyPlayerPaysKeyedForEffect` queues the offer. A link carries no frame: the key rides `pickThen` and the card's IDs `optionCarry`, the offer is the `AnyPlayerOffer` already on the link (the controller, the order, who has declined, the two flags), the payment is `payCost` or `payActionClause`, and the step and the guarded spell are `owedInStep` and `guardsStackItem`. `anyPlayerFrameOf` rebuilds the link's decline (record the refusal, ask the player after this one in the order, or run the nobody branch) and its payment (record it, run the paid branch) from that data, and `payUnlessFrameOf` hands it to every reader, so the answer, the standing-answer exclusion, the departure table's decline and the enumerator treat a keyed link as they treat a framed one. Card side: `effects.AnyPlayerPays{Then, Carry}` and `effects.AnyPlayerConsequences`; the four "when you cast this spell, any player may ... to counter it" cards share `anyPlayerMayCounterThisSpell` and two keys (`any-player/counter-this-spell`, and `-half-life` for Temporal Extortion's per-payer price).

**Not keyed.** The per-object offer (`QueueAnyPlayerPaysForEachForEffect`, decision 5 of ADR 0187's amendment), whose legs are `choose_cards` prompts, and the other any-player cards, whose `OnPaid` / `OnNobody` read the stack item (Wild Might's creature, Soul Strings' cards) or one Go value. An older binary refuses a file with a keyed link open (`checkEffectKeys`, `any-player consequences <key>`).

**Tests** (`effects/keyed_pay_unless_shapes_test.go`): Dash Hopes restored at the first link and again at the second, with the first player's decline on the offer, the first payer countering the spell; Temporal Extortion's restored offer pricing each player's own half; an unknown key refused. **Back-out:** returning nil from `anyPlayerFrameOf` fails `TestAnyPlayerOfferIsARestorePointAtEveryLink` and `TestAnyPlayerPerPayerPriceIsARestorePoint`; making `KnownAnyPlayerPays` accept every key fails `TestAnyPlayerOfferWithAnUnknownKeyIsRefused`.
