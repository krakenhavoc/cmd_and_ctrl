# ADR 0095 — Deck coverage checks and deck requests, on the site and in Discord

**Status:** Accepted · 2026-09-25 · Post-S30 — Rolling deck-driven catalog growth (outside a numbered sprint)
**Issue:** [#1631](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1631)
**Numbering:** swept with the AGENTS.md §4 check on 2026-09-25: `git fetch --all --prune`, then the
ADR file names in the history of every remote branch. The highest number present anywhere is
**0094** (`0094-card-location-index.md`); `0095` appears on no branch.

**Related:** [ADR 0004](0004-discord-identity.md) (the bot, its guild allow-list, its admin
session), [ADR 0017](0017-bug-report-button.md) (filing GitHub issues from the app, redaction),
[ADR 0042](0042-card-catalog-page.md) (`Completeness` and the catalogue),
[ADR 0037](0037-unimplemented-card-signal.md) (`game.Unimplemented`),
[ADR 0051](0051-user-database.md) (users, identities, the SQLite store),
[ADR 0092](0092-public-roadmap-and-site-portal.md) (the public roadmap page, whose rules on what a
public page may show this reuses).

---

## Context

The catalogue grows deck by deck. `docs/decklists/` holds hand-written triages: Aang, Hashaton,
the pirates and the top-100 staples. Each sorts one deck's cards into full, caveated, buildable and
blocked, and #1306 turned one of them into a per-card checklist issue. That workflow works, but a
maintainer has to start it. A player has no way to ask for their deck, and no way to see before
sitting down how much of it the engine will automate.

Most of what is needed already exists:

- **Fetching a deck.** `deck.FetchFromURL` (`server/internal/deck/fetcher.go`) fetches Moxfield and
  Archidekt links, `deck.ParseText` reads a pasted list, and `deck.Resolve` resolves entries against
  the card index.
- **Per-card status.** `game.Unimplemented` answers per card, and the deck-upload response already
  lists `Unimplemented` names. `catalog.Build` gives each catalogued card its `Completeness`: full,
  caveats or unreviewed.
- **Filing an issue.** `internal/github.Client.CreateIssue` files issues behind `POST /bugreport`,
  and the server holds `CMDCTRL_GITHUB_TOKEN`.
- **The bot.** It already calls the game server over loopback with an admin session. It registers
  guild-scoped commands (`/cc-invite`, `/cc-games`, `/cc-end`) and answers every interaction
  immediately, within Discord's 3-second deadline.

Nothing puts a whole decklist into buckets, and nothing lets anyone outside the repo ask for cards.

## Owner decisions (2026-09-25, on #1631; ADR accepted the same day, including the sign-in and links-only assumptions)

1. **Every bot command moves to the `/c2-` prefix:** `/c2-invite`, `/c2-games`, `/c2-end`, and the
   two new ones below. The old `cc-` registrations are removed.
2. **The coverage checker lives on the site and in Discord.** The site page is public, like the
   roadmap.
3. **Any member of an allowed guild may file a deck request.** Requests are rate-limited per
   person. Requesting the same deck again adds a comment to the open issue instead of filing a
   second one.
4. **The issue is a checklist of the missing cards,** with the requester's Discord display name.
5. **The site checker has a button that files the same request** as `/c2-deck-req`.

## Decision

### 1. One report: `deck.Coverage`

A new function in `server/internal/deck` (or a small `deckcoverage` package, if the imports require
one) turns a resolved decklist into a report. It is the only place that decides what a card's
bucket is. The site, both Discord commands and the filed issue all render this one report.

```go
type CoverageBucket string // "automated" | "caveats" | "unreviewed" | "manual" | "no_effect"

type CoverageCard struct {
    Name     string
    OracleID string
    Count    int            // copies in the list
    Bucket   CoverageBucket
    Caveats  []string       // bucket == caveats only: the card's player-facing Caveats
}

type CoverageReport struct {
    DeckName   string
    Source     string         // "moxfield" | "archidekt" | "text"
    SourceURL  string         // empty for pasted text
    DeckKey    string         // "moxfield:<id>" | "archidekt:<id>"; empty for text
    Commanders []string
    Counts     map[CoverageBucket]int  // by distinct card
    Cards      []CoverageCard          // sorted: bucket, then name
    Unknown    []string                // names the index could not resolve
}
```

**The buckets:**

| Bucket | Meaning | Test |
|---|---|---|
| `no_effect` | Nothing to automate: a vanilla creature, a basic land. | `!game.NeedsCatalogEffect(c)` |
| `automated` | In the catalogue and `CompletenessFull`. | catalogue entry, full |
| `caveats` | In the catalogue with declared simplifications. | `CompletenessCaveats` |
| `unreviewed` | In the catalogue but not audited. | `CompletenessUnreviewed` |
| `manual` | Not in the catalogue. It plays as a sandbox card. | `game.Unimplemented(c)` |

`manual` is `game.Unimplemented` exactly. This ADR adds no second definition of "unimplemented",
because the bot's improvisation (ADR 0033 §8), the stack overlay's `manual` chip and the upload
summary already share that bit.

**Validation.** The report runs `deck.Validate` and returns the violations beside the buckets.
They are informational only: a 60-card list or a deck with an illegal commander still gets a
report. The checker answers "how much of this is automated", not "is this legal".

### 2. Server routes

- **`POST /deck-coverage` is public and needs no session.** The body is `{url}` or `{text}`, and the
  response is the `CoverageReport` as JSON. It carries names, oracle IDs, buckets and caveat
  sentences, and no art or oracle text: the same line ADR 0092 drew for `GET /roadmap`.
  - The route fetches a third-party URL on the caller's behalf, so it is rate-limited per client
    IP. It uses the lobby's `newLimiter` at about 1 request per 10 seconds, with a burst of 3.
  - Reports are cached for 10 minutes, keyed by `DeckKey`. Two checks of the same deck fetch it
    once.
  - It reuses `FetchFromURL`'s 10-second timeout and 2 MiB cap, and its typed fetch violations
    (Cloudflare block, unknown host, private deck) as the error text.
- **`POST /deck-requests` files or joins a request.** The caller must be either:
  - an admin session: the bot, which passes the requester as
    `{url, requester: {discord_id, display_name}}`; or
  - a signed-in user whose account has a Discord identity: the site button. The server reads the
    requester from the session.

  Anonymous and guest sessions get 401 and 403. Requests must be **URL-based**: pasted text has no
  stable identity to deduplicate on, so the site shows the button only for a link check.

### 3. Filing, deduplication and limits

A new migration adds a `deck_requests` table and a `deck_request_asks` table:

```sql
CREATE TABLE deck_requests (
    deck_key     TEXT PRIMARY KEY,   -- "moxfield:<id>"
    issue_number INTEGER NOT NULL,
    issue_url    TEXT NOT NULL,
    created_at   INTEGER NOT NULL
);
CREATE TABLE deck_request_asks (
    deck_key     TEXT NOT NULL,
    requester    TEXT NOT NULL,      -- "discord:<snowflake>"
    at           INTEGER NOT NULL
);
CREATE INDEX deck_request_asks_requester_at ON deck_request_asks (requester, at);
```

**What a request does:**

1. **Rate limit.** Each requester gets **3 asks per rolling 24 hours**. The bot and the site share the
   same key, `discord:<snowflake>`, so switching surfaces does not reset the limit. The limit is
   counted from `deck_request_asks`, not from an in-memory limiter, so it survives a deploy. An ask
   over the limit is refused with a message saying when the next one is allowed.
2. **Nothing to add.** If the report has no `manual` and no `unreviewed` cards, no issue is filed.
   The requester is told the deck is already fully in the engine; the caveated cards are listed.
3. **Duplicate deck.** If `deck_key` has an issue and that issue is **open**, the request adds a
   comment instead of a new issue: "Also requested by \<display name\>", plus the current counts,
   since the deck may have changed. It adds no comment when the same requester asks twice. If the
   issue is **closed**, the request files a new issue and repoints the row.
4. **The issue itself.** It is filed through `internal/github` with the ADR 0017 redaction applied.
   - **Title:** `[deck-request] <deck name>` (with the commander's name when the deck has none).
   - **Labels:** `enhancement` and `deck-request`. On a label rejection the issue is re-filed with
     no labels, the way `fileBugIssue` handles it.
   - **Body:**
     - the deck link;
     - the requester's Discord display name, with no snowflake or avatar;
     - counts per bucket;
     - `## Cards to add`, one checkbox per `manual` card as `- [ ] Name (oracle id)`;
     - `## Cards to review`, one checkbox per `unreviewed` card;
     - `## Automated with caveats`, a plain list with each caveat;
     - a footer saying the issue was filed by `/c2-deck-req` or the site.

   This is the #1306 shape, and a catalogue batch can work straight from it. The existing
   `issue-triage.yml` adds `needs-triage`, and that is intended.

The GitHub token stays on the server. The bot never holds one. That is ADR 0004's split, the one
that already keeps the bot's Discord secrets out of the server's env file.

### 4. The bot

- **Rename.** The three command constants become `c2-invite`, `c2-games` and `c2-end`.
  `RegisterCommands` switches from one `ApplicationCommandCreate` per command to
  `ApplicationCommandBulkOverwrite` per guild. Bulk overwrite replaces the whole guild command set,
  so the old `cc-` commands disappear on the first boot of the new binary, with no cleanup step.
  AGENTS.md, `deploy/README.md` and the bot docs are updated in the same PR.
- **Deferred replies.** A deck fetch plus a GitHub call can take longer than 3 seconds. Both deck
  commands answer at once with
  `InteractionResponseDeferredChannelMessageWithSource`, ephemeral, and then edit the reply when
  the work finishes. This is the bot's first deferred reply. The per-interaction context becomes
  about 20 seconds for these two commands only.
- **`/c2-deck-check <link>`** replies ephemerally with the counts per bucket and up to about 15
  `manual` names, then "…and N more". It includes a link to the site's `#/deck-check?url=…` for
  the full list. When there is something to request, it adds a **Request these cards** button.
  The button runs the same handler as `/c2-deck-req`, keyed to the invoker who presses it.
- **`/c2-deck-req <link>`** calls `POST /deck-requests` with the invoker's ID and display name. It
  replies in the channel with the issue link when an issue is filed or joined, so the table can
  see it. It replies ephemerally when the request was refused or there was nothing to add.

### 5. The site: `#/deck-check`

The page is public: it is added to `isPublic` and linked from the `#/home` portal and the shared
header (ADR 0092).

- **Input.** A link field and a "paste a list" tab. `?url=` pre-fills the link and runs the check,
  which is how the Discord reply's link opens the full report.
- **Result.** Counts first, then the cards grouped by bucket. Caveated cards show their caveat. For
  signed-in viewers, card names link to `#/catalog?q=` as the roadmap's names do. Signed-out
  viewers see plain text.
- **Request these cards.** The button shows only for a link check that has `manual` or
  `unreviewed` cards.
  - A signed-in user with a Discord identity can press it; it calls `POST /deck-requests` and shows
    the issue link.
  - A signed-out visitor sees "Sign in with Discord to request these cards". The page is public
    but filing is not: an anonymous filer would be a spam path onto a public repo, and signing in
    gives the issue a name and the rate limit a key.

## Consequences

- **Every deck request is also a catalogue work list.** Deck triages stop being something only a
  maintainer can start.
- **The server fetches third-party URLs for anonymous callers.** The per-IP limit, the cache, the
  fetch caps and the fixed host list (Moxfield, Archidekt) bound that exposure. The route never
  fetches an arbitrary host: `FetchFromURL` refuses unknown ones (`ErrUnknownSource`).
- **The repo is public, so the requester's Discord display name is published.** Owner decision 4
  accepts that. The snowflake is stored locally for the limit and is never published.
- **Renaming the commands breaks muscle memory once.** Bulk overwrite removes the old commands
  cleanly.

## Plan

1. **Server:** `deck.Coverage`, `POST /deck-coverage`, the migration, `POST /deck-requests` with
   filing, dedup and limits, and tests using a fake GitHub reporter. Document the routes in
   `docs/lobby.md`.
2. **Bot:** the `c2-` rename with bulk overwrite, deferred replies, `/c2-deck-check` with its button,
   `/c2-deck-req`, and tests against the fake server.
3. **Client:** `#/deck-check`, its portal and header links, and vitest for the rendering helpers.
4. Create the `deck-request` label on the repo.

PR 2 and PR 3 depend on PR 1's routes and can run in parallel with each other.

## Implementation notes (PR 1, server)

The routes are documented in [docs/lobby.md](../lobby.md), with the exact JSON shapes PR 2 and PR 3
consume. Where PR 1 differs from, or fills in, the text above:

- **Package and names.** The report lives in `server/internal/deckcoverage`, not `internal/deck`:
  the effects package's own tests import `deck`, so `deck` cannot import `catalog` back. The types
  are `deckcoverage.Report`, `Card` and `Bucket`. JSON fields are snake_case (`deck_name`,
  `source_url`, `deck_key`, `oracle_id`), and the report adds `violations` (§1's validation
  output). `deck.ParseDeckURL` derives the deck key without fetching, so a cached deck is not
  fetched at all.
- **Card order.** "Sorted: bucket, then name" uses the order `manual`, `unreviewed`, `caveats`,
  `automated`, `no_effect`: the most actionable first.
- **Buckets.** A catalogued card whose text needs no catalogue entry (Serra Angel's keywords) is
  `no_effect`, following the table's first row. Sideboard rows are resolved and validated but not
  bucketed. A card the engine has an entry for but the catalogue has no verdict for is
  `unreviewed`, never `automated`.
- **The bot's bucket on `POST /deck-coverage`.** An admin session is limited separately (1/s,
  burst 10) instead of per IP. The bot calls from loopback for every guild member at once, and the
  public 1-per-10-seconds bucket would starve it. PR 2 should send its admin session on this route.
- **Fetch errors.** An unsupported link is 400, a missing deck 404, a private or unreadable deck
  422. A Cloudflare block or an unreachable deck site is **502**, not a 4xx: the caller's link was
  fine. Every one carries a sentence to show the player, the violation code and a `violations[]`
  entry.
- **What counts as an ask.** Only an ask that reaches GitHub, an issue filed or a comment added,
  is recorded. `nothing_to_add` and a repeat ask on the same open issue are not. The limit is
  checked before the deck is fetched, and again under the filing lock.
- **Repeat asks.** A requester who already asked on the deck's current issue gets `joined` with
  `already_requested: true` and no comment. "Current" means asked since the row's `created_at`,
  so a refiled issue starts a fresh set of requesters.
- **A deleted issue.** A 404 or 410 from GitHub for the tracked issue counts as closed: a new issue,
  and the row repointed.
- **Off switches.** `POST /deck-requests` answers 503 without `CMDCTRL_GITHUB_TOKEN`, and also
  without a database (`CMDCTRL_DATA_DIR`), since the limit and the deduplication live there.
- **The issue.** It adds a `## Names the card index could not resolve` section when the deck has
  any. The deck name, the display name and unresolved names are redacted and kept to one line, and
  their `@` is broken with a zero-width space so a name cannot ping anyone. With no deck name and
  no commander, the title falls back to the deck key.
