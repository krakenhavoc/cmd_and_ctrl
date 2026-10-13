# ADR 0150 — Other players pay: "unless any player pays" and assist

**Status:** Proposed · 2026-10-13 · S53 — Faces, zones, and shared payment (ADR first), tracker [#1910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1910)
**Issues:** [#1903](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1903) (seam: unless any player pays — Rhystic Circle) and [#1857](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1857) (seam: assist — another player helps pay the generic mana — Fan Favorite). Registry rows: `unless-any-player-pays` and `assist`, both open; this ADR does not touch them (they close with the code).
**Owner decisions:** none yet. The [questions](#questions-for-the-owner) at the end need answers before any engine code is written.
**Numbering:** assigned by the S53 orchestrator. `origin/develop` at `9a9372749` ends at **0149**, so this is **0150**.
**Builds on:** [ADR 0018](0018-triggers-on-the-stack.md) §6 (the pay-unless prompt, and the latitude that lets a table walk past it), [ADR 0013](0013-replacement-effects.md) §5af and #1397 (a cost announcement parked on another player's answer), [ADR 0060](0060-leaving-the-game.md) (the departure table), [ADR 0108](0108-turn-scoped-effects-object-history-and-damage-shields.md) §5 (non-mana pay-unless payments), [ADR 0111](0111-action-dock.md) (the dock), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic and `Purpose`), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers), [ADR 0129](0129-energy-getting-and-paying-it.md) §3 (the `pay_amount` prompt), [ADR 0146](0146-voting.md) (one prompt per player in turn order, and a read-only view of it for everyone else).

This ADR was written plan-first. No code changed with it. Every claim about the code below was checked on `origin/develop` at `9a9372749`.

---

## Context

Two seams with one gap: **a payment made by a player other than the one the engine already asks.**

- **#1903.** "Unless any player pays {2}" and "Any player may pay {1}. If no one does, …" offer one payment to the whole table. The engine's pay-unless prompt asks one named player.
- **#1857.** Assist lets the caster pick another player, who may pay some of the spell's generic mana. The cast path has one payer.

They are one group because they need the same three things: a prompt addressed to a player who is not acting, a payment from that player's own mana (or life, or cards) that pays for an effect someone else controls, and a way for the rest of the table to see who is being asked. They are not one mechanism. One happens while an effect resolves and the other while a spell is cast, so the engine work is two pieces with a shared client and bot pattern.

### The rules

Rules quoted below are paraphrased from the pinned Comprehensive Rules. Where an exact reading matters, it is called out.

- **CR 118.12** reads "[Do something] unless [a player does something else]" as "[a player may do something else]. If they don't, [do something]". **CR 118.12a**: the payment is a cost paid as the spell or ability resolves. So an "unless" payment follows the cost rules: **CR 118.3** (you can't pay a cost you lack the resources for) and **CR 119.4** (life can be paid only if the life total is at least the amount; 0 life is always payable).
- **CR 101.4** (APNAP): when several players make choices at the same time, the active player chooses first, then each other player in turn order. The Gatherer rulings on the Rhystic cards and on Browbeat read "any player may" the same way: starting with the active player, each player in turn order gets the chance, and once one player pays, nobody after them is asked and the effect doesn't happen.
- **"Any player" includes the controller.** "Any opponent" (Innocent Traveler, Distant Memories) does not.
- **CR 702.132a (assist):** a static ability that changes how the spell is paid for (CR 601.2g-h). If the spell's *total* cost has a generic mana part, then before mana abilities are activated during the cast, the caster may choose another player. That player gets the chance to activate mana abilities, then the caster does. Before the total cost is paid, the chosen player may pay any amount of the generic mana in the total cost. The reminder text ("up to {3}") is the printed generic. The rule's cap is the generic in the **total** cost (CR 601.2f), so commander tax and Thalia-style increases raise it and reductions lower it.
- **CR 601.2h**: the total cost is paid in any order, but all of it is paid at once. Partial payment isn't allowed.
- **CR 605.3a**: a mana ability resolves immediately and gives no player priority. That's why Rhystic Cave is out of scope (see [Out of scope](#out-of-scope)).
- **CR 800.4a**: when a player leaves the game, their choices go with them. A departed player's offer is a decline, and a departed helper paid nothing.

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
| Rhystic Cave | {1} | **mana ability** | out (CR 605.3a) |
| Leonin Arbiter | {2} | a static's special action, not an offer | out |

**The same offer with a non-mana payment ("any player may [do X]. If a player does / If no one does"): 14 more cards, none catalogued.** Book Burning, Breaking Point, Browbeat, Barbarian Bully (have this deal N damage to them); Argothian Wurm, Shivan Wumpus (sacrifice a land); Brain Gorgers (sacrifice a creature); Prowling Pangolin (sacrifice two creatures); Phantasmagorian (discard three cards); Carrion Rats, Carrion Wurm (exile cards from their graveyard); Innocent Traveler (any opponent sacrifices a creature); Worms of the Earth (sacrifice two lands *or* take 5 damage); Distant Memories (any opponent may let you have the card, which costs nothing). They use the same fan-out with a different payment. Which of them come in this ADR's PRs is [question 3](#questions-for-the-owner).

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

- **One prompt at a time, each an ordinary pay-unless.** The prompt kind, its validator, the payment path, `legal/choices.go`'s moves, the inline dock prompt and its keys, the heuristic and the departure table's `dropDecline` all answer it unchanged. A player who leaves while asked declines (`declineDepartedChoiceLocked`), so the offer moves on (CR 800.4a). Players who left before their turn are skipped when the next prompt is built.
- **The first payer ends the offer.** Paying runs `OnPaid` with that player and nobody after them is asked. An unfundable "Pay" is a decline (unchanged) and moves the offer on.
- **The offer is plain data on the open prompt:** `PendingChoice.AnyPlayerOffer *AnyPlayerOffer{Controller, Order []uuid.UUID, Declined []uuid.UUID, Opponents bool}`. That gives the view (decision 5) and the bot (decision 6) the whole picture without a new kind. The branches stay closures on `payUnlessFrame`, as every pay-unless's do, so a table waiting on an offer is not a restore point and the snapshot census already says so (`snapshot.go`, `"payUnlessResume"`). The new field is additive (`pendingChoices[].anyPlayerOffer` in the shape file).
- **Card helpers.** `effects.AnyPlayerPays{Cost, Question, OnNobody}` covers the "unless any player pays" wording and `{…, OnPaid}` covers "if a player does". Both are bound to the stack item the way `PayUnless` and `MayPay` are (`NewContext(g, item)` in the closures), so `OnNobody` reads the resolving item's targets and X (Soul Strings' {X}, Wild Might's target).

### 2. It blocks the table, in the step it was asked in

Rhystic Study's prompt is the one kind a table may walk past (ADR 0018 §6; `choice_gate.go`'s allowlist). That latitude only makes sense when nothing on the board waits on the answer. Here the effect is paused partway through resolving: Rhystic Shield's +0/+2 and Wild Might's +4/+4 must land before damage, and Rhystic Tutor's search is the rest of the spell. So every link of the chain is queued with `OwedInStep` set to the current step (the anchor `QueueMayPayInThisStepForEffect` uses), and, for a cast trigger whose outcome counters the spell (Dash Hopes, Temporal Extortion, Brain Gorgers, Phantasmagorian), with `GuardsStackItem` set to that spell (`counter_unless_paid.go`'s rule).

A consequence: `payUnlessCovered` (`auto_answer.go`) leaves a step-owed prompt uncovered, so **no standing answer (ADR 0127) applies** to an any-player offer. A remembered "Always pay Rhystic Study" never silently pays an opponent's Rhystic Tutor. This is [question 2](#questions-for-the-owner).

### 3. Payments other than mana

`PayAction` (`pay_unless_action.go`) already has discard, sacrifice and energy. Two kinds are added so the 13 cards and the life-paying ones work:

- **`PayActionLife`** (`Count` life): payable when `CanPayLifeLocked(p, Count)` (CR 119.4, and CR 119.8's "can't lose life" through the same gate). Paid with `payLifeAsCostLocked`. Temporal Extortion's "half their life, rounded up" is worked out **per payer** when their link is queued, since each player's half is different. That's why the frame's action is built per link, not once.
- **`PayActionDamage`** (`Count` damage from `Source` to the payer): always choosable. Being dealt damage isn't a cost the payer can lack the resources for, so "Pay" always counts as paying even if the damage is then prevented. It's dealt through the ordinary damage path, so prevention and replacement apply. Needed only for the Browbeat family ([question 3](#questions-for-the-owner)).

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

**Assist, the caster.** The cast flow gains an "Assist" row exactly where gift's opponent picker sits (`OptionalCostView`-shaped, from a new per-viewer `CastSurfaceView.Assist{Max, Players []string}` in `protocol/view.go`, beside `optional_costs`): "Ask a player to help pay (up to {3})", with a player pick. The cast is sent with `assist_player`. While it's parked, the caster's dock shows "Waiting for Bob to help pay for Fan Favorite", with a **Cancel request** button ([question 5](#questions-for-the-owner)).

**Assist, the helper.** A `pay_amount` with `resource: "mana"` is drawn as the existing stepper (ADR 0129 §3) with a new header: "Alice asks you to help pay for Fan Favorite ({3}{B}). Pay up to {3} of its generic mana." The stepper runs 0 to `max` and shows the mana the answer will tap, from the auto-tapper's plan. The buttons are "Pay {n}" and "Don't help" (0). The helper pays from their pool and auto-tapper only, with no tap picker in the first PR ([question 6](#questions-for-the-owner)).

### 6. The bot

**As an offeree (any player pays).** The card declares what the offer is worth to its controller in the existing `Purpose` (Rhystic Tutor: `Tutors: 1`; Rhystic Shield, Wild Might: `Answers: AnswerPump`; Icy Prison: a removal row, since nobody paying keeps the creature exiled). The heuristic reads `any_player_offer` and:

- **is the controller** (or, for "if a player does" cards, the player the payment hurts): never pays;
- **is an opponent of the controller:** pays when it can and the price is below what not paying gives the controller, scored against `giftsValue`. A mana price goes through `taxAgainstPlan`'s check (don't pay a tax that strands the turn plan's next spell). A life price uses the Phyrexian floor (`phyrexianLifeFloor`, ADR 0131 §5). A sacrifice or discard uses the existing least-valuable-cards score.

Whether a bot opponent should **free-ride**, declining when a later opponent can still pay, is [question 7](#questions-for-the-owner).

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

1. **Rhystic Cave.** "Add one mana … unless any player pays {1}" is a mana ability (CR 605.3a), which resolves at once and can be activated in the middle of paying a cost. Pausing it for three other players' answers would pause a cast mid-payment, which the engine doesn't do (decision 4 parks *before* payment). It stays uncatalogued, and the registry row can say so.
2. **Leonin Arbiter.** "Any player may pay {2} for that player to ignore this effect until end of turn" is an action a player takes whenever they could cast an instant. It isn't an offer made during a resolution. It's its own seam, closer to ADR 0106's any-player activations.
3. **Cleansing**, unless the owner picks the batched form in [question 4](#questions-for-the-owner).
4. **Carrion Rats, Carrion Wurm, Worms of the Earth**: two more `PayActionKind`s (exile from your graveyard, and an "either" payment), for a later slice.
5. **Two-Headed Giant teams.** Assist's natural home. C2 has no team format, and "another player" is any other player.

## Consequences

**Cost.** One door and one card primitive for any-player offers, two `PayActionKind`s, one `PendingChoice` field for each seam (`AnyPlayerOffer`, `Assist`), both additive in the snapshot shape. One `CastSpellParams` field, a `pay_amount` resource (`mana`) and unit (`assist`), two wire fields on the pending-choice view, one on the cast options, two log events, and the read-only dock request. No new prompt kind: every question either seam asks is a `pay_unless` or a `pay_amount`.

**Cards.** PR 1 and PR 2 ship the 10 cards the table marks PR 1 or PR 2. Fan Favorite loses its caveat in PR 3, and the other 15 assist cards become card work. Question 3 decides how much of the 14-card "any player may [do X]" family comes along.

**The thing most likely to be got wrong later.** Treating an any-player link like Rhystic Study's prompt. It must block its step (decision 2), and a standing answer must not reach it. A new door that skips `QueueAnyPlayerPaysForEffect` and queues a bare `PayUnless` to "each player" would lose both, which is why there is one door.

## Delivery

- **PR 1: the any-player offer with mana.** `QueueAnyPlayerPaysForEffect`, `AnyPlayerOffer`, the gate and guard wiring, the view field, the offeree header and read-only dock request, the bot's offeree rule, the log events. Cards: Rhystic Circle, Rhystic Shield, Rhystic Tutor, Wild Might, Soul Strings, Nakaya Shade, Icy Prison. Closes the `unless-any-player-pays` row as Partial.
- **PR 2: life (and the question 3 family).** `PayActionLife` (and `PayActionDamage` if chosen). Cards: Aether Rift, Dash Hopes, Temporal Extortion, plus whatever question 3 adds.
- **PR 3: assist.** `assist_player`, the park and re-run, `PayResourceMana`, `CastSurfaceView.Assist`, the caster's and helper's client, departure, the bot. Fan Favorite drops its caveat, and the `assist` row closes.
- **Follow-up issues:** the remaining assist cards (one batch), Cleansing (if question 4 says later), Rhystic Cave and Leonin Arbiter (documented as out).

## Questions for the owner

Each question lists the recommended option first.

1. **Who is offered "any player may pay", and in what order (§1; CR 101.4)?**
   - (a) **Recommended:** every player still in the game, controller included, in APNAP order from the **active** player, stopping at the first who pays. It's what the rules and the Rhystic rulings say, and on your own turn you are asked first and simply decline.
   - (b) The controller's opponents only, in turn order from the controller. Fewer clicks (the controller never wants to pay), but it isn't the card, and on a "you may pay" card the controller is a legal payer.
   - (c) Ask everyone at once and take the first "Pay" to arrive. Fastest at a live table, but it's a race, not APNAP, and bots would always win it.

2. **Does an any-player offer block the table, and can a standing answer reach it (§2; ADR 0018 §6, ADR 0127)?**
   - (a) **Recommended:** it blocks the step it was asked in (and guards the spell for the "counter it" cards), and no standing answer applies. The effect is mid-resolution, so a table walking past it would be answering it by doing.
   - (b) It keeps Rhystic Study's latitude (doesn't block) and standing answers apply. Fewer waits, but Rhystic Shield's toughness can arrive after combat damage.
   - (c) It blocks, but a "Never" standing answer applies (so a player can mute all Rhystic offers). Convenient, but it needs a new key family in ADR 0127 for one prompt shape.

3. **Which payments ship with #1903 (§3)?**
   - (a) **Recommended:** mana and life (PR 1 and PR 2: the 10 cards), plus `PayActionDamage` and the existing sacrifice and discard payments in PR 2. That adds Book Burning, Breaking Point, Browbeat, Barbarian Bully, Argothian Wurm, Shivan Wumpus, Brain Gorgers, Prowling Pangolin, Phantasmagorian and Innocent Traveler for one new payment kind. Distant Memories (no payment) comes too.
   - (b) Mana and life only: the 13-card family as #1903 counted it, minus Rhystic Cave and Leonin Arbiter. The smallest change. The rest is filed as a follow-up.
   - (c) Everything, including the graveyard-exile and "either" payments. Most cards, but two more payment kinds for three cards.

4. **Cleansing ("for each land, destroy that land unless any player pays 1 life")?**
   - (a) **Recommended:** later, as a follow-up issue. One card, and it needs its own prompt shape.
   - (b) Batched: each player in APNAP order picks any number of the lands not yet saved and pays 1 life each, in one `choose_cards` prompt. Equivalent in outcome, with far fewer clicks than one offer per land.
   - (c) Literally one offer per land, around the table. Faithful but unplayable: 40 lands at a four-seat table is up to 160 prompts.

5. **Can the assist caster withdraw a parked request (§4, §5)?**
   - (a) **Recommended:** yes. A "Cancel request" button withdraws the prompt, and nothing was moved or paid, so it's a clean undo of an announcement that never happened. Without it, a helper who walks away holds the caster's cast hostage.
   - (b) No. The caster waits for the answer (the helper can always answer 0), as a commander owner's CR 903.9 answer is waited on today.
   - (c) No button, but a timeout answers 0 after N seconds. Needs a timer the engine has nowhere else.

6. **How does the assist helper pay (§5)?**
   - (a) **Recommended:** the amount stepper, paid from their pool and then their auto-tapper. One number and one click, with the planned taps shown, and the same posture as a mana pay-unless.
   - (b) The stepper plus a tap picker, so the helper chooses which lands to tap. More control and more UI, for a prompt that comes up rarely.
   - (c) The helper floats mana by hand first and pays from the pool only. Most like paper and the slowest.

7. **Does a bot opponent free-ride on an any-player offer (§6)?**
   - (a) **Recommended:** no. A bot pays when it is worth it to the bot, whoever is still to be asked. Predictable, needs no model of the other seats, and matches how it answers Rhystic Study today.
   - (b) Yes. A bot declines when a later opponent could still pay and would want to, and pays only as the last opponent asked. Sharper in theory, but it needs a guess about the others, and at a table of bots everyone declines until the last.
   - (c) Configurable, a `heuristic.Config` switch on in `DefaultConfig()`, off in `BaselineConfig()`. Measurable in the arena, at the cost of one more switch.

8. **Can an assist caster ask a bot (§4, §6)?**
   - (a) **Recommended:** yes. Any other player still in the game can be chosen (CR 702.132a), and the picker marks bot seats with "(always declines)", since the heuristic answers 0.
   - (b) Only seats played by people. Saves a wasted click, but it hides a legal choice and an MCP seat (ADR 0122) might help.
   - (c) Yes, and the bot sometimes helps (for example, a symmetrical Game Plan or Huddle Up that names it). That needs a new `Purpose` reading for little gain.

9. **Delivery.**
   - (a) **Recommended:** three PRs as listed: mana offer, life (and the question 3 cards), assist. Each is reviewable alone, and assist doesn't wait on the offer.
   - (b) Two PRs: the whole any-player offer, then assist.
   - (c) One PR for both seams. Fewer reviews, but a large diff across engine, protocol, client and bot.
