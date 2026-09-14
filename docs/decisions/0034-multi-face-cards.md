# ADR 0034 — Multi-face cards: MDFC, transform, adventure

**Status:** accepted — the spine and the MDFC picker shipped together. See "What shipped" below.
**Motivated by:** [#278](https://github.com/krakenhavoc/cmd_and_ctrl/issues/278) (this spike), [#265](https://github.com/krakenhavoc/cmd_and_ctrl/issues/265) (Sea Gate Restoration entered as a land with no prompt), [#276](https://github.com/krakenhavoc/cmd_and_ctrl/issues/276) (transform commander has no colour identity).
**Depends on:** [ADR 0010](0010-card-effect-catalog.md) (oracle-ID-keyed specs), [ADR 0012](0012-layer-system.md) (`Effective()`), [ADR 0013](0013-replacement-effects.md) (the entry-choice pause), [ADR 0022](0022-impulse-exile.md) (`ExilePlay`).

Every structural claim below was checked against the tree at `f26c961`
and against the Scryfall dump at `data/scryfall/default-cards.json`
(refreshed 2026-09-09). Line numbers are from that commit.

---

## Context

### What is actually broken

`game.Card` has one `Name`, one `TypeLine`, one `ManaCost`
([card.go:14-251](../../server/internal/game/card.go)). `cards.CardFace`
carries `Name`, `TypeLine`, `OracleText`, `ImageURIs`, `Loyalty` and
nothing else — **no `ManaCost`, no `Colors`, no `Power`/`Toughness`**
([index.go:124-135](../../server/internal/cards/index.go)). And
`deck.toGameCard` copies Scryfall's **top-level** fields only
([deck.go:273-298](../../server/internal/deck/deck.go)).

Scryfall ships two structurally different shapes under the umbrella
"multi-face", and conflating them is why this looked like one problem:

| | `transform`, `modal_dfc` | `adventure`, `split`, `prepare` |
|---|---|---|
| top-level `mana_cost` | **`null`** | **`"{4}{U} // {1}{U}"`** |
| top-level `colors` | `null` | populated |
| top-level `image_uris` | absent | present |
| top-level `type_line` | `"Sorcery // Land"` | `"Enchantment // Instant — Adventure"` |
| face `mana_cost` | present | present |
| face `colors` | present | `null` |
| face `image_uris` | present | absent |

Both shapes are wrong on arrival, in two different ways.

**Back-to-back DFCs arrive costless and colourless.** Replaying
`cards.Index.Load`'s exact insert order over the dump, **1,503 of the
40,070 `byName` keys** resolve to a printing whose top-level
`mana_cost` is `null` (1,203 `transform`, 300 `modal_dfc`). Every one
of those imports as `game.Card{ManaCost: "", Colors: nil}`. `ParseCost("")`
succeeds and yields the **zero cost**
([mana_cost.go:82](../../server/internal/game/mana_cost.go)), so
`applyCastCostLocked` charges nothing and emits no warning
([mutations.go:839-870](../../server/internal/game/mutations.go)).
Every transform and modal DFC in the game is **castable for free,
silently.** #276 reports one downstream symptom of this (commander
identity); the free casts are not in any issue.

**Split-style cards arrive with an unparseable cost.** **1,047 `byName`
keys** resolve to a printing whose top-level `mana_cost` contains
`" // "` (350/350 split printings, 443/459 adventure, 99/99 prepare).
`ParseCost` rejects the bare `/`
([mana_cost.go:90-91](../../server/internal/game/mana_cost.go)) and
`applyCastCostLocked` **treats an unparseable cost as free**
([mutations.go:842-852](../../server/internal/game/mutations.go)), with
an `EventCostWarning` nobody is reading. `legal/cast.go:91` mirrors the
same posture for the bot. Also not in any issue.

**#265 is a type-line substring collision.** There is no `PlayLand`
verb; the land drop is a branch inside `CastSpell` gated on
`card.IsLand()`
([mutations.go:604](../../server/internal/game/mutations.go)), and
`IsLand()` is a case-insensitive substring search for `"land"` over
the whole type line
([card.go:353-355](../../server/internal/game/card.go),
`typeLineHas` [card.go:410](../../server/internal/game/card.go)).
Sea Gate Restoration's imported `TypeLine` is the literal string
`"Sorcery // Land"`. `IsLand()` returns true, the land branch runs,
and the card goes straight to the battlefield as a 7-mana sorcery
wearing a land's clothes. `IsSorcery()` also returns true. Exactly
the report.

**The type parser produces junk types on every DFC.**
`ParseTypeLine` splits on the first em-dash and whitespace-tokenises
the rest ([characteristic.go:159-182](../../server/internal/game/characteristic.go)).
`"Sorcery // Land"` yields `Types: ["Sorcery", "//", "Land"]`. Aang's
`"Legendary Creature — Human Avatar Ally // Legendary Creature — Avatar
Spirit Ally"` yields seven bogus subtypes including the literal `"//"`
and `"—"`. Every subtype-matching effect is quietly wrong on all 401
transform and 100 modal_dfc oracle IDs.

### What is not broken, and should be left alone

Deck **resolution and validation are already face-aware and correct.**
`Index.Load` indexes the composite name *and* each face name, with a
guard that stops an art-series `"X // X"` record hijacking the real
`"X"` entry ([index.go:228-240](../../server/internal/cards/index.go),
mirrored in `Put` at [index.go:405-423](../../server/internal/cards/index.go)).
`FindByName` additionally falls back across a single-slash separator
([index.go:280-291](../../server/internal/cards/index.go)).
`isLegalCommander` walks every face
([validate.go:221-249](../../server/internal/deck/validate.go)).
And the colour-identity check reads Scryfall's **top-level**
`color_identity` ([validate.go:184-191](../../server/internal/deck/validate.go)),
which is correct per CR 903.4c for every layout — verified directly:
Aang, Swift Savior // Aang and La has `color_identity: ["U","W"]` at
the top level even though `colors` is `null`.

So the seam is narrow. **`toGameCard` is the only import-path function
that has to change.** Everything upstream of it already has the data.

### Three corrections to the record

1. **`class` is not a multi-face layout.** All 73 `class` printings in
   the dump have **no `card_faces`**. Fortune Teller's Talent is
   `layout: class`, `mana_cost: "{U}"`, single-faced. The Aang triage's
   "Multi-face cards (7)" is **6**, and #278's list of face-carrying
   layouts should drop `class`.
2. **`room` does not exist in our data.** Zero printings, any set.
   Out of scope by absence, not by judgement — no sizing required.
3. **#276's line reference has drifted.** The stale comment it cites at
   `mutations.go:2926` is now at
   [mutations.go:3053-3058](../../server/internal/game/mutations.go);
   line 2926 is an unrelated `ErrorMsg: err.Error()`.

### Cards at stake

| Layout | In Aang list | In Hashaton list | Oracle IDs in all of Magic |
|---|---|---|---|
| `modal_dfc` | Sea Gate Restoration, Sink into Stupor | — | 100 (**60 with a land back**) |
| `transform` | Aang Swift Savior, The Legend of Kuruk | Jace Vryn's Prodigy, Rona Herald of Invasion | 401 |
| `adventure` | Virtue of Knowledge | Altar of Bhaal | 170 |
| `prepare` | Skycoach Conductor | — | 57 |
| `split` | — | — | 137 |
| `class` | ~~Fortune Teller's Talent~~ (not multi-face) | — | 38 (no faces) |

**Nine cards across the two tracked decklists**, not ten. One of them
(Aang) is a commander, which is what makes #276 visible.

---

## Decisions

### 1. Representation: `Faces []Face` + `ActiveFace int`, with the existing flat fields as the materialised active face

```go
// server/internal/game/card.go

// Face is one printed side of a multi-face card, in engine terms.
// Values are already parsed (P/T as ints) — the importer does the
// Scryfall string work once.
type Face struct {
	Name            string
	TypeLine        string   // the FACE's type line: "Sorcery", never "Sorcery // Land"
	ManaCost        string
	Colors          []string // face colors, or color_indicator, or derived from cost
	Power           int
	Toughness       int
	StartingLoyalty int
}

type Card struct {
	// ... every existing field unchanged ...

	// Layout is Scryfall's printing layout, copied verbatim at
	// import ("normal", "transform", "modal_dfc", "adventure",
	// "split", "prepare", ...). The cast path branches on it to
	// decide what a face CHOICE means; see SetFace / faceOnResolve.
	Layout string

	// Faces is every printed face, front first. nil for the 33,363
	// single-faced oracle IDs, which keep today's behaviour to the
	// byte.
	Faces []Face

	// ActiveFace indexes Faces. INVARIANT: the flat printed fields
	// above (Name, TypeLine, ManaCost, Colors, Power, Toughness,
	// StartingLoyalty) always equal Faces[ActiveFace]. Maintained
	// only by SetFace; never assigned directly.
	ActiveFace int

	// ColorIdentity is CR 903.4 identity for the WHOLE card — both
	// faces and rules text — copied from Scryfall's top-level
	// color_identity, which already computes it correctly for every
	// layout. NOT the same thing as Faces[0].Colors. See #276.
	ColorIdentity []string
}

// SetFace switches the card to face i and re-materialises the flat
// printed fields from it. The only writer of ActiveFace. No-op for
// single-faced cards. Callers on the battlefield must follow with a
// layer-staleness mark — the face change is a printed-value change.
func (c *Card) SetFace(i int) { /* ~25 lines */ }
```

**Why the flat fields stay fields.** The alternative — turning `Name`,
`TypeLine`, `ManaCost` into methods that read `Faces[ActiveFace]` — is
the "purer" model and it is the wrong trade here. There are **293
`TypeLine:` struct-literal sites** and **49 `ManaCost:`** across
`server/`, almost all in tests, plus 389 `Card{` literals. Every one
would need rewriting. Keeping them as fields makes the change **additive
for all 74 non-test `Is*()` call sites, all 30 `.TypeLine` reads and all
22 `.ManaCost` reads** — they keep compiling and start being *right*,
because "the characteristics of the face that is currently up" is
exactly what CR 711.2 says a transformed permanent has.

**Why not separate `game.Card`s per face.** A transform is not a zone
change. The permanent keeps its counters, its damage, its
`EnteredBattlefieldAt` timestamp (which drives CR 613 ordering,
[card.go:212](../../server/internal/game/card.go)), its
`SummonedThisTurn` flag, its `KnownBy` set, its `InstanceID`. Two Card
values means either swapping (losing all of it) or mirroring (two
sources of truth for one permanent). Rejected.

**Why not front-face-only with a `BackFace *Face` for lands.** It
solves the `modal_dfc`-land subset and nothing else. Transform is 4 of
the 9 blocked cards and the one with a live production bug, and a
`BackFace` pointer has nowhere to record "the back is what is on the
battlefield now". Rejected as a shipping shape — it is what the model
degenerates to if you only fill in face 1 for land backs, which is fine
as a *sequencing* choice but not as a *representation*.

### 2. What `TypeLine` means, and what `Effective()` returns

**`Card.TypeLine` is the active face's own type line, verbatim from
Scryfall's face record.** Never the `"A // B"` concatenation. For
`Faces == nil` it is the top-level line, unchanged.

**`Effective()` needs no change at all**
([characteristic.go:131](../../server/internal/game/characteristic.go)).
It layers over the flat printed fields, which now hold the active
face — so `printedCharacteristic` parses one clean type line, the
`"//"` junk type disappears, and a transformed Jace reports
`Types: ["Planeswalker"]` with the layer engine none the wiser. That is
the whole payoff of decision 1.

One three-line fix rides along: `printedCharacteristic` sets
`Colors: printedColorsFromCost(c.ManaCost)`
([characteristic.go:84](../../server/internal/game/characteristic.go)),
**ignoring `Card.Colors` entirely.** That is why Jace's back face
(cost `""`, `colors: ["U"]`, `color_indicator: ["U"]`) would still read
as colourless. Prefer `c.Colors` when non-empty; fall back to the cost.
This also fixes Devoid and colour indicators generally, which is a
latent bug independent of faces.

### 3. The cast path: one `SetFace` call covers the whole announce sequence

`CastSpell` takes a **value copy** of the card out of the source zone
([mutations.go:412-423](../../server/internal/game/mutations.go)) and
then reads it ten more times before anything moves: `IsLand()` at
:437, `validateAlternativeCost(card.OracleID, …)` at :446,
`TargetModeFor`/`TargetSpecFor` at :466-467, `ModeSpecFor` at :484,
`castTargetSpec` at :495, `AdditionalCostFor` at :552,
`TapPermanentsCostFor` at :569, the sorcery-speed gate at :596-601, the
**land branch at :604**, and `printedCostLocked` at :1063.

So:

```go
// CastSpellParams gains one field.
Face int `json:"face,omitempty"` // which printed face to cast/play; 0 = front

// ... and CastSpell gains one line, right after the card copy at :423:
card.SetFace(params.Face)   // validated against len(card.Faces) and Layout
```

Because `card` is a copy, **every one of those ten downstream reads
picks up the chosen face for free.** #265's land branch stops firing on
Sea Gate's front face, the cost parser gets `{4}{U}{U}{U}` instead of
`""`, and the catalog key (decision 5) resolves to the right half. This
is the single highest-leverage line in the whole model.

**Face choice is an announce-time parameter, not a `PendingChoice`.**
Every other announce decision — alternative cost, modes, X, additional
cost, tap cost, targets — is a client-side prompt whose answer rides
the `cast_spell` action
([targeting.ts:72-81](../../client/src/lib/targeting.ts),
`Board.svelte:174-186` documents the whole chain). The
`PendingChoice` machinery resumes *replacement*, *search* and *trigger*
frames ([pending_choice.go:256-369](../../server/internal/game/pending_choice.go));
it has no frame for a half-validated cast and should not grow one.

### 4. Per-layout semantics: two different meanings for "choose a face"

| Layout | Face chosen at announce? | Permanent's `ActiveFace` after resolution |
|---|---|---|
| `modal_dfc` | **yes** — the faces are independently playable | the chosen face; the other never returns |
| `transform` | no — always face 0 (CR 712.4) | 0, until an effect calls `SetFace(1)` in place |
| `adventure` | **yes** — but only to pick the *spell* | **always 0**; the adventure half exists only on the stack |
| `split`, `prepare` | (deferred) face 0 only | 0 |

The unifying rule is one small function at the two push sites — the
land branch ([mutations.go:637](../../server/internal/game/mutations.go))
and `resolveTopOfStackLocked`'s permanent branch
([mutations.go:1242-1249](../../server/internal/game/mutations.go)):

```go
// faceOnResolve returns the face the PERMANENT keeps, given the face
// that was cast. MDFC keeps what you chose; everything else is 0.
func faceOnResolve(layout string, castFace int) int
```

Adventure additionally reroutes the non-permanent half: the adventure
spell resolves and is **exiled instead of graveyarded**, with a
permission to cast the creature half later. That permission model
already exists — `Card.ExilePlay`
([card.go:239](../../server/internal/game/card.go),
`exile_play.go`), and `castSourceZoneLocked` already accepts
`from_zone: "exile"`
([mutations.go:1113](../../server/internal/game/mutations.go)). The one
gap: an adventure exile has **no end-of-turn expiry**, so
`ExilePlayPermission` needs a "does not expire" flag. The reroute goes
in `routeStackCardToGraveyardLocked`
([mutations.go:1439](../../server/internal/game/mutations.go)).

### 5. Catalog keying: one oracle ID, two faces

`registry` is `map[string]Spec` keyed on the oracle ID, and `Register`
**panics on a duplicate**
([registry.go:15, :26-33, :78, :88](../../server/internal/cards/effects/registry.go)).
Scryfall issues one `oracle_id` per *card*, so the two faces of Sea
Gate Restoration collide onto one spec slot. There are ~25 production
sites that pass a bare `card.OracleID` to a catalog hook, spread across
`mutations.go`, `activated.go`, `keywords.go`, `characteristic.go`,
`effect_hooks.go`, `targets.go`, `view.go` and `legal/cast.go`.

**Decision: a composite key, derived in exactly one place, where face 0
keeps the bare oracle ID.**

```go
// server/internal/game/effect_hooks.go
//
// CatalogKey returns the catalog key for a card's ACTIVE face. Face 0
// keys on the bare oracle ID, so all ~241 registered specs and every
// single-faced card are untouched. A back-face spec registers as
// "<oracle_id>#1".
func CatalogKey(c Card) string {
	if c.ActiveFace == 0 {
		return c.OracleID
	}
	return c.OracleID + "#" + strconv.Itoa(c.ActiveFace)
}
```

`Register`, `Lookup`, `Has` and `Spec` are unchanged — they already take
and hold opaque strings. The work is swapping ~25 call sites from
`c.OracleID` to `CatalogKey(c)`, which is mechanical and reviewable.
Because `CastSpell` has already called `SetFace` by then, announce-time
hooks and battlefield-time hooks both resolve to the correct half with
no further plumbing.

Rejected alternative: `Spec` grows a `BackFace *Spec`. That forces
every one of the 16 `Catalog*` function pointers
(`effects/wire.go:21-217`) to learn about faces at its own call site.

### 6. The wire and the client

```go
// server/internal/protocol/view.go — CardView (currently :502-669)
Layout     string         `json:"layout,omitempty"`
Faces      []CardFaceView `json:"faces,omitempty"`
ActiveFace int            `json:"active_face,omitempty"`

type CardFaceView struct {
	Name     string `json:"name"`
	TypeLine string `json:"type_line,omitempty"`
	ManaCost string `json:"mana_cost,omitempty"`
	Image    string `json:"image,omitempty"` // "/cards/{id}/image?face=N"
}
```

**`CardView.Name` / `.TypeLine` / `.ManaCost` keep meaning "the active
face's".** That is deliberate and it is what makes the client change
small: all twenty-odd client type checks — `cardTypes.ts:17-37`,
`bucketForBattlefield` `cardTypes.ts:51-55`, the duplicated
case-sensitive regexes in `Card.svelte:131-132`, the cast-timing gate
in `timing.ts:153-172`, the mana-source estimator in
`Board.svelte:161`, the impulse-exile verb in `zoneBrowser.logic.ts:110`
— keep working with **zero edits**, because they now receive one clean
type line instead of a concatenation.

`faces` is purely additive: it feeds the picker and the hover overlay.

**Images need a face parameter.** `GET /cards/{id}/image` accepts only
`?size=` ([http.go:46-82](../../server/internal/cards/http.go)) and
`ImageURI` falls back to `CardFaces[0]`
([index.go:428-448](../../server/internal/cards/index.go)). Add
`?face=N`; `ImageCache.Fetch`'s on-disk key must include it. Client side
this is free — the service worker already treats the query string as
part of the cache key
([service-worker.js:58-61](../../client/src/sw/service-worker.js)).

But the client builds that URL **inline in five places with no shared
helper**: `Card.svelte:101-103`, `HoverZoomOverlay.svelte:31-32`,
`StackOverlay.svelte:124-125`, `PlayerIdentity.svelte:77`,
`Game.svelte:1014-1019` (the mulligan grid). **Extract a single
`cardImageURL(card, size)` helper first**, as its own commit, before
anything face-aware lands.

**The face picker** slots into the head of the existing announce chain
at `Board.svelte:346-352` (`handlePlayCard`), ahead of the alternative-cost
prompt. `ModePickerModal.svelte` is the closest structural precedent — a
face pick is `min=1, max=1` over two `{label, mana_cost, type_line,
image}` options — and `AlternativeCostModal.svelte` is the closest
*semantic* one. Either is a fine base; a dedicated `FacePickerModal`
that shows both card images side by side is better UX for the ~150
lines it costs. The chosen index rides `cast_spell` through
`applyCastChoices` ([targeting.ts:72-81](../../client/src/lib/targeting.ts)).

`timing.ts:153-172` is the one client file that genuinely changes
behaviour: an MDFC in hand is legal to play if **either** face is
legal right now, so the gate becomes a union over `faces`.

`cardMetaCache.ts` already types `card_faces` (`:27-32`) and never reads
it; `GET /cards/{id}` already serves the whole `cards.Card` including
`card_faces` and `layout` ([http.go:42](../../server/internal/cards/http.go)),
so the hover overlay's back-face panel is a client-only change.

### 7. Deck import and validation

`toGameCard` is the only function that changes
([deck.go:273-298](../../server/internal/deck/deck.go)): build `Faces`
from `c.CardFaces`, stamp `Layout` and `ColorIdentity`, call
`SetFace(0)`. `printedLoyalty`'s face fallback
([deck.go:261-271](../../server/internal/deck/deck.go)) folds into the
per-face parse and can go.

`cards.CardFace` grows the five fields it is missing
([index.go:124-135](../../server/internal/cards/index.go)):
`ManaCost`, `Colors`, `ColorIndicator`, `Power`, `Toughness`. Face
colours resolve as `colors` → `color_indicator` → derived from
`mana_cost`, because the two Scryfall shapes populate different ones.

`Resolve` ([deck.go:137](../../server/internal/deck/deck.go)) and
`Validate` ([validate.go:104](../../server/internal/deck/validate.go))
need no change. Singleton keys on the composite `Name` consistently;
commander legality already walks faces; colour identity already reads
the correct top-level value.

**Two additions.**

`isPlayablePrint` ([index.go:321-331](../../server/internal/cards/index.go))
should exclude `reversible_card` alongside `token` / `double_faced_token`
/ `art_series`. All 81 such printings have `mana_cost: null` and
`type_line: null`, and **71 `byName` keys currently resolve to one**.
The face-insert guard keeps the real single-faced entries safe today —
they are all `"X // X"` doubled keys — but the class of record belongs
with the other placeholders, and it is one line.

**The honest fallback for out-of-scope layouts: import face 0 and say
so.** Add a non-fatal `Violation{Code: CodeUnsupportedLayout}`, the same
shape as the existing sideboard warning
([validate.go:196-202](../../server/internal/deck/validate.go)), naming
the cards and the simplification. Refusing the import would reject a
whole deck over a card that is merely cosmetically wrong; silence is
what produced #265. A yellow banner is the correct middle.

---

## Sizing

All estimates are production lines, excluding tests, which for this
codebase historically run 1–1.5×.

### Layer 0 — the shared spine (required by every layout)

| Change | file:line | ~lines |
|---|---|---|
| `cards.CardFace` +5 fields | [index.go:124](../../server/internal/cards/index.go) | 15 |
| `reversible_card` filter | [index.go:321](../../server/internal/cards/index.go) | 1 |
| `game.Face`, `Card.{Layout,Faces,ActiveFace,ColorIdentity}`, `SetFace` | [card.go:14](../../server/internal/game/card.go) | 80 |
| `printedCharacteristic` prefers `Card.Colors` | [characteristic.go:84](../../server/internal/game/characteristic.go) | 5 |
| `toGameCard` builds faces | [deck.go:273](../../server/internal/deck/deck.go) | 45 |
| `CatalogKey` + ~25 call-site swaps | `effect_hooks.go` + 8 files | 40 |
| `CardFaceView`, `viewOfCard`, redaction | [view.go:502/1626/1695](../../server/internal/protocol/view.go) | 55 |
| `?face=N` on the image route + cache key | [http.go:46](../../server/internal/cards/http.go), [images.go:81](../../server/internal/cards/images.go), [index.go:428](../../server/internal/cards/index.go) | 40 |
| `protocol.ts` types; `cardMetaCache` `layout` | [protocol.ts:582](../../client/src/lib/protocol.ts), `cardMetaCache.ts:17` | 30 |
| `cardImageURL` helper + 5 call sites | `Card.svelte:101`, +4 | 40 |

**~350 lines. 2–3 days.** Ships **no new gameplay** and turns on **no
cards** — but it fixes, on its own: the 1,503 free-casting DFCs, the
1,047 unparseable-cost-is-free split/adventure/prepare cards, the `"//"`
junk types on every DFC, and `IsLand()` on Sea Gate Restoration.

**It also breaks something, deliberately.** After the spine, Sea Gate
Restoration is a 7-mana sorcery that can no longer be played as a land,
because nothing yet lets the player choose the back face. That is
correct but *less playable* than today. The spine must ship **in the
same sprint as** the MDFC picker, or behind a flag.

### Per layout, on top of the spine

**MDFC — ~250 lines + a generated cycle file. ~2 days. Unlocks 2 cards
in the Aang list, 100 oracle IDs (60 land-backed) overall.**

`CastSpellParams.Face` + `SetFace` (10) · `actions.go` param passthrough
(5) · `faceOnResolve` at the two push sites (20) ·
`legal/cast.go` enumerating both faces (25) · `FacePickerModal.svelte`
(~150) · `Board.svelte:346` chain + `targeting.ts` param (20) ·
`timing.ts` union-over-faces (15).

The back face's *rules* cost nothing. "As this land enters, you may pay
3 life. If you don't, it enters tapped" is **byte-identical** to the
shockland clause, and
`EntersTappedUnlessYouPayLife(name, 3)`
([shocklands.go:75](../../server/internal/cards/effects/shocklands.go))
already implements it end to end — `EntryLifeCost`
→ `offerEntryLifePaymentLocked`
([entry_choice.go:67](../../server/internal/game/entry_choice.go))
→ `errReplacementPending`
→ `ResolveEntryPayLife` ([entry_choice.go:164](../../server/internal/game/entry_choice.go))
→ `executeEntryToBattlefieldLocked`. And the pause only works where
`ReplacementEvent.entryResumable` is set, which is **exactly and only**
the land branch at
[mutations.go:624](../../server/internal/game/mutations.go) — the path
an MDFC land back takes. Sixty land-backed MDFCs become ~5 lines each
in one cycle file, registered under `<oracle_id>#1`.

**Transform — ~150 lines for a sandbox flip. ~1 day. Unlocks identity
for 401 oracle IDs immediately; play for 4 cards across both lists.**

Identity is **free from the spine**: face 0's cost and colours finally
reach the engine. What remains is the flip itself: a `transform` action
verb (there is none — `ActionType` at
[protocol.ts:60-95](../../client/src/lib/protocol.ts) has no such
member), an `Effect` primitive that calls `SetFace(1)` and marks layers
stale, and CR 712 hygiene — a transform is *not* a zone change, so
counters, damage, auras, `EnteredBattlefieldAt` and `SummonedThisTurn`
all persist untouched, which falls out of the in-place `SetFace`.

Rules-correct transform — "transform this creature" as a printed
triggered ability, day/night, the exile-and-return-transformed pattern
Jace and Kuruk use — is **much** larger and is deferred. A manual flip
in the admin context menu is honest and unblocks playtesting.

**Adventure — ~120 lines. ~1.5 days. Unlocks 2 cards, 170 oracle IDs.**

`ExilePlayPermission` gains a no-expiry flag (10) ·
`routeStackCardToGraveyardLocked` adventure branch
([mutations.go:1439](../../server/internal/game/mutations.go)) (40) ·
`faceOnResolve` already returns 0 (0) · the face picker doubles as the
adventure-half picker (10) · `ZoneBrowserModal.svelte:86-92` already
casts from exile (0) · tests.

Caveat worth stating plainly: **both adventure cards in the tracked
lists are independently blocked on their effects** — Virtue of Knowledge
needs ETB-trigger doubling, Vantress Visions needs ability-copying,
Altar of Bhaal needs reanimation, Bone Offering needs a token. Adventure
support makes them castable, not functional.

**Split — 0 lines beyond the spine. 0 cards in either list.**

137 oracle IDs; **none in either tracked decklist.** The spine alone
makes them cost the *left half* correctly instead of being free, which
is a strict improvement. Declare the simplification (left half only)
via the `CodeUnsupportedLayout` warning and move on. Fusing is not
designed here.

**Room — out of scope. Zero printings in the dump.**

**Prepare — out of scope beyond the spine.** 57 oracle IDs, 1 card
(Skycoach Conductor). Its data shape is adventure's, so the spine fixes
its cost; the *mechanic* ("while it's prepared you may cast a copy of
its spell") needs spell-copying, which does not exist. It is blocked on
that, not on faces.

**Flip (26), meld (21), class (38, not multi-face) — out of scope.**

---

## Execution order

1. **#276 alone, now.** See the verdict below. Independent of everything
   here; ~15 lines; strictly restrictive.
2. **The spine.** Nothing else can start without it, and it fixes four
   silent correctness bugs on its own.
3. **MDFC face picker.** Ships in the same sprint as (2) — the spine
   regresses Sea Gate's playability until this lands. This is the #265
   fix and the first thing that turns cards on.
4. **Transform, identity half.** Free from the spine; verify it.
5. **Transform, sandbox flip.** A `transform` verb and an in-place
   `SetFace(1)`.
6. **Adventure.** On top of `ExilePlay`.

### Explicitly do NOT attempt first

- **Do not start with transform mechanics.** Day/night, "transform
   this", and the exile-and-return-transformed pattern are a sprint on
   their own and unblock nothing the identity fix does not.
- **Do not start with adventure.** Both adventure cards in the tracked
   lists are blocked on unrelated effects anyway.
- **Do not touch split or fusing.** Zero cards, and the spine already
   makes them cost something.
- **Do not refactor the flat printed fields into methods.** 293
   `TypeLine:` literal sites say no.
- **Do not route the face choice through `PendingChoice`.** It has no
   frame for a half-validated cast and should not grow one.
- **Do not ship the spine without the picker in the same sprint.**

---

## Verdict on #276: confirmed, it stands alone, ship it first

The narrow fix proposed in the issue is correct, small, and does **not**
become redundant when the full model lands.

**The data is present and already trusted.** `cards.Card.ColorIdentity`
exists ([index.go:57](../../server/internal/cards/index.go)), is
correct for every layout — verified directly against the dump: Aang,
Swift Savior yields `["U","W"]` at the top level with `colors: null` —
and `deck/validate.go:184-191` already enforces the deck against it.
It simply never reaches `game.Card`.

**Both branches of `commanderIdentityFor` are the bug**
([mutations.go:3099-3114](../../server/internal/game/mutations.go)). It
reads `c.Effective().Colors`, which is
`printedColorsFromCost(c.ManaCost)`
([characteristic.go:84](../../server/internal/game/characteristic.go)),
then falls back to `distinctColorsInManaCost(c.ManaCost)`
([mutations.go:3119](../../server/internal/game/mutations.go)). For a
transform DFC `ManaCost` is `""`, so both yield nil.

**The change:** add `ColorIdentity []string` to `game.Card`; copy it in
`toGameCard` ([deck.go:295](../../server/internal/deck/deck.go), one
line next to the existing `Colors` copy); prefer it in
`commanderIdentityFor`; correct the stale comment at
**[mutations.go:3053-3058](../../server/internal/game/mutations.go)**
(not 2926 — the file has grown). **~15 lines.**

**Blast radius is three call sites, all narrowing:**
`materializePlanLocked` ([mutations.go:937](../../server/internal/game/mutations.go)),
`filterPipeByCommanderIdentity` ([mutations.go:3059](../../server/internal/game/mutations.go),
itself called once from `ActivateManaAbility` at :2953),
and `gatherTapSources` ([autotap.go:146](../../server/internal/game/autotap.go)).
All three only *shrink* a set of mana colours. The change can therefore
only remove colours Command Tower should never have offered — it cannot
withhold mana a player is entitled to. **No protocol change, no client
change, no zone change.** `ColorIdentity` does not need to reach the
wire.

**It survives the full model.** CR 903.4c identity is a property of the
whole card — both faces plus rules text plus hybrid symbols — which is
exactly what Scryfall's top-level `color_identity` computes and is
**not** the same as `Faces[0].Colors`. The field is still the right
field after `Faces` lands. Fixing it now costs nothing later.

**Two caveats to write into that PR.** `commanderIdentityFor` returns
the first commander's identity only; a partner pair needs the union —
moot today, since partners are refused at import
([deck.go:156-158](../../server/internal/deck/deck.go) raising
`UnsupportedMechanicError` off the face-walking sniff at
[deck.go:190-201](../../server/internal/deck/deck.go), with a second
net at [validate.go:110-121](../../server/internal/deck/validate.go)).
And it
does **not** fix `printedCharacteristic` dropping `Card.Colors`
([characteristic.go:84](../../server/internal/game/characteristic.go)) —
that stays broken for Devoid and colour indicators until the spine.
Using `ColorIdentity` sidesteps it rather than fixing it, which is the
right call for a 15-line PR.

---

## Consequences

**Good.** Two silent free-cast classes (2,550 `byName` keys between
them) stop being silent. Every DFC stops carrying a literal `"//"` in
its type list. The land-drop branch stops firing on sorceries. The
client's twenty-odd type checks start receiving clean data without a
single edit. Transform commanders get a colour identity a week before
the rest of the model.

**Bad.** `CatalogKey` puts a derived string between every card and its
spec — cheap, but it is a new indirection in the hottest lookup in the
engine, and a call site that forgets it will silently resolve to face
0's spec rather than erroring. The `Faces[ActiveFace]` invariant is
enforced by convention plus one setter; a direct write to `ActiveFace`
desynchronises a card and nothing will catch it. Both want a
`go vet`-adjacent test that walks every `Card` in a finished game and
asserts the invariant.

**Declared simplifications, to be surfaced in the import banner.**
Split cards cast as their left half only. Prepare cards cast as their
creature half only. Flip and meld import face 0 only. Transform requires
a manual flip. A fetched or effect-placed MDFC land takes the unpaid
branch and enters tapped, inheriting the existing shockland limitation
([shocklands.go:42-58](../../server/internal/cards/effects/shocklands.go)).

---

## What shipped

Steps 2 and 3 of the execution order, in one PR, as the sizing
demanded — the spine alone would have left Sea Gate Restoration a
seven-mana sorcery you cannot play as a land.

### The design held

Every load-bearing call survived implementation unchanged:

- **The flat fields stay fields.** Zero of the 293 `TypeLine:` literal
  sites and zero of the 389 `Card{` literals needed rewriting. Three
  *assertions* changed, all because `Card.Name` is now the active
  face's name rather than Scryfall's composite — which is the
  behaviour the design asked for.
- **`Effective()` needed no change**, exactly as predicted. The
  `"//"` and `"—"` junk types disappeared without the layer engine
  learning anything.
- **`CastSpell`'s value copy was the whole game.** One
  `card.SetFace(params.Face)` after the copy made all ten downstream
  reads face-correct.
- **`CatalogKey` with a bare face-0 key** left all ~241 existing specs
  and every single-faced card untouched. 25 production call sites
  swapped mechanically.
- **The MDFC land backs really did cost nothing in rules code.** All
  sixty are a table in `mdfc_lands.go`: fifteen
  `EntersTappedUnlessYouPayLife(name, 3)`, thirty-five
  `SelfEntersTapped()`, ten with no entry clause at all.

### Three things the design did not anticipate

1. **The face must be written into the SOURCE ZONE, not just onto the
   copy.** The CR 614 pipeline resolves the entering card by ID
   through `LookupCardForEffect`, so a land back whose `ActiveFace` is
   still 0 on the card in hand never finds its own spec. `CastSpell`'s
   land branch now stamps the face on the zone card before running
   the pipeline — and restores it on the refusal paths, because a cast
   that does not happen must not leave a card in hand wearing its back
   face. Same reasoning for the stack push.
2. **`printedLoyalty`'s face fallback was not merely redundant, it was
   wrong.** "The first face that prints a loyalty" gave Nissa,
   Vastwood Seer — a 4/4 Elf Scout — her back face's 3. Per-face
   loyalty fixes it; the whole-card fallback is gone.
3. **The entry-prompt resume path never counted the land drop.** A
   pre-existing bug, live since #268, affecting every shockland: the
   `LandsPlayedThisTurn` tally is bumped in `CastSpell`'s land branch
   but not in `executeEntryToBattlefieldLocked`, so a land that paused
   on a prompt was invisible to the legal-move enumerator. Found by
   the MDFC back-face test and fixed alongside.

### One decision the spike left open, settled here

`CastSpellParams.Face` **rejects** a face the card does not offer
(`ErrInvalidFace`) rather than clamping to the front. Silently casting
the wrong half — a seven-mana sorcery when the player meant a land —
is the worst available failure. `SetFace` itself still clamps, because
a setter's job is to never leave a card incoherent; refusing is the
action layer's.

### Resolved

- **#265 / #289** — both halves. `IsLand()` reads the active face, so
  the land branch stops firing on a sorcery front, and the cost gate
  finally sees a real cost.
- **#325** — Aang, Swift Savior's ETB airbend, now that the card
  imports as a {1}{W}{U} 2/3 instead of a free colourless 0/0.
- **#343** — partially: the 0/0 is fixed and the back face is imported
  and on the wire. "Waterbend {8}: Transform Aang" still needs the
  `transform` verb, which is step 5.
- The 1,503 free-casting DFCs, the 1,047 unparseable-cost casts, the
  `"//"` junk types on 501 oracle IDs, and 71 `byName` keys resolving
  to a `reversible_card` placeholder.

### Still outstanding

Steps 4–6: transform's flip verb, adventure's exile-and-recast, split
fusing. All are declared in the import banner
(`CodeUnsupportedLayout`) rather than left silent.

---

## Addendum (S32): a per-instance face on the exile-play grant

This ADR's §4 table says a `transform` card is "always face 0
(CR 712.4)" at announce and always face 0 as a permanent. The first
half is right and stays. **The second half was wrong**, and it is what
blocked S27's battles: a Siege prints *"When it's defeated, exile it,
then **cast it transformed**"*, which is an effect casting a
`transform` card's BACK face. `CastableFaces` correctly refuses to
offer that face — nobody may choose it — but `faceOnResolve` returned
0 unconditionally, so even a back-face cast would have resolved into
a front-face permanent. For a Siege that means the defeated battle
re-entering the battlefield, which is worse than not casting it.

Three changes, all small, all in the direction §4 already pointed:

1. **`ExilePlayPermission` grows `Face int`** (`game/exile_play.go`).
   Zero means the grant does not speak about faces, which is every
   grant before S32 and leaves impulse exile, airbend, warp and
   cascade untouched. Non-zero names the ONE face the grant opens.
   §4's adventure row wanted the same slot — *"an adventure exile has
   no end-of-turn expiry, so `ExilePlayPermission` needs a 'does not
   expire' flag"* — and this is the other half of that same
   permission growing up.

2. **A face-naming grant SETS the face rather than permitting it**
   (`faceForCastLocked`, `game/face.go`). This is the one deliberate
   departure from the "reject, never clamp" rule settled above. That
   rule protects a CHOICE: silently casting the wrong half of a modal
   DFC is the worst failure available. A face-naming grant offers no
   choice — there is exactly one legal cast of that card by that
   player — so `params.Face` is ignored rather than being a second
   place the only possible answer has to be spelled. An out-of-range
   granted face is still refused, because clamping it would land on
   face 0 and hand the player a free cast of the battle.

3. **`faceOnResolve` keeps a cast `transform` face.** CR 712.4 is a
   rule about casting and is enforced where it belongs, by
   `CastableFaces`. A non-zero cast face can now only have arrived
   through an effect that said "cast it transformed", and that
   permanent is the back face.

The catalog side needed nothing new: a Siege back face registers under
`"<oracle_id>#1"`, which is §5's keyspace, and the sixty MDFC land
backs were already living in it.

**A fourth change the first three made necessary: CR 712.8.** A
double-faced card is front face up in every zone except the
battlefield and the stack, and the engine did not enforce it anywhere.
That was latent while the only back faces on the battlefield were MDFC
lands — Sea Gate, Reborn dying left a *land card* in the graveyard,
which nothing in the catalog reads. A Siege back face makes it
reachable and makes it matter in the wrong direction: a Refraction
Elemental that died would leave a **creature card** where a battle
card belongs, and "return target creature card from your graveyard"
would reanimate a 4/4 off a card that is not a creature card at all.
Stronger than printed.

The reset lives in `MoveCard`, keyed on the DESTINATION, next to the
CR 400.7 "new object" resets that were already there (tapped state,
counters, attachments, the S27 protector). Destination rather than
source because that is how the rule reads — it is a property of where
the card *is* — and because a spell countered off the stack owes the
same reset as a permanent that died. This is the family
[#372](https://github.com/krakenhavoc/cmd_and_ctrl/issues/372) and
#539 are in: a zone-change rule that was honoured on some routes and
not others is honoured on the one primitive they all share.

**The declared simplification is timing, not faces.** The free cast is
a GRANT bounded to the turn the Siege was defeated, not an inline cast
during the trigger's resolution — cascade's trade, for cascade's
reason (the announce path has no frame for a half-validated cast). A
Siege defeated on somebody else's turn is therefore lost, which is
weaker than printed and never stronger. `effects.SiegeTransformedCastCaveat`
is the one sentence every Siege in the catalog publishes about it.
