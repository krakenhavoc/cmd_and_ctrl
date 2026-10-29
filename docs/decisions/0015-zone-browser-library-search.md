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

## Amendment 2026-10-16 — search a player's graveyard, hand, and library (#3988)

Kotose, the Silent Spider prints the extraction search: "Search that
player's graveyard, hand, and library for any number of cards with the
same name as that card and exile them. Then that player shuffles." The
Surgical Extraction family prints the same clause. The amendment above
left "searches that also need the hand" uncovered.

### The rule (CR 701.23)

- **All three zones are searched.** The clause says "and", not "and/or",
  so the library is always searched. The search event fires and "then
  that player shuffles" happens whatever was taken, including when every
  card taken came from the graveyard or the hand, and when nothing was
  taken. This differs from the "and/or graveyard" search above, which does
  not shuffle a library it did not look in.
- **The searcher looks at the whole hand.** Searching a zone is looking at
  every card in it. The library's cards are forgotten at the shuffle, but
  the hand's are not: every card in the searched hand becomes known to
  the searcher, the ones that do not match too. Only the searcher learns
  them; no one else does.
- **Any number, none included.** The extraction cards say "any number", so
  the floor is zero (`Unbounded`). A hand is a hidden zone, so a
  stated-quality search could fail to find there in any case (CR 701.23b).
  For the floor rule, a hand candidate counts beside the graveyard's.

### The shape

- **`SearchLibrarySpec.AlsoGraveyardAndHand`** (catalog
  `SearchLibrary.AlsoGraveyardAndHand`), with `LibraryOwner` naming the
  searched player. The graveyard and hand are the library owner's. The
  graveyard half reuses the "and/or" plumbing (`searchesGraveyard()`), and
  `libraryUntouched` is never set, so the shuffle is unconditional. The
  destination is a zone the cards leave to. A battlefield entry from the
  hand opens the CR 614 window with `OldZone` hand, though no card needs it
  yet.
- **`PendingChoice.SearchHandCards`** (snapshot `searchHandCards`, additive
  in v9 with its shape recorded; `Clone` copies it) is the subset of
  `SearchCards` in the hand. As with the graveyard, `checkSearchPicksLocked`
  checks a candidate against the zone it was offered from, and
  `searchFloorLocked` counts the live hand.
- **Wire.** `search_hand` (omitempty) marks the hand candidates, and
  `search_all_zones` tells the client that every zone is searched (read
  through `PendingChoice.SearchesGraveyardHandAndLibrary`, not stored). Both
  are chooser-only, like `options[]`. The search sheet captions hand tiles
  "Hand", and its hint says that the library is shuffled either way. The
  enumerator needs no change, for the reason the amendment above gives.

### Back-out

`TestExtractionSearch*` and `TestCloneCarriesTheHandCandidates` (game),
`TestSearchPromptMarksTheHandHalf` (protocol) and the Kotose card tests
each fail without the shape they name.

## Amendment 2026-10-16 (b) — "all cards" in the extraction search (#4033)

Kotose and Surgical Extraction say "any number of cards with the same
name". The rest of the family (Eradicate, Extirpate, Quash, Scour,
Splinter, Sowing Salt) says "**all** cards with the same name". The
amendment above read every extraction card as "any number", which is
right for those two only.

### The rule (CR 701.23b, 400.2)

CR 701.23b lets a searcher fail to find cards with a stated quality in a
**hidden** zone. A graveyard is a public zone (CR 400.2), so the rule
does not reach it: every card with that name in the graveyard is found
and exiled. The hand and the library are hidden zones, even though the
searcher looks at the whole hand, so the searcher may still leave any of
their matches. There is no count, so "all" has no cap.

### The shape

- **`SearchLibrarySpec.TakeAll`** (catalog `SearchLibrary.TakeAll`), on
  an `AlsoGraveyardAndHand` search. It implies `Unbounded`. Without
  `AlsoGraveyardAndHand` it is ignored: "and/or graveyard" lets the
  searcher skip the graveyard, so nothing in it is forced.
- **The graveyard's matches are not candidates.** Nothing about them is
  a choice, so the prompt offers only the hand's and the library's
  matches, any number, none included. The graveyard's matches ride on
  the spec in the continuation frame (unexported, never set by a card).
  The take moves the ones still in the graveyard together with the
  searcher's picks, so `Then` sees one found list. A match that left the
  graveyard while the prompt was open is not found.
- **No prompt when there is nothing to choose.** With no match in the
  hand or the library, the graveyard's matches are taken without a
  prompt, and the library is still searched and shuffled.
- **No wire, snapshot or enumerator change.** The prompt is the
  "graveyard, hand, and library" prompt above with no graveyard tiles,
  its floor stays zero, and the card's prompt text says that every match
  in the graveyard is exiled.

The alternative was to offer the graveyard matches as locked picks: a
new wire field naming the required candidates, a client that preselects
them, and an enumerator floor. It reaches the same outcome with more
moving parts, and a client that ignored the field could submit an answer
the server must refuse.

### Back-out

`TestTakeAll*` (game) and the extraction family's card tests
(`TestExtirpate*`, `TestEradicate*`, `TestScour*`, `TestSplinter*`,
`TestSowingSalt*`, `TestQuash*`) fail without the shape.

## Amendment 2026-10-29 — "and/or graveyard" records which zones were searched (#4516)

> Superseded in part by the second amendment of 2026-10-29 (#4519): the zones
> are asked first, not with the answer, so `library_searched`, `ResolveSearchLibraryZones`
> and the "I searched my library too" toggle below are gone.

The second amendment of 2026-10-13 declared one simplification: the
engine did not ask which zones an "and/or graveyard" search looked in, so
a search that found nothing was treated as a library search. The printed
text shuffles only "if you search your library this way" (CR 701.23), so
a player who searched the graveyard alone and found nothing was shuffled
anyway, and the cards were stamped Full over it. This closes it.

### The rule (CR 701.23)

- "Library and/or graveyard" lets the searcher choose their zones. The
  answer says whether the library was searched: **a pick from the library
  is a library search**, and a player who took a graveyard card, or
  nothing, **says whether they looked in the library too**. Unsaid, it is
  the graveyard alone.
- The library is shuffled, and `EventSearchLibrary` fires, only when it
  was searched. A graveyard-only search that found nothing neither
  shuffles nor announces; the instruction after the search still runs.
- "Graveyard, hand, and library" names the library, so it is always
  searched and the answer is not asked. A search with no graveyard half is
  a library search whatever it finds.

### The shape

- **`Game.ResolveSearchLibraryZones(choice, chooser, picks, libraryToo)`**;
  `ResolveSearchLibrary` is the `libraryToo = false` form. It reaches the
  existing `libraryUntouched` flag through `executeSearchTakeLocked`, which
  now sets it for an "and/or graveyard" search unless `libraryToo` is true
  or a library card was taken. The flag is already written with the take's
  continuation, so **no snapshot change**: a table restored while the
  prompt is open answers through the same frame (`alsoGraveyard` is in
  `pendingChoices[].searchPrompt`).
- **Wire.** `resolve_choice` takes `library_searched` (bool, omitted is
  false) beside `card_ids`; the prompt carries `search_library_optional`
  (omitempty, chooser-only) on an "and/or graveyard" search, read through
  `PendingChoice.SearchAsksLibraryZone`. Both are additive.
- **Client.** The search sheet shows "I searched my library too" under the
  hint. It reads on, and is disabled, while a library card is picked; it
  is the player's word otherwise, off each time a prompt opens.
- **Bots and the enumerator** are unchanged: their answers never say the
  library was searched, so a bot's "and/or graveyard" search that takes
  nothing, or a graveyard card, does not shuffle.
- The unprompted take (an unconditional search that must find every
  match) has no question to ask and keeps the old reading: library cards
  taken mean a library search.

### Cards

Tower Winder, Delivery Moogle, Finale of Devastation, Ajani's Aid,
Goldmane Griffin, Chandra's Firemaw, Teferi's Wavecaster and Yanling's
Harbinger print "search your library and/or graveyard" and now match
their text; each stays Full and its "declared simplification" comment is
gone. "Search your hand and/or library" (Last Light of Durin's Day, #4515)
is the same question for the hand and is left for a follow-up once that
seam is on develop.

### Back-out

`TestGraveyardOnlySearchFindingNothingDoesNotShuffle`,
`TestGraveyardPickShufflesOnlyIfTheLibraryWasSearched`,
`TestLibraryPickShufflesWhateverTheAnswerSays` and
`TestARestoredAndOrGraveyardSearchStillAsksWhichZones` (game) fail
without the rule; the card tests are `TestTowerWinderWithNoTowerAnywhere`,
`TestTowerWinderBothZonesOfferedMayFindNothing` and
`TestFinaleOfDevastationGraveyardTooBigLeavesFailToFindLegal`.

## Amendment 2026-10-29 (second) — an and/or search asks the zones first (#4519)

The amendment above (#4516) asked which zones were searched *with the
answer*: the prompt showed the library's candidates, and the player then
said whether the library had been searched. That leaks. A player could look
at the library's matches, answer "graveyard only" and skip the shuffle, and
looking through a library is searching it (CR 701.23a, "look at all cards in
that zone"). It also left "search your hand and/or library" (Last Light of
Durin's Day, #4515) without the question. This closes both: the zones are
asked **first**, and the choice is what shows the library and what shuffles
it.

### The rule (CR 701.23, 701.23a)

- An and/or search ("your library and/or graveyard", "your hand and/or
  library") opens on a question: the library, the other zone, or both. No
  library card is shown, made known to the searcher or counted until the
  library is chosen.
- The library is shuffled, and `EventSearchLibrary` fires, **if and only
  if** the library was chosen, whatever is then taken, nothing included. A
  search of the graveyard (or hand) alone never shows, shuffles or announces
  the library. "Graveyard, hand, and library" names all three zones and
  asks nothing; a plain library search has no other zone.
- Under Opposition Agent (CR 723.2, "while they're searching their
  libraries") the control begins when the library is chosen, not before,
  and a search of the other zone alone is neither controlled nor have its
  finds exiled. The question itself is the searcher's, like #4134's.

### The shape

The smallest faithful design is a first step of the same prompt, in the
slot #4134 already built for "you may search".

- **`searchResumeFrame.zones`** (`searchZones`: unasked, asking, library,
  other) beside `undecided`. A new prompt for an and/or spec opens `asking`
  (`SearchLibrarySpec.asksZones`), with all its candidates held on the
  `PendingChoice` as before and none shown or granted.
  `queueSearchChoiceLocked` skips the knower grant and the control record
  for it; the forced take-all shortcut is skipped too (the zones are a
  decision even when one card would be forced).
- **`Game.ChooseSearchZones(choice, chooser, zones)`** answers it. It
  trims the prompt in place to the chosen zones (fresh slices, and a new
  frame: an undo snapshot shares both), recomputes the floor from them
  (`floorZones`), and, if the library was chosen, marks its candidates known
  to the searcher and begins the Opposition Agent control. The next answer
  is the ordinary pick. `checkSearchPicksLocked` refuses a pick while the
  question is open. The prompt is asked even when the chosen zones hold no
  match, so another seat cannot tell from the prompt closing whether the
  library had one (#4140). An optional and/or search would ask #4134's
  "do you search?" first; `BeginSearchLibrary` leaves the zone question
  open.
- **`ResolveSearchLibraryZones(…, libraryToo)` and `library_searched` are
  gone** (both were a day old): the recorded zones decide
  `libraryUntouched` in `executeSearchTakeLocked`, and a prompt that never
  asked keeps the #4516 reading by its take.
- **Wire.** `PendingChoiceView` gains `search_zones_pending` (the question
  is open; `options[]`, `search_min`, `search_graveyard` and `search_hand`
  are withheld), `search_other_zone` (`"graveyard"` or `"hand"`, on the
  prompt before and after) and `search_library_skipped` (answered with the
  other zone alone). `search_library_optional` is gone. The answer is
  `resolve_choice {choice_id, search_zones: ["library", "graveyard"]}`.
  All chooser-only; another seat sees the same prompt before and after.
- **Snapshot.** `pendingChoices[].searchPrompt.zones` (`"asking"`,
  `"library"`, `"other"`; absent when the search does not ask), additive
  in v9, recorded with `-update-shape`. A prompt carrying it is written
  under `search-library/prompt/zones` on `pickThen` (`SearchPromptZonesKey`),
  so a binary from before it refuses the file instead of restoring the
  prompt as an ordinary one that shows the library. A prompt written without
  it restores as one that does not ask. No fixture is edited, and no closure
  route is added: the frame is data.
- **Client.** The search sheet's primary on the question is "Library and
  graveyard" (or "Library and hand"); the secondaries are "Library only" and
  "Graveyard only" (or "Hand only"). Enter searches everything. The "I
  searched my library too" toggle (`L.searchedMyLibrary`) is gone, replaced
  by five labels in the registry (`docs/labels.md`). Once answered, the
  hint says whether the library will be shuffled.
- **Bots and the enumerator.** The question is three moves, none naming a
  card; searching both is the `AlwaysLegal` one, so a stuck seat has an
  answer. The heuristic seat scores a zone answer by its zone count, so it
  searches everything on offer (it cannot see what is where, and a missed
  zone is a missed card). The agent-seat move list carries the same moves.

### Cards

Tower Winder, Delivery Moogle, Finale of Devastation, Ajani's Aid,
Goldmane Griffin, Chandra's Firemaw, Teferi's Wavecaster, Yanling's
Harbinger and Last Light of Durin's Day all stay Full: each matches its
printed text, and the "simplification" note on `AlsoHand` is gone.

### Back-out

Each of these makes the named tests fail: not asking (`initialZones`)
fails `TestAnAndOrSearchAsksTheZonesFirst` and the restore tests; granting
the library's knowledge up front fails the same first test; accepting a
pick before the answer fails it and
`TestARestoreAtEachStepOfTheZoneQuestion`; keeping the library's cards when
only the other zone is chosen fails
`TestGraveyardOnlyNeverOffersTheLibraryAndDoesNotShuffle` and
`TestAHandAndOrLibrarySearchAsksTheZonesToo`; shuffling a graveyard-only
search fails the same pair and `TestAnEmptyAndOrSearchStillAsksBothQuestions`;
not shuffling a chosen library fails
`TestChoosingTheLibraryOffersItAndShuffles`; taking a one-card unconditional
search unasked fails `TestAnUnconditionalAndOrSearchIsNeverTakenUnasked`;
exiling a graveyard-only find, or beginning the control before the library
is chosen, fails the `TestOppositionAgent…` zone tests (`opposition_agent_zones_test.go`); and the view,
enumerator, dispatcher, snapshot and client each have their own
(`TestSearchPromptAsksTheZonesBeforeShowingAnything`,
`TestEnumeratorAsksTheZonesBeforeOfferingAnyCard`,
`TestDispatchResolveChoiceSearchZones`, `TestTheZoneStepIsWrittenUnderItsOwnKey`
and the "asked first (#4519)" render tests).
