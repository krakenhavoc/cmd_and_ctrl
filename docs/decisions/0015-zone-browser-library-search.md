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

## Amendment 2026-10-13 — a search may find nothing; "up to N" may take fewer (#3053, #3081)

§4 above deferred the library-search prompt to S22, which built it
(`PendingChoiceSearchLibrary`, `SearchLibrary`, `ResolveSearchLibrary`).
S22 kept one shortcut from the old auto-pick: when the library held no
more matches than the effect could take, the search resolved
synchronously and took them, unasked. This amendment removes it for the
searches it was wrong for. The owner's decision on #3053 (option 1 of
the three the issue lists) is the rule; this records how the engine
implements it.

### The rule (CR 701.23b, 701.23d)

- A search of a **hidden zone** (a library) for cards with a **stated
  quality** — a type, a name, a land type, a mana value — may find
  nothing even when matches are there, including exactly one match. An
  "up to N" search with a stated quality may take any number from 0 to N.
- A search **simply for a quantity** ("a card", "three cards": Demonic
  Tutor, Entomb, Intuition) must find that many, or as many as the zone
  holds. "Up to N cards" with no quality may take fewer: "up to" is what
  the clause says, and CR 701.23d speaks only of a plain quantity.
- A search of a **public zone** must find a valid card when one is
  present. The engine has no public-zone search; the rule is encoded in
  `searchFloor` and tested there for the day one lands.
- Whatever is found, the rest of the instruction happens: the shuffle,
  `EventSearchLibrary`, the continuation (`Then`), life loss.

### The shape: an engine rule, not a per-card flag

`SearchLibrarySpec.statedQuality()` is `Pred != nil || Validate != nil`.
`spec.floor(matches)` is the fewest cards the searcher may take, from
`searchFloor(hiddenZone, statedQuality, mayFindFewer, limit, matches)`:
zero when the clause allows a short count (`Optional`, `UpTo`,
`Unbounded`) or the search names a quality in a hidden zone; otherwise
`min(limit, matches)`.

- **The shortcut** now takes the matches unasked only when the floor
  equals the match count — an unconditional search against a library of
  exactly that many. Every stated-quality search prompts, however few
  matches there are.
- **`PendingChoice.SearchMin`** (wire `search_min`, omitempty) carries the
  floor; `SearchMax` is unchanged. `checkSearchPicksLocked` refuses a
  short answer, with the floor lowered by candidates that left the library
  since the prompt opened (milled by another seat's trigger) — a rule that
  says "as many as possible" must not wedge the table on an answer nobody
  can give. `SearchFloorLocked` exposes it to the enumerator, which
  offers "fail to find" (the choice's `AlwaysLegal` answer) only when the
  floor is zero and enumerates picks of size `max(floor, 1)`..`SearchMax`.
- **`Optional` is demoted to its real meaning**: "you MAY search". It was
  doing double duty as the only way to get "fail to find" onto a
  stated-quality search. Review of the 42 `Optional: true` search sites
  removed it from the 9 sites on 7 cards that were only standing in for that (Cultivate,
  Point the Way, Reshape the Earth, Shared Summons, Titania's Command,
  Urza's Saga III, Axgard Armory) and kept it on every card whose text
  says "may search". A new field `UpTo` covers the one case the quality
  rule does not: "up to N" with no quality.
- **Seven catalog cards passed `func(game.Card) bool { return true }`**
  for "a card" (Case of the Stashed Skeleton, Demonic Counsel's
  delirium half, Hour of Victory, Lagomos, Razaketh, Tamiyo's Journal,
  Vile Entomber). The engine reads that as a stated quality, so they
  could have failed to find. They now omit `Predicate`.

### #3081

The 17 "up to N" cards the issue lists (Blighted Woodland, Burnished
Hart, Circuitous Route, Cycle of Renewal, Encroaching Dragonstorm,
Explosive Vegetation, Harrow, Myriad Landscape, Roiling Regrowth,
Skyshroud Claim, Emergent Ultimatum, Gifts Ungiven, Saheeli Rai, Entish
Restoration, Scapeshift, Traverse the Outlands, Fblthp Knows the Way)
all search for a stated quality, so the engine rule covers them with no
per-card change; the oracle text of each says "up to". The issue's
suggested `UpTo` flag exists but is for the unconditional case only.

### Client and bots

One prompt per tutor is the accepted cost; there is no setting. The
search sheet opens with the matches already taken when they all fit under
the ceiling (the cases that used to resolve silently), so Enter takes
them; "Find nothing" is a separate secondary button, shown whether or not
a card is picked, and absent when `search_min` is above zero. The
heuristic bot's score for a take is positive and for "find nothing" is
zero, so it takes the best card unless there is a reason not to; the
enumerator offers both answers.

## Amendment 2026-10-13 (second) — search your library and/or graveyard (#3217, #3218)

The amendment above wrote the public-zone half of CR 701.23b into
`searchFloor(hiddenZone, …)` "for the day one lands". Three catalog cards
(Finale of Devastation, Tower Winder, Ajani's Aid) print "search your
library **and/or graveyard**", and shipped with a caveat because the
engine searched one library only. This is that day.

### The rule

- The searcher may search the library, the graveyard, or both for ONE
  qualifying card. The engine does not ask which zones: it offers every
  match from both in one prompt, the way a player with the graveyard face
  up on the table would look there first.
- "And/or" lets the searcher choose which zones to look in, so the
  graveyard half adds **no must-find**: a player may search only the
  library and fail to find even with a match in the graveyard (the
  rulings on Ajani's Aid and Finale of Devastation). The floor is the
  library-only floor of the amendment above: zero for a search with a
  stated quality, `min(limit, cards in both zones)` for an unconditional
  "a card" search (no catalog card has one today). An earlier draft of
  this amendment made a graveyard match mandatory (`min(limit, graveyard
  matches)`); that removed a choice the cards print and was reverted
  before it landed.
- "If you search your library this way, shuffle": **a take that comes
  wholly out of the graveyard did not search the library**, so it does
  not shuffle and does not fire `EventSearchLibrary` (an Archivist of
  Oghma must not trigger). Taking any library card is a library search
  and shuffles. A card taken from the graveyard is a public card, so it
  is never revealed as a library card.
- **Declared simplification: finding nothing counts as a library
  search** (shuffle plus `EventSearchLibrary`), the common case. The
  engine does not ask which zones were searched, so a player who looked
  only in the graveyard and found nothing would, by the rules, not have
  shuffled. Finding nothing in the graveyard alone is a no-op difference
  for play (a shuffle of an unknown library), and the card comments on
  Finale of Devastation, Tower Winder, Ajani's Aid and Delivery Moogle
  say so.

### The shape

- **`SearchLibrarySpec.AlsoGraveyard`** (catalog `SearchLibrary.
  AlsoGraveyard`). The graveyard searched is the library owner's, so it
  is not meaningful with `LibraryOwner`; the destination is hand,
  battlefield or exile (not `ZoneLibrary`).
- **`spec.floorZones(lib, gy)`** is the floor: `searchFloor` over the
  library's rule, with an unconditional search counting both zones'
  cards. `searchFloorLocked` recomputes it from the live zones when the
  prompt carries graveyard candidates, so an unconditional search's
  floor falls if its candidates leave while the prompt is open.
- **`PendingChoice.SearchGraveyardCards`** (wire `search_graveyard`,
  omitempty, chooser-only like `options`) is the subset of
  `SearchCards` that sits in the graveyard; the library's candidates
  stay first. `checkSearchPicksLocked` checks a candidate against the
  zone it was OFFERED from, so a library card milled mid-prompt cannot
  be claimed by turning up in the graveyard. The snapshot field is
  additive (`searchGraveyardCards`), and `Clone` copies it.
- **The take.** `executeSearchTakeLocked` sets an unexported
  `libraryUntouched` on its copy of the spec when every card taken is in
  the graveyard; `finishSearchLocked` then skips the event and the
  shuffle and still runs `Then`. The hand and exile destinations route
  through the one exit primitive, which finds the card's zone itself; the
  battlefield destination stamps the controller in the zone the card is in
  and opens the CR 614 entry window with `OldZone` graveyard.
- **Client.** The search sheet tiles graveyard cards separately from
  library ones (a Command Tower in each is two picks) with a "Graveyard"
  caption, and its hint says what the shuffle depends on. The enumerator
  needed no change: it offers picks over `SearchCards` and asks the engine
  to accept each, and the fail-to-find answer it marks always-legal
  appears whenever the floor is zero, which for a stated-quality search is
  always.

### Back-out

Each shape has a named test that fails without it; the table is in the PR
body for #3217.
