# ADR 0078 — Tokens get real artwork: a Scryfall token printing, resolved at runtime

**Status:** Proposed · 2026-09-19 · S35 — Playtest stabilisation, round 2
**Tracking issue:** [#1115](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1115)
**Depends on:** [ADR 0010](0010-card-effect-catalog.md) (the oracle-ID-keyed
catalog and the hook-variable seam the engine reads it through),
[ADR 0034](0034-multi-face-cards.md) (faces, `?face=N`, and `cards.Card`'s
shape), [ADR 0041](0041-game-persistence.md) (what a restore point carries),
[ADR 0061](0061-token-creation-and-discard-are-replaceable-events.md) (the one
shared token-creation path this stamps on).
**Related, and deliberately after this:**
[#521](https://github.com/krakenhavoc/cmd_and_ctrl/issues/521) (S33 sub-PR 5,
synthetic token catalog keys).
**Split out of this, and independent in both directions:**
[#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127) — 21 token
templates declare no colour where the printed token is coloured. Found by this
ADR's matching rule, and a rules bug on its own terms; see decision 3d.

**Numbering:** `git fetch origin` then `git ls-tree docs/decisions/` over every
remote head (`git branch -r`, 385 heads, HEAD excluded) on 2026-09-19. The
highest number present anywhere is `0077-opponent-board-summary.md`; 0078
appears on no branch. 0005, 0024, 0029 and 0030 stay permanently unused per
AGENTS.md §4.

**Every structural claim below was checked against `origin/develop` and against
the Scryfall dump at `data/scryfall/default-cards.json` on 2026-09-19.** That
local dump's `last-refresh` stamp is `2026-04-19T11:06:04Z`, so the record
counts are that snapshot's; the shape is what matters and the shape is stable.

---

## Context

### What a player sees

A token renders as its name in text on a card-shaped div
([`Card.svelte:437`](../../client/src/lib/components/board/Card.svelte), the
`.name-fallback` branch). Treasures, Clues, Food and every creature token are
grey rectangles with words on them while every real card has art. That is the
whole of #1115.

The mechanism is one missing string. `cardImageURL` returns `null` when a
`CardView` has no `scryfall_id`
([`cardImage.ts:41`](../../client/src/lib/cardImage.ts)), and a token is a
`game.Card` created with `ScryfallID` and `OracleID` empty
(`CreateTokensThenForEffect`, reached from
[`effect_api.go:2223`](../../server/internal/game/effect_api.go)). No id, no
request, no picture.

### The comment that has been wrong for a while

[`tokens.go:10-13`](../../server/internal/cards/effects/tokens.go) says:

> Tokens have an empty ScryfallID because they're not a Scryfall printing.

They are. The dump holds **2,957** records with `"layout":"token"`, of which
**2,922** are English, and **every one of those 2,922 carries the full six-size
`image_uris` map**. A token printing is an ordinary record with an id, a name, a
type line, printed P/T, `colors`, `keywords` and art.

`Index.Load` keeps every record it decodes
([`index.go:263`](../../server/internal/cards/index.go)); `isPlayablePrint`
(`index.go:491`) only *demotes* token layouts when a name index entry collides
with a real printing. So `idx.Get(<token printing id>)` already succeeds, and
**`GET /cards/<token-printing-id>/image` already serves the art today** — the
handler ([`http.go:48`](../../server/internal/cards/http.go)) accepts any UUID
the index knows, and `ImageCache.FetchFace`
([`images.go:111`](../../server/internal/cards/images.go)) reads
`ImageURIForFace` off whatever record comes back. The service worker's
`CARD_IMAGE_PATH` already matches that URL shape
([`service-worker.js:67`](../../client/src/sw/service-worker.js)), so the
offline cache needs no change either.

The missing piece is **choosing an id and stamping it on the card**. Nothing
else in the image path moves.

### What the wire cannot say

`CardView` has no `is_token` ([`view.go:973`](../../server/internal/protocol/view.go)).
The client infers token-ness by looking for the word in `type_line`, which is
the *effective* type line and is redacted for a card the viewer does not know.
The engine already has the predicate — `game.Card.IsToken()`
([`card.go:1012`](../../server/internal/game/card.go)) — and it is the same test
the CR 704.5d state-based action uses. The wire simply does not carry it.

### Corrections to the survey in #1115

Four claims in the issue have drifted or were off; the conclusions all survive.

| Claim | Actual |
|---|---|
| `CreateTokenForEffect` at `effect_api.go:1545` | `effect_api.go:2223`. Line 1545 is `ReturnFromGraveyardForEffect`. |
| "96 tokens live in `tokens_table.go`" | **123** rows today, plus 9 behaviour templates in `tokens.go` and 7 further `Token …` literals elsewhere in `cards/effects`. |
| "2,958 `layout:"token"` records" | **2,957**. A `reversible_card` record carries the string `"layout":"token"` inside a face and inflates a naive `grep -c` by one. |
| The deferral comment at `tokens.go:9-13` | Lines **10-13**. |

And two that were not in the survey at all, both load-bearing, in
[§ Two predicates that read `ScryfallID != ""`](#two-predicates-that-read-scryfallid--) below.

---

## Decision 1 — The art id is a Scryfall **printing**, stamped on `Card.ScryfallID` at token creation

One field, the one that already exists, written once, at creation, in the single
shared creation path (`CreateTokensThenForEffect` → `token_create.go`), after
the CR 701.7b replacement window and before the battlefield entry.

`OracleID` stays empty. That is deliberate and it is what makes this cheap:
every catalog-shaped thing in the engine keys on the **oracle** id —
`game.CatalogKey`, `game.IsAutoCard`, `TargetModeFor`, `Unimplemented`,
`catalog.IsCatalogPrinting` — so stamping only the printing id changes what the
client can *draw* and nothing about what the engine can *do*. The AUTO badge
does not appear on a Treasure, the catalog does not gain 123 entries, and
[#521](https://github.com/krakenhavoc/cmd_and_ctrl/issues/521)'s synthetic key
namespace is untouched.

**Not** a new `Card.ArtScryfallID` field. Reusing `ScryfallID` is what makes the
client, the image route, the on-disk image cache, the service worker and the
snapshot all work with no change: `ScryfallID` is already classified `carried`
in [`snapshot_drift_test.go:182`](../../server/internal/game/snapshot_drift_test.go),
so a token's art survives capture and restore for free and **the drift test
needs no new row**. A new field would have to be classified there, serialised,
projected onto the wire separately, and taught to `cardImageURL`. The cost of
reuse is the two predicates in the section below; that cost is two lines.

## Decision 2 — Resolution happens at **creation**, not at view-build time

The alternative — resolve in `viewOfCard` and never persist — was considered and
rejected on one consequence: a token's art would be recomputed from whatever
dump the process has, every frame, forever. A game restored after a Sunday
refresh would come back with different art on the same physical token, and a
replay would not reproduce what the players saw.

Stamping at creation gives the opposite, and it is the property the owner
should want: **the id is on the object**. It goes into the restore point,
because `ScryfallID` is `carried`; it goes into the replay; it is redacted for a
non-knower by the existing `out.ScryfallID = ""` line
([`view.go:4194`](../../server/internal/protocol/view.go)) with no new code. A
token in a saved game keeps whatever art it had when it was made, for as long as
it exists, across deploys and dump refreshes.

The cost is honest and small: **the rule can change what it picks between one
process and the next**, so two Soldiers made in the same game, either side of a
deploy that followed a dump refresh, can carry different art. In a game where a
Soldier is made every turn that is visible. It is also exactly what paper does
when you run out of one token card and reach for another, and it is bounded by
decision 4.

## Decision 3 — The matching rule

Given a token template (a `game.Card` with `Name`, `TypeLine`, `Power`,
`Toughness`, `Colors`, `Keywords`), the resolver runs three stages: **eligible
pool**, **identity match**, **preference**, then decision 4's tie-break.

### 3a. The eligible pool, precomputed once at `Index.Load`

A record joins the pool when **all** of:

| Field | Test | Why |
|---|---|---|
| `layout` | `== "token"` | `double_faced_token` is out of scope (decision 9); every other layout is a real card. |
| `lang` | `== "en"` | The dump carries 35 Japanese token records. The table is written in English and the client renders English. |
| `image_uris` | non-empty | The whole point. 2,922 of 2,922 English token records pass; the test is there because a future record may not. |
| `border_color` | `== "black"` | Excludes **106** records: 83 gold-bordered (World Championship / Collectors' Edition reprints), 17 silver-bordered (Un-sets), 4 white, 2 borderless. |
| `set_type` | not in `{memorabilia, minigame, funny}` | Excludes **178** records of the same character. `masters`, `box`, `promo`, `duel_deck` and `from_the_vault` stay in: those are ordinary token printings. |

**2,722 records** survive. This is the one full pass; see decision 5.

### 3b. Identity — five fields, all required, exact

A pool record is a **candidate** only if every one of these matches. The order
below is the order to compare in, cheapest and most selective first:

1. **`name`**, case-insensitively, after whitespace collapse — `normalizeName`
   already exists in `index.go:521` and is the right function.
2. **`type_line`**, compared as *(card types, subtypes)* rather than as a
   string. Parse both sides with the engine's own `game.ParseTypeLine`, discard
   supertypes, compare the two sets. String equality is wrong and the dump says
   so: the Halfling token prints **`"Tolkien Creature — Halfling"`**, not
   `"Token Creature — Halfling"`. Discarding supertypes also means the
   template's `"Token "` prefix costs nothing.
3. **`power`** and **`toughness`**, as Scryfall's *strings*, compared against
   `strconv.Itoa` of the template's ints — and empty on both sides for a
   non-creature token. A `"*"` never equals a number, which is correct: a
   characteristic-defining token is not the same object as a printed 0/0 one.
4. **`colors`** as a set (order-insensitive), against the template's `Colors`.
   Scryfall's `colors`, not `color_identity` — a Treasure's identity is empty
   but so are its colours, and for a coloured token with an activated ability
   the two diverge.

`cards.Card` carries none of `colors`, `keywords` or `released_at` today
(`index.go:261`). `colors` and `keywords` are added by this work; `released_at`
is not needed, which is part of decision 4's argument.

### 3c. Preference — keywords, as a tie-break and not a filter

Among the candidates, prefer those whose Scryfall `keywords` set equals the
template's `Keywords` set, case-folded. If none does, keep the whole candidate
set rather than failing.

Soft, not hard, because the two lists answer slightly different questions: the
template's `Keywords` are what the engine grants (`printedCharacteristic` folds
them into the layer engine), Scryfall's are what is printed on that piece of
card. They agree almost always and the difference is usually art-irrelevant.
It earns its place anyway: **on the 100 table rows that resolve, the keyword
pass narrows the candidate set on 20 of them**, and it is what separates the 52
flying Spirits from the 1 non-flying one, and the 5 lifelinking Soldiers from
the 59 plain ones.

### 3d. What this rule actually does, measured

Simulated against all 123 rows of `tokens_table.go` and the 2,722-record pool:

| | |
|---|---|
| Distinct identity buckets in the pool | **714** |
| Largest bucket | **86** (Treasure) |
| Table rows that resolve | **100 of 123** |
| …to exactly one candidate | 20 |
| …to more than one, so decision 4 decides | 80 |
| Rows the keyword pass narrows | 20 |
| Rows where no candidate matches keywords exactly | 2 |
| Table rows that resolve to **nothing** | **23** |

**The 23 misses are a `tokens_table.go` bug, not a resolver failure**, and
finding them is the most useful thing this work does before it produces a single
picture. 21 of them are rows whose `Colors` is empty where every printed version
of that token is coloured — `1/1 colorless Soldier`, `1/1 colorless Human`,
`2/2 colorless Zombie`, `3/3 colorless Beast`, `4/4 colorless Angel with flying
and vigilance` and sixteen more. The other two are sharper:

- **`3/3 colorless Phyrexian Wurm artifact`** — Wurmcoil Engine's token. The
  printed token card is named **"Wurm"**; the current oracle text of the card
  that makes it says "Phyrexian Wurm". Scryfall's record keeps the printed name.
  So *name* is not always a reliable key: a token card printed before a rename
  keeps the old name forever.
- **`0/0 white Spirit Cleric`** — printed as **`*/*`**, so no numeric P/T can
  match it.

### They are not fixed here — they are [#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127)

**Owner decision, 2026-09-19: leave all 23 templates exactly as they are, ship
art for the 100 that match, and fix the colours separately.** This ADR edits no
template and this work changes no token's characteristics.

The reason is that a colourless template is a **rules** bug that exists today
whether or not tokens ever get art. Colour is a characteristic the engine reads:
a colourless Soldier is not hit by "destroy target white creature", is not
pumped by a white lord, dodges protection from white and colour-based cost
reduction, and counts wrongly for devotion and anything else reading `Colors`.
Correcting 21 of them changes what those tokens **are**, and that review belongs
in a PR whose subject is the rules and whose reviewer is looking at each calling
card — not in a PR about pictures. It is filed as
[#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127), which carries
the full list, the two non-colour cases and the per-card review it needs.

Two consequences for this ADR, both good:

1. **Until #1127 lands, those 23 templates render as text, and that is expected
   rather than a failure.** Decision 8's test treats them as a named, listed set
   (see there); the suite is green while #1127 is open.
2. **When #1127 lands, their art resolves under this rule with no further work
   here.** A corrected `Colors` is exactly what the identity match in 3b needs —
   a white Soldier lands in the 69-printing white Soldier bucket and decision 4
   picks from it. No resolver change, no new override, no follow-up PR in this
   ADR's scope. The two non-colour cases (`Wurm`, `*/*` Spirit Cleric) are
   decided in #1127; if it chooses to keep either name as-is, that template
   stays on the text fallback permanently and takes a pinned id through the
   override seam in decision 6 whenever somebody wants it to.

Neither case wants the *rule* bent to reach them, which is why the rule stays
strict.

## Decision 4 — The tie-break is the **lowest Scryfall UUID** among the best candidates

Deterministic given a dump, needs no new index field, and — the reason it beats
both obvious alternatives — it does not systematically prefer old art or new.

- **Newest printing** (`released_at` descending) is the intuitive answer and the
  worst on stability. Treasure has 86 eligible printings and picks up more every
  few months; "newest" means the Treasure art changes with almost every set. It
  also needs `released_at` on `cards.Card`.
- **Oldest printing** is stable by construction, needs the same new field, and
  systematically selects the least attractive art in the pool — the 2003
  printing, every time.
- **Lowest UUID** is stable *in expectation*: Scryfall ids are random, so a
  refresh that adds one printing to an N-printing bucket has a `1/(N+1)` chance
  of displacing the pick. Across 123 templates and a handful of new printings
  each per year, that is a small number of tokens changing art per year, not a
  small number per set.

**The trade-off the owner accepted, stated plainly:** a rule that reads a
refreshing dump can change its mind. This is the price of decision 1's runtime
resolution versus pinning ids in the repo, and it is paid in exchange for the
override seam (decision 6) that the user-selectable-art feature needs.

**Stability, precisely scoped.** Three levels, and the useful one is the first:

- **Per token instance: permanent.** Decision 2 stamps the id on the object. A
  token that exists keeps its art until it stops existing — across frames, undo,
  restore points, replays, deploys and dump refreshes.
- **Per process: total.** The pool and the per-template answer are computed from
  the index the process loaded and memoised (decision 5). Two tokens from the
  same template made in the same process are always the same art.
- **Across processes: best-effort.** Only a dump refresh *and* a restart can
  change a template's answer, and only for tokens made after it.

A restored game's snapshot already carries `ScryfallID`, so this is not a
migration concern: old restore points hold tokens with no art, they come back
with no art, and they render exactly as they render today.

## Decision 5 — Cost: one pass at load, one bucket map, one memo

A naive scan of 2,722 records per token creation is not free, and a board with
an Academy Manufactor makes three tokens per trigger.

- **At `Index.Load`**, in the pass that already walks every record: build
  `map[identityKey][]uuid.UUID` — the 714 buckets of decision 3d, values sorted
  by the UUID string so decision 4's answer is the head. Swapped in under
  `i.mu` with `byID` / `byName` / `byOracle`. Memory is one map of 714 entries
  and 2,722 UUIDs; the index already holds 40,000 full `Card` structs.
- **At resolve time**, a memo keyed by the template's identity key, behind an
  `RWMutex`, holding the final answer. The keyspace is bounded by the number of
  distinct templates the catalog can produce — 123 table rows plus ~16 others —
  so it is a few hundred entries and never evicted.
- **Steady state is one `RLock` and one map read per token creation.** The
  keyword pass runs once per template, on the first token of that kind in the
  process, over a bucket whose median size is 3 and whose maximum is 86.
- **A negative result is memoised too.** The 23 misses must not re-scan their
  bucket on every Treasure-adjacent trigger for the rest of the process.

## Decision 6 — The resolver is a hook variable, and the override sits in front of it

`internal/game` does not import `internal/cards`, and `internal/cards/effects`
does not either. The engine reads the catalog through the exported function
variables in [`effect_hooks.go`](../../server/internal/game/effect_hooks.go),
which is the dependency inversion that breaks the cycle. This takes the same
shape:

```go
// game/effect_hooks.go
//
// TokenArtResolver returns a Scryfall PRINTING id for a token about
// to be created, or "" for "no art — render the name".  Nil means no
// resolver is wired, which is the state of every game-package test and
// of a server built without the cards index; tokens then behave
// exactly as they did before ADR 0078.
var TokenArtResolver func(req TokenArtRequest) string

type TokenArtRequest struct {
    Template   Card      // name, type line, P/T, colors, keywords
    Controller uuid.UUID // whose token this is
    GameID     uuid.UUID // which table
}
```

`main.go` assigns it from a small `internal/cards/tokenart` package built over
the loaded `*cards.Index`, the same way `main.go` blank-imports `effects` today.

**`Controller` and `GameID` are on the request from the first commit and are
unused by the rule.** They are the whole of the seam. The later
user-selectable-token-art feature is a lookup that runs *before* the rule and
falls through to it:

1. A per-game override for this template, set by the host — the table agrees a
   Treasure looks like *that* one.
2. A per-player override for this template, from the `users`-scoped store
   [ADR 0051](0051-user-database.md) built.
3. The rule.

The repo's own escape hatch is the same mechanism at a fourth, lowest
precedence: an optional pinned id on a `tokens_table.go` row, consulted before
the rule for a template the rule cannot reach. **Nothing here builds it**, and
it is not what decision 3d's 23 misses use — 21 of those are a colour bug that
[#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127) fixes at the
source, after which the rule reaches them unaided. A pinned id is for the
residue: a template #1127 decides to leave mismatched, such as the Wurmcoil
`Wurm` rename, if somebody later wants art on it.

## Decision 7 — `CardView` gains `is_token`

```go
IsToken bool `json:"is_token,omitempty"`
```

Set in `viewOfCard` from `c.IsToken()`
([`card.go:1012`](../../server/internal/game/card.go)), which is
`typeLineHas(c.TypeLine, "token")` — the **printed** type line, not the
effective one, and the same predicate the CR 704.5d state-based action and
`effects.IsToken` already use. **Today that is the only signal there is**: a
token's token-ness lives entirely in the string `"Token "` at the front of a
type line stamped by `tokens_table.go` and by `TokenCopyTemplate`
([`token_copy.go:146`](../../server/internal/cards/effects/token_copy.go)).
When [#521](https://github.com/krakenhavoc/cmd_and_ctrl/issues/521) gives tokens
a synthetic catalog key there will be a second, better signal; `is_token` is
projected from one predicate, so swapping what that predicate reads is a
one-line change in the engine and no change on the wire.

**Public, and it survives the redaction.** A token is distinguishable from a
card at a paper table, CR 111.8 and CR 704.5d already make its token-ness
everyone's business, and the client needs it to style a token as a token
regardless of whether the viewer knows the face. So it goes on the public side
of the `known` early-return in `viewOfCard` and must be added explicitly to the
allowlist in
[`face_down_view_test.go`](../../server/internal/protocol/face_down_view_test.go),
which pins that list. A face-down token still shows the card back: `Card.svelte`
tests `face_down` before `imgSrc`, and `scryfall_id` is redacted for a
non-knower anyway.

`docs/protocol.md`'s **CardView** paragraph gains `is_token` and a sentence
saying that a token's `scryfall_id` is a *token printing*, chosen by the server,
and is not a card the player could own.

## Decision 8 — A token that matches nothing keeps today's text fallback

`ScryfallID` stays empty, `cardImageURL` returns `null`, `Card.svelte` renders
`.name-fallback`. Byte-identical to today. No placeholder art, no "?" card, no
error toast: the failure is invisible to players and that is correct, because it
is a *repo data* problem and not a player problem.

It is not invisible to us:

- **Logged once per distinct template key per process**, at `WARN`, naming the
  key and the four identity fields. Once per key, not once per creation — a
  Treasure-heavy game would otherwise write a line per trigger.
- **Counted** on the resolver, as a plain expvar-style counter beside the
  memo, so "how many token creations rendered as text today" is answerable
  without grepping logs.
- **Tested against an expected-miss list, not against zero misses.**
  `TestEveryTokenTemplateResolvesToAPrinting` walks `effects.TokenKeys()` and
  the behaviour-token constructors, resolves each, and compares the set of
  template keys that resolved to nothing against a checked-in
  `knownUnresolvedTokens` list — the 23 keys of decision 3d, each with a
  one-line comment naming
  [#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127) and why that
  row misses. It fails on a **set difference in either direction**, and reports
  both halves by key rather than failing on the first:
  - a template **not** on the list that stops resolving — a new row added
    without checking it, or a printing that left the dump — is a failure, which
    is the guard this test exists to be;
  - a template **on** the list that starts resolving is also a failure, with the
    message "remove it from `knownUnresolvedTokens`" — so #1127's fix cannot
    land and leave a stale exemption behind.

  The list shrinks to empty as #1127 corrects templates and is deleted with the
  last entry. **The suite is green the whole time #1127 is open**, and a
  regression is still caught the day it happens.

  It lives in a `*_manual_test.go` gated on `CMDCTRL_SCRYFALL_DUMP`, the
  convention
  [`realdump_manual_test.go`](../../server/internal/deck/realdump_manual_test.go)
  set, because the 510 MB dump is not in the repo and ordinary CI has no copy.
  **`e2e-nightly.yml` does have one** — it caches `data/scryfall` by date and
  runs `scripts/scryfall-refresh.sh` — so the test runs nightly at 08:30 UTC,
  three and a half hours after the Sunday 05:00 UTC refresh. A template that
  goes dark because a printing left the dump is caught within a day, not at the
  next playtest.
  A hermetic unit test over a small hand-written pool pins the *rule*; the
  manual test pins that the rule still reaches the real data.

**No template is edited to make this test pass.** The 23 known misses go on the
list as data, `tokens_table.go` is untouched by this ADR's work, and
[#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127) empties the
list on its own schedule. That is the difference between a guard and a
permanently-failing reminder, and it is what lets the art ship for the 100
templates that already match without waiting on a per-card rules review of the
other 23.

## Decision 9 — `GET /catalog/image/{id}` does not change

It refuses token printings today because `IsCatalogPrinting`
([`catalog.go:421`](../../server/internal/catalog/catalog.go)) demands that the
id be the *representative printing of an oracle id the catalog registers*, and a
token printing's oracle id is a token's, which `effects.Has` does not know.

It keeps refusing them, and the route's own doc comment says why: it is
unauthenticated, and *what makes that safe is the scope, not the content* — "a
few hundred UUIDs, fixed at build time and already published by `GET /catalog`".
Admitting 2,722 token printings multiplies the reachable set by roughly ten and
turns a bounded showcase into a general-purpose token-art proxy that fetches and
caches arbitrary art on demand for anyone who asks. It also buys nothing: the
catalog page renders the *cards* the engine automates, and a Treasure is not one
of them.

The in-game route `/cards/{id}/image` is session-gated behind
`auth.Middleware` and already accepts any id in the index, tokens included. That
is the route every token in every game uses, and it needs no change at all.

If the catalog page later wants to show what a card's token looks like, the
shape is to widen the scope predicate to "a representative printing of a
registered oracle id, **or** a printing this build's resolver actually picks" —
still a fixed, enumerable set of at most one id per template, publishable from
`GET /catalog` exactly as the current set is. Named here so it is not
reinvented; not built.

## Decision 10 — Double-faced tokens are out of scope

The dump holds **119** `"layout":"double_faced_token"` records, all English.
Three reasons, and the third is why this needs saying out loud rather than
merely going unbuilt:

1. **Nothing in the repo makes one.** No row of `tokens_table.go` and no
   behaviour constructor produces a two-faced token.
2. **Their art is not where the resolver looks.** **0 of the 119** carry a
   top-level `image_uris`; all 119 carry one per entry in `card_faces`. Reaching
   it means stamping `Layout` and `Faces` on the token as well as `ScryfallID`,
   which drags ADR 0034's whole face apparatus — `ActiveFace`, `CastableFaces`,
   `faceOnResolve`, the `?face=N` URL — onto an object that cannot be cast and
   whose transform trigger nobody has coded.
3. **It would half-work by accident.** `viewOfFaces`
   ([`view.go:5226`](../../server/internal/protocol/view.go)) already builds
   `/cards/{ScryfallID}/image?face=N` for any card with two or more `Faces`. The
   moment a token carried both a stamped `ScryfallID` and a `Faces` array it
   would start shipping face URLs with no code written and no decision taken.
   Being explicit is what stops that happening by omission.

**Interaction with ADR 0034**, for whoever picks this up: a **token copy** of a
double-faced card is a different thing and already works the way it should —
`TokenCopyTemplate` copies the original's `ScryfallID`
([`token_copy.go:131`](../../server/internal/cards/effects/token_copy.go)) per
CR 707.2, so the copy renders the copied card's art through the ordinary path.
Only *printed* double-faced tokens are excluded. When one is wanted, the
resolver changes very little: the identity key is built from the **front face's**
name, type line, P/T and colours, and `Faces` is populated from `card_faces`
exactly as `deck.ToGameCard` does for a real DFC. The resolver is not the
blocker; the engine's transform path for an uncastable object is.

## Decision 11 — This ships before #521, and they are separable

[#521](https://github.com/krakenhavoc/cmd_and_ctrl/issues/521) gives token
templates a **synthetic catalog key** (`token:treasure`, …) so
`CatalogManaAbilities` / `CatalogActivatedAbilities` can re-derive a token's
abilities after a restore. It exists because a live Treasure makes
`ContinuationCensus.IntrinsicAbilityCards` non-zero and **blocks every restore
point** while it is on the battlefield (`snapshot.go:721`). That is a
**persistence** problem about un-rebuildable Go closures.

This ADR is an **art** problem about an empty string on the wire. The two touch
the same objects and share no code path: #521 writes an *oracle-shaped key* and
reads the ability registry; this writes a *printing id* and reads the image
index. Decision 1's "stamp `ScryfallID`, leave `OracleID` empty" is what keeps
them apart — neither the synthetic key namespace nor the census sees a stamped
art id, because both key on `OracleID`.

When #521 lands, the natural home for a *pinned* art id is its registry entry,
beside the token's abilities — registry data rather than a table column. Nothing
here forecloses that: the override fall-through in decision 6 gains a fourth
source and the rule stays the floor.

Order matters only in one direction: shipping art first means #521's registry
arrives with a place to put an id it already knows it wants. Shipping #521 first
would not have made this easier.

## Two predicates that read `ScryfallID != ""`

Not in the survey, and both are hit by decision 1. Neither is a reason not to do
this; both are two-line changes that must be made *in the same commit* as the
stamping.

**`Card.ToughnessIsKnown()`
([`card.go:911`](../../server/internal/game/card.go))** ends:

```go
return c.LostLastCounter || c.ScryfallID != ""
```

Its own doc comment names the set this protects — *"test fixtures that left the
body at 0, and 0/0 token templates"* — and reads `ScryfallID` as "a printing
stands behind this object, so its 0 is a real printed 0". Stamping a token flips
that. `tokens_table.go` has genuine 0/0 rows (`0/0 colorless Construct
artifact`, `0/0 white Spirit Cleric`), and after this change a 0/0 Construct
with no counters would become a permanent the CR 704.5f state-based action
sweeps into the graveyard. **Fix:** make branch 5 `c.ScryfallID != "" &&
!c.IsToken()`, keeping the skip exactly where the comment says it belongs. The
rules question of whether a 0/0 token *should* die is real and is **not settled
here** — it is a separate change to a separate decision, and folding it into an
art PR would be the worst way to make it.

**[#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127) does not
change this requirement, and does not substitute for it.** #1127 corrects
`Colors`, which is a different field on a different set of templates; the two
0/0 rows this paragraph names are not colour bugs, and `&& !c.IsToken()` is
needed the moment *any* token carries a stamped `ScryfallID` — which is sub-PR
2, regardless of what #1127 has or has not done by then. Both predicates are
fixed in the same commit as the stamping, as stated.

**`Card.fromScryfallPrinting()`
([`mutations.go:5659`](../../server/internal/game/mutations.go))**, read by
`printedIdentityOf` for commander colour identity, asks "was this instance
stamped from a real Scryfall record, or conjured — the demo seed, **a token
template**, a test fixture". After this, a token answers yes. The consequence is
narrow (a colourless token would report `known: true` instead of `known: false`,
and tokens are not commanders) and the safe change is the same: `&&
!c.IsToken()`, so the predicate keeps meaning what its comment says.

Both are exactly the class of thing a `snapshot_drift_test.go`-style guard
cannot catch, because no field is being *added* — an existing field is gaining a
population it never had. The unit tests for them go in the same commit.

## Attribution and Fan Content policy — an unaddressed housekeeping item

The repo carries **no Scryfall attribution and no Wizards Fan Content Policy
notice**, anywhere: not in `README.md`, not in the client, not on the public
`/catalog` page. That is a gap today and this ADR records it rather than fixing
it.

**Serving token art does not change the posture.** The same
`GET /cards/{id}/image` route, the same `ImageCache`, the same Scryfall CDN and
the same on-disk store already serve art for roughly 35,000 real printings to
the same private playgroup. Token printings are more records of the same kind
from the same source through the same code. Nothing about the nature of what is
served changes; only the count of ids does.

**Recommended, not decided here, and not legal advice:** a short notice naming
Scryfall as the source of the card data and images, and the project as
unofficial Fan Content, belongs in two places — `README.md`, where a reader
lands first, and a footer or About panel in the client, where a player does.
The public `/catalog` page is the one surface an unauthenticated stranger can
reach, so if only one place is done, do that one. It is a small chore, it is
independent of every decision above, and it should be its own issue rather than
a checkbox on this one.

---

## Sub-PRs

One PR each, in this order. Sub-PRs 2 and 3 are independent and can run in
parallel.

| # | Scope | Depends on |
|---|---|---|
| 0 | This ADR, the AGENTS.md §3 range line, and the `docs/sprints.md` S35 section | — |
| 1 | `cards.Card` gains `Colors` and `Keywords`; `Index.Load` builds the eligible pool and the 714 identity buckets; `cards/tokenart` with the rule, the memo and the counter; unit tests over a hermetic fixture pool | 0 |
| 2 | `game.TokenArtResolver` + `TokenArtRequest` in `effect_hooks.go`; `main.go` wires it; stamping in `token_create.go`; `ToughnessIsKnown` and `fromScryfallPrinting` gain `&& !c.IsToken()`, with tests | 1 |
| 3 | `CardView.is_token`, `viewOfCard`, the `face_down_view_test.go` allowlist, `client/src/lib/protocol.ts`, `docs/protocol.md`; the client styles a token as a token | 0 |
| 4 | `TestEveryTokenTemplateResolvesToAPrinting` with its `knownUnresolvedTokens` list in a `*_manual_test.go`, and its wiring into `e2e-nightly.yml`. **No template edits** | 2 |
| 5 | Delete the stale comment at `tokens.go:10-13`; a soak-game screenshot pass confirming Treasure, Food, Clue, Blood, Gold, Eldrazi Spawn and the common creature tokens all render art | 2, 3, 4 |

**No sub-PR here touches `tokens_table.go`.** The 23 templates that resolve to
nothing keep today's text fallback and are corrected under
[#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127), which is a
rules-correctness bug with its own per-card review and its own sprint slot.
#1127 and this work are independent in both directions: art ships for 100
templates without it, and when it lands the remaining art appears with no PR in
this ADR's scope. Sub-PR 2 is the one to review carefully — it is the one that
changes engine predicates.

## Consequences

- Every token a player makes gets a picture, through code that already existed;
  the client's image path, the on-disk image cache and the service worker are
  untouched.
- The wire gains one boolean and tokens gain one previously-empty string. No
  protocol version bump: a pre-0078 client ignores `is_token` and renders a
  token's `scryfall_id` as art without being told it is a token.
- `snapshot_drift_test.go` needs no new row, because no `Card` field is added.
  Token art survives capture, restore, undo and replay for free.
- A restored game keeps the art it had. A dump refresh never rewrites history —
  only what the *next* token looks like, and only after a restart.
- Two engine predicates stop meaning "conjured object" and have to be told so
  explicitly. That is a small, permanent tax on reusing `ScryfallID`, and it is
  cheaper than a second id field.
- **23 templates keep the text fallback until
  [#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127) lands**, so
  the board is mixed for a while: most tokens have art, a colourless Soldier
  does not. Accepted deliberately — the alternative was holding every token's
  art behind a per-card rules review.
- **No token's characteristics change in this work.** The 21 colourless
  templates are a rules bug this work *found* and handed to #1127; correcting
  them there makes their art appear under this ADR's rule with no further change
  here.
- The catalog page gains nothing and its unauthenticated surface does not grow.
- CI does not verify that every template resolves; the nightly e2e run does,
  because only it has the dump. A regression is caught within a day, not within
  a merge.

## Deferred

- **User-selectable token art.** The feature decision 1 exists for. The request
  struct carries `Controller` and `GameID` from day one; the store, the picker
  and the precedence between a per-game and a per-player choice are all
  unspecified. Build it on top of ADR 0051's `users` tables, not beside them.
- **Double-faced tokens** (decision 10) — blocked on an engine transform path
  for an uncastable object, not on the resolver.
- **Pinned ids as registry data** — the natural home once
  [#521](https://github.com/krakenhavoc/cmd_and_ctrl/issues/521) lands.
- **Widening `/catalog/image/{id}`** to the resolver's chosen ids (decision 9),
  if the catalog page ever renders tokens.
- **The Scryfall / Fan Content notice** — its own issue, independent of all of
  this.
- **Whether a 0/0 token with no counters should die.** Exposed by the
  `ToughnessIsKnown` change and deliberately left exactly as it is today.
- **The 21 colourless token templates**, and the two other mismatches — handed
  to [#1127](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1127) as a
  rules-correctness bug in its own right. Deferred *out of* this work rather
  than *by* it: #1127 can land before, after or alongside, and the only coupling
  is that `knownUnresolvedTokens` shrinks as it goes.
- **Art for emblems** (ADR 0064) and for the face-down 2/2 (ADR 0069). Both are
  objects with no printing and neither is a token; neither is touched here.

## Alternatives considered

**Pin a Scryfall id per token in the repo.** A column on each `tokens_table.go`
row. It is stable by construction — the id changes when somebody edits the file
and never otherwise — it needs no pool, no bucket map, no memo, no `Colors` or
`Keywords` on `cards.Card`, and no resolver at all. The 23 rows that resolve to
nothing would simply be 23 rows somebody looked at once. **This was the
recommendation.**

The owner chose runtime resolution anyway, and the reason is decision 6: *"in
the future we will add a feature that lets users select token images"*. A pinned
column is a constant, and a constant is a poor thing to put an override in front
of — the override would end up being the only real resolver, arriving later,
with the pinned value as a default nobody had designed a fall-through for.
Building the seam first and the rule behind it means the override slots in at
one call site. The cost is everything in decisions 3 to 5, and a pick that can
move when the dump refreshes. Recorded honestly: the recommendation was pinning,
the decision is the rule, and the reason is a feature that does not exist yet.

**Common tokens only** — art for the dozen that actually show up (Treasure,
Food, Clue, Blood, Gold, Soldier, Zombie, Spirit, Goblin, Saproling, Eldrazi
Spawn, Beast) and text for the rest. Roughly a tenth of the work for most of the
visible benefit, and it is the version that ships in an afternoon. Not chosen
for two reasons. The tail is not rare — 123 templates exist because 123 of them
came up, and a table where the Treasure has art and the Rhino does not looks
*more* broken than one where neither does, because the inconsistency reads as a
bug rather than as a missing feature. And the work it saves is the work that
finds bugs: the 23-row audit in decision 3d only exists because the rule was run
over every row.

**A new `Card.ArtScryfallID` field** rather than reusing `ScryfallID`. Keeps
`ScryfallID` meaning exactly "a printing this object *is*", so neither predicate
in the section above needs touching. Rejected on cost: a new field must be
classified in `snapshot_drift_test.go`, serialised, projected onto the wire as a
second image id, and taught to `cardImageURL`, `viewOfFaces` and the service
worker — five places changed to avoid changing two.

**Resolve in `viewOfCard`, never persist.** Rejected in decision 2: art would be
recomputed every frame from whatever dump the process holds, restored games
would come back looking different, and replays would not reproduce what the
players saw.

**Newest printing as the tie-break.** Rejected in decision 4: the most visible
tokens are the ones reprinted most often, so "newest" changes exactly the art
players see most, roughly every set.

**Match on `oracle_id` instead of on characteristics.** Scryfall gives token
records oracle ids, and several printings of the same token share one — 69 of
the 91 Soldier records share `eac25f12-…`. It is a cleaner key than five fields
compared by hand. Rejected because the engine has no way to *get* one: a
template in `tokens_table.go` is a name, a type line, P/T and colours, and
picking an oracle id for it is the same matching problem one level up, plus an
oracle id to pin. If templates ever carry an id, they should carry a printing id
and skip the rule entirely — which is the pinning alternative above.
