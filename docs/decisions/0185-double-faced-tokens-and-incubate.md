# ADR 0185 — Double-faced tokens and incubate (CR 701.53, CR 111.10i)

Date: 2026-10-14
Status: Accepted
Sprint: none (Reality Fracture set, tracker #2795; outside docs/sprints.md at the owner's request)
Issues: #2417 (seam: incubate), #2832 (Reality Fracture engine gaps), #2795 (set tracker).
Builds on: [ADR 0079](0079-transforming-a-permanent.md) (the transform verb, whose out-of-scope item 1 named this), [ADR 0034](0034-multi-face-cards.md) (`Card.Faces` / `ActiveFace` / `SetFace`), [ADR 0083](0083-token-abilities.md) (a token's abilities live in the catalog under a token key), [ADR 0078](0078-token-art.md) (token art).
Numbering: the number was reserved in the claim comment on #2417. After `git fetch`, the highest number on `origin/develop` was 0183; the other open branches carried no ADR files at the time.

## Context

CR 701.53a: to incubate N, create an Incubator token that enters with N +1/+1 counters on it. CR 111.10i, restated in 701.53b: an Incubator token is a double-faced token. Its front face is a colorless Incubator artifact with "{2}: Transform this token." Its back face is a 0/0 colorless Phyrexian artifact creature named "Phyrexian Token."

ADR 0079 listed double-faced tokens as out of scope because `game.Card` carried faces only for a card with an oracle ID, and a token template described one face. Reading the code again, the gap is smaller than that:

- `Card` already has `Layout`, `Faces` and `ActiveFace` with no dependence on an oracle ID, and `SetFace` materialises the face that is up. `cloneCard`, the snapshot and the wire carry all three.
- `CanTransform` keys on `Layout` and `len(Faces)`, not on the oracle ID, so a token with the transform layout and two faces is already a transformable permanent.
- A token's abilities are found through `game.CatalogKey`, which answers the token key for an object with no oracle ID. That is the one place that did not know about faces: the back face of a card keys as `<oracle id>#1`, and a token had no such suffix.
- Counters are not part of a characteristic, so a transform keeps them (CR 712.18) with no work.

The Phyrexian Token is a 0/0, so its size is its counters. The engine skips the toughness-zero state-based action for a token whose 0 it cannot vouch for (`Card.ToughnessIsKnown`), so the template has to say its 0/0 is real.

## Decisions

### 1. A token template can carry a back face

`tokenTemplate.Back *tokenTemplate` (`effects/token_catalog.go`) is the second face. `tokenFromCatalog` builds `Layout = transform`, `Faces = [front, back]` and `ActiveFace = 0` from the two templates, so a template never spells them and the invariant "the flat fields equal `Faces[ActiveFace]`" holds by construction. The back is a template of its own so that a face can carry the abilities the front does; a back that prints nothing registers nothing. `checkTokenBack` refuses a back with a back of its own, a face with no type line, a face carrying a catalog identity or instance closures, and a template that spells its own `Layout` or `Faces`.

### 2. The back face keys as `<token key>#1`

`catalogKeyOf` gains one case: a token (no oracle ID, a token key) on a non-zero face of a transform layout keys as `token:<slug>#<face>`. The oracle ID still wins, so a token copying a real double-faced card resolves to that card. A back face that declares abilities files them under `tokenBackKey(slug)` at boot; the Phyrexian Token declares none, so it resolves to no entry, as a vanilla token does.

### 3. Incubate is a creation with a counter entry clause

`effects.Incubate{Controller, N}` calls `CreateTokensForEffect` with the Incubator template and `TokenEntryOptions.Counters{"+1/+1": N}`. The counters ride the CR 614 entry pipeline, the same path any "enters with counters" token takes, so Doubling Season and the other replacements see the creation and the counters. `Controller` defaults to the resolving ability's controller and is set for "its controller incubates X" (Excise the Imperfect). N of zero or less still creates the token: "incubate 0" is legal. No event is emitted for the keyword action; no card in the catalog or in the dump needs "whenever you incubate".

### 4. The Incubator

`IncubatorToken()` is the template: `Token Artifact — Incubator`, colorless, one activated ability "{2}: Transform this token." (`Do(TransformThis{})`, declared as `AnswerAnimate` for the bot's response pricing), back face `Token Artifact Creature — Phyrexian` named "Phyrexian Token", 0/0. `PrintedPTKnown` is set on the template so a transformed Incubator with no counters is a real 0/0 and CR 704.5f buries it. `TransformThis` already applies CR 701.27f, so an ability that went on the stack before the token turned over does nothing afterwards.

### 5. A token copy of a token copies its key and its faces

`TokenCopyTemplate` copied a token's name, type line and stats but not its token key, so a copy of any token with a catalog ability lost it. For a source with no oracle ID and a token key it now copies the key and `PrintedPTKnown`, and for a transform-layout source both faces and which one is up. CR 707.8a: the copy of a double-faced token is a double-faced token with both faces, entering with the same face up. Counters are not copied (CR 707.2), so a copy of a transformed Incubator is a 0/0 back face and dies at once, which is the printed result. The faces are a deep copy.

## Not changed, and why

- **`MoveCard` still puts a token on its front face outside the battlefield** (CR 712.8a, `game/zone.go`). A transformed Incubator that dies sits in the graveyard for a moment as a front-face Incubator before CR 704.5d removes it. The dies triggers read last-known information stamped at the exit (`Event.LastKnownTypes`, `leftAsType`), not that card, and `TestATransformedIncubatorDyingIsACreatureDying` pins it with Blood Artist.
- **Clone-style copies.** `CopiableValuesOf` copies `Layout`, `Faces` and `ActiveFace` for every source. A Clone of a double-faced permanent therefore becomes double-faced, which CR 707.8 does not do for a non-double-faced copier. This predates the ADR and covers every transforming card, not tokens in particular; it is filed as #3512 rather than fixed here (with the token-copy half of CR 707.8a for a real double-faced card).
- **Token art.** The token art pool takes only Scryfall's `token` layout, and the Incubator printings are `double_faced_token` with a combined name, so the Incubator has no art and renders as a text token. Filed as #3511.
- **Triformer's third side** (Incubation Triformer, "transform it into a third side"), **"incubate N, X times"**, and the Incubator-aware cards (Elvish Vatkeeper, Progenitor Exarch, Phyrexian Cytoshaper, Norn's Inquisitor's "transforms into a Phyrexian" trigger) are later card slices. The first needs a third face; the rest are card files on this seam.

## Consequences

- A second double-faced token is a template with a `Back`. `IsDoubleFacedForEffect` now answers true for an Incubator, which is right: Tetzin, Gnome Champion's "another double-faced artifact" can name it.
- A snapshot taken with an Incubator on the battlefield is still a restore point; the token key, the faces and the active face are plain data, and the back face's catalog entry is empty.
- Cards shipped on it: Sunfall, Traumatic Revelation, Converter Beast, Eyes of Gitaxias, Merciless Repurposing and Excise the Imperfect, all `full`.
