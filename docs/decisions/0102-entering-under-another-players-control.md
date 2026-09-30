# ADR 0102 — A permanent that enters under another player's control

**Status:** Accepted · 2026-09-30 · Post-S30 — Rolling deck-driven catalog growth
**Issue:** [#1759](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1759). It was found building ADR 0097's card batch (#1749). It relates to #1745, gaining control of a spell.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune` and read every `docs/decisions/` file name on all 37 remote heads. Numbers 0098–0101 are taken by ADRs written in parallel: 0098 (discard as a permanent would enter), 0099 (discover), 0100 (delve, either-or and variable sacrifice costs) and 0101 (keyword counters). No branch has 0102, so this one takes **0102**.
**Builds on:** [ADR 0013](0013-replacement-effects.md) (the CR 614 window, its paused entries and the CR 616 ordering prompt), [ADR 0061](0061-token-creation-and-discard-are-replaceable-events.md) (token creation), [ADR 0063](0063-durations-and-control.md) (layer-2 control), [ADR 0043](0043-copy-effects.md) (entering as a copy), [ADR 0060](0060-leaving-the-game.md) (CR 800.4a), [ADR 0033](0033-ai-bot-seat.md) (the bot) and [ADR 0041](0041-game-persistence.md) (restore points).
**Owner decisions:** 2026-09-30. All eight open questions are answered; see [Owner decisions](#owner-decisions-2026-09-30) at the end.

---

## Context

Issue #1759 names three kinds of card that need a permanent to end up under a player other than the one whose card or spell put it there:

1. **Self-entry under an opponent.** Captive Audience: "This enchantment enters under the control of an opponent of your choice."
2. **Another player creates a token.** The Hunted cycle ("target opponent creates …"), Beast Within ("its controller creates …") and similar cards.
3. **An effect puts a card onto the battlefield under another player's control.** Kenrith ("… under its owner's control"), The Beamtown Bullies ("target opponent … puts … onto the battlefield under their control").

I checked each kind against the code and the Scryfall dump. **Kinds 2 and 3 already work. Only kind 1 is missing.** The seam is narrower than the issue suggests.

### Kind 2 works: another player creates a token

CR 111.2 says the player who creates a token owns it, and it enters under that player's control. The engine does this today:

- `effects.CreateToken{Controller: …}` passes the creating player to `CreateTokenForEffect`.
- `mintTokenLocked` (`game/token_create.go`) sets both `Owner` and `Controller` to that player.
- The creation event's `TokenController` is the CR 616 affected player for token doublers (`affectedPlayerForEvent`).
- Each token's entry event carries that player as `Actor`, so the permanent lands under them and their "whenever a creature you control enters" triggers fire.

About 180 cards name another player, or each player, as the token creator. 25 of them are already catalogued: Beast Within, Forbidden Orchard, Generous Gift, Pongify, Rapid Hybridization, Swan Song, Terastodon, Saw in Half and others.

**The Hunted cycle is not waiting on this ADR.** That is Hunted Horror, Hunted Troll, Hunted Phantasm, Hunted Lammasu, Hunted Dragon and Hunted Bonebrute. Each one is an ETB trigger with a "target opponent" clause over `CreateToken{Controller: target}`. The only missing parts are rows in `tokens_table.go`: a 3/3 green Centaur with protection from black, a 1/1 blue Faerie with flying, a 4/4 black Horror, a 2/2 white Knight with first strike and a 1/1 white Dog. Those rows ship with the cards.

### Kind 3 works: an effect puts a card under another player's control

CR 110.2a says an object an effect puts onto the battlefield enters under the control of the player the effect instructs, "unless the effect states otherwise". The engine already takes the stated player at every effect-side entry:

- `ZoneEntryOptions.Controller` (`game/battlefield_put.go`) covers the hand, library and command-zone puts and `PutOntoBattlefieldTogetherThenForEffect`.
- `ReturnFromGraveyard.Controller` covers the graveyard, where zero means "its owner".
- The `controller` argument of `ReturnToBattlefieldForEffect` and `ReturnFromExileToBattlefieldForEffect` covers exile.

The batch door stamps each card's controller in its source zone before the first CR 614 window opens. The comment says why: "so a replacement consulted for one card of the batch reads the right controller on the others (Authority of the Consuls asks whose permanent is entering)" (`entry_batch.go`).

Coverage in the dump:

- About 10 cards say "put … onto the battlefield under its owner's / their control": Kenrith, Bilbo, Trove Warden, Game Preserve, Savior of Ollenbock, Glyph of Reincarnation, Caldera Breaker, The Beamtown Bullies, Endless Whispers and Dubious Challenge.
- About 257 say "return … to the battlefield under its owner's control". 19 of those are catalogued.

None of these needs new machinery.

### Kind 1 is the gap: a permanent that says whose control it enters under

Four cards print this. All four are legal in Commander and none is catalogued:

| Card | Text | Also needs |
|---|---|---|
| Captive Audience | "This enchantment enters under the control of an opponent of your choice." | nothing (ADR 0097 shipped its modes) |
| Pendant of Prosperity | "This artifact enters under the control of an opponent of your choice." The ability then names "This artifact's owner". | nothing |
| Abby, Merciless Soldier | "When you cast this spell, create … tokens … equal to the amount of mana spent to cast it. Abby enters under the control of an opponent of your choice." | nothing in the 99. As a *Partner—Survivors* commander, the deck validator refuses it (`hasUnsupportedPhrase`). |
| Xantcha, Sleeper Agent | "Xantcha enters under the control of an opponent of your choice." | "Any player may activate this ability", and "can't attack its owner or planeswalkers its owner controls". Neither exists. |

Two more cards replace the entering controller from *another* source. They share the rule this ADR builds, but they are out of scope unless the owner says otherwise (open question 1):

- **Gather Specimens:** "If a creature would enter the battlefield under an opponent's control this turn, it enters under your control instead."
- **Crafty Cutpurse:** "… each token that would be created under an opponent's control this turn is created under your control instead."

Two neighbours look similar but already work and are **not** this seam. Akroan Horse ("When this creature enters, an opponent gains control of it") and Sleeper Agent are an ETB trigger over `GainControl`. They really do enter under the caster, whose enters triggers see them, and then change control. That is the printed difference from Xantcha.

### Today a permanent spell always enters under its caster

A permanent spell's entry event is built in `resolveTopOfStackLocked` (`game/mutations.go`) with `Actor: item.Controller`. `StackItem.Controller` is fixed at cast, and nothing rewrites it (#1745). `landEntryLocked` (`game/entry_choice.go`) then writes `Card.Controller = ev.Actor`, or the owner when there is no actor. **So a permanent spell's controller is always the permanent's controller at entry.** This ADR adds the first exception.

### How the entering controller is carried today

The "would-be controller" (the player the permanent is about to enter under) has two spellings during the window. They agree only because every door writes both:

- **`ev.Actor` on the `RepEventMove`.** `EnteringPermanentChooser` (the shockland payer, the reveal-land chooser, the Clone chooser), `SelfEntersTappedUnless` (checklands), the `EntersAsCopyOf` candidate list and the landing itself all read it. The helper's comment says why: "the battlefield-entry path stamps Card.Controller only AFTER the replacement pipeline has run".
- **The entering card's `Card.Controller` in its source zone.** Authority of the Consuls, Kismet and the CR 616 chooser for a move (`affectedPlayerForEvent`, `case RepEventMove`) all read it.

`ev.Actor` also means "who did it" in one place. The land-play branch bumps `LandsPlayedThisTurn[ev.Actor]`, and `EventZoneMove` and `EventETB` copy `ev.Actor` as their actor.

Other facts this design relies on:

- `Card.BaseController` is captured lazily by the first layer pass after entry, from `Card.Controller`. It is cleared on the way out. That makes it CR 110.2's "player under whose control it entered", with no write site to change.
- `SummonedThisTurn` is stamped on every entry and cleared in the untap step, for the active player's permanents. So a creature is sick until its controller's next turn, as CR 302.6 requires.
- **CR 616.1's sub-steps (616.1a–e) are not implemented.** The ordering prompt offers every applicable effect as an equal (`queueReplacementOrderPromptLocked`).
- `ChoosePlayerAsEnters` (#980, True-Name Nemesis) asks from the `AsEnters` hook, which runs *after* the permanent has landed and `EventETB` has fired. That is acceptable for protection. It is too late for control, because by then the wrong player's enters triggers have already fired.

### What the rules and the rulings say

Every rule below was checked against the pinned text (`MagicCompRules 20260819.txt`):

- **CR 110.2.** A permanent's controller is, by default, the player under whose control it entered the battlefield. **CR 110.2a** is the effect-side rule quoted above. **CR 110.2b** separates "controls the permanent that spell becomes" from the default controller, which is the player who put the spell on the stack. That becomes relevant with #1745, not here.
- **CR 108.3.** Ownership is who started the game with the card. Nothing in this ADR changes an owner.
- **CR 614.1d.** "[This permanent] enters …" is a replacement effect. **CR 614.12** adds that such an effect may come from the permanent itself. To decide which apply, "check the characteristics of the permanent as it would exist on the battlefield, taking into account replacement effects that have already modified how it enters". **CR 614.12a:** "If a replacement effect that modifies how a permanent enters the battlefield requires a choice, that choice is made before the permanent enters the battlefield."
- **CR 616.1.** The affected object's controller chooses the order. **CR 616.1b:** "If any of the replacement and/or prevention effects would modify under whose control an object would enter the battlefield, one of them must be chosen." That comes after self-replacement effects (616.1a) and before copy effects (616.1c). **CR 616.1f** repeats the process with only the effects that still apply. **CR 614.5** applies each effect once.
- **CR 603.3a.** A triggered ability is controlled by whoever controlled its source when it triggered. **CR 603.6a** checks every permanent, "including the newcomers", for ETB triggers.
- **CR 302.6.** A creature can't attack or use {T} unless its controller has controlled it continuously since their most recent turn began.
- **CR 800.4a.** When a player leaves, their owned objects leave the game, then anything they still control is exiled.

The official rulings add the rest:

- **Xantcha (2018-07-13):** "Xantcha's first ability is a replacement effect that modifies how it enters the battlefield, not a triggered ability. Players can't take actions … while Xantcha's on the battlefield before it's controlled by another player." Also: "If a player creates a token that's a copy of Xantcha, the player who creates the token is its owner, not the player under whose control it enters the battlefield."
- **Captive Audience (2019-01-25):** once an opponent controls it, "its ability triggers during that player's upkeep and that player makes all choices for it." And: "if Captive Audience's owner leaves the game, Captive Audience leaves the game with them. If Captive Audience's controller leaves the game, Captive Audience is exiled." Pendant of Prosperity has the same pair of rulings.
- **Gather Specimens (2008-10-01)**, the canonical CR 616.1b example: its effect "is applied before any other replacement effects that would also modify how the creature enters." "If Clone would enter under an opponent's control, you choose which creature it copies as it enters under your control." "Any 'enters' triggered abilities will trigger after the creature is on the battlefield under your control." With two copies, "the creature's would-be controller … chooses one … Then the new would-be controller of the creature repeats this process among the remaining" effects. CR 614.5 makes that terminate.

The ordering is observable. Here is a four-player example, with Kismet ("Artifacts, creatures, and lands your opponents control enter tapped") controlled by player C:

- A casts Xantcha and chooses C.
- CR 616.1b applies the control change first. The creature now enters under C, so Kismet no longer applies, and Xantcha enters untapped.
- If Kismet were ordered first, as today's equal-weight prompt allows, Xantcha would enter tapped under C. That is wrong.

---

## Options

### A. Gain control after entry (the Akroan Horse shape) — rejected

This would enter the permanent under its caster and then register a layer-2 `GainControl` record. It is a different card:

- The caster's "whenever an enchantment enters under your control" triggers fire. The chosen opponent's do not.
- `EventControlChanged` fires, so Khârn-style triggers would see a theft that never happened.
- For a short window the caster controls the permanent, which the Xantcha ruling rules out.
- `BaseController` becomes the caster, so the permanent would go back to the caster if the control record ever ended.

### B. Choose on cast, and resolve with the stored answer — rejected

This would stamp the chosen opponent on the `StackItem` at announce and seed `Actor` from it at resolution. It fails in three ways:

- It covers casting only. Xantcha reanimated, blinked, put by Genesis Wave or copied by a Clone must ask as well. The ability is on the permanent, not on the spell.
- CR 614.12a puts the choice at entry, not at cast. The chosen player can leave while the spell is on the stack.
- It leaves a Gather Specimens cast in response with nothing to interact with.

### C. Ask from the `AsEnters` hook and rewrite the controller afterwards — rejected

This is `ChoosePlayerAsEnters` plus a write. The answer arrives after `EventETB`, so it has Option A's trigger problems without the event. It also creates a window in which the caster controls the permanent.

### D. An entry replacement with a controller selector — recommended

The ability becomes what the rules say it is: a CR 614.1d replacement effect from the permanent itself. It is asked inside the CR 614 window, ordered first under CR 616.1b, and it rewrites the would-be controller that every later effect in the window reads. The rest of this ADR describes it.

---

## Decision

### 1. A declaration on `ReplacementEffect`, built by one constructor

Add a selector next to `CopySelector` and `EntryHandReveal`: `ReplacementEffect.EntryController *EntryControllerChoice`. Card files never build it by hand:

```go
Replacements: []game.ReplacementEffect{
    EntersUnderTheControlOfAnOpponentOfYourChoice("Captive Audience", game.ControlForHarm),
},
```

The constructor writes a self-scoped `AppliesTo` (`ev.CardID == src.InstanceID`, `NewZone == ZoneBattlefield`), the selector, and the CR 616.1b tier flag from Decision 3. It watches `EventZoneMove` like every other entry replacement.

Because the declaration is catalog data on the card's own row, it follows the card into every entry:

- a cast;
- a put, return or blink;
- a token copy (the Xantcha ruling);
- a Clone that copied it (Decision 6).

`asksItsOwnQuestion` learns the new selector, so the "nobody can answer" and must-settle paths treat it like the other entry prompts.

### 2. One would-be controller, written in both places

When the effect applies, it sets the chosen player as the would-be controller:

- It rewrites **`ev.Actor`**. That is the field the landing, `EnteringPermanentChooser`, the checkland conditions and the copy candidates already read.
- It **re-stamps `Card.Controller` on the entering card in its source zone**, through the existing `setControllerInZoneLocked`. For a token, which comes from no zone, it re-stamps the staged copy in `Game.enteringTokens` instead. That is the field Authority of the Consuls, Kismet and `affectedPlayerForEvent` read.

It remembers the prior value and puts it back if the entry does not happen. That is the batch's `priorController` rule. `StackItem.Controller` is never touched: the *spell* is still the caster's while it resolves, which is CR 110.2b's distinction.

The one reader for whom `Actor` means "who did it" is the land-play tally. It captures its player on the event before the window opens (an unexported `landPlayer`), so the two meanings cannot drift apart. No printed land has this clause today; this change is insurance.

The alternative is a new exported field, `ReplacementEvent.EntersUnder`, with every reader moved over to it. That is cleaner but touches about a dozen readers. It is open question 2.

### 3. CR 616.1b is ordered first, without a prompt

A constructor-set flag on `ReplacementEffect`, `ChangesEntryController`, marks the tier. It is declared, never inferred from `Replace`. The apply-loop narrows the applicable set before its existing branches:

- If any applicable effect has the flag, only flagged effects are candidates on this pass.
- With exactly one, it is applied, and no ordering prompt is shown.
- With several (two Gather Specimens, or Xantcha plus Gather Specimens), the existing `replacement_order` prompt is asked over the flagged effects only. It goes to the would-be controller, because Decision 2 has made `affectedPlayerForEvent` read the right player.
- CR 616.1f's re-gather then runs as it does today. It sees the new controller, so Kismet drops out and Authority of the Consuls comes in. The once-per-event map (CR 614.5) is what ends the two-Gather-Specimens back and forth.

616.1a (self-replacements), 616.1c (copy) and 616.1d (back face up) are not tiered by this ADR (open question 5). Adding only 616.1b is enough for the four cards and for the Kismet example.

### 4. The prompt: `entry_controller`, asked as the permanent would enter

**When.** Inside the CR 614 window (CR 614.12a):

- For a spell, that is at resolution, not at cast (Option B).
- For a put, return, blink or token copy, it is at that entry.

The permanent is not on the battlefield while the question is open, so no player can act "while Xantcha's on the battlefield before it's controlled by another player".

**Who and what.**

- The chooser is the would-be controller as the effect applies: normally the caster, or for "return it under your control" the player who returned it.
- The options are the chooser's opponents still in the game (CR 102.3), listed in turn order starting after the chooser. The answer is a seat ID.
- The prompt names the entering card and the effect's purpose (Decision 8).
- It resumes through the pipeline's existing contract: a `replacementResume` frame, then `finishSettledReplacementLocked` and the landing. That is the same path `copy_target` and `entry_reveal_from_hand` take.

It is a **new kind** rather than `option_pick` for `entry_reveal_from_hand`'s reason: its continuation is a *paused event*, not a card's next sentence. `ChoosePlayer`'s seat frame speaks the option-pick chain's contract, and bending that to resume an event is the coupling #1198 declined.

**Edge cases, each a declared rule:**

- **One eligible opponent, as in any two-player game:** the choice is forced, so no prompt is shown and that opponent is used (open question 3).
- **No eligible opponent:** the effect does nothing, and the permanent enters under the would-be controller. In a live game this cannot happen, because the last player standing has already won.
- **An event that cannot pause (`mustSettleNow`):** the replacement is mandatory, so it is **not** skipped the way an optional question is. The engine picks the first eligible opponent in turn order after the chooser (open question 4). The only entry that cannot pause today is the sandbox `move_card` verb, which is a manual move. So this is mostly a rule written down in advance.

**The three answers #730 / #794 / #902 require of a new kind:**

- **Choice gate (`choiceGateDecisions`): blocks the table.** An entry is waiting on it, as with `copy_target`.
- **Enumerator (`choiceMoves`):** one move per offered seat. Every offered seat is a legal answer, and the queue refuses to raise the prompt with no options, so it cannot become the #544 wedge.
- **Departure (`choiceDepartureDecisions`): dropped.**
  - If the chooser leaves, the entering card is theirs and has left the game with them (CR 800.4a step 1), or the effect putting it has ceased to exist. The entry is abandoned and the card stays where it was, as `pausedZoneChangeStaleLocked` would leave it.
  - If an *offered* opponent leaves, `pruneDepartedSeatOptionsLocked` removes that option. If none are left, the no-opponent rule applies.

### 5. Landing, ownership, sickness and triggers need no new code

`landEntryLocked` already writes `Controller = ev.Actor`. What follows from that:

- **Owner** is never written. Pendant of Prosperity's "This artifact's owner draws a card" and Xantcha's "can't attack its owner" read `Card.Owner` (CR 108.3).
- **Default controller (CR 110.2).** The lazy `BaseController` capture records the chosen player. If some later layer-2 theft ends, the permanent goes back to that player and not to the caster.
- **Summoning sickness (CR 302.6).** The entry stamps `SummonedThisTurn`. The flag clears in the untap step of the chosen player's own turn, so Xantcha can attack on that player's next turn. No control *change* happened, so `EventControlChanged` does not fire and `materialiseControlLocked`'s control-change stamp is never reached.
- **Triggers:**
  - The permanent's own triggers are controlled by the chosen player (CR 603.3a), since the harvester reads `source.Controller`. Captive Audience's "At the beginning of your upkeep" is that player's upkeep.
  - "Whenever an enchantment / creature enters under your control" fires for the chosen player, not for the caster. Every such predicate goes through `enteredUnderYourControl`, which reads the landed controller.
  - The caster's "whenever a creature an opponent controls enters" fires.
  - `EventZoneMove` and `EventETB` carry the chosen player as `Actor`.
  - Cast triggers stay the caster's: the spell was theirs. Abby's cast-trigger tokens are created under the caster.

### 6. Copies and tokens follow from CR 614.12

- **A Clone copying Xantcha.** CR 616.1c's copy applies first, because the control effect is not on the Clone yet. CR 616.1f then re-gathers against the permanent "as it would exist on the battlefield", and finds Xantcha's copied self-replacement. So the Clone enters under an opponent of its would-be controller's choice. This needs the re-gather to see catalog replacements on the *copied* oracle key for an entering card. Checking that is the first test in the implementation PR.
- **A token copy of Xantcha** (Kiki-Jiki, Followed Footsteps). The creation window runs first, and `mintTokenLocked` makes the creator both owner and would-be controller. The token's own entry window then runs Xantcha's effect, which rewrites the controller only. That gives the ruling's result: the creator owns it, and it enters under someone else.
- **Gather Specimens beside a self-entry card** (if it is in scope). Both effects are flagged. A self-entry effect applies only to its own permanent, and Gather Specimens only to creatures, so the two meet only on a creature with the clause, such as Xantcha or Abby. There the would-be controller orders the two, per the ruling.

### 7. Leaving the game needs no new code

`leaveGameObjectsLocked` already does what both rulings say:

- If the **owner** leaves, step 1 removes everything they own, Captive Audience included.
- If the **controller** (not the owner) leaves, step 2's recompute leaves the permanent with its `BaseController`, which is the departing player. Step 4, `exileStillControlledLocked`, then exiles it.

The implementation PR pins both rulings with a test each, and adds nothing else here.

### 8. The bot: one move per seat, scored by a declared purpose

The enumerator offers each opponent. A policy cannot tell from the move list alone whether giving a permanent away hurts or helps. So the prompt carries a **purpose** that the constructor declares, following #780's `color_purpose`:

- **`ControlForHarm`** (Captive Audience, Xantcha): the heuristic gives it to the **strongest** opponent by `SeatEval.Strength` (`aiseat/heuristic/score.go`). It deliberately does not use `Threat`, which ranks a seat by how close it is to dying. A player at 5 life loses almost nothing to "your life total becomes 4".
- **`ControlForBenefit`** (Pendant of Prosperity): the heuristic gives it to the **weakest** opponent.

The purpose rides the wire as `control_purpose` on the prompt, so `aiseat/` still imports nothing from `internal/game`. The model tiers see the prompt rendered with its purpose line. A new wire field in `aiseat` needs its `boteval suite render` case in the same PR.

### 9. The wire, the client and the log

- The prompt projects as `entry_controller` with the entering card, the seat options and the purpose.
- The client renders one button per seat, like the `option_pick` player chooser, headed by the card's own sentence.
- The answer is one new event, `EventEntryControllerChosen` (chooser, chosen player, card). Under #984's gate it owes an arm in `projectEvent`: "Alice chose Bob to control Captive Audience." The zone-move line alone would show only who controls the permanent, not that a choice was made. The card is public by the time it lands, so the line redacts nothing. Silencing it instead is open question 7.

### 10. Snapshot impact: none on the card, and the prompt is not a restore point

- **The card.** `Controller`, `Owner` and `BaseController` are already carried (`snapshot.go`). No new `Card` field is needed: "entered under this player" is exactly `BaseController`.
- **The declaration.** It is catalog data, rebuilt from the card's row by the running binary.
- **The open prompt.** It holds a `replacementResume` frame, which the census already counts under `ChoiceResumeFrames`. So a table with the question open is not a restore point, the same as a table with `copy_target` or `entry_reveal_from_hand` open. That is not a regression. It is the current rule for every paused entry.
- **The prompt's seat list.** This is a new `PendingChoice` field (`EntryControllerOptions []uuid.UUID`, plus the purpose). It is an **additive** change to `pendingChoiceSnapshot`, recorded with `-update-shape`. No `SnapshotSchemaVersion` bump.
- **The new event kind.** It is an ordinary `EventKind` value in the replay stream. No fixture changes.

### 11. Delivery

**PR 1, engine only.** Nothing in the catalog changes behaviour. It contains:

- the declaration and the constructor;
- the Decision 2 write-back and the `landPlayer` capture;
- the 616.1b tier;
- the `entry_controller` kind with its three table rows and its `choiceMoves` case;
- the purpose on the wire, the heuristic and the log line.

Its tests, each through the real entry door, not a hand-built event:

- a spell resolves under the chosen opponent, and that opponent's "enters under your control" trigger fires while the caster's does not;
- Kismet is ordered after the control effect;
- a blink asks again, and a reanimation "under your control" is asked by the reanimating player;
- a token copy is owned by its creator;
- a Clone copying Xantcha is asked (Decision 6);
- a departed controller's copy is exiled and a departed owner's copy leaves the game;
- the enumerator offers only seats still in the game;
- no `EventControlChanged` fires.

**PR 2, cards.** Captive Audience, Pendant of Prosperity and Abby, Merciless Soldier (in the 99), with their oracle fixtures. The same PR flips the registry row `enters-under-an-opponents-control` to implemented, with a closed-seam fragment. Xantcha is added to the row's `Waiting` list behind its other two blockers, which get rows of their own.

---

## Cards unblocked

| Card | By this ADR | Notes |
|---|---|---|
| Captive Audience | **yes** | This clause is the whole blocker (#1759). |
| Pendant of Prosperity | **yes** | "Put a land card from your hand onto the battlefield" and "owner draws" both exist. |
| Abby, Merciless Soldier | **yes, in the 99** | The cast-trigger count reads `ManaSpentToCast`. As a *Partner—Survivors* commander it is still refused by deck validation. |
| Xantcha, Sleeper Agent | partly | It also needs "any player may activate" and "can't attack its owner". |
| Gather Specimens | only if in scope | It reuses Decision 3's tier. It needs a turn-scoped `ScopedEffect` replacement mod kind (ADR 0041 phase 3 says it must be data). |
| Crafty Cutpurse | not by this ADR | It rewrites `TokenController` on the creation event. `mintTokenLocked` sets `Owner = Controller`, which would give the Cutpurse player ownership. Per CR 111.2 the creator should keep it, so the owner and controller need splitting first. |

The following already work today, and this ADR does not change them:

- **The Hunted cycle** (6 cards). It needs only token rows.
- **About 180 "another player creates a token" cards.** 25 are catalogued.
- **About 267 "put / return … under its owner's / their control" cards.** 19 are catalogued.
- **Akroan Horse and Sleeper Agent.** These are an ETB trigger over `GainControl`, and they enter under the caster as printed.

---

## Consequences

- A permanent spell no longer always enters under its caster. Any new code that reads "who controls this permanent" during the entry window must use the would-be controller (`ev.Actor`, or `EntersUnder` if question 2 goes that way), never `StackItem.Controller` or the permanent's owner.
- The CR 616 ordering prompt stops treating every effect as an equal for the first time. The tier mechanism is general, so 616.1c and 616.1d can be added the same way later.
- There is one more entry-window prompt that holds a table out of restore points while it is open. That is the same trade every paused entry already makes.

## Out of scope

- **Gaining control of a spell** (#1745) and CR 110.2b's "default controller of a permanent from a stolen spell".
- **Tribute** (12 cards: "an opponent of your choice may put N +1/+1 counters on it"). It uses the same pool of opponents, but it is a choice about counters, not about control.
- **Desertion's** "put that card onto the battlefield under your control instead". This is a counter-spell redirect. It is already expressible through the effect-side `Controller` doors.
- **CR 616.1a, 616.1c and 616.1d** ordering tiers.

---

## Open questions for the owner

1. **Scope.** Should PR 1 cover only the four self-entry cards, or also Gather Specimens, the rules' own example of CR 616.1b, which would need a turn-scoped replacement mod kind? Crafty Cutpurse is recommended out, because it first needs owner and controller split at token mint.
2. **Where the would-be controller lives.** Should the effect rewrite `ev.Actor` and re-stamp the card in its source zone (recommended: small, and every current reader already agrees with it)? Or should `ReplacementEvent` get an explicit `EntersUnder` field, with every reader moved over to it (cleaner, about a dozen call sites)?
3. **Two-player games.** With only one eligible opponent, should the engine skip the prompt and use that opponent (recommended), or ask anyway so the choice is visible as a click?
4. **An entry that cannot pause.** Should the mandatory choice default to the first opponent in turn order after the chooser (recommended), or should the effect be skipped the way optional entry questions are? No entry reaches this today.
5. **The other CR 616.1 tiers.** Should self-replacement (616.1a), copy (616.1c) and back-face (616.1d) ordering be added with 616.1b now, or later when a card needs them?
6. **The bot's pick.** Should "give it away to hurt" go to the opponent with the highest `SeatEval.Strength` (recommended), or to the seat the heuristic's `Threat` ranks highest?
7. **The log.** Should the choice get its own narrated line (recommended: "Alice chose Bob to control Captive Audience"), or be silent, with the zone-move line left to show the new controller?
8. **Xantcha's other two blockers.** Should "any player may activate this ability" and "can't attack its owner or planeswalkers its owner controls" each get a registry seam row now, so Xantcha has somewhere to wait?

---

## Owner decisions (2026-09-30)

1. **Scope: the four self-entry cards only.** These are Captive Audience, Pendant of Prosperity, Abby, Merciless Soldier and Xantcha, Sleeper Agent. Gather Specimens and Crafty Cutpurse stay out. The `ChangesEntryController` tier is still built general, so either can join later without a second change to the ordering.
2. **The would-be controller lives in `ev.Actor`.** The effect rewrites `ev.Actor` and re-stamps `Card.Controller` on the entering card in its source zone, or on the staged token. It restores the prior value if the entry does not happen. There is no `EntersUnder` field.
3. **One eligible opponent: no prompt.** The choice is forced, so the engine uses that opponent.
4. **An entry that cannot pause: the first opponent in turn order after the chooser.** The replacement is mandatory and is never skipped.
5. **The other CR 616.1 tiers wait.** 616.1a, 616.1c and 616.1d are added when a card needs them.
6. **The bot gives a harmful permanent to the opponent with the highest `SeatEval.Strength`,** and a helpful one to the opponent with the lowest.
7. **The choice gets its own narrated log line:** "Alice chose Bob to control Captive Audience."
8. **Xantcha's other two blockers get registry seam rows and issues:** "any player may activate this ability" and "can't attack its owner or planeswalkers its owner controls". Xantcha is listed in both `Waiting` lists.

---

## Amendment (2026-09-30, #1759): what PR 1 built, and where it differs

PR 1 is the engine half. Three details differ from the text above, all
in the direction of reusing what exists:

- **The seats ride `PickOptions`, not a new `EntryControllerOptions`
  field.** An `option_pick` over players already carries seats as
  `ChoiceOption{Label, Player}`, and that list is already snapshotted,
  cloned for undo, projected to the wire as `pick_options` and pruned
  when a seat leaves (`pruneDepartedSeatOptionsLocked`, which now visits
  this kind too). The only new `PendingChoice` field is
  `ControlPurpose`, recorded in the snapshot shape as an additive field.
  The answer is `{option_index}`, the same payload `option_pick` takes.
- **An offered seat that leaves the game while the prompt is open** is
  pruned. If every offered seat has gone, the paused entry is resumed
  with no control change (the no-opponent rule, reached late) rather
  than dropped.
- **Decision 6's copy-aware gather is general.** Once a Clone has chosen
  what to copy, the self-replacement gather also collects the COPIED
  card's entry replacements (CR 614.12), in their own ID stride so a
  copied slot is never mistaken for the Clone's own applied selector. A
  copied copy selector is not collected. This is what makes a Clone
  copying Abby enter under an opponent; it equally makes a Clone copying
  any creature with its own "enters tapped" or "enters with" clause
  honour it.

The log line is `choose_controller` ("P1 chose P2 to control Captive
Audience"), from the new `EventEntryControllerChosen`. It is emitted
however the choice was reached, a forced one included.
