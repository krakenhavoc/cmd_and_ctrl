# ADR 0158 — Signed-in pages: Tables with History, a deck grid, and a catalog with a filter sidebar

**Status:** Accepted · 2026-10-13 · S71
**Issue:** [#3177](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3177) (tracker); the slices are filed under S71.
**Numbering:** every ADR file on `origin/develop` and on every remote head was listed. The highest number anywhere was **0157**, so this is **0158**.
**Builds on:** [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) (the site shell: `Page.svelte`, the grouped header, the shared `.card` / `.btn` primitives), [ADR 0112](0112-signed-in-home-player-mode-and-one-decks-page.md) (the signed-in home is Tables; one Decks page), [ADR 0042](0042-card-catalog-page.md) (the catalog), [ADR 0149](0149-privacy-notice-and-account-deletion.md) and #2982 (artist credit on an art crop).
**Amends:** [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) §3 item 3: "My games" leaves the header, because it is folded into Tables.

---

## Context

S71 put every page in one frame, but the pages behind sign-in kept their old insides. The owner's screenshots of cmd-dev (2026-10-13) show the problems:

- **Tables.** Every table carries five equal-weight actions (new invite link, new spectator link, table settings, Replay, archive), whether it is waiting in the lobby or ended a day ago. Ended games still draw two empty "Open seat" tiles. The join and create boxes are squeezed side by side, and "show archived tables" and "refresh" float at the bottom.
- **My games.** A narrow, flat list with active, won and archived games mixed together, repeating what Tables shows for the current ones.
- **Decks.** Each saved deck is a tall stack of text over three or four full-size buttons, so five decks fill a screen. The coverage figure, the thing a player scans for, is small green text.
- **Catalog.** Filters sit above a grid of full-size card images, so the cards start a screen down and the filters scroll away.

The owner reviewed mockups (`tables`, `decks`, `catalog`, 2026-10-13) and answered three questions.

## Decision

### 1. Tables has two sections: Now and History

The owner chose "build it as shown". My games is folded in.

1. **The page head.** Its actions slot holds the join box (invite code or link, then **Join**, with today's behaviour) and a primary **+ New table** button. **+ New table** opens a one-line form under the head: table name, **Use my last setup** (with the setup summary), and **Create**. The form stays open after a create fails, so the error can be read.
2. **Now** lists the tables the person is seated at or created that are in the lobby or in play. Each table is a card:
   - **The card head:** the name; a status chip (**Lobby**, or **In play · turn N**); and one line of context ("you host · created 2h ago", or "Slysloth hosts · your seat 2").
   - **The actions:**
     - One primary button: **Open** in the lobby, **Resume** in play.
     - **Copy invite**, for the host or creator in the lobby. It copies the existing invite link and creates one when there is none. The button is gone once the game is in play.
     - A **⋯** menu with the rest: new invite link, new spectator link, table settings, archive.
   - **The seats** sit in one row of four:
     - A taken seat shows the avatar, the name, and a second line: the deck in the lobby, or the deck and life total in play.
     - A bot seat shows its tier and deck.
     - An empty lobby seat shows **Add bot** inline, where the host may add one.
   - **The card footer** keeps today's lobby-only controls: Invite a tablemate, Change deck, and the "Needs at least two seats" note.
3. **History** lists every table the person has sat at that has ended or been archived. This is today's `GET /me/games`, minus the tables shown in Now.
   - **One compact row per game:** name, result chip (**You won**, **<name> won**, **Archived**), who else was there, when, **Replay**, and a **⋯** menu (open, archive or unarchive).
   - **Filters:** a segmented control, **All · Won · With friends · Archived**. "With friends" means at least one other seat was a person, not a bot. Filtering happens in the client, over the list it already fetched.
   - **Length:** the first 6 rows, then "Show N more", which adds 20 at a time.
   - **Order:** most recent first.
4. **The Now cards need a turn and life totals**, and `GET /games` carries neither. It gains, for an active game only:
   - `turn`: the turn number;
   - `life` on each `players[]` entry.

   They are given only to a caller who already sees that table in the list. Both are public at the table: every seat and spectator sees them. A game in the lobby or ended carries neither field. This is additive to the wire, documented in docs/lobby.md, and pinned by the existing route tests.
5. **My games goes.** The header loses **My games** (ADR 0156 §3 item 3 is amended), `#/my-games` redirects to `#/lobby` scrolled to History, and `MyGames.svelte` is deleted. Seat reclaim (`POST /me/games/{id}/session`) is what a History row's **Open** and a Now card's **Resume** call, as My games did.
6. **Gone:** the bottom "show archived tables" and "refresh" links. Archived is a History filter, and the page refreshes as it does today, on focus and on the lobby's events.

### 2. Decks is a grid led by the commander's art

The owner chose "use commander art for the main show, and an aura of the commander's colour identity".

1. **The check bar** is one line: **Link** | **Paste a list** toggle, the input, then **Check this deck**. The Moxfield export hint becomes the input's placeholder, or the line under it when pasting. The report under it is unchanged.
2. **Your decks** is a grid of three columns at 1200 px and two at tablet width.
   - **Sorting:** a segmented control, **Recent · Name · Coverage**.
   - **The art is the main visual.** Each deck card is led by the commander's art crop (`/cards/{id}/image?size=art_crop`, the existing image route and cache). `GET /me/decks` carries what the page needs for it: each commander's printing id (`commander_cards[].scryfall_id`, a printing with an image, absent when none resolves) and colour identity, and the deck's identity as their union (`color_identity`). Commander names alone could not be drawn, because no route outside a game resolves a name to a printing, about 140 px tall, faded into the card at the bottom. A partner pair shows both crops side by side.
   - **The artist is credited on the art** (#2982, `artistCredit.svelte.ts`). This is required, as it is everywhere an art crop is the main view of a card.
   - **The aura is the colour identity.** The card's border and a soft outer glow take the commander's colour identity:
     - W, U, B, R and G each map to one fixed colour token;
     - two or more colours make a gradient through them;
     - colourless takes a neutral steel.

     The colours are fixed per identity and do not follow the skin, because they mean the deck's colours, not chrome.
   - **Under the art:** the deck name, the commander and card count, and a **coverage ring** with "98/100". The ring is green at ≥ 96 % and amber below. A notes line ("2 simplified · from archidekt.com") follows.
   - **The actions:** **Coverage report**, plus **Request missing** when something is manual. A **⋯** menu holds Rename and Delete; Delete keeps its confirm.
   - A deck whose commander cannot be resolved to an image shows the aura over a plain band, never a broken image.
3. **No deck yet?** The pre-built picker link stays under the grid. A signed-out visitor sees the check bar only, as today (ADR 0112 §3).

### 3. The catalog gets a filter sidebar and a denser grid of real card art

The owner chose "as shown, as long as it still shows the card art".

1. **Layout:** two columns, a sticky sidebar about 250 px wide and the grid. Below about 1000 px the sidebar collapses into a **Filters** button that opens it as a sheet over the grid (tablet).
2. **The sidebar** holds:
   - the search box;
   - **Completeness** (Complete, Has caveats, Unreviewed), each with its count;
   - **Colour identity** as the six pips;
   - **Type**, each with its count.

   The counts are computed in the client from the catalog the page already loads. The filters are the same as today's; only their place changes.
3. **The bar over the grid:** the result count, a removable chip for each active filter, and a sort control (**Name · Newest**, where Newest uses the catalog's existing added order if it has one, and is left out if it has none).
4. **The grid:** six cards per row at 1200 px, four at tablet width. **Each card is its real card image**, loaded lazily, with the name and one completeness chip under it. A card with caveats shows them on hover or focus, and on tap at tablet width, rather than always printing them under the card.
5. `#/catalog?q=` keeps working, and so do the search's contract label and the catalog hint.

### 4. Admin in the header

Owner, 2026-10-13: in admin mode, **Admin** shows in the header nav as an accent-tinted link after Roadmap, and in the account menu (#3221).

## Consequences

- One page fewer: History and Now share a single fetch of `GET /me/games` plus `GET /games`. A bookmark to `#/my-games` still lands, through the redirect.
- `GET /games` grows two fields for active games. Nothing in them is private, and old clients ignore them.
- Every deck card fetches one art crop through the existing cache, so a library of nine decks makes nine image requests on first view and none after. The artist credit rule applies, and the test that pins credits covers the new place.
- The aura's five colours are new fixed tokens. They are not skin tokens, because a skin must not recolour a deck's identity.
- Labels that the e2e suite or the tutorial reads go through `labels.ts`. Specs that select on "My games", "show archived tables" or the old Tables buttons are updated in the same PR, and the nightly is run on the branch.

## Slices

| # | Slice | Model |
|---|---|---|
| 1 | This ADR | — |
| 2 | **Tables: Now and History** (§1): the `GET /games` additions (server), the page, the My games fold and redirect, the header change, labels and e2e | Opus |
| 3 | **Decks grid** (§2): the art, credit, aura, ring, sort and menus | Opus |
| 4 | **Catalog sidebar** (§3): the sidebar, the sheet at tablet width, chips, a denser grid of real art | Sonnet |

Each client slice attaches before and after screenshots at 1440 and 820 px, and runs the nightly e2e on its branch.
