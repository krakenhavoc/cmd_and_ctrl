# ADR 0197 — Optional triggers ask "you may" as they resolve

**Status:** Accepted · 2026-10-15 · Backlog
**Issue:** [#2494](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2494)
**Numbering:** every remote head was listed (`develop`, `main`, `cost-ledger`, `design/cool-pass`, `docs/issue-audit`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2947-resolution-copy-casts`, `fix/2950-darkness-replacement-continuation`, `fix/caddy-reload-admin-off`, `fix/creature-row-empty-board`, `wip/836-one-click-default`, `pr/2326`) and open PRs. The highest in use is **0196**; 0190 is a gap that #3699 and #3734 both left alone, so this is **0197**.
**Builds on:** [ADR 0018](0018-triggers-on-the-stack.md) (triggers on the stack, the CR 603.3b drain, the #1529 held batch), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers; its owner decision 10 filed this change), [ADR 0065](0065-modal-and-multi-target-clauses.md) §4 (the order mode → targets), and the #1289 / #3613 resolution pause and tail.

---

## Context

CR 603.5 (2026-09-25 edition): an optional triggered ability "go[es] on the stack when [it] trigger[s], regardless of whether their controller intends to exercise the ability's option or not. The choice is made when the ability resolves."

The engine asked the question as the ability triggered. A "No" dropped the trigger before it reached the stack, and a "Yes" went on to the mode and target picks and then to the stack. So:

- the player decided earlier than the rules allow, on a board that could still change before the ability resolved;
- nobody could respond to the trigger before the decision, and a response that should have informed it came after it;
- the prompt could interleave with another prompt from a spell that was still resolving (the issue's Kodama's Reach and Avenger of Zendikar report), because the harvest asks as the event happens, mid-resolution.

The 2026-10-13 audit counted 309 catalogued cards with a "When/Whenever/At …, you may" line. ADR 0127's owner decision 10 chose to make the engine faithful here and to do it on its own.

## Decision

### 1. The ability goes on the stack; the question waits for resolution

`dispatchTriggerInstanceLocked` no longer stops at `OptionalPrompt`. An optional trigger takes the same path as a mandatory one: the CR 603.3d "no legal target" removal, the CR 603.3c mode pick, the CR 603.3d target pick, `Build`, the drain. `buildTriggerItemLocked` stamps the built item with `StackItem.Optional` (`OptionalTrigger`): the question, the chooser when it is not the controller (`TriggerOptionalPrompt.Chooser`, evaluated as the ability is put on the stack), the ADR 0127 key and labels, and the Perplexing Chimera trade subject.

`resolveTopAbilityLocked`, after the CR 608.2b target check and `EventResolve`, asks an `Optional` item's question instead of running its effect: a `trigger_prompt` whose `OptionalItem` names the resolving item. The prompt is queued inside the resolution, so it is stamped `midResolution` and holds the CR 704.3 boundary (#1289), and the resolution's tail (its log line) is parked behind it (#3613). "Yes" runs the effect and the chosen modes inside the resolution; "No" runs nothing. Either way the tail and the boundary run after the answer.

A chooser who has left the game cannot take the option: `QueueChoiceForEffect` drops the prompt and the item resolves doing nothing. A chooser who leaves while the question is open is handled by ADR 0060's table as before (`trigger_prompt` is reassigned).

### 2. The one "may" that stays early: miracle's reveal

CR 702.94a: "You may reveal this card from your hand as you draw it … When you reveal this card this way, you may cast it." The reveal is a static ability's choice made as the card is drawn, and it is what creates the trigger. `TriggerOptionalPrompt.BeforeStack` marks it, and only it keeps the old harvest-time prompt (`queueTriggerPromptLocked`). The cast's own "may" was already a resolution-time permission.

### 3. Targets and modes are still chosen as the ability goes on the stack

CR 603.3c and 603.3d are unchanged, and the engine already applied them in that order. A targeted "you may" (Reclamation Sage) now asks for its target first, goes on the stack, and asks "you may" as it resolves. If its required target clause has nothing to point at, the ability is removed as it would be put on the stack (CR 603.3d), with no question at all, as before. `HasLegalTarget`'s "Yes would do nothing" warning is computed when the question is asked.

### 4. Intervening "if" (CR 603.4)

Each card checks its condition in `AppliesTo` (as it triggers) and again as it resolves. Both checks still run.

**Amendment (2026-10-15, [#3907](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3907)): the check comes before the question.** A row can now declare its resolution-time check: its Effect is a `Do` whose first step is an `If{Cond}` (the #3033 shape), or a hand-written Effect wrapped by `effects.InterveningIf(cond, row)`. The engine reads the declaration through `game.CatalogTriggerInterveningIf`. That is a package hook, not a field, so nothing reachable from `Game` holds a closure. When an optional item resolves, the declared condition is checked after the CR 608.2b target check and before the "you may". If it no longer holds, CR 603.4 removes the ability. It does nothing and asks nothing, so a standing answer has nothing to answer. The engine emits an `EventFizzle` with `Call: "condition"`. The log reads "X did nothing: its condition no longer held (CR 603.4)". The client's stack linger badges it "No effect", not "no legal targets". If the condition still holds, the question is asked as before. Valakut, the Molten Pinnacle, Emeria, the Sky Ruin and Land Tax declare their conditions this way.

An optional row whose condition is written only inside a hand-written Effect still checks it after the "yes". Each such card can be moved to `InterveningIf` when it is next touched. The game state is the same either way; only the needless question remains.

### 5. "Do this only once each turn"

Nykthos Paragon, Tidus's Cheer, Ondu Spiritdancer and Pantlaza read "the trigger was put on the stack this turn" or "the ability resolved this turn" as "the player said yes". That stops being true when every trigger reaches the stack and every one resolves. They now note the instruction as done on a "yes" under a label of their own (`Game.NoteDoneThisTurn`, the Corruption of Towashi pattern) and read that label as they trigger and as they resolve. A side effect is that Nykthos Paragon's declared simplification is gone: two gains at once are two triggers, and the controller chooses which to accept.

### 6. Standing answers (ADR 0127) at the new moment

The keys name the ability row, not the moment of asking, so they are unchanged. The resolution-time prompt carries the same `AutoAnswerKey`, the room answers it the same way, and the automatic answer's log line names the resolving item. An automatic "No" to a Consecrated Sphinx is now visible on the stack for the hold before it resolves doing nothing, as owner decision 10 expected.

The coverage rule (`triggerPromptCovered`) is unchanged: a targeted or modal optional trigger still takes no standing answer. Its targets are now chosen before the question, so that rule could be widened; that is a product question for a follow-up, not part of this change.

### 7. Trigger order (CR 603.3b) and remembered orders

More items now reach the stack: a trigger its controller would have declined is in the batch. Optional rows have no footprint, so a batch with one in it still asks for an order, as a "yes" batch did before. Remembered orders (ADR 0018's eighth amendment, #3879) are keyed by rows and apply unchanged.

One case is read differently. An optional item whose chooser has a standing "Never" for it resolves doing nothing, so the independence check (`triggerAccessLocked`) describes it as touching nothing. A seat that declines a trigger every time is not asked to order it.

### 8. Restore points

The question is plain data: `PendingChoice.OptionalItem` (the resolving item's ID) and `OptionalTradeFor`. The item itself rides the parked resolution tail, which Clone and the snapshot already carry. So, unlike the old harvest-time prompt (a `triggerResume` closure frame that the census counted), an open "you may" is a restore point when the item's effect can be rebuilt (a catalog row or a keyed body). `StackItem.Optional` and the two choice fields are additive snapshot fields; a binary from before them refuses a file that holds one (an unknown key), which is the designed rollback case.

### 9. Seats

- **Bot (heuristic):** it sees the same `trigger_prompt` kind, now at resolution, and prices it as before (its "you may" default and the ADR 0104 trade check). The board it reads is the board the ability resolves on.
- **MCP seat:** the prompt arrives as the trigger resolves. Nothing in the seat's protocol changes.
- **Client:** the dock already renders `trigger_prompt`; the prompt now appears while the trigger is shown as the resolving item.
- **Smart auto-pass:** nothing relied on the old moment. An optional trigger on the stack is a stack item like any other, so a seat that stops on stack items now stops on it.
- **Loops (ADR 0055 / 0193):** a recording replays answers by prompt shape in order; the prompt moved, and a loop recorded under this binary replays under it. An answer is still a player decision for the loop breaker; an automatic one is not (ADR 0127 §7).

## Consequences

- The table matches CR 603.5: an opponent can respond to an optional trigger before its controller decides, and the controller decides with the board the ability resolves on.
- Every optional trigger costs one more priority window: the stack shows it and players pass over it. That is the rule, and it is also what makes the Kodama's Reach interleaving impossible: only the resolving spell's own prompts are open until it finishes.
- A declined trigger leaves an `EventResolve` (and counts toward the CR 732 loop breaker's resolution tally), where it used to leave nothing.
- Card tests that answered the prompt straight after the event now pass priority to the trigger's resolution first (`awaitTriggerPrompt` in the card tests). A test that says "yes" to a targeted optional trigger picks the target first; a response that should change the outcome now goes between the trigger reaching the stack and the answer.
- The snapshot corpus board `reflexive_trigger` (Undead Butler's "you may exile it. When you do, …") was written by a history this engine no longer makes: the parent's question now pauses its resolution, which adds one event. Its fixtures are untouched and still restored; the board is registered as `reflexive_trigger_asked_at_resolution`, a new v9 file, built the CR 603.5 way.

## Out of scope

- Moving every hand-written optional intervening-if to `InterveningIf` (§4's amendment covers the mechanism and three cards).
- Standing answers for targeted and modal optional triggers (§6).
