# ADR 0143 — Gameplay settings overhaul

**Status:** Proposed · 2026-10-09 · S60 — Table clarity: a stack you can follow. The design waits for owner review, starting with the [questions](#questions-for-the-owner). No code changes with this ADR. The changes land in the PRs under [Delivery](#7-delivery).
**Issue:** [#2886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2886). In flight: [#2871](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2871) / PR #2879 (stop at a ticked step only when you can act; combat abilities in combat), [#2884](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2884) (trigger order skips independent triggers), [#2881](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2881) (Pass turn must walk every step), [#2880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2880) / PR #2885 (board picks).
**Amends, once accepted:** [ADR 0009](0009-smart-priority-autopass.md): the precedence list, the safety belt (§7), and `smartAutoPass`. [ADR 0111](0111-action-dock.md) §5 and §7: the dock's toggles and Pass turn. [ADR 0119](0119-a-stack-you-can-follow.md) §2: who sets the stack hold.
**Builds on:** [ADR 0110](0110-remember-me.md) §4 (synced settings), [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) (strict payment), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers), [ADR 0075](0075-table-settings-and-host-controls.md) §2 (table settings).
**Numbering:** the AGENTS.md §4 sweep was run on 2026-10-09: `git fetch --all --prune`, then every ADR file on any remote ref (`git log --remotes --name-only -- docs/decisions/`). The refs were `origin/develop`, `origin/main`, `origin/cost-ledger`, `origin/docs/issue-audit`, `origin/feat/2871-smart-step-stops`, `origin/feat/2874-choose-a-background`, `origin/feat/2880-pick-on-battlefield`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/playmats`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default` and `pr/2326`. The highest number on any of them is 0142. This ADR takes **0143**.

---

## Context

The owner's goal for auto-pass, verbatim: *"make the autopass as smart as possible and as convenient as possible so most of the time players are not thinking why do I have to click to pass or thinking I missed my window to respond"*. On settings: *"The whole gameplay settings need an overhaul so let's review it as a whole and make it better/smarter."*

Since 2026-10-04 the project is automation-first: smart defaults, with sandbox actions kept as fallbacks. The owner plays at tablet size and up.

The Gameplay tab has grown one PR at a time to 24 controls. Once PR #2879 merges, `gameplay` has 25 fields. A per-setting audit is in the [appendix](#appendix-the-audit). It covers each field's key, label, default, sync scope, every read site, what the field changes, its couplings, and whether a player would touch it. Its findings, in order of how much they work against the owner's goal:

1. **The step grid does not know whose turn it is.** The default ticks are main 1, declare attackers, declare blockers, main 2 and end. They apply on every opponent's turn as well as your own. With any castable instant in hand, `hasPlay` is true outside your own sorcery window. So by default you stop at each opponent's main 1, main 2 and end step. At a four-seat table that is up to nine stops a round where you almost never act. This is the main source of "why do I have to click to pass".
2. **Turning smart auto-pass off removes stops as well as adding them.** Rule 7 of `autopassDecision.ts` (the combat and opponent-end-step key windows) runs only with `smartAutoPass` on. A player who switches it off "to see everything" loses the stops at combat and at an opponent's end step, even with an instant in hand. That is the missed-window half of the goal. #2871 split off the ticked-step half; this half remains.
3. **Greyed-out controls are still read.** With smart off, the "Stop for" categories are disabled in the UI. They still decide the session autopass toggle's "answerable spell" hold (rule 4) and whether a ticked step has something in it (rule 8).
4. **The `autopass` toggle and the "Auto-pass priority" setting share a name and work against each other.** The toggle overrides five settings but not the categories. Its safety belt is set by a danger-styled setting (`autopassPersistThroughTurns`). It overlaps with **Pass turn**, which means the same thing ("I'm done for now") and is a sandbox verb. Pass turn skips the end step, its triggers and the cleanup discard (#2881).
5. **There are two stack holds.** Each person's `stackHoldMs` and the host's `bot_pace` both delay a pass on someone else's spell. An item resolves only once every seat has passed ([ADR 0119](0119-a-stack-you-can-follow.md), "The rules"). So the table always waits for the longest hold anyone has set. The personal setting can only make the table slower, never faster.
6. **Bluffing has three gates and a silent conflict.** It needs the setting, the in-game arm, and smart auto-pass on. "Always stop for opponents' spells" makes "represent a counterspell" do nothing, and nothing says so.
7. **Several controls are near-dead.** These are `autopassPersistThroughTurns`, `autoPassOwnStack = false` (which `hold` does better, one stack at a time), `autoPassPriority = false` (used on purpose only by the tutorial), the two bluff-pause sliders, and the five response categories, whose defaults suit nearly everyone.
8. **Non-priority controls are mixed in with priority ones.** These are `adminOverrides` (which changes what right-click does, and is available to any player despite its name), `showBotReasoning`, `confirmExit` and `highlightLegalActions`.

Two things are already right, and this ADR keeps them. `triggerOrder` and `autoAnswers` are decided by the server and reconciled every frame. Standing answers are made in the prompt itself, not in Settings.

## Decision

### 1. Three promises

Each decision below serves one of these, and the copy says them in a player's words.

- **No pointless stop.** Auto-pass stops only where you can do something you would plausibly want to do. (Owner goal, first half.)
- **No missed window.** If you can respond to an opponent's spell, to an attack or at an opponent's end step, auto-pass stops. No setting turns that off except Manual. (Owner goal, second half.)
- **One place for each thing.** Every behaviour has one control, and the in-game buttons are temporary versions of a setting, never a second copy of it.

### 2. How passing works

**2.1 One "Auto-pass" choice.** The "Auto-pass priority" checkbox, the "Smart auto-pass" checkbox and "Always stop for opponents' spells" become one three-way choice, stored as a new key `gameplay.passMode`:

| Mode | What it does | Today's equivalent |
|---|---|---|
| **Smart** (default) | Stops at your ticked steps when you can do something there. Stops at an opponent's spell, at combat and at an opponent's end step when you can respond. Passes everything else. | `autoPassPriority` on, smart on, always-stop off |
| **Careful** | Smart, and also stops at every opponent spell and ability even when you can't respond, so you see each one and a pause never gives anything away. | Always-stop on |
| **Manual** | Never passes for you. Every time you get priority, you click `next`. | `autoPassPriority` off |

`smartAutoPass` is removed. What it still did after #2871 was one of two things: "stop at every opponent item", which Careful is, or the accidental switch-off of the key windows, which no promise allows.

**2.2 The key windows are always on.** Rule 7 runs in Smart and Careful alike:

- an opponent's item on the stack;
- combat once attackers are declared (with #2871's combat abilities);
- an opponent's end step.

It holds when you have a response, so a key window never costs a click to someone with nothing to say. Manual holds everything anyway.

**2.3 Stops by whose turn it is.** The grid becomes two columns, **My turn** and **Opponents' turns**. `gameplay.stepStops` keeps its key and becomes the My-turn map. A new `gameplay.stepStopsOpponents` holds the other column. Rule 8 reads the column for the active player.

- **My turn, default:** main 1, declare attackers, declare blockers, main 2. Your own end step is no longer ticked, because you rarely act there.
- **Opponents' turns, default:** nothing ticked. The key windows (§2.2) already stop you whenever an opponent's turn gives you something to answer.

A player who wants a stop at an opponent's upkeep or beginning of combat ticks it there.

**2.4 Skip empty stops stays on, and moves to Advanced.** `stepStopsOnlyWhenCanAct` (PR #2879) keeps its meaning and its default (on).

**2.5 The precedence, after this ADR.** This list replaces the numbered list in the `autopassDecision.ts` header. ADR 0009 gets a pointer here.

1. Hard blocks (unchanged): no priority, the table busy, a pending choice, an owed block or attack, the loop breaker.
2. A running **Skip to my turn** (§4.2) clears itself at the start of your own main 1 and holds. There is no setting that keeps it on.
3. A manual phase pin holds (unchanged, #526).
4. A running **End turn** or **Skip to my turn** passes. It still holds for an opponent's item you can answer, or bluffs there if bluffing is armed.
5. Manual mode holds.
6. A non-empty stack:
   - `hold` armed → hold.
   - The stack is all yours → pass (`autoPassOwnStack`, Advanced).
   - Careful → hold.
   - You have a response → hold.
   - Bluff armed → bluff.
   - Otherwise → pass.
7. Empty stack, in a combat or opponent-end-step key window: a response → hold; an instant bluff → bluff. This runs in Smart and Careful.
8. The active player's column ticks this step → hold, unless Skip empty stops finds nothing to do. `engineMayMissMana` still counts as something to do (ADR 0118 decision 8).
9. Otherwise → pass.

The categories that decide "a response" in rules 4, 6 and 7 are the Advanced "What counts as a response" list. They are read in every mode, and the UI no longer greys them out.

**2.6 The stack hold becomes the table's pace.** `stackHoldMs` is removed. The client reads the hold from the table's existing `bot_pace`: Fast 0 s, Normal 2 s, Slow 3 s. These are `botPacePresets`' own numbers, so bots and people hold the same item for the same time. The host's "Bot speed" control is renamed **Table pace**. No protocol field is added, because `settings.bot_pace` is already in every frame. A person who wants longer on one spell has `wait` (on the countdown) and `hold`.

### 3. The page

**3.1 Layout.** The Gameplay tab has five short sections and one collapsed **Advanced** section. Everything in a section's first line is something a typical player might change.

```
┌ Gameplay ───────────────────────────────────────────────────────────────┐
│ PASSING PRIORITY                                                        │
│  Auto-pass   (•) Smart   ( ) Careful   ( ) Manual                       │
│  Smart: stops where you can do something, and whenever you can respond  │
│  to an opponent. Passes everything else.                                │
│                                                                         │
│  Stop at these steps                    My turn    Opponents' turns     │
│   Upkeep                                  [ ]           [ ]             │
│   Draw                                    [ ]           [ ]             │
│   Main 1                                  [x]           [ ]             │
│   Beginning of combat                     [ ]           [ ]             │
│   Declare attackers                       [x]           [ ]             │
│   Declare blockers                        [x]           [ ]             │
│   Combat damage                           [ ]           [ ]             │
│   End of combat                           [ ]           [ ]             │
│   Main 2                                  [x]           [ ]             │
│   End step                                [ ]           [ ]             │
│  A ticked step is skipped when you have nothing to do there.            │
│  You always get a chance to respond to an opponent's spell, an attack,  │
│  and an opponent's end step, ticked or not.                             │
│                                                                         │
│ READING TIME                                                            │
│  This table's pace is Normal: other players' spells stay on the stack   │
│  for 2 s before they resolve. The host sets it in Table settings.       │
│  Click wait on the countdown to keep priority and respond.              │
│                                                                         │
│ PROMPTS                                                                 │
│  Order my triggers   [Ask only when the order matters ▾]                │
│  Saved answers       Rhystic Study: never pay        [Remove]           │
│                                                                         │
│ MANA                                                                    │
│  [x] Charge mana costs                                                  │
│                                                                         │
│ TABLE                                                                   │
│  [x] Highlight what you can do right now                                │
│  [x] Ask before leaving a game in progress                              │
│                                                                         │
│ ▸ Advanced                                                              │
│   [x] Skip a ticked step when I have nothing to do there                │
│   What counts as a response                                             │
│     [x] Counterspells   [x] Instants and flash   [x] Targeted and       │
│     protective abilities   [ ] Value abilities   [x] Special actions    │
│   [x] Pass my own spells and triggers straight away                     │
│   Bluffing   kinds [ ] counterspell [ ] instant · style (•) Timed       │
│              ( ) Manual · pause 1.5–4 s                                 │
│   [ ] Manual card controls on right-click                               │
│   [ ] Show bot reasoning                                                │
└─────────────────────────────────────────────────────────────────────────┘
```

At tablet width the two stop columns stay side by side. The grid is ten rows of two checkboxes and fits a 768 px panel. Advanced is a `<details>` element, closed by default, and its open state is remembered per device in `localStorage` (a convenience, not a setting).

**3.2 Copy.** This is the exact text for each control. Help text is one or two sentences and names no engine concepts.

| Control | Label | Help |
|---|---|---|
| `passMode` | **Auto-pass**: Smart · Careful · Manual | Smart: "Stops where you can do something, and whenever you can respond to an opponent. Passes everything else." Careful: "Smart, and also stops at every opponent spell and ability, even when you can't respond. Slower, but you see everything, and a pause never gives anything away." Manual: "Never passes for you. Click next every time you get priority." |
| `stepStops`, `stepStopsOpponents` | **Stop at these steps**, columns **My turn**, **Opponents' turns** | "A ticked step is skipped when you have nothing to do there. You always get a chance to respond to an opponent's spell, an attack, and an opponent's end step, ticked or not." (Manual: "In Manual every step stops, so these do nothing.") |
| reading time (no key) | **Reading time** | "This table's pace is {Fast/Normal/Slow}: other players' spells stay on the stack for {0/2/3} s before they resolve. The host sets it in Table settings. Click wait on the countdown to keep priority and respond." |
| `triggerOrder` | **Order my triggers**: Ask only when the order matters (default) · Always ask · Never ask: order them for me | "When several of your triggers go on the stack together, you choose which resolves first. By default you are asked only when the order can change what happens." (After #2884 the examples in today's help no longer describe the rule, so they are dropped.) |
| `autoAnswers` | **Saved answers** | "Answers you saved from a prompt with 'Always' or 'Never'. The game answers these for you. Remove one to be asked again." (The list itself is ADR 0127's, unchanged.) |
| `strictMana` | **Charge mana costs** | "On: a spell or ability costs what it says, and clicking a card taps your lands for it. A card you can't pay for is dimmed; right-click it for 'Cast anyway (don't pay)', which the log shows to the table. Off: the sandbox, where mana is tracked on paper." |
| `highlightLegalActions` | **Highlight what you can do right now** | "While you owe a decision, cards you can play get a ring and permanents with abilities get a small pip. A card you can't play is still greyed out with this off." |
| `confirmExit` | **Ask before leaving a game in progress** | (none) |
| `stepStopsOnlyWhenCanAct` | **Skip a ticked step when I have nothing to do there** | "On your own main phase anything you can play counts, a land included. Anywhere else it takes a response from the list below. Off: every ticked step stops." |
| `respond*` | **What counts as a response**: Counterspells · Instants and flash · Targeted and protective abilities · Value abilities (Mind Stone, fetch lands, cycling) · Special actions (foretell, suspend, face-up) | "Auto-pass stops for an opponent's spell, an attack or an opponent's end step only when you have one of these. Mana abilities and land drops never count. In combat, crewing, animating a land or granting a keyword counts as a targeted ability." |
| `autoPassOwnStack` | **Pass my own spells and triggers straight away** | "Casting is already the decision. To respond to your own spell, click hold before you cast." |
| `bluff*` | **Bluffing**: Represent a counterspell · Represent an instant · Timed / Manual · Shortest and longest pause | "Auto-pass passes the moment you have no answer, so a pause tells the table you have one. A bluff pauses anyway. Arm it during a game with the bluff button (B)." |
| `adminOverrides` | **Manual card controls on right-click** | "Right-click a card you control to move it, change its counters or damage, or declare it in combat by hand: the fallback when the engine gets a card wrong. Off: right-click shows the card's abilities." |
| `showBotReasoning` | **Show bot reasoning** | "Bots explain each move in the table feed. It can mention cards in the bot's hand. Improvised plays are always announced, whatever this says." |

The tutorial forces Manual (it forced `autoPassPriority` off) and restores the previous mode afterwards. `practiceTable.ts` captures and restores `passMode` in place of `autoPassPriority`.

### 4. The in-game controls

**4.1 Each one is a temporary version of a setting.**

| Dock control | Key | Lasts | Is the temporary form of |
|---|---|---|---|
| `next` | Space | one pass | — |
| `hold` / `wait` (one action, two places) | `h` | until the stack it held empties (#2853) | Careful, plus "don't pass my own stack", for one stack |
| **End turn** (active player only) | `t` | until your turn ends | Smart's passing, for the rest of your turn |
| **Skip to my turn** (replaces `autopass`) | Shift+P | until your next main 1 | Smart's passing, until you are next active |
| `bluff` chip | `b` | the game | the Advanced bluffing settings |
| phase-icon pin | — | the next time that step comes | one stop |

**4.2 End turn and Skip to my turn.** Both are yields with a fixed end point. Neither one is a mode a player has to remember to switch off.

- **End turn** replaces the sandbox **Pass turn** in the action bar, keeping its place and key. It passes priority for you until your turn ends. It walks every step, so step triggers fire and resolve and the cleanup discard happens. It stops for any decision those need (#2881) and for an opponent's item you can answer (rule 4). The `pass_turn` verb stays on the server and in the ⋯ menu as the sandbox fallback, named **Skip to next turn (sandbox)**, and it is logged as it is today. The rules half is whatever #2881 lands. This ADR asks only that the dock button be the rules-faithful one.
- **Skip to my turn** is today's autopass toggle with the safety belt built in. It clears itself at the start of your own main 1, always. `autopassPersistThroughTurns` is removed, along with its warning.
- Both obey the hard blocks (rule 1) and a phase pin (rule 3), as the toggle does today.

**4.3 The bluff chip** is disabled in Manual and in Careful, with the title "bluffing needs Smart auto-pass: in {mode} you stop anyway, so a pause gives nothing away". This replaces today's "needs smart auto-pass" condition.

**4.4 Labels.** "autopass" is a registered label and the tutorial's step-9 anchor (ADR 0076 §2.4, ADR 0125 §2). It is renamed in `labels.ts` to "Skip to my turn", and Pass turn becomes "End turn". Both tutorial steps, the dock hint (`ActionDock.hint.ts`) and the e2e specs that select on either label change in the same PR. That PR runs the branch E2E.

### 5. Migration

Every existing `gameplay` key, and where it goes. The versions assume PR #2879 has merged at v23, and each delivery PR takes the next number. A removed key is deleted in `migrate` explicitly, as `alwaysAskTriggerOrder` is today (`settings.ts:1054`), because the shallow merge keeps unknown keys. An account copy from an older client goes through the same `migrate` (`applySyncedCopy`), so a synced value moves exactly as a local one does.

| Key | Fate | Rule |
|---|---|---|
| `autoPassPriority` | **removed** (into `passMode`) | `false` → `manual` |
| `smartAutoPass` | **removed** (into `passMode`) | `false` (and not manual) → `careful`. The key windows come back for this player; that adds stops only where they can respond |
| `alwaysStopOpponentStack` | **removed** (into `passMode`) | `true` (and not manual) → `careful`. Otherwise `smart` |
| `passMode` | **new** | from the three rows above; a bad value → `smart` |
| `stepStops` | kept, now **My turn** | The stored map equals the old default (`stepStopsMatchDefault`) → the new My-turn default (end unticked). Otherwise kept as it is |
| `stepStopsOpponents` | **new** | The stored `stepStops` equals the old default → all off. Otherwise a copy of the stored map, so a player who tuned the grid stops exactly where they did before, on every turn |
| `stepStopsOnlyWhenCanAct` | kept (Advanced) | unchanged |
| `respondCounterspells`, `respondInstants`, `respondAbilities`, `respondUntargetedAbilities`, `respondSpecialActions` | kept (Advanced) | unchanged; no longer greyed out |
| `stackHoldMs` | **removed** | dropped; the table pace sets the hold (§2.6) |
| `autoPassOwnStack` | kept (Advanced) | unchanged |
| `autopassPersistThroughTurns` | **removed** | dropped (§4.2) |
| `bluffCounterspell`, `bluffInstant`, `bluffMode`, `bluffDelayMinMs`, `bluffDelayMaxMs` | kept (Advanced) | unchanged |
| `triggerOrder` | kept | unchanged (server-held) |
| `autoAnswers` | kept | unchanged (server-held) |
| `strictMana` | kept | unchanged; relabelled |
| `highlightLegalActions` | kept | unchanged |
| `confirmExit` | kept | unchanged; relabelled |
| `adminOverrides` | kept (Advanced) | unchanged; relabelled |
| `showBotReasoning` | kept (Advanced) | unchanged |

`SYNCED_FIELDS` gains `passMode` and `stepStopsOpponents` as `synced` and loses the removed keys. `settingsSyncFields.test.ts` enforces this. The `gameplay` group goes from 25 fields to 22, and the visible page from 24 controls to 8 plus Advanced.

Two consequences of the migration:

- A player on the untouched defaults loses their opponents'-turn main-phase and end-step stops, and their own end-step stop. That is the point of §2.3. The key windows still stop them whenever they can respond.
- Mixed client versions: a tab still running an older client reads an account copy with no `autoPassPriority` and falls back to its own default (on). It picks up the new client on its next reload, and the table has at most eight users, so this is accepted rather than handled with a compatibility window.

### 6. What moves to the server

- **The stack hold**, by reusing `TableSettings.BotPace` (§2.6). This needs no new field and no snapshot change. The client maps `bot_pace` to milliseconds with the same three numbers `aiseat/runner.go` uses. A test pins the two tables together, so that a change to `botPacePresets` fails a client test.
- **Nothing else moves now.** The gate chain stays client-side, as ADR 0009 put it, and `triggerOrder` and `autoAnswers` stay server-held as they are. Moving the stops and key windows to the server would let a seat keep passing while its tablet sleeps, and would take a round trip out of every automatic pass. It is a larger change, with new seat state, a snapshot field and a protocol action. It is [Q7](#questions-for-the-owner), not part of this plan.

### 7. Delivery

Each PR is one logical change, targets `develop`, and carries `Sprint: S60` and `Issue: #2886`. PR #2879 merges first.

1. **Key windows always on; one Auto-pass choice** (client; schema v24). This adds `passMode` and removes `autoPassPriority`, `smartAutoPass` and `alwaysStopOpponentStack`. It changes rules 5-7, the practice table capture, the bluff chip gate, and the "Stop for" fieldset (never disabled). Tests: the migration from each combination; rule 7 in Careful; the toggle reading categories; `highlightsLive` with the new verdicts. Branch E2E: yes (client interaction).
2. **Stops by whose turn it is** (client; v25). This adds `stepStopsOpponents`, makes rule 8 read the active column, and adds the two-column grid. Tests: the migration of a default and a tuned grid; a default player holding an instant passes an opponent's main 1 and end step but stops on an opponent's attack; a tuned player stops where they did. Branch E2E: yes.
3. **The page** (client). Sections, the Advanced `<details>`, the copy in §3.2, the reading-time line, and the help in `docs/` (ADR 0009 and ADR 0119 get pointers). No schema change. Tests: the render tests for each section and `settingsNarrow` at 768 px. Branch E2E: no, unless a label the specs select on moves.
4. **Table pace** (client; v26). This removes `stackHoldMs`, reads the hold from `bot_pace`, renames "Bot speed" to "Table pace", and adds the new hints (`tableSettings.ts`). Tests: the hold for each pace; the pinned preset table. Branch E2E: yes (the countdown is dock UI).
5. **End turn and Skip to my turn** (client; v27; after #2881 merges). This renames the toggle, makes the dock's End turn the rules-faithful yield, moves the sandbox verb to the ⋯ menu, and removes `autopassPersistThroughTurns`. It updates `labels.ts`, `docs/labels.md`, the tutorial, the hints and the e2e specs. Branch E2E: yes.

PRs 1 and 2 deliver most of the owner's goal on their own. They can ship before the page is redrawn, with the existing controls relabelled in place.

### 8. Rejected alternatives

- **Presets that also write the Advanced fields** (a "Custom" preset whenever anything differs). The three modes cover the only choice most players make. Making the response categories part of a preset would turn every Advanced tweak into "Custom" and hide the preset a player picked.
- **Keep `smartAutoPass` as a switch for the key windows.** It is the coupling behind finding 2. A switch that only takes away "you get a chance to respond" breaks the second promise, and Manual and Careful already cover everyone who wants more stops.
- **Per-step "only when I can act"** (one per row). #2871 rejected it for the reasons in its ADR 0009 amendment, and the two-column grid doubles the rows.
- **Advanced as its own tab.** A tab is a place a player has to know to look. A collapsed section under the controls it modifies is found by the player who goes looking.

## Consequences

- At a four-seat table with the defaults, a player holding an instant stops only when an opponent casts something they can answer, attacks while they hold a response, or reaches their end step while they hold one. Their own turn stops at their main phases and combat declarations, when they can act there.
- The visible Gameplay tab goes from 24 controls to 8. Everything removed is either covered by a mode or the table pace, or was a footgun.
- Two in-game buttons that meant "I'm done" become two yields with a stated end point. The one in the action bar becomes rules-faithful.
- The table pace sets reading time for everyone. A host who picks Fast gets a fast table, and nobody can slow it down by accident through a personal setting.
- Costs: four schema bumps, a label rename the e2e suite feels, and a migration that changes behaviour for untouched-default players (deliberately, as v15, v19 and v22 did).

## Questions for the owner

Each question lists the recommended option first.

1. **Q1. One Auto-pass choice.**
   - **(a) Recommended:** Smart / Careful / Manual replaces "Auto-pass priority", "Smart auto-pass" and "Always stop for opponents' spells" (§2.1).
   - **(b)** Keep the three checkboxes, fix the couplings (rule 7 always on, categories never greyed), and relabel.
   - **(c)** Presets that also set the Advanced fields, with a "Custom" state.
2. **Q2. Stops by whose turn it is.**
   - **(a) Recommended:** two columns. Opponents' turns start empty, and the end step leaves the My-turn defaults. Players still on the old default grid move to the new one, and a player who tuned their grid keeps their ticks in both columns (§2.3, §5).
   - **(b)** Two columns, with everyone's current grid copied into both, so nobody's behaviour changes until they edit it.
   - **(c)** One grid, with just the main phases and end step unticked by default.
3. **Q3. The key windows.**
   - **(a) Recommended:** an opponent's spell, combat once attackers are declared, and an opponent's end step always stop you when you can respond, in Smart and Careful, with no switch (§2.2).
   - **(b)** As (a), and add an opponent's beginning of combat as a fourth key window. It is the last moment to tap or remove an attacker. It costs one stop per opponent turn for anyone holding removal.
   - **(c)** Keep a switch for them in Advanced.
4. **Q4. Reading time.**
   - **(a) Recommended:** the table pace (the host's "Bot speed", renamed) sets the stack hold for people and bots alike, and the personal setting goes away (§2.6).
   - **(b)** The table pace is a floor, and each player may choose a longer hold for themselves.
   - **(c)** Leave both as they are.
5. **Q5. The autopass toggle and Pass turn.**
   - **(a) Recommended:** **End turn** (the active player only; walks every step, per #2881) and **Skip to my turn** (replaces autopass; always clears at your main 1). "Autopass persists through your own turns" is removed, and the sandbox Pass turn moves to the ⋯ menu (§4.2).
   - **(b)** Keep both buttons as they are, and only rename autopass and fix Pass turn's rules (#2881).
   - **(c)** One button that means End turn on your turn and Skip to my turn otherwise.
6. **Q6. What goes under Advanced.**
   - **(a) Recommended:** skip empty stops, the response categories, passing your own stack, bluffing and its pause range, manual card controls, and bot reasoning, in a collapsed section at the bottom of Gameplay (§3.1).
   - **(b)** The same list on its own "Advanced gameplay" tab.
   - **(c)** No Advanced section: everything visible, regrouped under the new headings.
7. **Q7. Passing on the server.**
   - **(a) Recommended:** not now. The stops and key windows stay client-side, and only the stack hold reads a table setting (§6). Revisit if a sleeping tablet stalls real games.
   - **(b)** Move the stops, the mode and the key windows into the seat's server state, so the server passes for a seat whose tab is asleep or closed. This needs a new action, snapshot fields and a server copy of the response classes.

## Appendix: the audit

This is the audit behind the Context. It was taken on `origin/develop` at ee35531a4, plus PR #2879. Line numbers are develop's unless marked `PR 2879`. "Read at" lists runtime reads only. It leaves out each field's own declaration, its Settings control and its tests. Every `gameplay` field is synced (`SYNCED_FIELDS`, `settings.ts:575-659`).

### A. The auto-pass gate chain

`client/src/lib/autopassDecision.ts:141-223` decides every automatic pass, and
`routes/Game.svelte:473-543` assembles its inputs. The rules are numbered in the
file header (lines 18-46):

1. Hard blocks: no priority, mulligans, a pending choice, an owed block or
   attack, the CR 732 loop breaker.
2. The safety belt: the viewer's own precombat main clears the session autopass
   toggle.
3. A manual pin holds.
4. The session autopass toggle passes everything except an opponent's stack item
   the viewer can answer.
5. `autoPassPriority` off holds.
6. A non-empty stack. Hold is on → hold. All mine → `autoPassOwnStack`. Smart off
   or always-stop → hold. A response → hold. A bluff → bluff. Otherwise pass.
7. **Only with smart on**: the combat and opponent's-end-step key windows hold
   for a response.
8. A ticked step holds, unless `smartAutoPass` finds nothing to do there (PR 2879
   changes that to `stepStopsOnlyWhenCanAct`).
9. Otherwise pass.

A `pass` verdict is then delayed by the stack hold (`stackHold.ts`,
`Game.svelte:400-470`). A bluff verdict is delayed by the bluff timer.

### B. Per setting: `gameplay.*` (Settings → Gameplay)

| Key | Label (Settings.svelte line) | Default | Read at | What it changes | Depends on / overrides | Would a typical player change it? |
|---|---|---|---|---|---|---|
| `confirmExit` | "Confirm before leaving an active game" (795) | `true` | `Game.svelte:337` (beforeunload), `:778` (back to lobby) | Asks before leaving a live game | None | Rarely |
| `highlightLegalActions` | "Highlight what you can do right now" (806), long help (809-818) | `true`; forced on once at v15 | `Game.svelte:892` via `highlightsLive(…, autopassVerdict)` | Rings, pips and counts on playable cards | **Coupled to the autopass verdict**: on a frame auto-pass will pass, nothing lights | Rarely. It is accessibility- and taste-level |
| `triggerOrder` | "Order my triggers" select (821-830), help (833-840) | `"when_it_matters"` | `Game.svelte:842-846` → `triggerOrderPref.ts` → action `set_trigger_order_preference` → server `Player.TriggerOrder` (`server/internal/game/player.go:221-231`), decided in `seatNeedsTriggerOrder` (`mutations.go:5902`) | Whether the CR 603.3b order prompt appears | Server-held. #2884 is making "when it matters" skip independent triggers (Vivi + Ugin), which leaves "always" and "never" for very few players | Rarely, once #2884 lands |
| `autoAnswers` | AutoAnswersSettings list (842); rules are made in the prompt itself (`ChoicePromptModal.svelte:1119-1124`, `AutoAnswerNotice.svelte:85`) | `[]` | `Game.svelte:871` → `autoAnswerPref.ts` → action `set_auto_answers` → server `auto_answer.go:86` | Standing yes/no answers ("never pay for Rhystic Study") | Server-held. Made in the flow, not in Settings | Yes, but from the prompt. Settings is only where you review or delete one |
| `autoPassPriority` | "Auto-pass priority through unstopped steps" (850), help (853-859) | `true` (v2/v3 migrations forced it) | `Game.svelte:519` → rule 5 (`autopassDecision.ts:183`). The tutorial forces it off (`practiceTable.ts:68-71, 102-117`) | Off means every priority window waits for a click (a "manual" mode) | **Off makes about 15 other settings moot.** The session autopass toggle (rule 4) overrides it | Almost never. Off is miserable at a four-seat table |
| `stepStops` | "Stop on these steps" grid (861-880). Help: "When auto-pass is on, priority stops here…" | main 1, declare attackers, declare blockers, main 2, end: on. Upkeep, draw, beginning of combat, combat damage, end of combat: off (`settings.ts:376-390`) | `Game.svelte:528` (`stopKeyFor(step)`, so first-strike damage shares combat damage) → rule 8 | A ticked step holds (subject to only-when-can-act) | **Not split by whose turn it is.** The ticks apply on every opponent's turn too. With an instant in hand, `hasPlay` (`responseWindow.ts:224-242`) is true outside your sorcery window, so the default grid stops you at each opponent's main 1, main 2 and end step. Ticking "end" duplicates rule 7's opponent-end-step window | Some. It is the most-tuned control, and it's the wrong shape |
| `stepStopsOnlyWhenCanAct` (PR 2879, v23) | "Only stop at my ticked steps when I can do something" | `true`, copied from `smartAutoPass` by the v23 migration | PR 2879 `Game.svelte:530` → rule 8 | Skips a ticked step with nothing to do | Reads the "Stop for" categories through `hasPlay`. `engineMayMissMana` (ADR 0118 decision 8) can force a hold | Almost never. On is right for everyone |
| `smartAutoPass` | "Smart auto-pass (stop only when you can do something)" (888), help (891-899; rewritten by PR 2879) | `true` | `Game.svelte:529` → rule 6 (`:193`), rule 7 (`:202`), rule 8 (`:218`, moved by PR 2879). Also `bluff.ts:70` (`pressBluff` is a no-op when off) and `BluffChip.svelte:27` (chip disabled) | Develop: three things. After PR 2879: two (opponent stack and key windows) | **Off has opposite effects.** It stops on every opponent stack item (more stops), but it also **turns rule 7 off**: an unticked combat or opponent end step with a response in hand now passes (fewer stops, missed windows). It disables the "Stop for" and "Bluff" fieldsets (901, 1003) and the dock's bluff chip | Rarely, and a player who does turn it off gets something other than what the label says |
| `respondCounterspells` | "Counterspells (anything that targets a spell or ability on the stack)" (910) | `true` | `Game.svelte:482` → `cats` → `hasResponse`, `hasPlay` | Whether a stack-targeting move counts as a response | **The fieldset is disabled when smart is off, but the categories are still read**: by rule 4 (the toggle's "answerable" hold) and by rule 8 (`hasPlay`). PR 2879 enables the fieldset whenever either toggle is on, which narrows this but does not close it | Almost never |
| `respondInstants` | "Instants and flash spells" (919) | `true` | `Game.svelte:483` | Same, for any other cast | Same | Almost never |
| `respondAbilities` | "Abilities that target or protect something…" (928) | `true` | `Game.svelte:484` | Same, for targeted or `interacts` activations. PR 2879 adds `combat_interacts` in combat windows | Same | Almost never |
| `respondUntargetedAbilities` | "Value abilities (Mind Stone, fetch lands, Clues, cycling)" (938) | `false` (v22, #2853) | `Game.svelte:485` | Same, for untargeted value activations | Same | Rarely |
| `respondSpecialActions` | "Special actions (foretell, suspend, turning a card face up)" (949) | `true` | `Game.svelte:486` | Same, for CR 116.2 special actions | Same | Rarely |
| `alwaysStopOpponentStack` | "Always stop for opponents' spells and abilities" (959), help (976-979) | `false` | `Game.svelte:530` → rule 6 (`autopassDecision.ts:193`) | Every opponent stack item holds, answer or not | **It sits inside the "Stop for" fieldset, so it is disabled with smart off** (where it is moot). **It also makes `bluffCounterspell` moot**, and nothing in the UI says so. **The session toggle ignores it** (rule 4) | Some: a careful or new player, or one who wants no tells |
| `stackHoldMs` | "Let other players' spells sit on the stack for at least … before auto-pass lets them resolve" select 0/1/2/3 s (982-1001) | `2000` | `Game.svelte:406` (`holdLeft`) → `stackHold.ts` | Delays an automatic pass while someone else's top item is younger than this | **The table already has a second hold**: `TableSettings.BotPace` gives bots 0/2/3 s (`server/internal/aiseat/runner.go:143-146`). An item resolves only once every seat has passed (CR 117.4), so the table waits for the longest hold at the table. One player's 3 s is everybody's, and a 0 does not speed anything up. Combined with a timed bluff as max, not sum (`Game.svelte:576-578`) | Rarely. It is a table matter, not a personal one |
| `autoPassOwnStack` | "Auto-pass your own spells and triggers on the stack" (1083), help (1086-1095) | `true` | `Game.svelte:522` → rule 6 (`:190`) | Your own stack passes at once | **The `hold` toggle overrides it for one stack** (`holdPriority.ts`), which is the in-game way to get the same effect | Almost never. `hold` is the right control |
| `autopassPersistThroughTurns` | "Autopass persists through your own turns", danger-styled, with an all-caps warning (1097-1114) | `false` | `Game.svelte:513` → rule 2 (`autopassDecision.ts:156`) | Off: the session toggle clears itself at your own main 1. On: it never clears | **It only means anything while the session toggle is on.** It's a footgun with a warning, and it exists for an edge case | Almost never. Effectively dead |
| `bluffCounterspell` | "Represent a counterspell…" (1018). Also the dock chip's popover (`BluffChip.svelte:100-101`), and `pressBluff` turns it on (`bluff.ts:75-76`) | `false` | `Game.svelte:538` (ANDed with `$bluffArmed`) → rules 4 and 6 | Pauses on opponent stack items you can't answer | **It needs both this setting and the session `bluffArmed` switch**, three states in all. Moot with `alwaysStopOpponentStack`. Disabled with smart off | Some, from the dock |
| `bluffInstant` | "Represent an instant…" (1027), plus the chip popover | `false` | `Game.svelte:539` → rules 4, 6, 7 | Also pauses in combat and on an opponent's end step | Same as above. **Rule 7 is gated on smart**, so it does nothing with smart off | Some, from the dock |
| `bluffMode` | "Bluff style" Timed / Manual (1031-1040), plus the chip | `"timed"` | `Game.svelte:540` | Timed passes after a random pause. Manual waits for `next` | Bluff only | Rarely |
| `bluffDelayMinMs` / `bluffDelayMaxMs` | "Shortest pause (ms)" / "Longest pause (ms)" sliders (1042-1069) | `1500` / `4000`, clamped to 500-15000 in `bluff.ts:404-421` | `Game.svelte:577` | The range of a timed bluff | Bluff only. Combined with the stack hold as max | Almost never |
| `strictMana` | "Strict mana enforcement" (1122), help in ADR 0118's words (1126-1132) | `true` (forced once at v19) | `Game.svelte:658` (stamps casts and activations, `manaEnforcement.ts`); `Hand.svelte:144`, `ExileStrip.svelte:216`, `CardContextMenu.svelte:94`, `Board.svelte:999, 2545`, `contextMenu.logic.ts:1979, 2056` (Cast anyway); the tutorial forces it on | On: costs are charged and lands auto-tap. Off: the sandbox | None in auto-pass. `engineMayMissMana` is a separate rule-8 input that does not read it | Rarely. Off is the sandbox posture |
| `adminOverrides` | "Enable admin overrides (right-click menu)" (1140) | `false` | `Card.svelte:637, 667` | Right-click opens the manual override menu | None. The label says "admin" but any player may turn it on | Rarely. It's a fallback tool |
| `showBotReasoning` | "Show bot reasoning" (1158) | `false` | `BotFeed.svelte:37` | Bot per-move narration in the feed | None | Rarely. It's debug output |

### C. Per table: `TableSettings` (host or admin; `server/internal/game/settings.go:73-105`, client `tableSettings.ts`, `TableSettingsPanel.svelte`)

| Field | Label | Default | Read at | What it changes | Couplings | Changed? |
|---|---|---|---|---|---|---|
| `undo_limit` | Undo limit, "One take-back per player, per turn." | 1 | server undo budget; client `hasUndoBudget` (`tableSettings.ts:47`) | Undo budget | `undo_scope` | Sometimes |
| `undo_scope` | "Their own" / "Host can undo anyone" | `own` | server undo | Whose moves an undo may take back | — | Rarely |
| `starting_life` | Starting life | 40 | engine `Start` | Life at start, locked once active | — | Rarely |
| `commander_damage` | Commander damage | 21 | the SBA check | The lethal threshold | — | Rarely |
| `bot_pace` | "Bot speed": Fast / Normal / Slow (`TableSettingsPanel.svelte:237-251`; hints at `tableSettings.ts:146-164`) | `normal` | `aiseat/runner.go:143-146`, re-read before every bot decision | Bot think time, plus the bot's own stack hold (0 / 2 / 3 s) | **The same idea as each human's `stackHoldMs`, set in a different place by a different person** | Sometimes |
| `allow_spawn` | Spawning | `false` | the spawner gate | Host or admin may spawn cards | — | Rarely |

None of the per-table fields is a priority setting except `bot_pace`'s stack hold.

### D. In-game controls (session state, never persisted) and how they map

| Control | Where | Key | State | Maps to a setting? | Notes |
|---|---|---|---|---|---|
| `next` | Dock action bar | Space | — | No | Passes once. Cancels a running hold or bluff |
| **hold** | Dock toggles (`ActionDock.svelte:498-512`), and **wait** on the countdown line (`:476-483`) | `h` | `holdPriority.ts`. Clears itself once the stack it held empties (#2853) | Overrides `autoPassOwnStack`, `alwaysStopOpponentStack`, smart, and categories for one stack (rule 6) | `wait` and `hold` are the same action under two names |
| **autopass** | Dock toggles (`:516-534`) | Shift+P | `autopassEnabled` in `Game.svelte:376`. Clears at your own main 1 unless `autopassPersistThroughTurns` | Overrides `autoPassPriority`, the grid, smart, `alwaysStopOpponentStack`, `autoPassOwnStack` (rule 4). **Still reads the "Stop for" categories** to hold for an answerable opponent spell | The tutorial's step-9 anchor (labels contract). It's labelled "autopass", which collides with the setting "Auto-pass priority" |
| **bluff** split chip | Dock toggles (`BluffChip.svelte`) | `b` | `bluffArmed` (`bluff.ts:427`), armed at mount if either bluff kind is on | Writes `bluffCounterspell`, `bluffInstant`, `bluffMode`. Disabled while smart is off | The pause range stays in Settings |
| Phase-icon pin | PhaseDisplay (`PhaseDisplay.svelte:193`) | — | `priorityStops.ts`, consumed on step change | Beats everything except rules 1-2 | A one-shot step stop. **It does not say whose turn it applies to; it applies to the next time that step comes, whoever's turn it is** |
| **Pass turn** | Dock action bar (`ActionDock.svelte:762-771`) | `t` | — | No | Sends the sandbox verb `pass_turn`. `Game.PassTurn` (`mutations.go:9591-9632`) jumps to the next untap. **No end step, no end-step or other step triggers, no cleanup discard** (#2881). It's the most prominent "move on" control, and it's the least rules-faithful one |

### E. Hidden couplings, collected

1. **Smart off removes stops as well as adding them.** Rule 7 is gated on `smartAutoPass`, so turning it off to "see everything" quietly drops the combat and opponent-end-step stops for a player who holds a response. That's the missed-window half of the owner's goal. #2871 split off the ticked-step half; this half remains.
2. **Greyed-out categories are still read.** With smart off, "Stop for" is disabled, but `hasResponse` and `hasPlay` still read the five categories, through rule 4 (the toggle) and rule 8. PR 2879's `disabled` condition narrows the gap; it does not close it.
3. **The step grid ignores whose turn it is.** The defaults tick main 1, main 2 and end, so a player holding any castable instant stops at every opponent's main 1, main 2 and end step. In a four-seat game that's up to nine extra stops per round, most of them with nothing worth doing. The "end" tick duplicates rule 7.
4. **The session toggle and the setting share a name and fight.** The "autopass" toggle bypasses five settings but not the categories. Its safety belt is configured by a danger setting, and it overlaps with **Pass turn** (both mean "I'm done for now").
5. **Two stack holds, one effect.** `stackHoldMs` (each human) and `bot_pace` (the host, for bots) both delay passes on someone else's item. The table waits for the longest, so the per-person setting can't make anything faster and can make everyone slower.
6. **Bluff has three gates.** The setting, the session arm, and smart on. `alwaysStopOpponentStack` silently moots `bluffCounterspell`.
7. **`highlightLegalActions` reads the verdict.** Any change to the gate chain changes which frames light up (ADR 0105 §3). The redesign has to keep `highlightsLive` in its tests.
8. **The tutorial forces `autoPassPriority` off** (`practiceTable.ts:68-71`). If the setting is replaced by a preset, the practice table must force the equivalent and restore it the same way.
9. **The server already owns two prompt settings** (`triggerOrder`, `autoAnswers`), reconciled by actions on every frame. The priority settings are client-only. A backgrounded tablet tab therefore throttles the timed bluff and stack-hold timers, and a closed tab stops passing.

### F. Dead or near-dead

- `autopassPersistThroughTurns`: a danger setting for one session toggle's edge case.
- `smartAutoPass` (after PR 2879): what remains is "stop at every opponent item", which `alwaysStopOpponentStack` already does, plus the accidental key-window switch-off.
- `autoPassPriority = false`: a mode only the tutorial uses on purpose.
- `bluffDelayMinMs` / `bluffDelayMaxMs`: two sliders for a range almost nobody tunes.
- `autoPassOwnStack = false`: duplicates `hold`, which is better because it's scoped to one stack.
- The `respond*` categories: five switches whose defaults are right for nearly everyone. They're kept for the few who aren't.
- Outside Gameplay but worth noting: `display.expandStyle` is self-declared temporary (`settings.ts:133-145`).

### G. In-flight work and what it means here

- **#2871 / PR #2879**: adds `stepStopsOnlyWhenCanAct` (v23) and combat abilities in combat windows. The ADR assumes it merges first, and starts its schema work at v24.
- **#2884**: "when it matters" skips independent triggers. That makes the trigger-order setting a rarely-touched control; it moves to the Prompts section and is not removed.
- **#2881**: Pass turn must walk every step and fire step triggers. The ADR folds Pass turn and the autopass toggle into one "Pass until…" control whose turn-end half has #2881's semantics.
- **#2880 / PR #2885**: board picks. No settings. It confirms that confirmation lives in the dock, the same place the new pass control sits.
