# ADR 0127 — Answering repeated prompts for you

**Status:** Proposed · 2026-10-07 · S59 — Automated table: clicks that act, payment that counts (tracker [#2189](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2189))
**Issues:** [#1961](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1961) (this change). Related: [#1968](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1968) (accept the trigger order automatically, S60) and [#1530](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1530), whose item 2 shipped as PR [#2334](https://github.com/krakenhavoc/cmd_and_ctrl/pull/2334) (`Player.TriggerOrderAlwaysAsk`).
**Owner decisions:** the three answers of 2026-10-07 on #1961, quoted under [Owner decisions](#owner-decisions-2026-10-07). They are binding. Everything else below is a call this ADR makes, and the calls the owner should rule on are numbered under [Questions for the owner](#questions-for-the-owner).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-07. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote head: `origin/develop`, `origin/main`, `origin/docs/issue-audit`, `origin/feat/750-conditional-block-restrictions`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default` and `pr/2326`. The highest number on any of them is 0126 (`0126-bots-that-play-their-decks.md`). No open pull request adds an ADR. This ADR takes **0127**.
**Builds on:** [ADR 0018](0018-triggers-on-the-stack.md) (triggers, the trigger prompt, and the 2026-10-05 amendment for #1530 item 2), [ADR 0055](0055-loop-breaker.md) (the CR 732 loop breaker), [ADR 0110](0110-remember-me.md) §4 (settings on the account), [ADR 0111](0111-action-dock.md) §2 (small prompts inline in the dock) and §10 (labels), [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) (strict payment and the auto-tapper's pool top-up), [ADR 0119](0119-a-stack-you-can-follow.md) §2 (the stack hold), [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §2 (the label registry).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

Some cards ask the same question many times a game. An opponent's Rhystic Study asks "pay {1}?" on every spell you cast. Your own Consecrated Sphinx asks "draw two cards?" on every card an opponent draws, so a wheel asks seven times in a row. The reporter on #1961 wants to answer once: "no, yes, and yes if able".

Every claim below was checked on `origin/develop` at `c032a7a1`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

### Owner decisions (2026-10-07)

On #1961:

1. **It fits strict payment.** Write the ADR first.
2. **Ask, Always or Never.** Each player picks Ask, Always or Never for each prompting card.
3. **"Always pay" pays only when auto-pay can cover the cost, and asks otherwise.**

### The rules

- **CR 603.5:** an optional triggered ability goes on the stack "regardless of whether their controller intends to exercise the ability's option or not. The choice is made when the ability resolves." The same rule says the "unless" part of an ability "is dealt with when the ability resolves."
- **CR 118.12:** in "[A player] may [do something]. If [that player] [does, doesn't, or can't], [effect]", the action "is a cost, paid when the spell or ability resolves." **CR 118.12a:** "[Do something] unless [a player does something else]" means the same as "[A player may do something else]. If [that player doesn't], [do something]." Rhystic Study's tax is this cost.
- **CR 118.3:** "A player can't pay a cost without having the necessary resources to pay it fully." An "Always pay" that cannot be covered is not an answer the player can give, so the server asks instead (owner decision 3).
- **CR 605.3a:** a player may activate a mana ability "whenever a rule or effect asks for a mana payment, even if it's in the middle of casting or resolving a spell or activating or resolving an ability." This is what lets the auto-tapper pay a tax during a resolution.
- **CR 608.2d:** choices an effect offers are announced "while applying the effect". A player "can't choose an option that's illegal or impossible", except that an empty library does not make drawing impossible (CR 121.3). Drawing from an empty library loses the game the next time state-based actions are checked (CR 704.5b).
- **CR 732.1 and 732.1a:** players "typically make use of mutually understood shortcuts rather than explicitly identifying each game choice", and "As long as each player in the game understands the intent of each other player, any shortcut system they use is acceptable." A standing answer is such a shortcut. That is why every automatic answer is written in the game log for the whole table (§6).
- **CR 800.4f:** a player who has left the game does not pay. Nothing here changes it: a departed seat has no prompts to answer.

### What exists

**How the prompts are asked.** Four kinds carry the yes/no questions this issue is about.

| Kind | Asked when | Example | Queued by | Answered by |
|---|---|---|---|---|
| `trigger_prompt` | before an optional trigger goes on the stack | Consecrated Sphinx: "draw two cards?" | `queueTriggerPromptLocked` (`game/pending_choice.go:3049`), from a `TriggeredAbility.OptionalPrompt` (`effects.Optional`, `triggers_common.go:123`) | `ResolveTriggerPrompt` (`:3386`) |
| `pay_unless` | during the resolution | Rhystic Study, Smothering Tithe, Esper Sentinel, Mystic Remora's tax; also `MayPay` ("you may pay {2}. If you do") | `queuePayUnlessLocked` (`:3471`), `queueMayPayLocked` | `resolvePayUnless` (`:3810`) |
| `confirm` | during the resolution, or in a branch of an earlier answer | Rhystic Study's "draw a card?" (`MayChoice`, `effects/may_choice.go`) | `QueueConfirmForEffect` (`game/chained_choice.go:251`) | `ResolveConfirm` (`:400`) |
| `commander_return` | a state-based action (CR 903.9a) | "put it into the command zone?" | `commanderReturnSBALocked` (`game/commander_return.go`) | its resolver |

Rhystic Study is two prompts to two players: the caster is asked to pay, and if they do not, the Study's controller is asked whether to draw (`effects/rhystic_study.go`). The second is queued from inside the first's decline branch, after the trigger has finished resolving.

**The engine asks CR 603.5's question early.** A `trigger_prompt` is asked when the ability triggers, and a "No" drops the trigger before it reaches the stack (`ResolveTriggerPrompt`). CR 603.5 puts it on the stack and asks at resolution. This ADR does not change that; see [Out of scope](#out-of-scope) and question 7.

**What a pay-unless answer does.** `resolvePayUnless` dequeues the prompt first. On "Pay" it calls `payCostLocked` (`:3984`), which pays from the pool and auto-taps the rest. If that fails it taps and spends nothing, and the answer becomes a decline: the "unless" branch runs. So a "Pay" that cannot be covered is a "Don't pay" today. A non-mana payment (`PayAction`, discard or sacrifice; `game/pay_unless_action.go`) needs the payer to name cards.

**Facts on the prompt that say the answer depends on the board.** Three fields already exist, and the engine reads each of them live:

- `GuardsStackItem` (`pending_choice.go:838`): the spell on the stack whose fate the answer decides. Set for ward, Mana Leak, Daze and Dazzling Denial (`game/counter_unless_paid.go`).
- `OwedInStep`: the step the answer is owed in. Set for upkeep pay-or-else (Stasis, Pact of Negation, cumulative upkeep including Mystic Remora's, echo; `game/upkeep_pay_unless.go`) and for Hellkite Charger's in-combat may-pay.
- `PayAction`: a non-mana payment.

**A decision resets the loop breaker.** `dequeueChoiceLocked` (`:1446`) removes the prompt and calls `notePlayerDecisionLocked` (`game/loop_breaker.go:270`), because "answering a prompt is a player decision, so it restarts the CR 732 loop run". `removeChoiceAtLocked` (`:1495`) removes a prompt without that. While `LoopNotice` is set, automatic passing stops for clients and bots alike.

**What the resolving item knows.** `g.resolving` (`game/game.go:735`, `game/resolving_item.go`) is the item resolving now. A catalog ability's stack item names its row as `Params.Ability`, an `AbilityRef` of `{key, slot, ref, name}` (`game/ability_ref.go:74`, ADR 0041 P9). A trigger prompt's frame carries the whole `TriggeredAbility` and the source card (`triggerResumeFrame`).

**A per-seat preference the server holds.** PR #2334 added `Player.TriggerOrderAlwaysAsk` (`game/player.go:221`). It is set by the `set_trigger_order_preference` action (`actions/actions.go:176`), which mints no undo entry, and the enumerator never offers it. It is cloned, snapshotted additively as `triggerOrderAlwaysAsk`, carried across an undo by `RestoreFrom` (`game/clone.go:982-988`), and shown only on the seat's own `PlayerView` (`protocol/view.go:1534`). The client keeps the setting in `gameplay.alwaysAskTriggerOrder` (synced) and reconciles it against the seat's view on every frame (`client/src/lib/triggerOrderPref.ts`), so a reconnect, a second device or a server restart all converge by one rule.

**Commits and undo.** `Room.apply` (`server/internal/ws/room.go:282`) clones the game, runs the action, pushes an undo entry stamped with the caller, and captures a frame. Undo pops only the top entry, and only its caller (or the admin) may pop it (`:518-559`). An entry can be `freeUndo`, which skips the per-turn undo budget (`:185-201`, used for bot improvisations). Every commit goes through this room, including the bots' (`aiseat/runner.go:1126`).

**The client.** The yes/no prompts and `pay_unless` are drawn inline in the dock (`client/src/lib/choiceDock.ts`, `YES_NO_KINDS` at `:51`; ADR 0111 §2), with the buttons Yes / No and "Pay {N}" / "Don't pay". Smart auto-pass (`client/src/lib/autopassDecision.ts`) stops whenever the viewer has a pending choice (rule 1) and never answers one. The stack hold (ADR 0119 §2, `gameplay.stackHoldMs`) delays automatic *passes* only. Nothing in the client or the server answers a prompt for a human today.

---

## Decision

### 1. Which prompts are covered

A prompt is **covered** when one standing answer can be the whole answer, and nothing on the prompt says the answer depends on the board. The server decides this once, when it queues the prompt, from the prompt's own facts. Nothing is declared per card, so a new card that uses these primitives is covered without a change. The lesson of #951 (a per-card flag that card authors kept forgetting) does not apply here, because a missing key fails safe: an uncovered prompt is simply asked.

**Covered:**

| Kind | Covered when | Always means | Never means |
|---|---|---|---|
| `pay_unless` (mana) | `GuardsStackItem`, `OwedInStep` and `PayAction` are all empty, and the cost has no X | pay, when auto-pay can cover it (§5) | don't pay |
| `pay_unless` from `MayPay` | the same | pay, when auto-pay can cover it | don't pay |
| `trigger_prompt` | the ability has no target clause, no modes, and no `Trade`, and the prompt is not a miracle reveal | Yes | No |
| `confirm` | no `LifeCost`, and the labels are the default Yes / No (a "you may", not a choice between two named things) | Yes | No |

**Never covered**, because the right answer changes with the board or the answer is more than a yes or a no:

- **A pay-unless that guards a spell** (`GuardsStackItem`: ward, Mana Leak, Daze). Whether to pay depends on which spell is at stake.
- **An answer owed in this step** (`OwedInStep`: Stasis, Pact of Negation, cumulative upkeep, echo, Hellkite Charger). The cost grows, or the permanent, the combat or the game hangs on it.
- **A non-mana payment** (`PayAction`). It needs the payer to choose cards.
- **A targeted, modal or trading trigger** (Artisan of Kozilek's "you may return target creature card", Perplexing Chimera). The answer leads into a choice of target or a trade.
- **A cast** (`may_cast`: cascade, discover, miracle). A cast has its own costs, targets and timing.
- **A life payment** (`confirm` with `LifeCost`, `entry_pay_life` on a shockland). Life is the board.
- **Every pick**: cards, targets, colours, options, numbers, damage assignment, the trigger order, the CR 732 shortcut.

`commander_return` and the commander's `optional_replacement` are plain yes/no questions about one card. Whether they are covered is question 3; this ADR is written as though they are not, and covering them changes nothing else below.

### 2. The key: card and prompt

A preference is keyed by **card and prompt**, not by card alone. Rhystic Study asks two questions of two players, and one player can meet both: the tax from an opponent's Study and the draw from their own. "Rhystic Study: Always" cannot mean both "always pay" and "always draw".

The key is derived by the server and is opaque to the client:

- **A trigger prompt:** the source's catalog key and the triggered row's label: `<catalog key>|triggered|<row key>`. For example `…|triggered|Consecrated Sphinx — draw two cards`.
- **A prompt queued while a catalog ability resolves:** the resolving item's `AbilityRef` (key, slot, row label), then the prompt's kind and its ordinal among the prompts of that kind this resolution queued: `<ref key>|<slot>|<row label>|pay_unless#1`. A resolving spell uses its card's catalog key and `spell` in place of the slot and label.
- **A prompt queued from a branch of an answered prompt:** the parent prompt's key, then `>`, then the kind and ordinal: Rhystic Study's draw is `<the tax's key>>confirm#1`. To do this, each resolver that runs a branch records the answered prompt's key as the branch's origin while the branch runs.
- **Anything else** gets no key and is never covered: a prompt the engine cannot trace to a catalog row (an engine-built delayed or reflexive trigger, a miracle reveal), or one whose source is a face-down card.

The key never contains the question text. Cards build questions from changing numbers ("Esper Sentinel — pay {3}?"), and a key must survive that. A card that is later rewritten so that its row label changes loses its stored preferences. The prompt is then asked again, which is the safe direction.

Every prompt carries `auto_answer_key` on the wire, empty when not covered, and only on the chooser's view (§8). It also carries the card's name and the question the client shows next to the setting.

### 3. Where the preference lives

**In the account's synced settings, mirrored to each seat on the server.** This follows the pattern #2334 set for `TriggerOrderAlwaysAsk`.

- **The client setting.** `gameplay.autoAnswers`, a list of `{key, card, prompt, answer}` where `answer` is `"always"` or `"never"`. Ask is the absence of an entry. `card` and `prompt` are display copies, so Settings can list the rules without the server. The field is `"synced"` (ADR 0110 §4, `SYNCED_FIELDS`), so a rule set on one device applies everywhere the person is signed in. A guest keeps it in the browser only. It defaults to an empty list, and the shallow merge fills it in, so `SETTINGS_VERSION` does not need to change. The list holds at most **100** rules. Adding a 101st is refused with a message to remove one in Settings. 100 rules of about 200 bytes each fit inside the account's 32 KiB.
- **The server copy.** `Player.AutoAnswers`, a map from key to answer. It is set by a new action, `set_auto_answers {rules: [{key, answer}]}`, which replaces the whole map. The action is player-scoped (a seat sets only its own; the admin may set any), mints no undo entry (`MintsNoUndo`), is allowed during the opening roll, and is refused above 100 rules or with a key over 256 bytes. The enumerator never offers it, so no bot sends it and a bot seat's map is always empty. The MCP seat does not expose it.
- **Reconcile, not send on join.** The client compares its setting with the seat's `auto_answers` on every live frame and sends the action when they disagree, through the same shape as `triggerOrderPrefToSend` (one pure module, `autoAnswerPref.ts`, with a `lastSent` guard). This covers a change in Settings, a reconnect, a second device, and a server restored from an older snapshot.
- **Not per table.** There is no separate per-table rule (question 4). A player who wants different answers at a different table changes the rule; it takes effect at the next frame.

The server copy is what makes it work with the browser closed: once a seat has connected once, its rules are in the game, and they are in every restore point.

### 4. How the server answers

**The server answers, not the client.** An answer the client sent would not happen while the tab is closed, asleep on a phone, or reconnecting, which is exactly when a long run of Sphinx prompts would stall the table. The server already holds the prompt, the payment planner and the log.

**Each automatic answer is its own commit.** After every commit, while still holding the room lock, the room asks the game for the first prompt it may answer automatically (`Game.NextAutoAnswer()`), and answers it as a new commit:

1. Clone the game for the undo entry.
2. `Game.AutoAnswer(choiceID)` checks the prompt again and answers it through the kind's own resolver body, with one difference: it removes the prompt with `removeChoiceAtLocked`, **not** `dequeueChoiceLocked`, so an automatic answer is not a player decision and does not reset the loop run (§7).
3. Push an undo entry stamped with the **chooser**, marked `freeUndo` (question 5), and with the answered prompt's ID recorded on the entry.
4. Capture the frame.

The room repeats this until no prompt qualifies, then returns the last frame to the caller as today. Every replay line and dump is written. A prompt raised by a bot's commit, an admin's or a lobby step is answered the same way, because they all commit through the room. The room clears up within one call, so no goroutine, timer or subscriber is involved, and a lockstep arena replays the same answers in the same order.

**What qualifies.** A prompt is answered automatically only when all of these hold:

- it has an `auto_answer_key`, and the chooser's `AutoAnswers` has an answer for it;
- the chooser's seat is a human seat that has not left the game;
- the prompt is not marked `asked_by_hand` (below);
- the enumerator would list an answer to it for that seat now (`legal` `choiceMoves`), so the answer goes in the order the table would take it;
- `LoopNotice` is not set (§7);
- for **Always**: the chooser's library is not empty, since the engine cannot tell which Yes draws (question 6);
- for **Always** on a `pay_unless`: auto-pay can cover the cost (§5).

**Asked by hand.** When a prompt has a rule but fails one of the last three checks, the server marks it `asked_by_hand` and does not look at it again. The player answers it in the dock, with a line saying why ("Always pay: not enough mana"). Without the mark, a prompt the player is thinking about could be answered under them the moment they tap a land. The same mark is set when an automatic answer is undone (§6).

### 5. Always pay

"Always pay" pays only when auto-pay can cover the cost, and asks otherwise (owner decision 3).

**Can it be covered?** A new read-only check, `canPayCostLocked(p, cost)`, asks the same question `payCostLocked` answers, with the same spend context, the same cost-as-paid rules (Chromatic Orrery, #1600) and the same planner, including ADR 0118 PR 3's pool top-up. It changes nothing. If it says yes, the server answers "Pay" and `payCostLocked` pays: floating mana first, then auto-tapped sources, ranked as the auto-tapper ranks them (Treasures and once-per-turn sources last). If it says no, the prompt is asked by hand.

The server never sends the "Pay" that today becomes a decline when it fails. With strict payment on (ADR 0118), an automatic "Pay" spends real mana or nothing.

**Strict and permissive seats.** A seat with `strictMana` off still pays a tax from real mana: `payCostLocked` has no permissive mode. So "Always pay" works the same at both settings.

**Never pay** declines at once, and the "unless" branch runs as for a clicked "Don't pay".

### 6. Seeing and undoing an automatic answer

**The game log, for everyone.** A new event, `EventAutoAnswer`, carries the chooser, the source card, the prompt kind and the answer. It projects to a new log kind, `LogAutoAnswer`, rendered on the server:

- "Bob paid {1} for Rhystic Study (automatic)"
- "Bob didn't pay for Rhystic Study (automatic)"
- "Alice answered Yes to Consecrated Sphinx — draw two cards (automatic)"

The card's name is redacted per viewer as on any log line. The answer is public, because the answer is public in paper. "(automatic)" is public too, because CR 732.1a asks that the table understand each player's shortcut. The event is part of the snapshot as an additive field of schema 7.

**The prompt, for the chooser.** Every covered prompt gets one more control under its buttons in the dock: a toggle named **`Remember this answer`**. With it on, the button the player presses also sets the rule: Yes or "Pay" sets Always, No or "Don't pay" sets Never. The toggle is off each time a prompt appears. It is not shown on a prompt with no `auto_answer_key`. Its title names the card and the question ("Remember for Rhystic Study — pay {1}?").

**A notice after the answer, for the chooser.** When the server answers for the viewer, the dock shows a short notice for about six seconds, or until another request takes the dock: "Rhystic Study: paid {1} for you". It has two buttons, **`Undo`** and **`Ask me next time`**. The notice is a status, not a question, so it takes no keys and blocks nothing.

- **Undo** sends the ordinary `undo`. It works while the automatic answer is the top undo entry, which is the chooser's (§4). Undo restores the game from before the answer, so the prompt is open again, and the room marks it `asked_by_hand` so the next commit does not answer it again. The player then answers in the dock. Once another commit sits on top, Undo is greyed with "Someone has acted since". This is the same rule every undo follows.
- **Ask me next time** removes the rule (sets Ask) in the setting, which the reconcile sends to the server. It does not change the answer just given.

**Settings.** Settings → Gameplay gets a section named **`Automatic answers`**: one row per rule, with the card name, the question, a select (Ask, Always, Never) and **`Forget`**. Choosing Ask in the select removes the rule. An empty list says how to add one: "Tick Remember this answer on a prompt". All new names go in `client/src/lib/labels.ts` (ADR 0125 §2).

### 7. Bots, the stack hold, the loop breaker and #1968

**Bots do not use it.** A bot answers its own prompts through the enumerator, as it does today. A bot seat never has rules, and the room checks the seat's kind anyway. A human's automatic answer reaches a bot as an ordinary state change.

**The stack hold (ADR 0119 §2) is unchanged.** The hold delays an automatic *pass* while another seat's item is on top of the stack. An automatic answer is not a pass. A Rhystic Study trigger still sits on the stack for the hold before it resolves and asks for the tax. A Sphinx trigger that an automatic Yes puts on the stack is the Sphinx controller's item, so the other seats hold on it as they hold on any other. The automatic answer itself is not delayed. The prompt never reaches the screen, so there is nothing to read, and the log line is the record.

**The loop breaker (ADR 0055).** An automatic answer is automatic, like an automatic pass. It does not call `notePlayerDecisionLocked`, so it does not restart the loop run. Two Consecrated Sphinxes set to Always would otherwise draw both libraries out with nobody clicking. With this rule, the breaker's existing count fires on the run of Sphinx resolutions, and while `LoopNotice` is set the room answers nothing automatically. Each prompt is then asked by hand. A real decision, including answering one of those prompts by hand, clears the notice as today.

**#1968, accepting the trigger order.** Separate, and compatible. #1968 is a per-seat preference about CR 603.3b's order prompt, not about a card, and it builds on #1530 item 1 (S60). It should follow the same server-held pattern as `TriggerOrderAlwaysAsk` and this ADR. Automatic answers come first in time: an automatic Yes on an optional trigger puts it into the waiting batch, which the drain holds until every optional prompt is answered (#1529). So this ADR makes the order prompt appear sooner. It does not answer it. The order prompt is never covered here (§1).

### 8. Snapshot, wire and docs

- `Player.AutoAnswers` is cloned, snapshotted additively as `autoAnswers` in schema 7 (recorded with `-update-shape`, no version bump; absent restores empty), and carried across an undo by `RestoreFrom`, like `TriggerOrderAlwaysAsk`. The action mints no undo entry, so an undo must not take the rules back.
- `PendingChoice` gains `AutoAnswerKey` and `AskedByHand`, both additive (`autoAnswerKey`, `askedByHand`).
- `EventAutoAnswer` and `LogEvent`'s new kind are additive. `log_event_kind_gate_test.go` gets the new kind.
- The view: `auto_answers` on the seat's own `PlayerView` only; `auto_answer_key`, `auto_answer_card` and `auto_answer_prompt` on the chooser's `PendingChoiceView` only. `FilterViewFor` clears them for everyone else and for spectators. They are not secret, but nobody else needs them.
- The undo entry gains the ID of the prompt it answered (`undoEntry.autoAnswered`). It is room state, not game state, and is never persisted.
- `docs/protocol.md` documents the action, the view fields and the log kind. `docs/bot.md` says bots ignore it.

---

## Tests

**Go.**

- **Keys** (`game`): Rhystic Study's tax and draw get different keys, and the draw's key is derived from the tax's; Consecrated Sphinx's key is its row; Esper Sentinel's key is the same at power 2 and power 3; an engine-built delayed trigger and a face-down source get none.
- **Coverage** (`game`): each "never covered" row in §1 has an empty key (ward, Mana Leak, Stasis, Mystic Remora's cumulative upkeep, a discard-to-pay echo, Artisan of Kozilek, cascade, a shockland, Court of Ambition's two-label confirm).
- **Answers** (`game`): Never on Rhystic Study's tax runs the draw prompt; Always with {1} available pays it and taps one land; Always with nothing available leaves the prompt open, marked asked by hand, and a later land drop does not answer it; Always on Consecrated Sphinx builds the trigger; Always with an empty library is asked by hand; an automatic answer does not restart the loop run, and two Sphinxes set to Always stop at the breaker's threshold; with `LoopNotice` set nothing is answered automatically; a prompt to a bot seat is never answered automatically.
- **State** (`game`): clone and snapshot round-trip of `AutoAnswers`, `AutoAnswerKey` and `AskedByHand`; `RestoreFrom` keeps the live rules; the shape file records the new fields.
- **The action** (`actions`): seat scope, admin, more than 100 rules, an over-long key, a bad answer, `MintsNoUndo`, allowed during the opening roll; the enumerator never offers it (`legal`).
- **The room** (`ws`): a commit that queues a covered prompt returns a frame where it is answered, with two undo entries and the top one stamped with the chooser and free; undoing it reopens the prompt marked asked by hand, and the next commit leaves it open; a room with no socket connected still answers (the browser-closed case); a bot runner's commit triggers the answer too.
- **The log** (`protocol`): the three renderings exactly, the redaction of a face-down source, and the view fields private to their seat.

**Vitest.** `autoAnswers` is classified as synced and filled by the merge; `autoAnswerPref.ts` reconcile cases (as `triggerOrderPref.test.ts`); the remember toggle sets Always or Never from the button pressed and is absent with no key; the 101st rule is refused; the notice builder's buttons and Undo's greyed state; the Settings rows and Forget; `labels.test.ts` with the new names.

**Playwright.** One new spec, `auto-answer-1961.spec.ts`: two seats, one with Rhystic Study. The other casts a spell, ticks **Remember this answer** and presses "Don't pay". They cast again: no prompt appears, both logs show "<player> didn't pay for Rhystic Study (automatic)", and the caster's dock shows the notice. **Ask me next time** brings the prompt back on the third spell. The nightly E2E runs on the client PR's branch before it merges.

---

## Delivery

Each PR goes into `develop`, Sprint S59, Issue #1961.

| PR | What | Needs |
|---|---|---|
| 1 | **This ADR** and the AGENTS.md §3 ADR range line. Docs only. | — |
| 2 | **Server.** Keys and coverage (§1, §2), `Player.AutoAnswers` and `set_auto_answers` (§3), `NextAutoAnswer`, `AutoAnswer` and the room's follow-up commits (§4), `canPayCostLocked` (§5), `EventAutoAnswer` and the log line (§6), the snapshot fields and the view (§8), `docs/protocol.md`, `docs/bot.md`; the Go tests. | 1 |
| 3 | **Client.** `gameplay.autoAnswers`, the reconcile, the remember toggle, the notice, the Settings section, the labels; the vitest and the Playwright spec; the nightly E2E on the branch. | 2 |

After PR 3: run the nightly E2E on `develop`, then close #1961 with evidence (the run, and a cmd-dev game where a remembered Rhystic Study answer showed in the other seat's log).

## Consequences

- A player answers a repeated question once. Rhystic Study, Smothering Tithe and Consecrated Sphinx stop interrupting them, and keep working while their browser is closed.
- Every automatic answer is in every player's log, so the table always knows what was chosen and that it was a standing answer.
- "Always pay" never fakes a payment. It pays real mana or asks.
- A rule follows the person across devices and tables.
- Prompts whose answer depends on the board are still asked every time, which keeps the feature small and the answers safe.
- An automatic answer is one more undo entry. A wheel against a Sphinx set to Always adds seven, and the undo ring holds 32.

## Out of scope

- **Asking CR 603.5's question at resolution.** The engine asks an optional trigger's question before the trigger goes on the stack. That is a separate engine change (question 7). The keys here name the ability row, not the moment of asking, so a rule set now still works after that change.
- **Bots.** They keep deciding for themselves.
- **A rule per table** (question 4).
- **Choosing which sources pay.** "Always pay" uses the auto-tapper's choice.
- **The trigger order** (#1968) and every other pick.
- **Telling the other players about a rule before it fires.** The log line is the disclosure.

## Questions for the owner

Each question lists the most CR-faithful option first, then the others. The recommendation is the option the Decision above is written with.

1. **Who answers: the server or the client?**
   - (a) **The server, as its own commit** (§4). It works with the browser closed, the chooser can undo it, and the loop breaker still sees it as automatic.
   - (b) The server, inside the commit that raised the prompt. Simpler, but the chooser can't undo it, because the entry belongs to whoever acted.
   - (c) The client. Nothing happens while the tab is closed.

   **Recommendation: (a).**

2. **The key: card and prompt, or card alone?**
   - (a) **Card and prompt** (§2). Rhystic Study's tax and its draw are different questions, often to the same player.
   - (b) Card alone. One rule would answer both, so "Always" could mean pay and draw at once.

   **Recommendation: (a).**

3. **Cover the commander return (CR 903.9a)?** It is asked every time a commander dies, and it is a plain yes/no about one card.
   - (a) **Not covered.** Sometimes the graveyard or exile is the right answer (reanimation, a card that cares about exile), and ADR 0115 made this a deliberate question.
   - (b) Covered, keyed by the commander, Ask by default.

   **Recommendation: (a)** for this ADR. (b) is a small follow-up on the same mechanism if players ask for it.

4. **Where the rules live.**
   - (a) **The account's synced settings, mirrored to each seat** (§3).
   - (b) Per table only: set again in every game.
   - (c) Both, with a per-table override.

   **Recommendation: (a).** The cards are in a person's decks, so the answers travel with the person. (c) can be added later without changing the server.

5. **Does undoing an automatic answer spend the undo budget?**
   - (a) **No (`freeUndo`).** The player didn't click it, and the budget polices taking back your own clicks. A bot improvisation is free for the same reason.
   - (b) Yes, like any other undo.

   **Recommendation: (a).**

6. **"Always" on a prompt whose Yes may draw, with an empty library.** CR 121.3 lets a player draw from an empty library, and CR 704.5b then makes them lose.
   - (a) **Ask by hand whenever the chooser's library is empty**, for every Always. This is coarse: the engine can't tell which yes draws. It costs one click, and only in that rare case.
   - (b) Answer anyway. That follows the rule literally and can lose the game with nobody clicking.

   **Recommendation: (a).** It is CR-faithful too, since the player is still free to draw. They just choose it themselves.

7. **File the CR 603.5 change?** Put optional triggers on the stack and ask when they resolve, as the rule says.
   - (a) **Yes: a separate issue and ADR**, not part of this one.
   - (b) Leave it.

   **Recommendation: (a).** It is the more faithful engine, and it would make an automatic "No" to a Sphinx visible on the stack for the hold before it resolves doing nothing. It touches every optional trigger, so it should not ride on this feature.

8. **The log's word for it.**
   - (a) **"(automatic)"** at the end of the line.
   - (b) No marker. The answer is public either way.

   **Recommendation: (a).** CR 732.1a's shortcuts are acceptable when everyone understands them, and a marker is how the table learns that a player has one.
