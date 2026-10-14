# ADR 0179 — Shop links for the real cards

**Status:** Accepted · 2026-10-14 · Backlog
**Issue:** [#3444](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3444) (this ADR), [#3445](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3445) (the follow-up placements, which reuse everything here)
**Owner decisions:** the rules in #3444's body, quoted under [Owner rules](#owner-rules). They are binding. The calls the issue left open and this ADR made are under [Calls made here](#calls-made-here).
**Numbering:** every ADR file added on any remote head was listed (`git log --remotes --diff-filter=A -- docs/decisions`, fourteen remote branches); the highest was **0178** (`fix/2950-darkness-replacement-continuation`), so this is **0179**. No migration: the hide setting rides in the synced settings body (ADR 0110 §4).
**Builds on:** [ADR 0110](0110-remember-me.md) §4 (synced settings), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the privacy notice that cannot drift), [ADR 0095](0095-deck-coverage-and-deck-requests.md) (the deck check), [ADR 0042](0042-card-catalog-page.md) (the Catalog).

---

## Context

A player who likes a deck at the table often wants the real cards. Pointing them at a card shop is useful to them, and a shop's referral programme can pay the site a commission on what they buy. Both only hold if the links are honest: plainly disclosed, never in the way of a game, easy to turn off, and with no tracking added by us.

## Owner rules

From #3444, verbatim in substance:

1. Shop links only appear when the operator has configured a shop: a server environment variable per retailer, a URL template holding `{url}`, so a retailer's referral wrapper or parameter can be set without a code change. With none configured, no link renders.
2. Links are built from the card name (and set where useful), never Scryfall's `purchase_uris`. No prices are shown.
3. Never on the game table while a game is running.
4. A clear note next to the links: "Shop links may earn CMD & CTRL a commission." The same in the Terms and the Privacy notice: clicking takes you to the retailer, whose own privacy terms and cookies apply, and CMD & CTRL does not track the click.
5. A per-account setting to hide shop links (default shown), and a local one for guests.

## Decision

### 1. Configuration: one template per retailer

| Variable | Retailer |
|---|---|
| `CMDCTRL_SHOP_TCGPLAYER` | TCGplayer |
| `CMDCTRL_SHOP_CARDKINGDOM` | Card Kingdom |

Each holds a template in one of two shapes (`server/internal/shoplinks`):

- **A wrapper:** an absolute `https` URL with `{url}` exactly once, after the host, e.g. `https://partner.example/c/123?u={url}`. The client puts the destination there percent-encoded.
- **A direct link:** `{url}` first, optionally followed by `&name=value` pairs, e.g. `{url}&partner=cmdctrl`. The client uses the destination as it is (every destination has a query string) and appends the pairs.

The template is plain printable ASCII with no spaces, quotes, backslashes, braces (other than the placeholder) or angle brackets, and at most 1 KiB. `{url}` after the host means a destination can never choose where a link goes.

**Unset is the default** and means no shop links anywhere. A malformed value **fails the boot**, naming the variable, like `CMDCTRL_DISCORD_INVITE_URL`: a typo is found at deploy, not as a link that goes somewhere unexpected. CD's "Sync server env (shop links)" step checks the same rules before writing, so a bad value fails the deploy and leaves the running server as it was. Both variables are written on every deploy, empty included, from the Actions variables of the same names (a GitHub environment's value overrides the repository's), so clearing one turns that shop off on the next deploy.

The valid set is served on the existing public `GET /config` as `"shops": [{"id", "name", "template"}]`, `[]` when none, on every deployment (it is not a dev surface). The client re-applies the same rules (`validTemplate`) and drops a template that fails, so a server and client that ever disagree fail closed.

### 2. Destinations

Built in the browser (`client/src/lib/shopLinks.ts`) from the card's **front-face name** (the part before ` // `, which both retailers find). Never from Scryfall's `purchase_uris`; no price is fetched or shown. Set is not used yet: the deck check and the Catalog name a card, not a printing.

| What | URL |
|---|---|
| TCGplayer card search | `https://www.tcgplayer.com/search/magic/product?productLineName=magic&q=<name>&view=grid` |
| TCGplayer list cart (Mass Entry) | `https://www.tcgplayer.com/massentry?productline=Magic&c=<line>\|\|<line>…`, each line `"<count> <name>"` percent-encoded |
| Card Kingdom card search | `https://www.cardkingdom.com/catalog/search?search=header&filter%5Bname%5D=<name>` |
| Card Kingdom list cart (Deck Builder) | `https://www.cardkingdom.com/builder?c=<lines joined by newlines, percent-encoded>` |

Checked on 2026-10-14: Card Kingdom's search returned Sol Ring results, and its builder page server-renders the `c` lines into its list box. TCGplayer's pages are a client-side app; its Mass Entry bundle reads `query.c` and `query.productline`, and its own "share" link builds exactly this `c` format (encoded lines joined by `||`); its search route builds `/search/<line>/product?q=…&productLineName=…`. A list is merged by name (copies summed) before it is built. The formats are pinned by `shopLinks.test.ts`.

### 3. Placements (#3444)

| Where | What |
|---|---|
| Decks: the deck check report | "Buy this deck" (a list cart per shop, every card in the report including the commanders), and a small per-shop Buy link on each card row |
| Decks: a saved deck's report | "Buy this deck" |
| Catalog | a small per-shop Buy link on each tile |

**Never on the game table while a game is running.** None of these routes is the table, and the in-game card zoom (`HoverZoomOverlay`, the only card zoom; there is none outside a game) has no shop links. A future placement on a table (#3445: an ended game) must check the game has ended first. `ShopLinks.svelte` carries the rule in its header.

"Buy missing cards" is not built: the site does not know which cards a person owns, so the deck check cannot know what is missing. "Request missing cards" is about cards the engine does not automate, which is a different thing. The placement waits for an owned-cards list, if one ever exists.

### 4. The disclosure

`SHOP_DISCLOSURE` = "Shop links may earn CMD & CTRL a commission." is shown with every group of links, with a "Hide shop links" button beside it (`ShopNote.svelte`). Where a surface has many one-card groups (deck rows, Catalog tiles), one note covers them, placed above them: the "Buy this deck" panel above the deck's rows, and a line above the Catalog grid.

The Terms (a "Shop links" section; `TERMS_UPDATED` 2026-10-14) and the Privacy notice (a "Card shops, if you follow a shop link" entry under who it is shared with) say the same, and add: a link opens the shop's own site, sometimes through its partner network, whose own privacy terms and cookies apply; the link holds card names and nothing about the person; it opens with `rel="noopener noreferrer sponsored"`, so the shop is not told which page it came from; and CMD & CTRL does not track the click (nothing is sent to this server, nothing is recorded).

### 5. The hide setting

`settings.display.shopLinks`, a boolean, default `true`, scope **synced** (ADR 0110 §4). For a signed-in account it is part of the synced settings body (`user_settings.body`, PUT `/me/settings`) and follows the person; for a guest it stays in this browser's `localStorage` (`cmdctrl.settings.v1`, read and written inside `try`/`catch`). No schema change, so no migration, no deletion-inventory entry and no `PRIVACY_SCHEMA_VERSION` bump: account deletion already deletes the settings row, and the privacy notice already lists synced settings. A missing or non-boolean value reads as the default, so `SETTINGS_VERSION` stands. Settings › Display shows the box only where a shop is configured; the "Hide shop links" button next to every disclosure turns it off.

### 6. For #3445

Reuse, unchanged: `shopLinks.ts` (`cardShopLinks`, `listShopLinks`, `SHOP_DISCLOSURE`), `shopLinksStore.ts` (`visibleShops`), `ShopLinks.svelte` and `ShopNote.svelte`. A new retailer is a line in `shoplinks.Retailers`, a case in each client builder, and a CD variable.

## Calls made here

- **Boot fails on a malformed template** rather than dropping it with a log line: the same posture as every other optional server URL, and CD checks first.
- **Two template shapes,** so a referral parameter needs no wrapper service.
- **Front-face names** for both search and list lines; set codes are left out until a placement knows the printing.
- **One disclosure per surface** where groups are one card each, placed above them.
- **No "Buy missing cards"** (§3).

## Consequences

- With no variable set nothing changes for anyone: `GET /config` gains `"shops": []` and every placement renders nothing.
- No request reaches this server when a link is followed; there is nothing to count clicks with, by design.
- A retailer changing its URL formats breaks links silently until `shopLinks.ts` is updated.

## Back-out

| To undo | Do |
|---|---|
| all links, at once | clear `CMDCTRL_SHOP_TCGPLAYER` and `CMDCTRL_SHOP_CARDKINGDOM` and redeploy |
| the code | revert the PR; the synced `display.shopLinks` key left in settings bodies is dropped by an older client's migrate (unknown keys are ignored) |
