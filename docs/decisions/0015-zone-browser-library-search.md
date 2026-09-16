# ADR 0015 — Zone browser modal shell (S18.5)

**Status:** Implemented · 2026-04-23 · Sprint S18.5

## Context

Two client-UX items were promised in earlier sprints and never
shipped:

- **S06 line 172** — "Graveyard / exile / command — clickable modal
  browser" slipped to S07 and then got squeezed out. Until S18.5,
  clicking a pile chip did nothing for graveyard / exile / command;
  library was the only wired pile, and only for its owner (draws a
  card).
- **S11 line 173** — "Library — searchable / reorderable." S14 later
  shipped the `SearchLibrary` primitive (Demonic Tutor, Vampiric
  Tutor, Cultivate) but the server auto-picks the first library
  match; there's still no player-facing picker today.

S20 will land predicate-filtered targeting (Doom Blade → picker
shows only non-black creatures). That picker will want a modal grid
with a card-name filter and click-to-submit — the same shape we'd
give a library-search modal. A reusable "modal shell + pick from a
filtered card list" pattern is worth more than two bespoke modals.

S18.5's task was to close both bullets in one mini-sprint. Only the
zone-browser bullet shipped.

## Decisions

### 1. Module-scope `zoneBrowser` store, one modal mounted in `Board.svelte`

The open/close state of the zone-browser modal lives in a tiny
writable (`client/src/lib/zoneBrowser.ts`) rather than threaded
through PlayerPanel / PileBar / CommandZone as props. Pile chips
call `openZoneBrowser({zoneKind, ownerID, ownerName})`; `Board.svelte`
mounts one `ZoneBrowserModal` bound to the store and passes
`view` + `viewerID` + `sendAction` in.

**Why module scope:** the PileButton API today is a bare `onClick` —
pushing a "browse this zone" handler through two intermediaries
(PileBar → PlayerPanel → Board) for every opener would bloat the
prop surface and duplicate the dispatch logic on every caller.
Writing to a store keeps PileButton unchanged and makes it trivial
for a future consumer (e.g. the stack overlay's "show all items
on the stack" affordance, or the S20 predicate picker) to open the
same modal with one call.

**Why one modal instance:** there is only ever one "what zone am I
looking at" question at a time. A multi-modal stack is more
complexity than the UX warrants.

### 2. Pure helpers in `zoneBrowser.logic.ts` for test coverage

`cardsForZone`, `canManageZone`, and `buildMovePayload` live
outside the `.svelte` file so vitest can exercise them without a
jsdom / Svelte component renderer. The client is node-only at
test time today (the existing test suite — `priority.test.ts`,
`settings.test.ts`, etc. — all test pure logic modules). Keeping
the derivations out-of-component means the wire shape for
`move_card` is pinned by tests rather than by a smoke pass through
the UI.

### 3. Client-side `canManageZone` mirrors `requireCardController`

The server's `actions.go` dispatcher gates `move_card` on the
caller being the card's controller. Non-owners attempting to move
cards out of someone else's graveyard will get
`ErrCardCallerMismatch`. The client-side `canManageZone` +
`buildMovePayload` guards are strictly UX cleanliness: a non-owner
simply doesn't see the action cluster, so they never fire a
speculative action the server would bounce. The server guard
remains authoritative.

### 4. Library-search modal deferred to when the server supports it

The stated M3 scope was a search modal hooked into a `SearchLibrary`
prompt frame. The investigation found that no such frame exists
today: `SearchLibraryForEffectWithOptions` picks the first library
match inline, no `PendingChoice` is emitted, and the catalog cards
(Demonic Tutor, Cultivate, etc.) resolve with zero player input.

Landing the player-facing picker requires new server work:

1. A new `PendingChoice` kind — `search_library` — with the
   viewer-known library slice filtered by the primitive's predicate.
2. A new resolve path on `resolve_choice` that moves the chosen
   card to `Dest`, optionally shuffles, and emits the existing
   `EventSearchLibrary`.
3. `SearchLibrary{}` grows a `PlayerPicks bool` (or similar) flag
   so "auto-pick first match" stays the default for Cultivate's
   two-for-one fetch and other effects that don't need the UI.

The S18.5 sprint brief drew a "client UI only" fence around this
work ("do not alter the Go server unless you discover the
SearchLibrary prompt frame is missing — in that case, STOP and
report back"). Since the frame is missing, the library-search
modal is deferred.

**Where it belongs instead:** S22 (card draw + library
manipulation) is already planned to add `ScryN`, `SurveilN`,
`RevealAndChoose`, and similar primitives with matching wire
frames. A tutor-picker is cousin to those modals and will share
~80% of the component code. Fold it into S22's scope when that
sprint kicks off.

### 5. Library-bottom destination intentionally not wired

The `move_card` action today doesn't distinguish top-of-library
from bottom — `ZoneLibrary` is a single zone. Library of Leng
(ADR 0013 replacement effect that redirects discards to the top)
and Vampiric Tutor (search to top) need the distinction at the
engine level; the browser modal punts and always sends library
moves to the top. Adding a top/bottom split is cleaner done
alongside S22's library-manipulation primitives than grafted onto
the browser now.

*Correction, 2026-09-16:* Library of Leng was never shipped (ADR 0013
§10a), and it puts the discarded card on **top** of the library, so it
never needed a bottom destination. The discard itself doesn't reach
the replacement pipeline yet
([#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650)).

## Consequences

### Good

- Every seated player can now inspect any other player's public
  zones at the table. This closes a UX gap that made mid-game
  "wait, what's in your graveyard?" into a verbal ask.
- Owner-action cluster gives a sandbox escape hatch for the common
  "oh I need to fetch that exiled card back" cases without leaning
  on the admin-undo button.
- The modal shell + `buildMovePayload` shape is a natural fit for
  the S20 predicate-picker and the eventual S22 tutor modal — both
  want "grid of cards, filter, pick".

### Neutral

- Library search deferred to S22; the sprint brief's M3 was
  partially descoped. Tracked here and in the sprint entry so it
  doesn't slip again.
- Stack is a browsable zone in the new modal but stack management
  stays on `StackOverlay` (counter / resolve). That's fine — the
  browser is "show me what's there"; the overlay is the interaction
  surface for the stack's ordering semantics.

### Bad

- Nothing structurally. The added module-scope store is tiny and
  singular; PileButton's surface is unchanged; no new wire paths
  or server mutations were introduced.

## Follow-ups

- S20 predicate picker should reuse `ZoneBrowserModal`'s layout
  (card grid + name filter) rather than build its own modal from
  scratch. Pull the grid + filter chunk into a sub-component if
  the re-use shape doesn't fit cleanly.
- S22 kickoff: add the `PendingChoice` kind + server plumbing
  described in §4 above, then wire `ZoneBrowserModal` (or a
  sibling `LibrarySearchModal` reusing the same primitives) to
  the new prompt frame.
- Library-top vs library-bottom destination split when S22 adds
  the engine distinction.
