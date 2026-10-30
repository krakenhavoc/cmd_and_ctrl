# ADR 0193 — Demonstrate a loop: record one iteration, propose N, replay it (CR 732.2)

**Status:** Accepted (owner decisions 2026-10-15) · 2026-10-15 · Sprint: none (Backlog — parked on a dependency or a single card). Proposed 2026-10-14; the owner took every recommendation, recorded under [Owner decisions](#owner-decisions-2026-10-15).
**Issue:** [#3382](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3382). Related: [#628](https://github.com/krakenhavoc/cmd_and_ctrl/issues/628) and [#804](https://github.com/krakenhavoc/cmd_and_ctrl/issues/804) (the loop breaker and its shortcut prompt), [#3016](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3016) ("use this order again" for trigger batches), [#2506](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2506).
**Builds on:** [ADR 0055](0055-loop-breaker.md) (the CR 732 loop breaker and `PendingChoiceLoopShortcut`), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers, answered by the room as their own commits), [ADR 0033](0033-ai-bot-seat.md) §8 (`Room.ApplyBundle`: one atomic commit, one undo entry), [ADR 0009](0009-smart-priority-autopass.md) (autopass lives in the client), [ADR 0111](0111-action-dock.md) (the action dock), [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) and [ADR 0126](0126-bots-that-play-their-decks.md) (agent and bot seats read `internal/legal`).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-14 after `git fetch --all --prune`: `origin/develop`, `origin/main`, `origin/cost-ledger`, `origin/docs/issue-audit`, `origin/feat/3536-graveyard-exile-cast-cost`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/playmats`, `origin/fix/2947-resolution-copy-casts`, `origin/fix/2950-darkness-replacement-continuation`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default` and `pr/2326`, plus every open PR's files. The highest number in use is 0192. 0190 is unused on every branch but sits below two merged numbers, so it may be held by a claim comment; this ADR skips it and takes **0193**.

This ADR is plan-first. No code changed with it.

---

## Context

A player who assembles Basalt Monolith and Rings of Brighthearth, Dramatic Reversal on an Isochron Scepter, or a sacrifice-outlet loop has to click every iteration. At a paper table they would do it once and say "and I do that 500 more times". That is a CR 732 shortcut.

What exists covers a different shape. ADR 0055's loop breaker notices a **triggered** ability (and, since #810, an activated one) resolving over and over, suspends automatic passing, and asks the ability's controller one question: resolve it K more times (`PendingChoiceLoopShortcut`, `MaxLoopShortcutIterations = 1000`). It repeats one stack item and nobody decides anything inside the loop. ADR 0127 answers one card's repeated prompt. Neither lets a player propose a **sequence of their own actions** — activations, casts, payments, targets, prompt answers — and repeat it.

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, checked for this ADR):

- **CR 732.1a:** "The rules for taking shortcuts are largely informal. As long as each player in the game understands the intent of each other player, any shortcut system they use is acceptable."
- **CR 732.1b:** "Occasionally the game gets into a state in which a set of actions could be repeated indefinitely (thus creating a 'loop'). In that case, the shortcut rules can be used to determine how many times those actions are repeated without having to actually perform them, and how the loop is broken."
- **CR 732.2a:** "At any point in the game, the player with priority may suggest a shortcut by describing a sequence of game choices, for all players, that may be legally taken based on the current game state and the predictable results of the sequence of choices. This sequence may be a non-repetitive series of choices, a loop that repeats a specified number of times, multiple loops, or nested loops, and may even cross multiple turns. It can't include conditional actions, where the outcome of a game event determines the next action a player takes. The ending point of this sequence must be a place where a player has priority, though it need not be the player proposing the shortcut."
- **CR 732.2b:** "Each other player, in turn order starting after the player who suggested the shortcut, may either accept the proposed sequence, or shorten it by naming a place where they will make a game choice that's different than what's been proposed. (The player doesn't need to specify at this time what the new choice will be.) This place becomes the new ending point of the proposed sequence."
- **CR 732.2c:** "Once the last player has either accepted or shortened the shortcut proposal, the shortcut is taken. The game advances to the last proposed ending point, with all game choices contained in the shortcut proposal having been taken. If the shortcut was shortened from the original proposal, the player who now has priority must make a different game choice than what was originally proposed for that player."
- **CR 732.3** (fragmented loops), **CR 732.4** (a loop of only mandatory actions is a draw), **CR 732.5** (no player can be forced to take an action that ends a loop, other than ones the loop's objects call for) and **CR 732.6** ("[A] unless [B]") are about loops nobody chooses to start. This feature does not touch them; ADR 0055 owns them.

What the rules settle for this feature:

1. **Who may propose:** the player with priority (732.2a). The record control is offered only to the priority holder, and the proposal is made at a point where the proposer holds priority.
2. **What is proposed:** a sequence of game choices **for all players** (732.2a), repeated a specified number of times. The recording therefore holds every seat's choices in the iteration, opponents' passes included, not only the proposer's.
3. **No branches:** the sequence "can't include conditional actions" (732.2a). A recording is one fixed list of choices; a replay that would need a different choice is not the proposed shortcut, so it stops there (§5).
4. **The ending point is a priority point** (732.2a).
5. **Every other player, in turn order after the proposer, accepts or shortens** (732.2b). This is a question to each of them, not a passive "interrupt at your stops"; it answers the issue's third question.
6. **The shortest ending point wins** (732.2b, 732.2c), and the game advances there with every choice before it taken.

### What exists in the code

- **Moves are wire actions.** `legal.Move` is `{type, player, params}` plus display fields, and the same `actions.Action` shape is what `actions.Dispatch` executes. Three sites dispatch for a seat: `ws/hub.go` (browsers and MCP seats), `aiseat/runner.go` (bots) and `aiseat/improvise.go` (bundles).
- **Object identity across zones.** An instance ID survives a zone change; CR 400.7's "new object" is `Card.ObjectEpoch`, which `MoveCard` bumps (`game/activation_tally.go`). A card that is sacrificed and returned keeps its instance ID. A token, a copy of a spell or ability, and a stack item get new IDs every time.
- **Atomic commits.** `Room.ApplyBundle` runs several mutations under one lock hold with one pre-clone: one sequence number, one snapshot, one replay line, one undo entry, rolled back whole on any error (ADR 0033 §8).
- **Standing answers** (ADR 0127) are answered by the room between commits (`autoAnswerThenCaptureLocked`), not dispatched by a seat.
- **The loop breaker** counts resolutions per ability and raises a blocking `loop_shortcut` prompt after `DefaultLoopThreshold` (25) of them with no decision in between.

## Decisions

Each decision says where it comes from: **[CR]** for a rule quoted above, **[0055]** etc. for an accepted ADR, and **[proposed]** for an engineering choice this ADR makes. Every product choice was put to the owner and is recorded under [Owner decisions](#owner-decisions-2026-10-15); a reference such as "owner decision 2" points there.

### 1. The flow: record, stop, propose, answer, replay

1. **Record.** The priority holder turns on "Record loop" (`loop_record_start`). From that commit on, the room records every action any seat dispatches, in order, until the recorder stops.
2. **Play one iteration by hand.** Everything happens for real: the iteration is ordinary play, each step its own commit and undo entry. Opponents pass (or do anything else) as they normally would.
3. **Stop** (`loop_record_stop`). The recorder must hold priority again. The server checks the recording closes (§3). If it does not, the recording is discarded and the recorder is told why in one sentence. If it does, the game holds a **ready recording**: its steps and its net change per iteration.
4. **Propose** (`loop_propose {count}`). The recorder, still holding priority, asks for N more iterations, 1 ≤ N ≤ the cap (owner decision 2). The proposal is the remaining sequence: the recording, N times, ending where the recorder has priority at the end of the N-th iteration [CR 732.2a].
5. **Answer.** Each other seat still in the game, in turn order after the proposer, gets a blocking `loop_proposal` prompt: **accept**, or **shorten** to a named place (§4) [CR 732.2b].
6. **Replay.** After the last answer the server replays the sequence to the earliest named ending point, as one commit (§5), and writes one log line (§7).

At any point before the replay the recorder can cancel (`loop_record_cancel`), which discards the recording and withdraws an outstanding proposal.

### 2. What is recorded, and how objects are named

A recording is a list of **steps**. A step is one dispatched action: its actor, its type and its params, captured after `actions.Decode` and after the dispatch succeeded. An action the engine refused is not recorded. **[proposed]**

Prompts answered by a standing answer (ADR 0127) are **not** steps. The room answers them again during the replay, the same way, because the setting cannot change under the room lock. **[0127]**

UUIDs in params are rewritten to **references** when the step is stored, so a replay can find the same thing in a later iteration **[proposed]**:

| What the ID names | Stored as | Found at replay as |
|---|---|---|
| A card or permanent that existed when recording started | its instance ID | the same instance ID (it survives zone changes; CR 400.7's new object is the epoch, not the ID) |
| A player | the player ID | the same player |
| An object **created** during the iteration (a token, a copy, a stack item, an emblem) | "the k-th object of that kind created in this iteration" | the k-th object of that kind created in the current iteration |
| A pending choice (a prompt answer's `choice_id`) | "the k-th prompt this iteration, to this seat, of this kind" | the same position in the current iteration |

The room keeps a per-iteration birth list while recording and while replaying: tokens from `EventTokenCreated`, stack items and copies by comparing the stack and battlefield IDs before and after each step. A reference that cannot be resolved at replay is a divergence (§5).

Recording state lives on `Game` as plain data (`Game.LoopRecording`: recorder, status, start sequence, steps as JSON, the closure signature) so it is cloned for undo and carried by snapshots, additively within the current `SnapshotSchemaVersion` (#522). The `game` package stores the steps opaquely; only `ws` and `actions` decode them. **[proposed]**

**A recording ends without a proposal** when: the recorder cancels; any undo is taken while recording (the history it describes is gone); the step, turn or active player changes; the game ends; or the recorder leaves the game. Each says why in the log. **[proposed; the same-step limit is owner decision 3]**

### 3. "Is it a loop?" — the closure check

At stop, the server compares a **loop signature** taken at record start with one taken now. The signature is the part of the state a loop must give back; the rest is what a loop is for. **[proposed; the boundaries follow CR 732.2a's "predictable results"]**

Must be **equal**:

- turn, step, active player, and the recorder holding priority;
- the stack, item by item (catalog ref and controller; a copy created and resolved inside the iteration is gone again by definition);
- no pending choice;
- every permanent present at start is present, with the same controller, tapped / face / phased / attached state. Counters, damage and the object's epoch may differ (a sacrificed-and-returned creature is the same card in the same state).

Allowed to **differ**: life totals, poison and other player counters, counters on permanents, the mana pool, cards moving between library, hand, graveyard and exile, new tokens, and per-turn tallies (storm count, spells cast, the loop breaker's runs).

A failed check names the first difference ("Basalt Monolith is tapped at the end and was untapped at the start"). The same signature is checked again after **every** replayed iteration (§5), so each repetition is provably the same loop and not only the first.

The **net change per iteration** is the diff of the allowed-to-differ fields (life per player, counters, pool, tokens created, cards drawn or milled). It is shown in the proposal and the log line. The replay never multiplies it: it runs the iterations (owner decision 1).

### 4. The opponents' answer: accept or shorten

`PendingChoiceLoopProposal` (`"loop_proposal"`) is queued to each other seat in turn order after the proposer, one at a time, skipping seats that have left the game [CR 732.2b]. It **blocks the table** (`ChoiceBlocksTable`), for ADR 0055's reason: the answer decides what the next N iterations are, and a table that could act through the question would be answering it by doing [0055].

It carries the proposer, the recording's summary (each step's label, as the log would render it), the net change per iteration, the proposed count and the current ending point (shortened by an earlier answer, if any).

The answer is **accept**, or **shorten to (iteration i, step s)** where `i` is 1…N and `s` is one of the answering seat's **own** choice points in the recording (CR 732.2b: "a place where they will make a game choice"). The default in the client is the first of their choice points in iteration 1 — usually "I respond to the first activation". A shortening later than the current ending point is refused; an earlier one replaces it [CR 732.2b].

After the last answer the replay runs to the ending point. If it was shortened, the shortening seat ends with priority at that point, and the log names who shortened and where [CR 732.2c]. The engine does not force them off the recorded choice (owner decision 7).

### 5. Replay: through the engine, one commit, stop on divergence

The replay is a room operation (`Room.ReplayLoop`) built like `ApplyBundle`: one lock hold, one pre-clone, one sequence number, one snapshot, one replay line, one undo entry **[0033 §8]**. For each iteration, for each step:

1. resolve the step's references (§2);
2. dispatch it with its recorded actor through `actions.Dispatch` — the same validation every live move gets, with the seat's own gates (priority holder, controller, costs, targets);
3. let the room answer standing prompts (ADR 0127), without capturing frames in between;
4. check the prompts now pending match what the recording saw next.

At the end of each iteration the loop signature (§3) is compared with the start.

**Divergence stops the replay** [CR 732.2a: no conditional actions]. It is any of: a dispatch refused (a target gone, a cost unpayable), a reference that resolves to nothing, a prompt the recording has no answer for, a recorded answer to a prompt that was not asked, a signature mismatch, the game ending, or the step budget (owner decision 2) running out. The room then **rolls back to the end of the last complete iteration** — it keeps one clone per iteration start for this — so the table is never left mid-iteration. The proposer gets priority there, and the log says "stopped after k of N: <reason>". A replay that completes no iteration commits nothing and says why.

Hidden information and randomness (a draw, a shuffle, a coin flip) are allowed inside a recording as long as the *choices* do not depend on them; when they do, the replay sees a different prompt or an illegal step and stops. This is owner decision 4.

**The loop breaker sits out.** A replay is a shortcut the whole table agreed to, so ADR 0055's detector must not raise its notice or queue a `loop_shortcut` prompt inside one — that prompt would be an unexpected prompt and stop every activation loop at iteration 25. `Room.ReplayLoop` sets a game flag that `loopSuspectedLocked` reads, and clears every run when the replay ends, as a decision does. **[0055 §3, proposed]**

### 6. Bounds

The cap on N is `MaxLoopShortcutIterations` (1000), ADR 0055's existing number for the same question, and a replay also has a step budget (dispatches, default 20 000) so a long iteration cannot hold the room lock for long. The proposal modal shows the cap and fills "max" with it. **[0055; the numbers are owner decision 2]**

### 7. The log, the snapshot and the wire

- **Log:** the per-step entries of a replay are folded into **one** public line: "Luke repeats *Basalt Monolith ⇄ Rings of Brighthearth* ×500: +1000 {C}" (the name is the first step's source and the first other source it touched; the delta is per-iteration × iterations completed). Engine events still fire, so triggers and tallies behave exactly as if clicked; only the log projection is folded. A shortened or stopped replay says so in the same line. **[proposed; owner decision 10]**
- **Snapshot:** `Game.LoopRecording`, the replay flag and the new pending-choice fields are additive and `carried` by `Clone`, `RestoreFrom` and the snapshot (ADR 0055's undo lesson: both copies, in the same change). **[#522]**
- **Wire:** four action types (`loop_record_start`, `loop_record_stop`, `loop_record_cancel`, `loop_propose`), one prompt kind (`loop_proposal`, answered with `resolve_loop_proposal {choice_id, accept | shorten: {iteration, step}}`), and `GameView.loop_recording` (recorder, status, step count, net change once ready) for every seat. Documented in `docs/protocol.md` in the PR that adds them.

### 8. `internal/legal`, bots and agent seats

Every new move and the new kind are enumerated in `internal/legal`, so bots and MCP seats see them; a kind with no case there is an empty move list (#499, #618) **[0055 §6]**:

- `loop_record_start` for the priority holder when no recording is open; `loop_record_stop` and `loop_record_cancel` for the recorder while one is; `loop_propose` with a bounded set of counts (1, 10, 100, the cap) for a ready recording.
- `loop_proposal`: accept first (always legal), then "shorten to iteration 1 at my first choice point". Bots take the first move, so **a bot opponent accepts**. **[0055 §6's bot rule: first in the list is what a bot takes; owner decision 9]**
- Bots never start a recording: the heuristic policies rank the record moves last and never pick them. MCP seats may, since an agent can play a loop by hand as a person can. **[owner decision 9]**

### 9. The client

- A **Record loop** toggle in the action dock (ADR 0111), shown while the player holds priority. While recording it reads "Recording · 3 steps" with Stop and Cancel; every seat sees a small "Luke is recording a loop" strip so the table understands the shortcut being built [CR 732.1a].
- On a ready recording, a modal: the iteration's steps, its net change, a count stepper (1 … cap, with a Max button) and Propose.
- The `loop_proposal` prompt for opponents: the same summary, Accept, and Shorten with an iteration number and a picker of their own choice points in the iteration.
- Names come from `client/src/lib/labels.ts` (`L.recordLoop`, …), with `docs/labels.md` updated in the same PR.

## Delivery

1. **Engine** (one PR): `Game.LoopRecording`, the reference rewriting, the closure signature, `Room.ReplayLoop` with divergence and roll-back to the last complete iteration, the breaker flag, with tests: Basalt Monolith + Rings of Brighthearth recorded and replayed ×N equals N iterations by hand; a recording that leaves Monolith tapped is refused with its reason; a replay whose target disappears at iteration k stops with k−1 iterations committed; a replay never trips the loop breaker; undo takes the whole replay back.
2. **Protocol, prompt and `internal/legal`**: the four actions, `loop_proposal` with turn-order answering and shortening, the moves for bots and agent seats, `docs/protocol.md`.
3. **Client**: the dock control, the propose modal, the opponents' prompt, labels.

## Consequences

### Good

- Every combo the catalog can execute can be demonstrated, with no per-card code; the replay uses the same dispatch as a click, so it can never do something a player could not.
- A replay is one undo, one snapshot and one log line, so the table, the replay file and the undo ring do not fill up with 500 copies of the same click.
- The CR 732.2b conversation is real: every opponent is asked, can shorten, and bots answer it.

### Tradeoffs

- Replay cost grows with the count. With the cap and step budget it is bounded, and it is paid once under the lock instead of 500 round trips.
- Loops that need a different choice each iteration (a different target each time, a growing X) cannot be recorded; CR 732.2a's "no conditional actions" excludes some of them, and "nested loops" (also 732.2a) are not built.
- A recording cannot cross a step or turn (owner decision 3), though CR 732.2a allows it.

## Out of scope

- "Until an opponent is dead" or "until X" as a count. The cap plus "the game ending stops the replay" covers the common case; a predicate is a later option.
- Nested loops, and loops that cross steps or turns.
- CR 732.4's draw (ADR 0055 amendment, owner decision 8).
- The trigger-order reuse of #3016. A trigger order answered by hand during the recording is a step and is replayed; a standing order from #3016, once built, is answered by the room like any standing answer.

## Owner decisions (2026-10-15)

The owner accepted every recommendation on 2026-10-15. Each is binding.

1. **Replay, not compute.** Every iteration is replayed through the engine; one iteration's net change is never multiplied. Triggers, replacement effects and "whenever you gain life" cards work without a determinism proof, and the cap bounds the cost.
2. **Bounds.** N is capped at 1000 (`MaxLoopShortcutIterations`, ADR 0055's number) and a replay at 20 000 dispatches. Raise them on measured cost, not before.
3. **Recording window.** A recording starts and stops in the same step, with the recorder holding priority at both ends and the same stack contents. Crossing steps or turns, which CR 732.2a allows, is not built for now.
4. **Hidden information and randomness are allowed.** Draws, shuffles and coin flips may be inside a recording; the replay stops at the first divergence.
5. **The opponents' prompt.** A blocking accept-or-shorten prompt goes to each other seat in turn order (CR 732.2b), with no standing "always accept loops" setting and no timeout of its own.
6. **Where an opponent can shorten.** At any (iteration, one of their own choice points in it), defaulting to their first choice point in iteration 1.
7. **CR 732.2c's "must make a different game choice" is not enforced.** The shortening seat gets priority at the ending point and the log shows the shortening (CR 732.1a keeps shortcuts informal, and enforcing it could strand a seat whose only legal move was the recorded pass).
8. **Undo.** The whole replay is one undo entry, costing one undo from the proposer's budget. Undoing it returns to the end of the demonstrated iteration with the recording discarded.
9. **Bots and agent seats.** Bot opponents always accept; bots never record or propose a loop; MCP seats may do both. A bot that finds and records combos is ADR 0126 strategy work and a separate issue.
10. **The log line.** One public line per replay, with the per-iteration net change times the iterations completed; the per-step entries are folded away.
11. **The control's place.** "Record loop" is in the action dock, shown only while the player holds priority, with no keyboard shortcut at first.

## Implementation notes (engine PR, 2026-10-15)

Delivery steps 1 and 2 shipped together (#3382). Where the code settles something this ADR left open:

- **Where steps are recorded.** In `actions.Dispatch`, the one function every seat's action passes through, so a browser's, an MCP seat's and a bot's moves are all recorded the same way. The loop verbs, the settings verbs, the table roll and a concession are not steps.
- **References.** An ID in a step's params is rewritten to `@loop:<kind>:<k>` when it names the k-th object (a stack item, a permanent, a waiting trigger or a pending prompt) that a scan before each step found to be new in the iteration. Everything else, a player or an object that existed at the start, is kept as written.
- **The prompts check.** Each step records the prompts open just before it (`kind@seat`), and a replayed step must find the same ones.
- **A failed closure.** `loop_record_stop` on a recording that does not close succeeds and discards it; the reason is the `loop` log line, since a refused action could not also discard.
- **Proposing.** `loop_propose` is refused unless the board still matches the recording's start signature, so play between the stop and the proposal cannot make the first replayed iteration start somewhere else.
- **A player leaving** during a recording or a proposal ends it, as a step change does.
- **The loop verbs mint no undo entry.** Only the replay does.
- **§8's "rank the record moves last".** The recording verbs are enumerated only for an agent seat (`Player.Agent`), so a bot is never offered them, and a browser's move list does not carry them: an unknown move kind would stop the client's smart autopass on every window. The client follow-up drives its dock control from `GameView.loop_recording`.

## Amendment (#3894, owner report, 2026-10-15): "Record loop" is in the dock's ⋯ menu

Owner decision 11 put "Record loop" in the action dock. Shipped as a chip on the toggles row, it made that row too wide: hold, Skip to my turn, bluff, loop, Undo and ⋯ do not fit in the dock's width at a desktop viewport, and "Skip to my turn" was cut off mid-word. Recording a loop is rare, and a player starts one before they press anything else, so a standing chip pays for itself in space nobody has.

- **Where it is.** "Record loop" is an entry in the dock's ⋯ menu (the Table group), offered under the same condition as before: the viewer holds priority, nothing is recording, no prompt is open (`loopControl`'s `record` mode). It stays in the action dock, as decision 11 says, one click further in. Its name is still `L.recordLoop`, now a `text` menu item, not an `aria` button.
- **While a loop stands** the dock shows it on the status row, for every seat, exactly as §9 says: the recorder's step count with Stop and Cancel, "propose…" once the loop closed, "waiting for answers" once proposed, and the strip other seats read. Those controls are not in the menu, because the player needs them in one click while they are playing the iteration.
- **Unchanged.** The state machine, the five loop actions, the sheets and the labels for Stop, Cancel and the sheets. No keyboard shortcut.
- **The row itself.** Moving the chip alone was not enough at a tablet-width dock, so the toggles row also lets Undo (now an icon with its count, named by its `aria-label`) and ⋯ wrap together onto a second row, instead of every label shrinking until it clips. The dock keeps one row at 1920×1080 and 2344×1377 and wraps at 1024×768 or when bluff and Skip to my turn are both at their widest.

## Amendment (#4736, owner report, 2026-10-30): start recording at any point; the proposal keeps priority

> "I was in the middle of resolving an ability and could not demo my loop as it was hidden. That should not be the case." (owner, 2026-10-30)

Owner decision 11 and the first amendment offered "Record loop" only while the player held priority with no prompt open, and the server agreed: `loop_record_start` refused without priority (`ErrLoopNeedsPriority`) and with any prompt or waiting trigger (`ErrLoopPromptOpen`). Partway through a resolution there is always a prompt open (it is what paused the resolution), so the entry was simply gone, with nothing saying why. The owner decided the design below; every choice in it is the most CR-faithful one on offer.

**The rule it rests on.** CR 732.2a: "At any point in the game, the player with priority may suggest a shortcut … The ending point of this sequence must be a place where a player has priority". Priority belongs to *suggesting* the shortcut, which is `loop_propose`. Recording is not a shortcut: it only captures what happens next, so nothing in CR 732 limits when a player may start one.

### Decisions

1. **Any seat may ask to record, at any point of an active game.** `loop_record_start` is accepted from a seat still in a game past its opening roll and mulligans, while no recording stands: with a prompt of their own open partway through a resolution, with another player's prompt open, or while another seat holds priority. Spectators, eliminated seats and a player another player controls (ADR 0166) still cannot.
2. **The recording itself runs from one quiet priority of the recorder's to the next.** "Quiet" is holding priority with no prompt open and no trigger waiting. When the request comes there, recording begins at once, as before. Otherwise it is **armed** (`status: "armed"`): it captures nothing, survives the step changing, and begins at the recorder's next quiet priority (`TendLoopRecording`, after every commit).
   - **Why not begin at the prompt.** Starting the iteration at a pending choice, so that it runs from that choice to the same choice recurring, is not sound. CR 732.2a requires the sequence to end where a player has priority, and nobody has priority partway through a resolution. A choice-to-choice iteration has no priority point at either end to propose from, replay to or shorten to. It would also need the closure check (§3) to compare a half-resolved stack item and its continuation, which the signature cannot see. The proposal and the closure check both already stand on the recorder's priority, so the recording does too.
   - **What that captures is one honest iteration.** The answer that finishes the current resolution is not a step. The iteration begins where the recorder next holds priority, whatever is still waiting on the stack, and it closes only when the board comes back to exactly that point (§3, unchanged). Recording that starts on the Rings of Brighthearth's "pay {2}?" begins over the Basalt Monolith's untap ability still on the stack, and the iteration from there (it resolves, the Monolith is tapped and activated again, the Rings asks again) closes and replays like any other loop.
3. **Stop ends the iteration at the recorder's quiet priority too.** A stop asked for there runs the closure check at once, as before. One asked for without priority, or with a prompt open, sets `stopping`. The recording goes on, capturing every step, and the check runs at the recorder's next quiet priority, the earliest ending point CR 732.2a allows. Before this amendment such a stop was refused without priority, and with priority but a prompt open it was discarded as "a prompt is still waiting for an answer".
4. **Proposing still needs priority** (CR 732.2a), unchanged on the server: `loop_propose` is refused without priority (`ErrLoopNeedsPriority`), with a prompt open, and on a board that is not the one the iteration started from. A ready recording whose recorder does not hold priority waits. The dock's line reads "Loop ready · N steps · propose when you next hold priority", with propose… disabled, and the propose sheet opens when they next hold priority.
5. **The menu entry is never hidden.** "Record loop" is drawn for every viewer, the command bar's spectator menu included. It is enabled for a seat that may ask, with a line under it, "Starts when you next hold priority", when it will be armed. Otherwise it is disabled, with a one-line reason as its visible note and accessible description: a spectator ("Only a player at the table can record a loop"), a seat that left, a game not active or still in its opening hands, a controlled player, another player's loop ("Ann is recording a loop") or the viewer's own ("You have a loop going: its controls are on the dock").
6. **The dock says what is waiting.** An armed recording reads "Loop recording starts when you next hold priority" with Cancel. A stopping one reads "Recording a loop · N steps · stops when you next hold priority" and has no Stop. Other seats read "Ann will record a loop from their next priority" (CR 732.1a).
7. **Smart auto-pass holds for the recorder's own demonstration.** While the viewer's loop is armed, recording or ready, nothing passes priority for them (`autopassDecision`'s `loopDemo`, rule 1). An armed recording begins at the priority it was waiting for, the iteration is played by hand (§1), and a ready loop needs that priority to propose. An auto-pass there would record a pass nobody chose, or end the step and discard the loop.

### Starting during another player's choice

Allowed. It is the same request, and nothing is captured until the asker next holds priority: the other player's answer is not a step, and their resolution finishes as it would have. Refusing it would only make the asker wait for a moment the server can wait for itself. Two recordings still cannot stand at once (`ErrLoopAlreadyRecording`).

### Unchanged

The closure signature, the proposal and its CR 732.2b answers, the replay, the bounds and the log. Owner decision 3 still holds for the captured iteration: it starts and stops in the same step, at the recorder's priority both times. Only the armed request may outlive a step. `internal/legal` still offers the recording verbs to an agent seat only in its priority windows, so a record move never makes a window it has nothing else to do in into a decision. The server accepts the request from an agent at any time.

### Wire

`loop_recording.status` gains `"armed"` (`steps` 0), and `loop_recording.stopping` (omitempty) is new. `Game.LoopRecording` gains `stopAtPriority` (`StopAtPriority`), additive within the snapshot version (#522). Both are in `docs/protocol.md`.

### Tests

- `server/internal/game/loop_recording_test.go`:
  - starting during the recorder's own paused resolution prompt, and the loop closing;
  - starting during another player's prompt;
  - an armed request outliving a step;
  - a stop without priority closing at the next priority;
  - a stop over a changed stack being discarded;
  - a proposal refused without priority and over a changed board, then accepted once priority is held;
  - cancelling an armed request;
  - refusals outside an active game.
- `server/internal/ws/loop_replay_test.go`: the Rings of Brighthearth case end to end, recorded mid-resolution and replayed ×3.
- Client: `loopShortcut.test.ts`, `autopassDecision.test.ts` and `gameMenu.render.test.ts` (the entry mid-resolution, disabled with its reasons, and the three status lines).
