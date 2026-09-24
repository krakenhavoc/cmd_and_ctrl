# ADR 0092 — A public roadmap from a CI-checked registry, and a site portal

**Status:** Accepted · 2026-09-24 · Post-S30 — Rolling deck-driven catalog growth (outside a numbered sprint)
**Issue:** [#1386](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1386) (Public roadmap page + site portal)
**Numbering:** swept with the AGENTS.md §4 check on 2026-09-24 — `git fetch --all --prune`, then the
ADR file names in the history of every one of the 408 remote branches. The highest number present
anywhere is **0091** (`0091-hideaway.md`); `0092` appears on no branch.

**Related:** [ADR 0042](0042-card-catalog-page.md) (the card catalogue and `Completeness`, whose tone
rule this reuses), [ADR 0037](0037-unimplemented-card-signal.md) (the closed keyword table),
[`docs/engine-seams.md`](../engine-seams.md) (the seam registry, whose open table this generates),
`server/internal/cards/coverage` (the census, the curated mechanic table and the oracle fixture).

---

## Context

Players can't see what the engine supports. The answer exists, but it is spread across four places,
all written for engineers and none checked against the others:

- **`docs/engine-seams.md`'s open table.** Hand-typed. When #1386 re-checked it against the code, four
  of its 21 rows had fully closed (dice and coin flips, layer invalidation, the pick-from-exile prompt,
  the post-departure record) and most of the rest had shipped the larger half of what they described.
  AGENTS.md's memory already warns that "the engine seams doc runs stale"; this was the measurement.
- **The census** in `docs/decklists/card-coverage-roadmap.md`: counts, not capabilities.
- **`coverage.Mechanics()`**: nineteen mechanics with a probe each. Its job is catching stale caveats,
  not describing the engine.
- **Each card's `Caveats`**: accurate and player-facing, but per card.

The site has a second, smaller problem. Its pages are reachable only through one another: the
catalogue is linked only from Login, My games from Lobby, Login and Game, and there is no shared
navigation at all.

The owner's decisions on 2026-09-24, recorded on #1386:

- The roadmap is **public**. The catalogue stays behind sign-in.
- Its data comes from a **curated registry checked by CI**. The same registry generates
  `engine-seams.md`'s open table.
- "Up next" is **ranked by the cards each item unblocks, with manual pins on top**.
- Two portal designs are to be **built**, a shared header and a `#/home` page, and the owner will pick
  one of them later.

## Decision 1 — The roadmap is public because it carries names, never art or text

`GET /catalog` was built unauthenticated and later put behind `auth.Middleware`. The comment at the
mount in `server/cmd/server/main.go` gives the reason: AGENTS.md §1 and §8 describe a private,
personal-use project, and "serving card art to anonymous visitors is a different posture from the one
the repo states".

The roadmap stays clear of that reason by construction:

- It publishes **card names and caveat sentences**. It publishes no image, no image URL, no Scryfall
  printing ID and no oracle text.
- A caveat is text this repository wrote, not text from a card.
- `TestBuildIsPlainPublishableData` fails if the serialised roadmap ever carries an `image`,
  `oracle_text`, `scryfall_id` or `engine_notes` key.

The catalogue does not change and stays signed-in. For a signed-in viewer an example card's name may
link into the catalogue, and the catalogue then shows the art as it does today. A signed-out viewer
sees plain text.

The comments that still called the catalogue public are corrected in this ADR's first PR:
`Catalog.svelte`, `Login.svelte`, `catalog/http.go`, `catalog/catalog.go`, and the endpoint list in
AGENTS.md §5.

## Decision 2 — `server/internal/roadmap` is the single source; `engine-seams.md`'s open table is generated from it

A new package, `server/internal/roadmap`, sits above `effects`, `coverage` and `catalog`, the way
`catalog` does. `registry.go` is one hand-written `Item` per keyword, mechanic or engine seam. Each
item has:

- a kind and a status (implemented, partial or missing);
- a player-facing `Summary`, and a `Missing` sentence when the status is not implemented;
- CR citations, a tracking issue and an ADR;
- a probe;
- curated `Examples`, the `Waiting` cards and `Unblocks`;
- an optional `Pin`;
- `EngineNotes`, the engineering prose.

A status is a judgement, so a person makes it. Deriving it would mean a probe claiming that something
is missing, and a probe cannot know about a gap no card has hit yet.

Probes come in three kinds, and each asks a question that has an honest answer:

1. **A Spec declaration.** For example: `PrintedKeywords` has the token, an `OptionalCosts` key is
   `game.KickerKey`, or a `SpecialAction` has kind foretell. No probe inspects a closure.
2. **A row of `coverage.Mechanics()`, borrowed by name.** Its phrases are always borrowed. Its probe is
   borrowed only when the row is `Exact`. A heuristic probe, such as surveil's "grew an entry hook", is
   never used to choose an example card.
3. **The printed text.** A regular expression runs over the card's oracle text, and a match counts only
   when the card file declares the card complete. This is the honest answer for mechanics that are
   closures (ward, cascade, hideaway): if a card prints "Cascade" and a person declared the card fully
   automated, the engine does cascade on that card. The text comes from
   `coverage/testdata/oracle_text.json`, which is now embedded through `coverage.PrintedOracle()` so that
   a running server can read it. It is the same file the oracle guard already reads and the nightly
   fixture test keeps current, and nothing serves it.

`Build()` runs once per binary behind a `sync.Once`, because the registry and the catalogue are both
fixed at init. It returns plain data:

- counts by status;
- the "up next" slugs;
- one entry per item, with at most three **examples** (complete cards, curated ones first) and at most
  three **partial examples** (`{name, caveat}` for cards that work with a gap this item explains).

"Complete" is `catalog.Build`'s merged verdict, so it means the same thing on both pages, with one
extra rule: a card whose front face has no spec is never shown as complete.

**The open-seams table** in `engine-seams.md` is now the block between two markers, generated by
`SeamsBlock()`. `TestSeamsTableIsCurrent` fails any PR whose table and registry disagree, and
`go test ./internal/roadmap/ -update` rewrites the block. The census is a function of the live catalogue,
so CI refreshes it (Discussion #1231). The seams table is a function of the registry alone, so it changes
only in the PR that edits `registry.go` and needs no refresh step. The Closed seams list stays
hand-written history.

The table drops the old **Count** column: it tallied recorded skips, and Cards waiting is now that list
itself. The **Audit** column becomes **Unblocks**: the audit's only-blocker figure, carried over where it
still describes open work and set to 0 where it counted a half of the seam that has since shipped.

## Decision 3 — CI checks what can be checked

`registry_test.go` enforces the following:

- Every row of `game.CanonicalKeywordTable()` has exactly one keyword item, and a keyword item names
  only tokens from that table. A keyword therefore reaches the page in the same change that adds it to
  the table.
- Every row of `coverage.Mechanics()` is borrowed by exactly one item.
- An implemented or partial item finds at least one complete card, or it says why none can exist
  (`NoCatalogExample`; for example, a creature whose only ability is fear needs no card file). An excuse
  that has stopped being needed fails the test too.
- Curated examples exist, are complete, and satisfy the item's own probe.
- **A card waiting on an item must not be registered as complete.** When one is, the failure says "the
  seam may have closed". This is the check the hand-typed table never had.
- A partial or missing item has a tracking issue and a `Missing` sentence, and an open seam has
  `EngineNotes`.
- `Summary`, `Missing` and `NoCatalogExample` pass the Caveats tone rule. The rule is one function,
  `effects.PlayerFacingProblems`, which the catalogue's caveat test now calls as well, so the two cannot
  drift apart.
- Slugs and pins are unique.
- As a report rather than a failure, the test lists caveat cards that no item explains. Those are
  where the next registry entries come from.

## Decision 4 — "Up next" is pins first, then cards unblocked

Pinned items come first, in `Pin` order. They are followed by every partial or missing item with a
non-zero `Unblocks + len(Waiting)`, highest score first, with ties broken by name, and the list stops
at six. There are no pins today; the owner will set them.

The ranking reuses two numbers the project already keeps: the audit's only-blocker count and the batch
skips. Neither is precise. The audit is an upper bound, and skips are recorded unevenly. Both are
honest about which gaps have the most cards behind them, and a pin overrides either when the owner
knows better.

## Decision 5 — Build both portals and choose later

The first of two portal experiments is **a shared header** (`SiteHeader.svelte`):

- the wordmark, then Lobby, My games, Catalog, Roadmap and Home, with Sign in or out on the right;
- signed-out visitors see only Roadmap, Home and Sign in;
- it replaces the copied top bars on Catalog, Join and Reclaim, and is added to Lobby, My games,
  Roadmap and Home;
- it stays off the game table.

The second is **a public `#/home` page**: a sitemap grid grouped into:

- Play: Lobby, My games, and joining with a code;
- Cards: Catalog and Roadmap;
- Help: the bot guide, the repository, and bug reports when they are enabled;
- Account.

Both ship. Landing after sign-in stays the lobby until the owner picks one, and the other is then
deleted.

## Delivery

- **PR 1** (this ADR): the registry, `Build`, the CI truth tests, the generated seams table and the
  re-triage of every open seam row, plus the corrected catalogue comments. It adds no route.
- **PR 2**: `GET /roadmap`, mounted without `auth.Middleware`, with `Cache-Control: public, max-age=300`,
  and the `#/roadmap` page.
- **PR 3**: the shared header and `#/home`.

## Consequences

- A batch PR records a skip by appending a name to `Waiting` in `registry.go` and running `-update`.
  It no longer edits a markdown table by hand. AGENTS.md §7 and the doc's intro say so.
- An engine PR that closes a seam finds out from CI: the first waiting card registered complete fails
  `TestWaitingCardsAreNotComplete`.
- The server binary embeds the oracle fixture, about half a megabyte, which it never serves.
- A registry status can still be wrong. The tests catch an item that claims too little, such as a
  waiting card that already works or an excuse that has gone stale. They cannot catch one that claims
  too much, such as a partial item whose remaining gap has quietly closed. That half is human review,
  which is also true of `Caveats`.
