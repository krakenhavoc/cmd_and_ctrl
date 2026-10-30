# ADR 0212 — A new table's lobby card as a character select

**Status:** Accepted · 2026-10-30 · S78 — Launch week. The owner approved the direction, then accepted every recommendation under [Owner decisions](#owner-decisions) the same day.
**Issue:** relates to [#4648](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4648) (the S78 tracker). [#4655](https://github.com/krakenhavoc/cmd_and_ctrl/pull/4655) fixes the picker's overflowing button, its lowercase labels and the "as printed" noise separately; nothing here repeats it.
**Numbering:** every remote head's `docs/decisions/` was listed on 2026-10-30. The highest number anywhere was **0211**, so this is **0212**.
**Builds on:** [ADR 0158](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) (Tables, Now and History, the deck grid and its commander art), [ADR 0075](0075-table-settings-and-host-controls.md) (table settings and the host), [ADR 0161](0161-an-lfg-board-in-discord.md) (#looking-for-game), [ADR 0182](0182-your-own-model-key-for-ai-seats.md) (a bot on your own model key), [ADR 0191](0191-a-sideboard-is-the-outside-the-game-pool.md) (the sideboard), [ADR 0105](0105-legal-action-highlights.md) (what `--ready` means), [ADR 0125](0125-a-walkthrough-that-keeps-up.md) (labels and hints), DESIGN.md.
**Amends:** [ADR 0158](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) §1 item 2 for a table **in the lobby** (its seats line, and "An empty lobby seat shows Add bot inline"), and that ADR's amendment of 2026-10-14 (#3499) for a table in the lobby: the toolbar of split buttons, the rules line and the "Waiting on …" note are replaced by §2–§5 below. A card **in play**, History, the page head and the right-click menu are unchanged.
**Mockup:** [mockups/0212-lobby-character-select.html](mockups/0212-lobby-character-select.html), one static page with three states. Open it from the repo so it finds the app's fonts; look at it at 820 px and 1280 px.

---

## Context

The owner's screenshot of a new table on the Tables page (`routes/Lobby.svelte`, 2026-10-30) shows a card that is mostly deck picker:

- **The seats are a thin strip.** The four seats are one row of small avatars. Under them, about 80 % of the card is the deck panel: `DeckUploadForm.svelte` mounts the saved decks (`YourDecksPicker.svelte` with `DeckRadioList.svelte`), then "or play a pre-built deck (8)" (`PrebuiltDeckPicker.svelte`), then an always-open paste textarea, a rules note and a Sideboard button, all at once.
- **Rows repeat themselves.** A saved deck's row shows its name and then its commander, which is often the same words.
- **Readiness is buried.** A seat without a deck says "deck pending" in small text, and the card says "Waiting on your deck" in grey at the far right. Nothing says how many of the table are ready.
- **Every action has the same weight.** Copy invite ↻, Copy spectator link ↻, Choose your deck ▴ and Invite by DM ▾ sit in one row, and the ↻ that makes a new link reads as "reload". The primary button in the head ("Open", or "Start") is an accent button that competes with them. The settings are a "Standard rules ▾" disclosure.

The approved direction is a **character select**: the seats are the hero, picking a deck fills your seat with your commander's art, one primary button follows the table's state, the picker is a grid of commander art behind tabs, the links live in one Invite menu, and the settings are a one-line summary.

### What the code does today (checked 2026-10-30, `origin/develop` at `2902916ed`)

**Who lands on this card.** Accepting a player invite, as a guest with a typed name or signed in, navigates to `#/lobby`, not to the table (`Join.svelte`, "lobby first so they can import a deck and see other seats' status"). A spectator invite goes straight to `#/games/{id}`. So the people who see this card are the creator, every seated player, guests included, and the admin. The table route keeps its own fallback: a seated player who reaches `#/games/{id}` in the lobby gets a "Choose your deck" dialog that mounts the same `DeckUploadForm` (`Game.svelte`, `viewerNeedsDeck`).

**Card art and guests.** ADR 0092 decision 1 keeps art and oracle text out of the public roadmap, and DESIGN.md §9 keeps card art off the front-door pages. "Signed out" there means **no session at all**. A guest seated by invite holds a session: `POST /games/{id}/join` sets the session cookie for a `player`-role principal (`server/internal/lobby/http.go`, `joinGame`). `GET /cards/{id}/image`, the route `CommanderArt.svelte` loads through `scryfallImageURL`, sits behind `auth.Middleware` with no role list (`server/cmd/server/main.go`, `newPublicHandler`), so any valid credential reads it: player, spectator, identified or admin. The same guest already sees card art on the board. **Commander art on this card is therefore allowed for a seated guest.** It is not allowed on the invite preview (`#/games/{id}/join`), which a visitor sees before they have a session; the preview stays text-only (§7).

**What each seat already carries.** `GET /games` returns every table to any authenticated caller; `redactMetaFor` strips only the invite tokens. Each `SeatInfo` (`server/internal/lobby/lobby.go`, mirrored in `client/src/lib/api.ts`) has:

| Field | Have it | Notes |
|---|---|---|
| seat number, name, Discord avatar, display name | yes | `seatColor(seat)` gives the colour |
| `is_host`, the viewer's own seat | yes | `mySeat(g)` in Lobby.svelte |
| `deck_uploaded`, `deck_name`, `deck_power` | yes | the deck's name is whatever the list or the library called it |
| `is_bot`, `bot_tier`, `bot_deck`, `bot_own_key` | yes | `bot_deck` is a pre-built deck id |
| `is_agent`, `agent_client` | yes | drawn by `AgentChip.svelte` |
| **the commander(s), a printing to draw, the colour identity** | **no** | needs §8 phase B |

The decks the picker shows are better off:

| Source | Commanders | Printing for art | Colour identity | Coverage |
|---|---|---|---|---|
| `GET /me/decks` (saved decks) | `commanders[]` | `commander_cards[].scryfall_id` (ADR 0158 §2) | `color_identity`, per commander and unioned | `coverage.as_printed_copies` / `deck_size` |
| `GET /decks` (pre-built) | `commander` (one name) | **no** | `colors` | `coverage` (full / caveats / unreviewed / unregistered) |
| `GET /bot/options` (bot decks, the same pre-built ids) | `commander` | **no** | `colors` | — |

**Readiness and Start.** There is no ready flag. A seat is ready when `deck_uploaded` is true, and a bot arrives with its deck. `Lobby.Start` refuses until every seat has a deck (`ErrDeckNotUploaded`), and the game refuses fewer than `game.MinPlayers` (2); the most is `game.MaxPlayers` (4). There is no table-size setting: a 2- or 3-player game is a table started with open seats. The client offers Start to a seated host or creator, or the admin, once two or more are seated (`mayStart`), and enables it when every seat has a deck (`canStart`). The page has no socket: it polls `GET /games` every 5 s and on focus, so another player's deck appears within about 5 s, and your own at once (the picker's `onSuccess` refreshes).

**Add bot.** Any seated player or the admin may add a bot to a lobby table (`canManageBots`). The picker needs `GET /bot/options`: the tiers (each with `available` and a reason when not), the bot decks, and `own_key` when the person may run a model bot on their own key (ADR 0182), with a model field. It posts `POST /games/{id}/seats/bot` with a tier and a deck id. Bot decks are the pre-built decks, so the Pre-built grid can serve both pickers.

**Links, DM, #lfg, settings.** Copy invite and Copy spectator link are offered to the creator, the host and the admin; ↻ (make a new link, after a confirm) to the creator and the admin (`canRotateInvites`). Invite by DM (`TablematePicker.svelte`) needs a signed-in person. Post to #looking-for-game needs `$community.lfg` and the host or admin. The settings line reads `GET /games/{id}/settings` (seats, spectators, the creator and the admin) and `settingsSummary` says only what differs from the defaults, so a default table reads "Standard rules".

**Screen size.** DESIGN.md §5: tablet (820 px) is the smallest screen; app screens get no phone layout. The Lobby has the site's one 819 px rule (for the empty-state cards) and a page-specific 1023 px rule that drops the seats to two by two. It has no phone layout.

## Decision

### 1. The seats are the hero

The four seats are four tiles in one row, at every supported width (820 px and up). A tile is about 170 px wide at 820 px and 280 px at 1280 px; the 1023 px two-by-two rule goes.

- **Every tile** is drawn in its seat's colour, `seatColor(seat)`, as the board and the front page's hero draw seats: a 3 px edge along the top, a faint wash of the colour, and a "Seat N" badge. The seat number is absolute, the same for every viewer.
- **A taken seat** shows, top to bottom:
  - **The art:** the commander's art crop, through `CommanderArt.svelte`, which credits the artist on the art (#2982). A partner pair or a commander with a Background shows two crops side by side, as the Decks grid does. A commander whose printing cannot be found shows the seat-colour wash, never a broken image.
  - **Who:** the avatar (or initial), the name, and `host` and `you` tags. An agent seat keeps `AgentChip`.
  - **What:** the commander's name, then the deck's name on a second, quieter line **only when it is not the commander's name**.
  - **A status row:** the colour-identity pips and the `DeckPowerBadge` on the left; on the right, **Ready** (a check on a mint disc, and the word) or "Choosing a deck…".
- **Your seat without a deck** is dashed in your colour, with "No deck yet". The whole tile is a button that opens the picker (§3). Once you have a deck, your tile is still a button: it reopens the picker to change your deck while the table is in the lobby.
- **While you browse the picker,** your tile previews the selected deck's art, commander and pips, marked "Preview" and "Not seated yet". Nothing is sent until you sit down.
- **An open seat** is dashed, with "Open seat" and two small buttons: **Invite**, which opens the Invite menu (§4), and **Add bot**, where you may add one.
- **A bot seat** shows the bot chip (dashed, seat-tinted, robot glyph; DESIGN.md §6) with its tier, "(own key)" when it runs on a player's key, its pre-built deck's commander art and name, and "Pre-built · <deck>" on the second line. A bot is always ready.
- **Removing a seat:** the × that removes a person or a bot, or leaves your own seat, stays on the tile with today's confirm (#2785, #3494). The admin's "seat link" moves off the tile into the card's right-click menu, one item per seat.
- **"Ready" is not `--ready`.** `--ready` means "this has a legal action now" on the board (ADR 0105). The check uses `--mint` (connected, success) and always carries the word, because seat 4's green is the same hue as mint.

### 2. One primary, driven by the table's state

The card has one primary button, in the footer's right corner. Its label says what happens next. The rules, in order (the first that applies wins):

| State | Primary |
|---|---|
| your seat is held through another session | **Open table** (reclaims it, as today's Open) |
| you created the table and are not seated | **Take a seat** |
| you are seated without a deck, picker closed | **Choose your deck** (opens the picker) |
| the picker is open with a deck selected | **Sit down with this deck** (#4655's wording); on Paste a list it is **Upload deck** or **Import deck** |
| seated with a deck, a seated player still choosing | **Waiting for N players**, disabled (N is the seated players without a deck) |
| everyone seated is ready but fewer than two are seated | **Waiting for players**, disabled, with "Invite someone or add a bot" in the footer |
| everyone seated is ready, two or more, and you may start | **Start game**, or **Start game · 3 players** while a seat is open |
| everyone is ready and you may not start | **Waiting for <host> to start**, disabled |
| the admin, not seated, and the table cannot start | **Open table** |

A disabled primary stays visible (DESIGN.md §6). Its tooltip says why. A table in play keeps **Resume** / **Open** as ADR 0158 says.

**Open table** is no longer an accent button. When it is not the primary, it is a ghost button in the footer, and it stays in the right-click menu.

**The readiness line** sits at the footer's left: "READY 2/3", four small segments in the seat colours (filled when ready, outlined while choosing, dashed when open), and a short sentence ("1 seat open · waiting on your deck", "Every deck is in"). It replaces the "Waiting on …" note. It counts seated players, not four, because open seats do not block Start.

**The ready-check moment.** When the last seated player's deck lands and two or more are seated (seen by the page as the count of seats without a deck falling to zero between two reads), each tile's check lands in seat order, the counter fills, and the primary turns to Start game with one soft ring that fades. The polite live region says once, "Every deck is in. You can start the game." (or "… Waiting for <host> to start."). With any of the three reduced-motion signals (DESIGN.md §4), nothing moves: the checks, the counter and the button simply change, and the announcement is the same. A remote deck is seen within one poll (5 s); this adds no socket.

### 3. The deck picker: a grid of commander art behind tabs

The picker opens **inline, under the seats**, with a notch pointing at your seat (open question 1 asks whether a sheet is better). One panel replaces the three stacked pickers.

- **Tabs:** **My decks** (with the count), **Pre-built** (with the count), **Paste a list**. My decks is shown to a signed-in person with saved decks. A guest, or a signed-in person with an empty library, gets Pre-built and Paste a list only, and opens on Pre-built.
- **Which tab opens first:** My decks, unless the last deck you seated was a pre-built one, or you have no library. This keeps today's rule (`DeckUploadForm`'s `prebuiltCollapsed`) and its preselect of the last deck (`lastDeck.ts`): the account's for a signed-in person, the browser's for a guest.
- **The grid:** tiles about 150 px wide, `repeat(auto-fill, minmax(150px, 1fr))`, so four columns at tablet width and seven at 1280 px. The grid scrolls inside a capped height. A tile shows the commander's art crop with the artist credit, the deck's name, the commander (only when it differs from the name; otherwise the card count or archetype), and the identity pips.
- **Coverage, flagged only when there is a gap:** a tile carries a small flag ("2 not as printed", a warning glyph and the words) only when the deck has cards that do not play as printed. A deck that plays as printed says nothing. The flag is neutral chrome, not a signal colour: `--priority` and `--danger` mean other things.
- **Choosing:** the grid is a radio group, as `DeckRadioList` is today (arrow keys move, Space selects, the filter box from `FILTER_FROM` = 8 decks, and a Recent · Name sort on My decks). Selecting a tile previews it in your seat. The footer's primary, **Sit down with this deck**, seats it.
- **The detail line under the grid** carries, for the selected deck, what a row cannot: the coverage sentence (`coverageLine`, `coverageDetail`, and the pre-built summariser in `prebuiltDecks.ts`), the source ("from archidekt.com"), a pre-built deck's summary, "Play this deck in paper" (ADR 0179 §7), and **Sideboard · N** for a saved deck (ADR 0191), which opens `SideboardList` as today.
- **Paste a list:** today's textarea and URL handling, the Moxfield hint, and the note "Every deck is validated against Commander rules …", which moves here from the top of the panel. Violations, warnings, the unimplemented list and the 200-deck library note render here, unchanged.
- **The table route's deck dialog** (`Game.svelte`) mounts the same picker, so both places keep one implementation, as `DeckUploadForm` does now.

### 4. One Invite menu

One secondary **Invite ▾** button in the footer, and the open seats' **Invite** buttons, open the same menu (`ContextMenu.svelte`'s keyboard rules: arrow keys, Escape gives focus back):

1. **Copy invite**: the player link, "whoever opens it takes a seat". It makes a link when none is held, as today, and says so.
2. **Copy spectator link**: "watch only; no seat".
3. **Invite by DM…**: opens `TablematePicker` under the seats, in the picker's place. Signed-in people only, as today.
4. **Post to #looking-for-game**, or its state ("In #looking-for-game", "Withdraw from #looking-for-game"), when the server offers it and you host (ADR 0161).
5. Under a "Make new links" heading: **New invite link…** and **New spectator link…**, each saying that the current link stops working. They keep today's confirm and then copy the new link. Creator and admin only.

When the browser refuses the clipboard, the card shows the link to copy by hand, as today. Who sees which item does not change; the menu shows only the items you may use, and is hidden for someone who may use none. Each item is still in the card's right-click menu, which keeps the rule that it is never the only way to an action.

### 5. Table settings as a one-line summary

The head's right side holds one quiet pill: a gear and **"40 life · 1 undo · standard rules"**. It always names starting life and undos, then what else differs from the defaults (commander damage, free mulligans, pace, spawning), or "standard rules" when nothing else does. It opens the existing `TableSettingsPanel` under the seats, editable by the host and the admin and read-only for everyone else, with the #lfg and results toggles as today. This needs a sibling of `settingsSummary` (`lib/tableSettings.ts`) that always prints the first two parts; `tableSettingsLine`'s accessible name keeps its shape.

### 6. The card at tablet width and up

- The seats stay four across at 820 px; the picker grid reflows by itself. The page keeps its one 819 px rule; the 1023 px seats rule is removed rather than replaced, and no phone layout is added (DESIGN.md §5).
- The card's head, seats, picker and footer are one column at every width; nothing moves into a side panel.
- Text you read to act is `--fg` or `--fg-muted`. Mono is kept for counts and badges (the seat badge, READY 2/3, tags).

### 7. What does not change

- A table **in play**: Resume, the life totals, the turn chip (ADR 0158 §1 item 4).
- **History**, the page head (join box, + New table), and the right-click menu, which gains one item per seat for the admin's seat link.
- **The invite preview** (`#/games/{id}/join`): it is seen before a session exists and stays text-only, with no art.
- The server's rules: who may start, who may add a bot, who may see a link, and the deck validation.

## Owner decisions

Asked as open questions, each with a recommendation (the more rules-faithful or complete option where one exists). On 2026-10-30 the owner accepted every recommendation, so each *Recommend* below is now part of the decision; the options not taken stay for the record.

1. **Inline picker, or a sheet over your seat?** Inline keeps the seats on screen, so the preview in your tile reads, and keeps one surface; a sheet (`.prompt-modal`) would give the grid more room at tablet width. *Recommend inline.*
2. **Does the picker open by itself?** Today the deck panel opens on its own when you are seated without a deck, except on a table you have just created, where Invite by DM opens first. *Recommend:* open the picker on its own whenever you are seated without a deck, the just-created table included, since Invite is now one click in the footer.
3. **Who sees a seat's commander before Start, and when?** `deck_name` and `deck_power` already reach every signed-in session through `GET /games`, and a deck name often is the commander. The Comprehensive Rules say nothing about a lobby: the command zone is public once the game begins. Options:
   - (a) like `deck_name`, everyone who can list the table;
   - (b) only the table's own people: its seats, its spectators, the creator and the admin;
   - (c) hidden from the other seats until Start, as a table setting ("reveal commanders at the start").

   *Recommend (b)* by default, and (c) later as a setting if the owner wants the paper habit of a simultaneous reveal.
4. **Is "ready" having a deck, or an explicit Ready toggle?** A toggle lets a player sit with a deck and keep browsing, but needs a new seat field and a new route. *Recommend deck in = ready*, which is Start's gate today and needs no server change.
5. **2- and 3-player tables.** Start works from two seated players. Should Start with open seats ask "Start with 3? The open seat closes", or should the host set the table's size so the counter reads "3 of 3"? *Recommend* the label "Start game · 3 players" and no confirm, and no size setting until someone asks for one.
6. **What does a spectator of this table see on Tables?** They normally land on the table, not here. *Recommend* the same tiles, read only, the Invite menu with only Copy spectator link, and a primary **Watch**.
7. **A guest without saved decks.** *Recommend* they open on Pre-built with the browser's last pre-built pick preselected, no My decks tab, and one line on Paste a list: "Sign in with Discord to keep the decks you paste." Signing in from the seat is not offered, because it would replace the guest's seat session.
8. **The bot picker.** *Recommend* that **Add bot** opens the picker region under the seats, titled "Add a bot to seat 3": a tier segmented control (unavailable tiers disabled with their reason), the own-key toggle and model field when offered, and the Pre-built grid. Today's two `<select>`s go.
9. **Commit the mockup?** This is the first HTML mockup kept in the repo; ADR 0156 and 0158's mockups were reviewed outside it. *Recommend* keeping it under `docs/decisions/mockups/` with its ADR, so a later change can be compared with what was agreed.

## Consequences

- The card reads as the table: who is sitting where, with what, and who is ready, in the seat colours a player meets again on the board.
- One primary per card again (DESIGN.md §11). Open, the links and the deck panel stop competing with it.
- More images: a lobby card loads up to four seat crops (once phase B lands) and the picker loads one crop per visible tile, lazily, through the existing image cache. A library of 200 decks loads only what scrolls into view.
- Every crop credits its artist; the pinning test (`artistCredit.render.test.ts`) gains the seat tile and the picker tile.
- The picker becomes one component shared by the Tables card and the table route's deck dialog, replacing the stacked `YourDecksPicker` / `PrebuiltDeckPicker` / textarea in `DeckUploadForm`.
- Phase B adds fields to `GET /games`, `GET /games/{id}`, `GET /decks` and `GET /bot/options`. They are additive and old clients ignore them. Nothing changes on the WebSocket.

## Phases

Each client slice attaches before and after screenshots at 820 and 1440 px and runs the E2E on its branch, since these are client-interaction changes.

### Phase A: the card, with no server change

| # | Slice | What | Model |
|---|---|---|---|
| A1 | **Seats, primary, Invite, settings** | §1 tiles without art (the art slot draws the seat wash; bot seats get pips from `GET /bot/options`' `colors`), §2 primary and readiness line, §4 Invite menu, §5 summary, the seat-link menu items, the 1023 px rule removed | Opus |
| A2 | **The picker grid** | §3 tabs and grid; My decks tiles with art (from `GET /me/decks`), Pre-built tiles with an identity wash and pips until B1; the preview in your tile; the detail line, Sideboard and Paste a list; mounted in `Game.svelte`'s dialog too; the bot picker (question 8) | Opus |
| A3 | **The ready-check moment** | §2's moment, the live-region line, and the reduced-motion path | Sonnet |

### Phase B: additive API fields, then the art

| # | Slice | What | Model |
|---|---|---|---|
| B1 | **Pre-built and bot decks carry their commanders** | `GET /decks` and `GET /bot/options` decks gain `commander_cards[]` (`name`, `scryfall_id`, `color_identity`), resolved by the same code as `GET /me/decks`. Pre-built tiles and bot seats get art. docs/lobby.md and the route tests | Sonnet |
| B2 | **Seats carry their commanders** | `SeatInfo` gains `commander_cards[]` and `color_identity`, set when a deck is seated (paste, library or pre-built) or a bot is added, and kept across a restart: a `seats` column with a migration, or derived from the restored game's command zone (the slice picks one and says why). Visibility as question 3 decides, in `redactMetaFor`. If it is a new column, check it against ADR 0149's inventory (it is deck data, like `deck_name`, which `seats` already holds) and bump the privacy schema version if AGENTS.md's rule applies. Human seats get art and pips | Opus |

### Phase C: not planned

The owner took the recommendations, so there is no explicit Ready toggle (question 4) and no table-size setting (question 5). A "reveal commanders at the start" setting (question 3c) stays a possible later slice, with a server change, if the owner asks for it.

### Labels, hints and tests that change

- **`labels.ts`** (the contract, ADR 0125 §2):
  - Four entries change: `copyInvite` and `copySpectatorLink` keep their names and become `menuitem`s; `newInviteLink` and `newSpectatorLink` become visible menu items ("New invite link…", "New spectator link…") instead of the ↻ buttons' accessible names.
  - About twelve are new: the Invite button and its menu ("invite to <table>"), Choose your deck, Sit down with this deck (if #4655 does not register it first), Start game, Waiting for …, Open table, the three tabs, the two grids ("your decks", "pre-built decks"), the seat tile ("seat N: <name>") and the readiness line.
  - `tableSettingsLine`'s example changes to the new summary. `docs/labels.md` is regenerated.
- **Hints:** `Lobby.hint.ts` names "Copy invite" and "your deck underneath", so its body is rewritten and its version goes to 3.
- **E2E:** one spec changes. `tests-e2e/tests/lobby.spec.ts` has about nine assertions across four of its eight tests that read the old toolbar (Copy invite, Copy spectator link and the two ↻ buttons as buttons), `Start` and `Open` by exact name, and "deck pending". `entry.spec.ts` and `join.spec.ts` read the invite preview and the `seats` list name, which stay. `full-game.spec.ts` seats decks through the API, and `hints.spec.ts` anchors on `L.newTable` and `L.lobbyTitle`, which stay. A new spec covers the picker as a guest (Pre-built tab, sit down, the counter moves) and the host's Start. Expect one changed spec, one new spec of three or four tests, and one branch E2E run per slice.
- **Client render tests:** `tableCard.render.test.ts` holds most of the churn (about 17 text queries on the toolbar, seats and primary); `playerTables`, `tablesPage`, `signedInHome` and `decksPage` render tests have one or two each; `deckPicker.test.ts` and `tableSettings.test.ts` gain cases for the tile subtitle rule, the gap flag and the always-two-part summary.
