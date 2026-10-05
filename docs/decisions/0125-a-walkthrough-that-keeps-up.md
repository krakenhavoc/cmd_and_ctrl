# ADR 0125 — A walkthrough that keeps up: first-use hints and a refreshed tutorial

**Status:** Accepted · 2026-10-05 · S65 — A walkthrough that keeps up: first-use hints and a refreshed tutorial. The owner accepted it unchanged in review on #2317 the same day, Calls made here included.
**Issues:** [#2313](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2313) (this change, and S65's tracker); relates to [#1073](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1073) (ADR 0076's tracker, whose sub-PRs 5 and 6 this finishes).
**Owner decisions:** the three answers of 2026-10-05, quoted under [Owner decisions](#owner-decisions-2026-10-05). They are binding. This ADR also makes calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before PR 2 lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-05. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote branch. The 37 remote heads are `origin/develop`, `origin/main` and 35 chore, docs, feat, fix, repro and wip branches. The highest number on any of them is 0124, on `origin/develop`. This ADR takes **0125**.
**Amends:** [ADR 0076](0076-tutorial.md) §2.1 (eleven steps become fourteen, §5), §2.2 and its sub-PR 2 note (the practice table opens the opening roll, §5.2), §2.4 (anchors come from a label registry and a unit test guards them, §2), §2.6 (the entry becomes a hint, and the settings field is a synced settings group, §4 and §6), and §5 (sub-PRs 5 and 6 are delivered here). [ADR 0121](0121-animated-dice.md) §1's "the practice table keeps calling `Start(nil)`" and §4 (a practice bot that wins the roll hands the first turn to the player, §5.2). [ADR 0110](0110-remember-me.md) §4 item 4 and the 412 merge, for one group only (§4). The AGENTS.md §5 paragraph "Labels are a contract" (§2.5).
**Builds on:** [ADR 0076](0076-tutorial.md) (the practice table, the coach, the anchor resolver, the bus), [ADR 0110](0110-remember-me.md) §4 (settings on the account), [ADR 0111](0111-action-dock.md) (the dock, the ⋯ menu, labels as a contract), [ADR 0112](0112-signed-in-home-player-mode-and-one-decks-page.md) (the signed-in home, the decks page, admin mode), [ADR 0119](0119-a-stack-you-can-follow.md) (the stack pile), [ADR 0120](0120-expand-a-players-board.md) (the expanded board and `boardAnchor`), [ADR 0121](0121-animated-dice.md) (the opening roll and table dice), [ADR 0092](0092-public-roadmap-and-site-portal.md) (the site header), [ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md) (the admin views), [ADR 0047](0047-keyboard-shortcuts.md) (the keymap and its overlay).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The client changes fast, and the only walkthrough it has goes stale as fast. The practice tutorial's nine middle steps landed on 2026-10-02 (#2002, #2005). In the next three days four more PRs had to change the tutorial's files to keep it true (#2195, #2222, #2225, #2248), and #2300 touched them again on 2026-10-05. There were 72 `feat(client)` and `fix(client)` commits in the two weeks before this ADR. The tutorial is one long script, so every one of those changes is a reason to reread all of it.

### What exists (checked on `origin/develop` at `4d86bfc9`)

**The tutorial** (ADR 0076).
- `client/src/lib/tutorial.ts` is the step machine: pure, six coach states, with detours (`first`), give-ups (`cannot`), hover steps and timeouts. `tutorialSteps.ts` is the script: eleven steps. `tutorialAnchor.ts` resolves an anchor. `components/tutorial/` holds `CoachCard`, `TutorialScrim` and `TutorialCoach`, which `Game.svelte` mounts on this tab's practice table only.
- The steps in the code today: welcome; read your hand; play a land; lands stack into piles; tap a land; cast a creature (with a "Now let it resolve" detour); abilities on right-click; move the turn along; let the bot play (the dock's autopass, the owner's choice of 2026-10-02); attack (with a "Watch the bot play" detour); hand-off. ADR 0076 §2.1's table predates the autopass step.
- **An anchor is an `aria-label` string written in `tutorialSteps.ts`**: `your hand`, `lands` and `creatures` within `your board`, `actions`, `autopass` within `actions`, and `<name> board` for the bot's panel. A card is anchored by `data-instance-id` and a portrait by `data-seat-id`, both through `boardAnchor` (ADR 0120 §3). `resolveAnchor` matches `[aria-label="…"]` exactly.
- **A missing anchor is silent to CI.** The step waits `ANCHOR_GRACE_MS` (1.5 s), then advances itself and logs `tutorial: step <id> has no anchor on the page; advancing` to the console. The player loses that step. The unit tests (`tutorialSteps.test.ts`, `tutorialCoach.render.test.ts`) build stub boards from the same strings, so they pass whatever the components render. `board-layout.spec.ts` asserts some of the labels, and PR CI runs no Playwright, so a rename is caught by the nightly at best, and only for the labels that spec reads.
- **Entry.** The only doors are the Lobby's empty-state link "Practice against bots" (`Lobby.svelte:806`), shown to a signed-in person with no tables, and the coach's own Replay. **ADR 0076 sub-PR 5 was never built:** `settings.ts` has no tutorial field, Settings → Advanced holds only Export and Import, and there is no first-visit offer. **Sub-PR 6 was never built:** `tutorial-layout.spec.ts` opens a practice table to measure the coach's layout at two sizes, then presses Skip tutorial. Nothing walks the steps.
- **Not covered at all:** Home, the Lobby's create form, the decks page and its deck check, Settings and its skins, the Catalog, the Roadmap, the admin views; and at the table the stack pile (ADR 0119), the command zone, the opening roll, the ⋯ menu and its dice (ADR 0121), the expanded board (ADR 0120), the attention strip, and the keymap.

**The practice table** (`server/internal/lobby/practice.go`). One human seat and one `random` bot on the two tutorial decks (`decks/tutorial.go`, each with a commander: Marwyn, the Nurturer and Azusa, Lost but Seeking). It calls `g.Start(nil)`, not `StartWithOpeningRoll`, so the human at seat 0 takes turn one and there is no roll (`practice.go:206`). Bots are started through `aiseat.SeatSpec` (`aiseat/manager.go:18`: player, tier, deck). A bot that wins the opening roll chooses itself through `aiseat.OpeningRollIndex` (`aiseat/opening_roll.go`), except the `random` tier, which picks uniformly (ADR 0121 §4).

**Settings** (`client/src/lib/settings.ts`, ADR 0110 §4).
- One versioned object, `SETTINGS_VERSION = 20`, in `localStorage`. `migrate` merges group by group over the defaults.
- `SYNCED_FIELDS` classifies every field as `synced` or `device`, and a test fails on an unclassified one. A signed-in person's synced fields go to `PUT /me/settings` (a JSON object of at most 32 KiB and depth 4, `If-Match` revision, 1 write per second). Guests, the admin token and a server with no database get 403 and stay browser-only.
- At sign-in the account's copy wins, and a toast offers "Keep this browser's instead". A 412 is merged field by field against the last agreed copy (`settingsSync.ts`).

**Help today.** The keymap overlay (`ShortcutsOverlay.svelte`, `dialog "keyboard shortcuts"`) is global, mounted by `ShortcutLayer` in `App.svelte`, and opens on `?` (`toggleHelp`). Home's Help group lists the bot guide, the source and the bug tracker. The site header (`SiteHeader.svelte`) has no help control. The table has no header: its ⋯ menu (`GameMenu.svelte`, `button "more actions"`, `menu "game actions"`) has Sandbox, Dice and Table groups.

**The labels contract** is a prose paragraph in AGENTS.md §5 that lists about forty names. Nothing checks it against the code.

### Owner decisions (2026-10-05)

From #2313:

1. **Shape:** first-use **hints** per page or feature (small, dismissible, owned by their feature, so a UI change touches one hint, not a long tour), **plus** a refreshed practice tutorial (stack lane, commander zone, opening roll, and so on).
2. **Trigger:** offered once, the first time someone meets the page or feature; remembered per account (synced settings) or per browser for guests; every hint and the tutorial replayable from a Help menu or Settings. This finishes ADR 0076's unbuilt offer and replay.
3. **Sync guard:** a fast unit test on **every** PR fails when any hint or tutorial step anchors to a label no component renders any more; one e2e walk of the tour runs nightly.

---

## Decision

### 1. The shape

```
 client/src/lib/labels.ts  ── the contract labels: one entry each, with the files that render it
        ▲            ▲                    ▲
        │            │                    └── components render L.<key>, never the literal
        │            └── tutorialSteps.ts anchors through it (typed)
        └── *.hint.ts beside each feature anchors through it (typed)
                     │
                     ▼
 lib/hints/  collects every *.hint.ts ── HintLayer (App.svelte) shows one at a time
                     │                          │
                     ▼                          ▼
 settings.help.seen (synced, or browser-only)   Help menu (header, ⋯ menu) and Settings → Advanced
```

- **One registry of contract labels** (§2). Hints and tutorial steps can only anchor to a label in it, and a unit test fails when a registered label's owner file stops rendering it.
- **Hints** (§3): one small file per hint, beside the feature it explains, collected automatically. One hint on screen at a time, never over a decision the player has to make.
- **Seen state** (§4): a `help` group in the existing settings object, synced for signed-in people, browser-only for everyone else.
- **The tutorial** (§5) grows from eleven steps to fourteen, and the practice table opens the opening roll.
- **Help** (§6): a Help menu in the site header and a Help group in the ⋯ menu, plus a Tips section in Settings → Advanced.

### 2. Contract labels: one registry, and the guard

#### 2.1 The registry

`client/src/lib/labels.ts` (new) exports one object, `L`. It imports nothing, so the e2e package can import it too. Each entry is one of:

```ts
// A fixed name.
yourHand: label({ name: "your hand", kind: "aria", owners: ["lib/components/board/Hand.svelte"],
                  doc: "the viewer's hand; tutorial steps 3, 4 and 7 anchor here" }),

// A name with a variable part. `stem` is the fixed part, and `match` says where it sits.
commandZone: dynamicLabel({
  stem: " command zone, ", match: "contains", kind: "aria",
  make: (seat: string, n: number) => `${seat} command zone, ${n} card${n === 1 ? "" : "s"}`,
  owners: ["lib/components/board/CommandZone.svelte"], doc: "a seat's command zone panel" }),
```

- `kind` is `aria` (an `aria-label`, or a dialog's or group's name given through a prop that becomes one) or `text` (a name a button or link gets from its visible text, like `Keep hand`). Only `aria` entries can be anchors, because `resolveAnchor` matches `aria-label`.
- `owners` lists the files, relative to `client/src`, that render the label. Most entries have one.
- `doc` is one line for the generated list (§2.5).
- **A component renders a registered label as `L.<key>`** (or `L.<key>(…)` for a dynamic one), never as a literal.

**What goes in.** Every label a hint or tutorial step anchors to, every label the e2e walk (§8) selects on, and every name in today's AGENTS.md list. Other `aria-label`s stay literals. The registry is the contract, not an inventory of the whole client.

#### 2.2 Anchors

`Anchor` (`tutorial.ts`) changes from `{ label: string; within?: string }` to `{ label: LabelRef; within?: LabelRef }`. A `LabelRef` is a branded type that only `L` can produce, so `npm run check` (svelte-check, in the `client` job) refuses a hint or a step that anchors to a raw string. A dynamic entry used as an anchor carries its `stem` and `match`, and `resolveAnchor` turns them into the CSS attribute operators `^=`, `$=` or `*=` instead of `=`. `cardID` and `seatID` anchors are unchanged.

Today's tutorial anchors move onto it: `L.yourHand`, `L.lands`, `L.creatures`, `L.yourBoard`, `L.actions`, `L.autopass`, and `L.seatBoard(name)` for the bot's panel.

#### 2.3 The guard

`client/src/lib/labels.test.ts` (vitest, so it runs in the `client` job on every PR that touches `client/` or `tests-e2e/`). It reads source files with `node:fs` and renders nothing, so it takes well under a second.

1. **Every entry is rendered by its owners.** Each file in `owners` exists and contains `L.<key>`. Delete the hand, or move it to a new component without updating the entry, and this fails, naming the entry, its owner and every hint and step that anchors to it.
2. **No literal copies of a registered `aria` name.** No non-test `.svelte` or `.ts` file under `client/src` other than `labels.ts` contains a registered static `aria` name as `aria-label="…"`, `aria-label={"…"}`, `label="…"` or `label: "…"`. This is what makes rule 1 sufficient: a component cannot keep rendering a contract name through a literal that the registry does not see. It also keeps contract names unique, which the anchor resolver needs: it takes the first match.
3. **Every anchor is registered and owned.** It walks every hint (§3) and every tutorial step, including each detour, by calling their anchor functions over the fixture boards the step tests already use, and checks each `label` and `within` against `L`. The types catch most of this at compile time; the test catches an anchor built by a function.
4. **The generated list is current** (§2.5).

**What it cannot see.** It proves that the owner file references the label, not that the element is on screen at the moment a hint needs it. A label behind an `{#if}` that never comes true passes. That gap is what the nightly walk (§8) is for, and why that walk fails on any "has no anchor" log line.

#### 2.4 Why a registry and not a scan

The owner's guard needs to know which labels the components render, without a browser. Three ways were weighed:

- **Scan `client/src` for every `aria-label`.** Cheap, and needs no change to components. Rejected because of what it gets wrong. Many contract labels are built at run time: `<name> goes first`, `Expand <name>'s board`, `stack: N on the stack`, `<seat> command zone, N cards`, `Discard N card`. Others are passed through props (`OpeningRollDock` gives the dock `label: "roll for the first turn"`; `Icon` takes `label=`), or chosen by a ternary (`joinBox === "code" ? … : …`). A scan would need a partial evaluator, would still miss some, and cannot tell a contract label from an incidental one, so it either reports noise or misses renames.
- **Render each owning component in vitest and look.** Accurate where it works, but most owners need a game view, stores and props to render at all, and the board takes seconds. That is the e2e suite's job, done slower.
- **A registry the components and the anchors both import — chosen.** A rename becomes an edit to one value, and every hint and step follows it with no further change. A removal fails rule 1. A literal copy fails rule 2. Its cost is one migration PR (Delivery PR 2) that changes about forty `aria-label` attributes from literals to `L.<key>`. The names do not change, so the e2e suite and screen readers see the same page.

#### 2.5 The AGENTS.md list is generated from the registry

`docs/labels.md` (new) is generated from `L`: one row per entry with its name, role, kind, owners and `doc`. Rule 4 of the guard fails when the file is stale, and `UPDATE_LABELS_DOC=1 npm test -- labels` rewrites it. The AGENTS.md §5 paragraph "Labels are a contract" shrinks to the rule and a pointer:

> **Labels are a contract** (ADR 0076 §2.4, ADR 0111 §10, ADR 0125 §2). The names the tutorial, the hints and the e2e suite rely on live in `client/src/lib/labels.ts`, listed in `docs/labels.md` (generated). Render one as `L.<key>`, never as a literal. To rename one, change its value in the registry; hints and tutorial steps follow on their own. The older e2e specs select on literals on purpose, so a rename turns the nightly red until they are updated: run it on the branch (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`). If you change how a feature is used, update its `*.hint.ts` in the same PR (ADR 0125 §7).

The existing specs keep their literals. They are the deliberate tripwire for a rename that was not meant. The new specs (§8) import `L`.

### 3. Hints

#### 3.1 What a hint is

```ts
export interface Hint {
  id: HintID;          // "<place>.<feature>", e.g. "table.stack". Stable; never reused.
  version: number;     // starts at 1; bumped when the feature changes materially (§3.3)
  place: HintPlace;    // a route name ("lobby", "decks", "catalog", …), "settings", "site" (any site page) or "table"
  order: number;       // which of a place's hints comes first
  anchor: Anchor | ((c: HintContext) => Anchor | null);
  title: Copy;         // tutorial.ts's Copy: a string, or a function of the player's key bindings
  body: Copy;
  when?: (c: HintContext) => boolean;   // extra conditions: signed in, admin mode, three or more seats, …
  action?: { label: string; href: string };  // one link button; only the tutorial offer uses it
}
```

`HintContext` carries the route, the session (signed in, admin mode, has a finished game), and at the table a read-only `TableMoment` (§3.5) and the game view. A hint reads nothing else.

#### 3.2 One file per hint, beside its feature

Each hint is a file `<Feature>.hint.ts` next to the component that owns the feature: `lib/components/board/StackLaneHost.hint.ts`, `routes/Decks.hint.ts`. `lib/hints/index.ts` collects them with `import.meta.glob("../**/*.hint.ts", { eager: true })`, which Vite and vitest both support.

- **No central list.** Two PRs that add hints never conflict, and the PR that changes `StackLaneHost.svelte` shows `StackLaneHost.hint.ts` right beside it in the file tree.
- **Retired ids.** A removed hint's id goes in `lib/hints/retired.ts`. A test fails when a live hint reuses a retired id, because someone who saw the old hint would never see the new one.

#### 3.3 Versions, and what "materially changed" means

`settings.help.seen` maps a hint id to the version the person last dismissed. A hint is unseen when its id is missing or its stored version is lower than the hint's own. Bumping `version` therefore offers the hint once more to everyone, including people who saw the old one.

**Bump it when a returning player would now be wrong:**
- the anchor moved to another part of the screen or to another control;
- the gesture changed (a left click became a right-click; an item moved from the ⋯ menu into the dock);
- the feature gained something the old copy would hide, and the new copy says so.

**Do not bump it** for a wording or typo fix, a restyle, or a label rename at the same place. Those change the copy or the registry and nothing else.

#### 3.4 Copy rules

- **A title of at most 32 characters and a body of at most 140**: one or two sentences, second person, present tense.
- **Name controls by their visible text** (`next`, `Pass turn`, `More actions`). A key is named through `CopyContext`, as the tutorial does, so a rebound key reads right.
- **No engine or project vocabulary.** No "engine", "server", "client", "snapshot", "seam", "oracle", "payload" or "websocket". The admin hint may say "database"; its readers run it.
- **No rules teaching.** The audience knows Magic (ADR 0076 §1). A hint says where something is and how to use it.
- `lib/hints/hints.test.ts` enforces the lengths, the banned words and a final full stop. It also checks that every id matches `<place>.<feature>`, that every place is known, and that every version is a positive integer.

#### 3.5 When a hint shows

**One on screen at a time, everywhere.** A second unseen hint waits.

**Site pages.** At most one hint per page visit: the first unseen hint for this place (then `site`) in `order` whose `when` holds and whose anchor resolves within 3 seconds of the route mounting. A hint whose anchor does not resolve is skipped for this visit, stays unseen, and logs `hint: <id> has no anchor on the page` to the console, as the tutorial does. A hint never makes the page wait.

**The table.** A table hint shows only in a quiet moment. `TableMoment` is computed in `Game.svelte` from what it already holds:
1. No dock request is open for the viewer: mulligan, target, payment, a choice, trigger order, attackers or blockers, the opening-roll sheet or its confirm, or a revealed-hand pick.
2. The viewer is not mid-gesture: no targeting, no drag, no open card menu.
3. The top of the stack, if there is one, is not another player's item while the viewer holds priority. That is the moment the viewer decides whether to answer it, and ADR 0119's stack hold gives them only a few seconds. A hint never draws the eye then.
4. The table has been on screen for at least 5 seconds, and no hint has closed in the last 20.

If a moment stops being quiet while a hint is up, the hint steps aside at once and stays unseen. It comes back at the next quiet moment. A hint is never on screen while the viewer has a decision to make, which is the CR-faithful reading of "never obscure a decision": the player always sees what the game is asking of them.

**The practice table.** Table hints are off while the coach is visible. After Skip tutorial or Finish they run as at any table. A tutorial step that completes marks the hints it teaches as seen (§5.3).

**"Hide tips".** Every hint card has it beside "Got it". It sets `settings.help.tipsOff`, and from then on no hint is offered until it is switched back on in Settings or the Help menu replays one. The Help menu's replays (§6) work with tips off, because the person asked.

#### 3.6 What it looks like

- **A card and a ring, no scrim.** The card is about 280px wide and points at its anchor. The anchor gets a 2px outline (`pointer-events: none`). Nothing dims, and nothing under the card loses its clicks except what the card itself covers.
- **Placement** is a pure function (`lib/hints/place.ts`, unit-tested on rectangles). It tries below, above, right and left of the anchor, and takes the first side where the card fits without covering the anchor, the action dock (`--dock-w`, `--dock-h`), the stack pile, an open dialog or sheet, or the coach (`--coach-w`, `--coach-h`). If none fits, the hint waits, as for a moment that is not quiet. It follows the anchor on resize and scroll at the coach's poll rate (`POLL_MS`).
- **A hint inside a dialog.** Settings is a modal that traps focus. A hint whose place is `settings` renders inside the dialog through a `<HintSlot>` the dialog places, so its buttons are inside the trap.
- **Phones (≤599px).** The card is a one-line strip, stacked on the dock bar at the table and on the bottom edge elsewhere, with the anchor ringed. A tap opens it to its two lines and buttons. This is the coach card's phone layout (ADR 0076 §2.3, second amendment).
- **Reduced motion** (`accessibility.reduceMotion` or the OS signal). No fade, no slide, and the ring does not pulse. The card simply appears.
- **Keyboard.** A hint never takes focus and never swallows a key; the keymap and the dock keep working. A new global shortcut, **Go to the tip** (`focusTip`, default `i`, rebindable, listed in the `?` overlay), moves focus to the card. There, Tab moves between "Got it", "Hide tips" and any action, Escape dismisses the card (it counts as Got it) and returns focus to where it was. With no tip showing, `i` does nothing and says nothing. The default follows ADR 0047: an unmodified letter that no table action uses.
- **Screen readers.** The card is `complementary "tip"` (a new registry label), so landmark navigation finds it. When it appears, the app's polite live region reads "Tip: <title>. <body>" once. While it shows, its anchor gets `aria-describedby` pointing at the body, so someone who tabs to the anchor hears it.

#### 3.7 The initial set

Copy is a draft. Each PR that adds a hint checks its copy against the page it ships with. "New" in the anchor column means the PR adds the label to the registry and its owner.

**Site pages**

| Id | Where | Anchor | When | Title — body |
|---|---|---|---|---|
| `site.help` | the first site page visited | `button "help"` (new, §6) | always first | **Tips show once** — Each page shows a tip like this the first time you visit. Help, up here, shows them again and starts a practice game. |
| `lobby.practice` | Lobby | `heading "cmd_and_ctrl · lobby"` (the title's existing label) | a guest, or a signed-in person whose `GET /me/games` has no ended game; never the admin token | **New here?** — A five-minute practice game against a bot shows you where everything is. Action: **Start practice** (`#/practice`); "Got it" reads **Not now**. |
| `lobby.create` | Lobby | `region "create game"` (new) | signed in | **Start a table** — Name it and create it. Its card then holds the invite link, the seats you can give to bots, and your deck. |
| `decks.check` | Decks | `textbox "deck link"` | always | **Check a deck** — Paste a link or a list. The report shows which cards the game plays for you and which you play by hand. |
| `decks.library` | Decks | `list "your decks"` (shown once the person has a saved deck) | signed in | **Your decks** — Save a checked deck here and it is offered when you sit down at a table. |
| `settings.display` | Settings | `navigation "settings sections"` | the first time Settings opens | **Skins are under Display** — Pick a skin and an accent colour there. Signed in, most settings follow you to your other browsers. |
| `catalog.search` | Catalog | `textbox "search the catalogue"` | always | **Every card the game plays** — Search for a card by name to see how much of it the game handles for you. |
| `roadmap.search` | Roadmap | `textbox "search the roadmap"` | always | **What comes next** — Search for a card to see what it is waiting on before the game can play it for you. |
| `admin.views` | Admin views | `navigation "admin views"` | admin mode on | **Admin views** — Live now, every table and every account, read from the live database. Archive, replay and revoke are linked from their rows. |

**The table**

| Id | Anchor | When (beyond a quiet moment) | Title — body |
|---|---|---|---|
| `table.dock` | `region "actions"` | the first quiet moment | **Your controls** — The next button moves the game on, Pass turn ends your turn, and anything the game needs from you opens here. |
| `table.opening-roll` | `group "opening roll"` | the viewer has rolled and others have not | **Rolling for the first turn** — Everyone rolls at once. The highest roll chooses who goes first, and a tie rolls again. |
| `table.stack` | the stack pile, `stack: N on the stack` (prefix `stack: `) | the stack holds an item (rule 3 of §3.5 keeps it off an opponent's live item) | **The stack** — Spells and abilities wait here, newest on top, until every player passes. Rest the pointer on one to read it. |
| `table.right-click` | `region "your board"` | the viewer controls a permanent with an ability | **Abilities live on right-click** — Right-click a permanent to see its abilities and use them. On a touch screen, tap its pip. |
| `table.commander` | the viewer's command zone (contains ` command zone, `, within `your board`) | a commander is in the zone | **Your commander** — It waits in the command zone and beside your hand. Click it to cast it, as you would a card in hand. |
| `table.attention` | `region "attention"` | the strip shows something for the first time | **News, top left** — What the bots are doing, cards revealed to you and the roll call show up here. None of it waits on you. |
| `table.more` | `button "more actions"` | the viewer's second turn | **More actions** — Dice and a coin, untap all, life history, table settings and Help are in here. |
| `table.expand` | the first opponent's board, `L.seatBoard(name)` | three or more seats | **See a board up close** — Rest the pointer on a board and its ⤢ opens it full size over the table. |
| `table.shortcuts` | `region "actions"` | a fine pointer and a hover-capable device, `table.dock` seen, the viewer's second turn; both keys are named from the player's bindings | **Keys for everything** — Press ? to see every keyboard shortcut. Space presses next. |

That is 18 hints: nine on site pages and nine at the table.

#### 3.8 Deliberately left out

- **Home.** It is itself the site map, and `site.help` points at Help from whichever page comes first.
- **My games, Join, Reclaim, Login and Practice.** Each does one thing and says so.
- **The Lobby's join box.** Its placeholder already says what to paste.
- **Self-explanatory dialogs:** keep or mulligan, targets, payment, the opening roll's sheet and confirm, the revealed-hand pick, votes. Each asks its own question.
- **Steps the tutorial covers that need no reminder at a real table:** the hand's peek, land piles, tapping a land.
- **Sandbox and host tools:** bluff, spawn, table settings, the game log drawer, life history. They are found in the menus that hold them, and a hint for each would turn hints into a tour.
- **Rules.** No hint explains commander tax, damage, priority as a concept or any other rule (ADR 0076 §1).

### 4. Where "seen" lives

**A new settings group, `help`:**

```ts
help: {
  seen: Record<string, number>;  // hint id → version dismissed
  tipsOff: boolean;              // "Hide tips"
}
```

- `SETTINGS_VERSION` 20 → 21, with a migration that adds the group with `seen: {}` and `tipsOff: false`. `SYNCED_FIELDS` classifies both fields `synced`, so a signed-in person's seen hints follow them to every browser, and a guest's, an admin token's or a no-database server's stay in this browser. That is owner decision 2 with no new storage: the per-account copy is ADR 0110's `user_settings` row.
- **The tutorial offer is a hint** (`lobby.practice`), so ADR 0076 §2.6's "settings-store field, moved to the user row when S34 lands" becomes this group. Its tracked obligation is discharged: the field is on the account through ADR 0110 §4, not only in a browser.
- **Size.** Each entry is about 30 bytes of JSON. The initial 18 hints come to about 0.6 KiB. The 32 KiB cap would need about a thousand hints. The body's depth is 3 (`help.seen.<id>`), inside the limit of 4.
- **No pruning of unknown ids.** Adding a hint does not bump `SETTINGS_VERSION`, so an older tab may meet ids it does not know. If it dropped them on its next write, a newer browser would offer those hints again. Only ids in `retired.ts` are dropped.
- **The practice record is untouched.** The tutorial forces four other fields (ADR 0076 §2.2); `help` is not one of them, and it syncs normally during practice.

**Merging, an exception to ADR 0110 §4 for `help.seen` only:**
- **At sign-in**, the account's copy still wins for every other field, `help.tipsOff` included. `help.seen` takes the **union** of the browser's map and the account's, with the higher version for an id in both. A hint someone dismissed as a guest on this browser does not come back when they sign in.
- **On a 412**, `help.seen` is merged the same way instead of field-wins.
- **The "Keep this browser's instead" toast** is not raised by a difference in `help.seen` alone. A union cannot lose anything either side had.

### 5. The tutorial refresh

#### 5.1 Fourteen steps

| # | Step | Anchor | Advances when | Teaches (§5.3) |
|---|---|---|---|---|
| 1 | A five-minute practice game | — | button | — |
| 2 | **Roll for the first turn** (new) | `dialog "roll for the first turn"`; detour to `dialog "choose who takes the first turn"` while the viewer is the chooser | the view has no `opening_roll` | `table.opening-roll` |
| 3 | Read your hand | `your hand` | hover ≥ 600 ms | — |
| 4 | Play a land | `your hand` | controlled land count +1 | — |
| 5 | Lands stack into piles | `lands` within `your board` | hover ≥ 600 ms | — |
| 6 | Tap a land for mana | `lands` within `your board` | the viewer's mana pool is not empty | — |
| 7 | Cast a creature | `your hand` | a creature of the viewer's is on the stack **or** on the battlefield | — |
| 8 | **Your spell is on the stack** (new; replaces step 6's "Now let it resolve" detour) | the stack pile (prefix `stack: `) | the stack is empty and the creature has arrived | `table.stack` |
| 9 | Abilities live on right-click | the newest creature with a menu | `ability-menu-opened` | `table.right-click` |
| 10 | **Your commander** (new) | the viewer's command zone | hover ≥ 600 ms | `table.commander` |
| 11 | Move the turn along | `region "actions"` | the step changed | `table.dock` |
| 12 | Let the bot play | `autopass` within `actions` | autopass on and the bot's turn running | — |
| 13 | Attack | the creature row and the bot's portrait | a creature of the viewer's is attacking | — |
| 14 | That is the whole interface | — | button | `table.more`, `table.shortcuts` |

Copy for the new steps (drafts, as in §3.7):
- **2, Roll for the first turn.** "Everyone rolls for the first turn at the start of a game. Press Roll." Its chooser detour: "You won the roll. Choose who goes first; for this game, take it yourself."
- **8, Your spell is on the stack.** "Your creature waits on the stack, on the left, until both players pass. Press next (Space) and it resolves." The key comes from `CopyContext`.
- **10, Your commander.** "Your commander waits here, in the command zone, and beside your hand. Once your lands can pay for it, click it to cast it."
- **14, the hand-off,** gains a sentence: "Help, in the ⋯ menu here and in the header elsewhere, replays any of this."

Other changes:
- **Step 7's** `done` also accepts the creature on the stack, so step 8 can teach the pile while the spell is there. Its `first` loses the resolve detour, which step 8 now is. Step 8 gives up (`cannot`) if the stack is already empty when it begins, for instance when the creature resolved before step 7 saw it on the stack.
- **Step 10 is a hover step with no bus event.** `hover.event` becomes optional. On a device with no hover, a step with no event moves on after `TOUCH_HOVER_STEP_MS` (12 s), which the coach already does. **The bus does not grow:** it keeps its three events (ADR 0076 §2.5).
- **`TUTORIAL_STEP_COUNT`** becomes 14, so the card reads "N / 14".
- The commander step needs no mana. Marwyn costs three, which the player cannot have in two turns, so the step teaches where the commander is and how it is cast, not the cast.

#### 5.2 The practice table opens the opening roll

`lobby/practice.go` calls `g.StartWithOpeningRoll(nil)` instead of `g.Start(nil)`. Both seats roll, ties roll again, and the winner chooses, exactly as at any table (CR 103.1).

So that the tutorial still follows the player's first turn, **a practice bot that wins the roll hands the first turn to the player.** CR 103.1 lets the winner choose any player, so this is a legal choice, not a rigged roll.
- `aiseat.SeatSpec` gains `FirstTurnTo *int`. Only the practice route sets it, to the human's seat.
- When it is set, the runner answers a window whose moves are all `choose_starting_player` with the move naming that seat. Every other window, rolling included, goes to the `random` tier unchanged. This is one window of one table type, not a tutorial tier, which ADR 0076 §2.2 rejected.
- When the player wins, step 2's detour asks them to take the first turn themselves. A player who hands it to the bot anyway gets the existing "First, your turn" detours on steps 4 and 7.

The roll is real, so the player sees the dice, the banner and the choice they will meet at their first real table. ADR 0121's practice exception goes, and with it the one place where a game here starts without a roll.

#### 5.3 Steps teach hints

A step may list `teaches: HintID[]`. When the step **completes** (its predicate fires), those hints are marked seen at their current version. A step that is skipped, gives up or advances itself marks nothing, so the hint is still offered at the player's first real table. Finishing the tutorial also marks `lobby.practice` seen; opening a practice table from anywhere does too.

### 6. Help, and the way in

**The site header** gets a **Help** button (`button "help"`, an icon with that name) between the site nav and the account control, for everyone, signed in or not. It opens `menu "help"`:
- **Tips for this page:** clears `seen` for this route's place and shows them one after another, ignoring the one-per-visit rule, because the person asked.
- **Show all tips again:** clears `seen` and switches tips back on. Tips then appear as pages are visited.
- **Practice game:** `#/practice`. Signed out, it reads **Practice game (sign in first)** and goes to `#/login`, since a practice table needs a session.
- **Keyboard shortcuts:** opens the existing overlay.

**At the table**, the ⋯ menu gets a **Help** group after Table:
- **Tips for the table:** clears `seen` for `table` and shows them at the next quiet moments.
- **Keyboard shortcuts.**
- **Replay the tutorial**, on the practice table only, which is what the coach's Replay does. "Practice game" is not offered at a real table: it would take the player out of their seat.

**Settings → Advanced** gets a **Tips and the tutorial** section above Export: a **Show tips** checkbox (the inverse of `help.tipsOff`), **Show all tips again**, and **Replay the tutorial** (`#/practice`). This is ADR 0076's "Replay lives in Settings → Advanced".

**Home**'s Help group gets a **Practice game** tile beside the bot guide.

**The tutorial offer** is the `lobby.practice` hint (§3.7): offered once, for the people ADR 0076 §2.6 named, with a permanent "Not now". The Lobby's existing empty-state link stays.

### 7. How a UI PR keeps hints current

The rule for authors, which PR 2 adds to AGENTS.md §5 (§2.5):

1. **You rename or move a contract label:** change its value or its `owners` in `labels.ts`. Hints and tutorial steps follow on their own. If the guard fails, it names the entry and everything that anchors to it.
2. **You change how a feature is used** (where it is, the gesture, what its control says): edit that feature's `*.hint.ts` in the same PR, and bump its `version` if §3.3 says the change is material. If the tutorial teaches it, edit the step too.
3. **You remove a feature:** delete its `*.hint.ts` and add its id to `retired.ts`.
4. **Your change touches a label or a step:** run the nightly E2E on the branch (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), as AGENTS.md already asks for a label rename.

**Why this keeps the per-change cost small.** A feature's hint is one file of about fifteen lines, sitting beside the component the PR already changes. No hint knows about any other, so no change reaches past its own feature. A rename costs nothing, because it happens in one place. The cost the owner wanted to avoid, rereading a long tour after every UI change, is now confined to the tutorial's fourteen steps, which cover only the practice table, and the nightly walk tells you which step broke.

### 8. Tests

**Client (vitest, every PR).**
- **The label guard** (§2.3): owners render their entries, no literal copies, every anchor registered, `docs/labels.md` current.
- **Hints** (`lib/hints/hints.test.ts`): unique ids in the `<place>.<feature>` form, no retired id reused, known places, positive versions, the copy rules (§3.4), and a non-empty anchor for each hint over a fixture context.
- **Showing** (`lib/hints/queue.test.ts`): one at a time; one per site-page visit; the order; `when`; `tipsOff`; a missing anchor skips and logs; replays ignore `tipsOff` and the one-per-visit rule.
- **The quiet moment** (`tableMoment.test.ts`): each rule of §3.5 against stub views, including an opponent's item on top while the viewer holds priority (not quiet) and the viewer's own item on top (quiet).
- **Placement** (`place.test.ts`): never over the anchor, the dock, the stack pile, a dialog or the coach; waits when nothing fits; the phone strip below 600px.
- **Seen state:** the v20 → v21 migration; `SYNCED_FIELDS` classifies the group; the sign-in union and the 412 union; no toast for a `help.seen` difference alone; retired ids dropped, unknown ids kept.
- **Rendering:** `HintCard` has `complementary "tip"`, the anchor gets `aria-describedby`, Escape dismisses and returns focus, `i` focuses the card, reduced motion drops the transition classes.
- **The tutorial:** the three new steps over stub boards (the roll and its chooser detour, the stack step and its give-up, the commander hover with no event), `teaches` marks only on completion, and the step count.

**Server (Go).**
- `TestPracticeTableOpensTheOpeningRoll`: a new practice table has an open opening roll and no turn yet.
- `TestPracticeBotHandsTheFirstTurnToThePlayer`: with the roll forced so the bot wins, the bot chooses the human's seat; with the human winning, the human is the chooser and the bot chooses nothing.
- `FirstTurnTo` unset changes nothing: the existing opening-roll tests for the random tier stay green, unchanged.

**E2E (nightly; ADR 0076 sub-PR 6).**
- **`tutorial-walk.spec.ts`:** signs in with the admin token as `tutorial-layout.spec.ts` does, opens `#/practice`, and walks all fourteen steps by doing each gesture: Start, Roll (and "I go first" if the page is the chooser), Keep hand, rest on the hand, play a Forest, rest on the lands, tap it, cast a lit creature, press next, right-click, rest on the command zone, press next, autopass on, wait for the bot's turn, attack, Finish. It asserts each step's title as it arrives. It **fails on any console line containing `has no anchor`**, which is the regression test for the anchors. It tolerates `cannot happen` lines, because the decks are shuffled and a hand may have no castable creature.
- **`hints.spec.ts`:** in a fresh browser context, visits each site page and asserts its hint appears beside its anchor, dismisses it, reloads and asserts it is gone, then uses Help → Tips for this page and sees it again. It starts a two-person table through `startGameAs`, sees `table.dock` in a quiet moment, reaches it with `i`, and checks that no hint is on screen while the mulligan dialog is open. It also fails on `has no anchor`.
- `tutorial-layout.spec.ts` presses Roll (and chooses) before Keep hand once the practice table rolls (Delivery PR 3).

---

## Delivery

Each PR goes into `develop` under Sprint S65 and Issue #2313.

| PR | What | Size | Needs | Parallel with |
|---|---|---|---|---|
| 1 | **This ADR**, the AGENTS.md §3 ADR range line, and the S65 section and index row in `docs/sprints.md`. Docs only. | S | — | anything |
| 2 | **The label registry and its guard** (§2): `lib/labels.ts`; about forty labels in some thirty files change from literals to `L.<key>`; the `Anchor` type; the tutorial's anchors; `labels.test.ts`; `docs/labels.md`; the AGENTS.md §5 paragraph. No name changes. Runs the nightly E2E on the branch. | M | 1 | 3 |
| 3 | **The practice table rolls** (§5.2): `StartWithOpeningRoll` in `practice.go`, `SeatSpec.FirstTurnTo` and its window, the Go tests, `docs/lobby.md`'s practice section, and `tutorial-layout.spec.ts`. The current eleven steps tolerate the roll, since it is the dock's own request, so this may merge before PR 8. | S | 1 | 2, 4–7 |
| 4 | **The hint engine** (§3.1–3.6, §4): `lib/hints/` (collection, queue, placement, `TableMoment`), `HintLayer` in `App.svelte`, `HintCard`, `HintSlot`, the `help` group with its migration and merge rules, `focusTip`, and the tests. It ships no hint except a test fixture. | L | 2 | 3 |
| 5 | **Help and the way in** (§6): the header's Help menu, the ⋯ menu's Help group, Settings → Advanced, Home's tile, and the `site.help` and `lobby.practice` hints. | M | 4 | 6, 7 |
| 6 | **Site hints** (§3.7, site rows except the two in PR 5), with the labels they add. | M | 4 | 5, 7 |
| 7 | **Table hints** (§3.7, table rows), with the labels they add. | M | 4 | 5, 6 |
| 8 | **The tutorial refresh** (§5.1, §5.3): the three new steps, the renumbering, `teaches`, the optional hover event, the copy. | M | 3, 7 | — |
| 9 | **The nightly walks** (§8, E2E): `tutorial-walk.spec.ts` and `hints.spec.ts`. ADR 0076 sub-PR 6. | M | 5, 6, 8 | — |
| 10 | **The exit check** below, with its evidence on #2313, and the amendment notes in ADR 0076 and ADR 0121. | S | 9 | — |

PRs 2 and 3 start at once. PR 4 needs PR 2's `Anchor` type. PRs 5, 6 and 7 run in parallel once PR 4 has merged. They touch some of the same files (`Lobby.svelte`, `Settings.svelte`, `labels.ts`), and whichever merges second rebases; the conflicts are mechanical. PR 8 needs PR 3's roll and PR 7's hint ids. PR 2 is the widest change but the most mechanical: it may be split by area (board, dock, site pages) if review prefers, with `labels.ts` and the guard in the first.

### Exit criteria

1. This ADR is accepted, and every PR in the Delivery table has merged into `develop`.
2. The nightly E2E on `develop` is green, including `tutorial-walk.spec.ts` and `hints.spec.ts`, with no `has no anchor` line in either.
3. A scratch branch that deletes `L.yourHand` from `Hand.svelte` fails the `client` job on its PR. The run is linked on #2313 and the branch deleted.
4. On cmd-dev, a signed-in person with an empty `help` group sees each site page's hint once, one at a time; reloads and sees none; signs in on a second browser and sees none there either. Help → Tips for this page shows the page's hints again.
5. On cmd-dev, Help → Practice game rolls for the first turn, walks all fourteen steps and finishes. At a real table afterwards, the hints the tutorial taught do not appear, the others appear in quiet moments, and none appears while a dock request is open or an opponent's item waits on the viewer.
6. The evidence is posted on #2313, and the work reaches `main` with the next promotion.

---

## Consequences

- **A renamed or removed label fails a PR, not a first session.** The guard runs in seconds on every PR that touches the client, and its message names what anchored to the label.
- **Some thirty files render labels through `L`.** That is a new import in each, and a new habit for authors. Rule 2 of the guard enforces the habit, so it cannot drift.
- **Every existing account sees each hint once after this ships,** including people who have played for weeks. "Hide tips" on any hint ends that in one click, and the tutorial offer still only goes to people with no finished game.
- **The practice table takes a few seconds longer to start:** a roll, and sometimes a choice. The practice bot never takes the first turn for itself.
- **Hints can be held back for a long time at a busy table.** A quiet moment may not come in a fast game. That is the intended trade: a missed tip costs nothing, and an obscured decision costs a game.
- **`SETTINGS_VERSION` 21,** with a migration; a v20 tab meeting a v21 account copy stops writing until it reloads, as ADR 0110 §4 already arranges.
- **The settings body grows by under 1 KiB.**
- **The labels list in AGENTS.md becomes a generated file.** Its prose context moves into each entry's `doc`.

## Out of scope

- **Teaching the rules,** in hints or in the tutorial (ADR 0076 §1).
- **A tutorial for anything but the table.** Site pages get hints only.
- **A per-hint analytics trail** (who saw what, when). Seen state is a map of versions and nothing more, and the tutorial bus stays at three events.
- **Localisation of hint copy.**
- **Hints for other people's actions** (a vote someone called, a spawn the host made). Those are announced where they happen.
- **A mobile-specific tutorial.** The coach's phone layout stays as ADR 0076 §2.3 has it.

## Alternatives considered

- **A longer tour that covers the site as well.** Rejected by owner decision 1: one long script is what goes stale.
- **A single registry file of all hints.** Easier to read in one place, but every hint PR would conflict there, and it puts each hint away from the feature that changes it. A generated list, like `docs/labels.md`, can be added later if one is wanted.
- **A server-side table of seen hints.** ADR 0110's synced settings already store per-account client state with a revision and a size cap, and a second store would need its own merge.
- **A tour library** (driver.js, Shepherd). The anchoring, placement and phone layout already exist for the coach. A library would bring its own anchoring, by CSS selector, outside the label contract.
- **Keeping the practice table's start without a roll** and teaching the roll only through `table.opening-roll`. Simpler, and the tutorial would never wait on a die. Rejected because the owner listed the opening roll among the tutorial's new subjects, and the practice table should start like the player's first real one.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review.

1. **One registry of contract labels, `client/src/lib/labels.ts`, that both components and anchors import**, over a static scan of `aria-label`s or a render-based check (§2.1, §2.4).
2. **The registry holds the anchored labels, the e2e walk's labels and today's AGENTS.md list**, not every label in the client (§2.1).
3. **Each entry names its owner files, and the guard checks that each owner references `L.<key>`**, plus a ban on literal copies of registered `aria` names anywhere in `client/src` (§2.3).
4. **Dynamic labels are matched by a fixed stem** with CSS `^=`, `$=` or `*=`, through the same anchor resolver (§2.2).
5. **`docs/labels.md` is generated from the registry, and AGENTS.md §5 points at it** instead of listing names (§2.5).
6. **Existing e2e specs keep their literals**, as a tripwire for renames; new specs import the registry (§2.5).
7. **One file per hint, `<Feature>.hint.ts` beside its component, collected by `import.meta.glob`**, with a list of retired ids that may never be reused (§3.2).
8. **A hint's `version` re-offers it, bumped only for a material change** as §3.3 defines it: a moved anchor, a changed gesture, or something new a returning player would miss.
9. **Copy limits of 32 characters for a title and 140 for a body, and a banned-word list**, enforced by a test (§3.4).
10. **One hint on screen at a time; one per site-page visit; at the table only in a quiet moment**, which excludes any open dock request, any gesture in progress, and another player's item on top of the stack while the viewer holds priority (§3.5).
11. **A hint that loses its quiet moment steps aside and stays unseen**, rather than counting as seen or staying over the decision (§3.5).
12. **Table hints are off while the tutorial coach is visible**, and a completed step marks the hints it teaches as seen (§3.5, §5.3).
13. **No scrim and no focus theft. A new `i` shortcut, Go to the tip, reaches the card, and it is `complementary "tip"` with a polite announcement** (§3.6).
14. **The initial set is 18 hints, nine on site pages and nine at the table**, with the omissions in §3.8.
15. **Seen state is a `help` group in the existing settings object (`seen`, `tipsOff`), synced for signed-in people, browser-only otherwise, at `SETTINGS_VERSION` 21** (§4).
16. **`help.seen` merges by union at sign-in and on a 412, and never raises the "Keep this browser's" toast**, an exception to ADR 0110 §4 for this one field (§4).
17. **This group discharges ADR 0076 §2.6's obligation to move the tutorial field to the account**, instead of a new `users` column (§4).
18. **The tutorial gains three steps (the roll, the stack, the commander) and becomes fourteen**, at the positions in §5.1, with no new bus event.
19. **The practice table opens the opening roll, and a practice bot that wins it hands the first turn to the player** through `SeatSpec.FirstTurnTo`, amending ADR 0121 (§5.2).
20. **The Help menu lives in the site header and in the ⋯ menu at the table**, with Practice game left out at a real table (§6).
21. **The tutorial offer is the `lobby.practice` hint**, for a guest or a signed-in person with no ended game, never the admin token (§3.7, §6).
22. **Settings → Advanced gets Show tips, Show all tips again and Replay the tutorial**, and Home's Help group gets a Practice game tile (§6).
23. **`tutorial-walk.spec.ts` fails on any `has no anchor` console line and tolerates `cannot happen` lines** (§8).
