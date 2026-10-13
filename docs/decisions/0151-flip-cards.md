# ADR 0151 — Flip cards: a status, not a face (CR 710, CR 110.5)

**Status:** Accepted (owner decisions 2026-10-13) · 2026-10-13 · S53 — Faces, zones, and shared payment (ADR first)
**Issues:** [#1892](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1892) (seam: flip cards, Rune-Tail, Kitsune Ascendant). Found while landing [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) §1 (#1858), which moved Rune-Tail onto this row. Registry row: `flip-cards` (to be closed with the implementation). Tracker: [#1910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1910).
**Owner decisions:** 2026-10-13, every recommendation accepted. See [Questions for the owner (answered 2026-10-13)](#questions-for-the-owner-answered-2026-10-13).
**Numbering:** assigned 0151 for S53 (0150–0155 are being written in parallel). After `git fetch --all --prune`, `git log --all --name-only -- 'docs/decisions/015*'` found no file numbered 015x on any ref, and every remote head (`origin/develop`, `origin/main`, `origin/cost-ledger`, `origin/docs/issue-audit`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/playmats`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default`, `pr/2326`) tops out at 0149.
**Builds on:** [ADR 0034](0034-multi-face-cards.md) (the face model: `Card.Faces`, `ActiveFace`, `SetFace`, the `#N` catalog key), [ADR 0079](0079-transforming-a-permanent.md) (turning a permanent over in place, and its "exile/graveyard, then return transformed" siblings), [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) §1 (state triggers) and §4 (per-face keywords), [ADR 0145](0145-meld.md) (the last layout to leave `layoutSimplifications`).

---

## Context

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`):

- **CR 710.1:** a flip card has a two-part frame on one card. The right-side-up half is the card's normal characteristics; the upside-down half is an alternative set. The back is an ordinary Magic back. **710.1a:** the top half has the normal name, text box, type line, power and toughness, and usually the ability that flips it. **710.1b:** the bottom half's alternative name, text box, type line, power and toughness "are used only if the permanent is on the battlefield and only if the permanent is flipped".
- **CR 710.1c:** "A flip card's color and mana cost don't change if the permanent is flipped. Also, any changes to it by external effects will still apply."
- **CR 710.2:** everywhere but the battlefield, and on the battlefield before it flips, a flip card has only its normal characteristics. Once flipped, the alternative name, text box, type line, power and toughness apply instead. Its example: Akki Lavarunner is nonlegendary and flips into the legendary Tok-Tok, Volcano Born, so "search for a legendary card" can't find the card and "legendary creatures get +2/+2" affects only Tok-Tok.
- **CR 710.3:** whether a permanent is flipped must be clear at all times, tapped or untapped.
- **CR 710.4:** flipping is one-way: a flipped permanent can't become unflipped. "However, if a flipped permanent leaves the battlefield, it retains no memory of its status. See rule 110.5."
- **CR 710.5 / 201.4c:** a player told to choose a card name may choose a flip card's alternative name.
- **CR 110.5:** a permanent's **status** is its physical state, in four categories: tapped/untapped, **flipped/unflipped**, face up/face down, phased in/out. **110.5a:** status is not a characteristic, though it may affect them. **110.5b:** permanents enter unflipped unless a spell or ability says otherwise. **110.5c:** a permanent keeps its status until something changes it, "even if that status is not relevant to it" (the example: a Dimir Doppelganger that became a copy of Jushi Apprentice and flipped keeps its flipped status as a copy of Runeclaw Bear, and is Stabwhisker the Odious when it later becomes a copy of Nezumi Shortfang).
- **CR 707.2:** status is not copied. **707.3:** "The copy's copiable values become the copied information, as modified by the copy's status." Its example: Tomoya the Revealer (a flipped Jushi Apprentice) becoming a copy of Nezumi Shortfang (unflipped) has Stabwhisker the Odious's characteristics.
- **CR 730.2h:** a merged permanent containing a flip card uses its alternative characteristics if the merged permanent is flipped.

Flipping is not transforming. CR 701.27 (transform) is about double-faced cards; a flip card is single-faced (CR 710.1), so "whenever this permanent transforms" never sees a flip, and an instruction to transform a flip card does nothing (CR 701.27c, 712.9).

### What the engine has

Most of the representation is already here, by accident of ADR 0034:

- Scryfall ships a flip card as `layout: "flip"` with a two-entry `card_faces` array, so `deck.printedFaces` already gives the card two `game.Face`s, and `Card.SettleImported` materialises face 0. `CastableFaces` falls through to `[]int{0}` for the layout, so a flip card is cast as its top half today, which CR 710.2 says is right.
- `catalogKeyOf` (`server/internal/game/effect_hooks.go`) appends `#<ActiveFace>` for any card with a non-zero face, so a flipped half registered as `"<oracle_id>#1"` would get its own triggers, statics and abilities the moment `ActiveFace` moves, exactly as Vault of Catlacan does for Storm the Vault.
- `MoveCard` (`server/internal/game/zone.go`) already sets a card back to face 0 when it lands anywhere but the battlefield or the stack (the CR 712.8a reset). That is CR 710.4's "retains no memory" for free on every route but one (decision 1).
- `CardNameMatches` (`choose_card_name.go`) already compares a chosen name against every face, so CR 710.5 holds today.
- `deck/validate.go`'s `layoutSimplifications` still holds exactly one entry, `"flip": "imports as its front face only"`, and the roadmap row says "A flip card can be cast as its top half, but it can't flip."

What is missing, and what is quietly wrong for the day a flip card is turned over:

1. **No verb.** Nothing sets `ActiveFace` on a flip permanent. `TransformPermanentForEffect` refuses it (`isDoubleFacedPermanent` allowlists `transform` and `modal_dfc`), which is correct and must stay so.
2. **Colour and mana cost would change (CR 710.1c).** Scryfall gives the flipped half `mana_cost: ""` and neither half a `colors` array (the colour is only at the top level). `faceColors` therefore derives face 0's colours from its cost (right) and gives face 1 none (wrong), so `SetFace(1)` would turn a flipped Kuon's Essence colourless with mana value 0.
3. **Keywords would not change.** `FaceKeywordsApply` answers only for `transform` and `modal_dfc`, so a flip card keeps its top half's keywords, narrowed from Scryfall's card-level union by `printedKeywords` (face 0). Flipped, Akki Lavarunner would keep haste and never get Tok-Tok's protection from red; Bushi Tenderfoot would never get Kenzo's double strike and bushido 2.
4. **Copies would copy the status.** `printedValuesOf` (`copy.go`) copies `Layout`, `Faces` and `ActiveFace` wholesale, which is right for a transformed permanent (CR 712.8 copies the face that is up) and wrong for a flip card: CR 707.2 does not copy status, so a Clone entering as a copy of Tok-Tok is an unflipped Akki Lavarunner.
5. **The image.** A flip card has ONE printed image, at Scryfall's top level; its `card_faces` carry no `image_uris`. `cards.ImageURIForFace(c, size, 1)` returns `""`, so `/cards/{id}/image?face=1` has nothing to serve, and the hover overlay's "other face" inset (`HoverZoomOverlay.svelte`, `otherFace`) already asks for it whenever a flip card in hand is hovered.

### The cards

Scryfall's `default-cards.json` (the file `scripts/go-docker.sh` tests read, refreshed 2026-09-24) has **26** oracle IDs with `layout: "flip"`. Five are Role-token pairs (Wicked // Cursed, Monster // Virtuous, Royal // Young Hero, Royal // Virtuous, Monster // Sorcerer), which are not cards a deck holds. Curse of the Fire Penguin is an Unhinged card, not legal in Commander. Erayo, Soratami Ascendant is banned in Commander. That leaves **19 Commander-legal flip cards**, the issue's "~19", listed in decision 8. None of them is in the catalog today.

## Decisions

### 1. Flipped is a status on the permanent, and the face follows it

`game.Card` gains one field:

```go
// Flipped is CR 110.5's flipped/unflipped status. Only a permanent has
// it (CR 110.5d); a card enters the battlefield unflipped unless the
// effect putting it there says otherwise (CR 110.5b), and it is cleared
// by every zone change (CR 710.4). It is not a characteristic (CR
// 110.5a) and is never copied (CR 707.2). A permanent that is not a
// flip card keeps it once set, irrelevant (CR 110.5c).
Flipped bool
```

and one invariant, held in one place: **for a card whose `Layout` is `flip`, `ActiveFace` is 1 exactly when `Flipped` is true.** Every other card's `ActiveFace` is untouched by `Flipped`. The face does the work, as it does for transform: `SetFace(1)` materialises the alternative name, type line, P/T and (decision 3) keywords, and `CatalogKey` becomes `"<oracle_id>#1"`. The status is what the rules name, and it is what survives a copy effect (decision 4) and a face-down turn (CR 708, 110.5c), which an `ActiveFace` alone cannot carry.

`LayoutFlip = "flip"` joins the layout constants in `face.go`, with `IsFlipCard(c Card) bool` (`c.Layout == LayoutFlip && len(c.Faces) == 2`).

The cleared-on-move half is one line beside the CR 712.8a reset in `MoveCard`, under the same condition: a move whose destination is not the battlefield clears `Flipped`, as the face reset already sets a flip card back to face 0 there. Only a permanent can be flipped (CR 110.5d) and a battlefield-to-battlefield move does not exist (a blink is two moves), so every card arrives on the battlefield unflipped (CR 110.5b) unless an entry deliberately set the status in its source zone first, which is decision 6 and nothing else.

Owner question 1 offered the cheaper alternative (no field, `ActiveFace` is the status); the owner chose the field.

### 2. The verb: `Game.FlipPermanentForEffect`, and `effects.FlipThis`

`server/internal/game/flip.go`:

```go
// CanFlip reports whether an instruction to flip this permanent does
// anything: a face-up flip card that is not already flipped (CR 710.4).
func CanFlip(c Card) bool

// FlipPermanentForEffect flips the named battlefield permanent (CR 710).
// Caller must hold g.mu.
func (g *Game) FlipPermanentForEffect(cardID uuid.UUID) error
```

It is `TransformPermanentForEffect`'s shape, deliberately, and for the same rule: flipping is not a zone change (CR 710.1c's "any changes to it by external effects will still apply" is CR 712.18 in other words). The permanent keeps its `InstanceID`, `ObjectEpoch`, counters, damage, attachments, tap state, controller, CR 613.7 timestamp and summoning sickness, and nothing re-triggers. The body sets `Flipped`, calls `SetFace(1)` for a flip card, nils `Card.effective`, and emits the event (decision 5).

What it refuses, returning nil with no event (an instruction that legally does nothing is not an `EventEffectError`, ADR 0079 decision 4):

- a permanent already flipped (CR 710.4: one-way, and flipping twice is nothing);
- a face-down permanent (it has no text, so nothing it prints could have asked; a face-up flip effect from elsewhere is not printed);
- a permanent that is not a flip card. Strictly, CR 110.5c lets any permanent be flipped and keep a status irrelevant to it, but every printed flip instruction says "flip this creature" of a flip card, and no card flips another permanent. Not setting the status on a non-flip card is the weaker direction and is recorded as out of scope 2.

`ErrCardNotFound` for an ID naming no battlefield permanent, as transform does.

Card side, `server/internal/cards/effects/flip.go`: `FlipThis{}` (every printed card says "flip this creature" or "flip <name>") reading the source off the `Context`, with the `isNewSourceObject` guard (#1432) so a flip trigger whose source left and came back flips nothing. The ki cycle's "you may flip it" is `MayChoice{OnYes: FlipThis{}.Apply}`; nothing new.

**Never `TransformPermanentForEffect`.** `CanTransform` keeps refusing a flip card, and `CanFlip` refuses a double-faced one. Two verbs, two rules, and neither reaches the other.

### 3. `SetFace` keeps a flip card's colour and mana cost, and reads its keywords per face

Two changes in `face.go`, both CR 710.1c / 710.2:

- **Colour and mana cost.** `SetFace` on a `flip` card writes `Faces[0].ManaCost` and `Faces[0].Colors` whatever face it is setting. That makes CR 710.1c true in every reader (the colour derivation, `manaCostForValue`, the wire's `mana_cost`) with no branch outside `SetFace`, and it is right for cards imported before this lands, whose face 1 carries no cost or colour. The importer is not changed to stamp face 0's cost onto face 1, because that would put a cost on the wire's face list that is not printed there and would leave every already-imported card and snapshot wrong.
- **Keywords.** `FaceKeywordsApply` answers true for `flip` too, so the importer stamps each half's own confirmed keywords (`printedKeywordsForFace`, which already narrows Scryfall's card-level union by a face's own keyword lines) and `SetFace` puts the flipped half's on the card. Flipped, Faithful Squire becomes a Kaiso that flies; Akki Lavarunner loses haste and Tok-Tok gains protection from red.

`manaCostForValue` needs nothing: the flat `ManaCost` is face 0's on both halves.

### 4. Copies: the copiable values are always the top half, and the copier's status picks the half

CR 707.2 and 707.3, in two lines:

- `CopiableValuesOf` normalises a flip card to face 0 before reading it (`whole := src; whole.SetFace(0)`, the shape the split-card branch already uses). A Clone of Tok-Tok copies Akki Lavarunner, with both halves in `Faces`, so it can flip itself later.
- `setPrintedValues`, after writing `Layout`, `Faces` and `ActiveFace`, re-materialises face 1 when the copied values are a flip card's and the **copier** is `Flipped`. That is CR 110.5c's Dimir Doppelganger example exactly: a flipped permanent that becomes a copy of Nezumi Shortfang is Stabwhisker the Odious. `restorePrintedSelf` (a copy effect ending) gets the same treatment, so a flipped Tok-Tok that was a Grizzly Bears until end of turn comes back as Tok-Tok.

`Flipped` is not in `PrintedValues`, and must never be: it is status (CR 707.2).

### 5. Flipping emits `EventFlipPermanent`, the layer version bumps, and the log narrates it

`EventFlipPermanent EventKind = "flip_permanent"` (`CardID`, `Actor` the controller, `Label` the name it flipped from). **Not** `"flip"`: `rngStreamFlip` and `LogFlip` already mean a coin flip (`EventFlipCoin`), and a reader of the log or a trigger author must not confuse the two. No printed card triggers on a permanent flipping, so no trigger constructor ships; the harvester can read the kind if one ever does.

- `layerVersionBump.OnEvent` gains an `EventFlipPermanent` arm beside `EventTransform`'s, for the reason that arm exists: the printed characteristics changed wholesale and `Effective()` returns its cache verbatim when warm.
- `LogFlipPermanent LogKind = "flip_permanent"`, narrated "Akki Lavarunner flipped into Tok-Tok, Volcano Born" (`TestEveryEventKindIsNarratedOrDeliberatelySilent` requires one or the other, and CR 710.3 says the table must be able to tell).

### 6. Entering flipped: Homura's "return it to the battlefield flipped"

Homura, Human Ascendant is the one card that says otherwise under CR 110.5b. ADR 0079 already has the shape for "return it from a graveyard transformed" (`ReturnFromGraveyardTransformedForEffect`, decision 10): set the face on the card in its source zone, run the ordinary graveyard entry, put the face back if the entry is cancelled or redirected. The flip form is the same function with a different guard and one more write, so the body moves into a shared `returnFromGraveyardOnFaceLocked(cardID, face, guard, ...)` and `ReturnFromGraveyardFlippedForEffect` calls it with `IsFlipCard`, setting `Flipped` alongside the face on the card in the graveyard. The move into the battlefield keeps both, because `MoveCard` clears them only for a destination that is not the battlefield (decision 1), and the CR 614 pipeline finds the flipped half's catalog key on the way in. A cancelled or redirected entry puts both back, exactly as the transform form puts the face back, so no card is left in a graveyard flipped.

No other entry site takes a status. A Clone, a reanimation or a blink of a flip card enters unflipped, which is CR 110.5b.

### 7. The wire, the client, and the bot

**The wire needs no new field.** `CardView.layout`, `faces` and `active_face` already reach the client; for a flip card `active_face == 1` **is** "flipped", because a flip card's face follows its status (decision 1). The flat `name`, `type_line`, `power`, `toughness` are the flipped half's, and `mana_cost` is the card's own (decision 3). Redaction is unchanged: a battlefield permanent is public.

**The image.** One printed image serves both halves.

- Server: `cards.ImageURIForFace` answers face 1 of a `flip` card with the top-level image, so `/cards/{id}/image?face=1` never 404s (it does today, from the hover inset).
- Client: `cardImageURL` (`client/src/lib/cardImage.ts`) always asks for face 0 of a `flip` card, so the browser and service-worker caches hold one image per printing. `Card.svelte` turns a flipped flip card's image 180° (composed with the tap rotation, so a tapped flipped permanent is 270°, which is CR 710.3's "clear when tapped") (owner decision 2). `HoverZoomOverlay.svelte`'s `otherFace` inset shows the same image turned the other way, labelled with the other half's name.
- `faces.ts`: `LAYOUT_FLIP = "flip"`, and `needsFacePicker` stays false (CR 710.2: a flip card is only ever cast as its top half).

**The bot.** Nothing new to decide. Every flip is a trigger or ability the card already resolves; the ki cycle's "you may flip it" is a `confirm`, which the heuristic's `choiceConfirm` arm (`aiseat/heuristic/choices.go`) already takes ("take the offer") unless the accept branch costs life. Flipping is never worse for the card's controller in the printed pool. The bot's permanent valuation reads the effective characteristics, which are the flipped half's after the flip.

### 8. Deck import, validation, and the cards

`flip` leaves `layoutSimplifications`, which is then empty. The map and `unsupportedLayoutViolations` stay (the next unsupported layout uses them); the `faces_test.go` cases that pin the flip warning move to asserting there is none, and `CodeUnsupportedLayout` keeps its pinning test against a synthetic layout. The roadmap row `flip-cards` closes and a `flip` mechanic row is added, with the implementation, not here.

The 19 Commander-legal flip cards, with what each is expected to need beyond this ADR (the implementation PR confirms each against the catalog before writing it):

| Card // flipped half | Flips when | Expected extra need |
|---|---|---|
| Rune-Tail, Kitsune Ascendant // Rune-Tail's Essence | state trigger, 30+ life | none expected: `WhenState` (ADR 0107 §1); the Essence is a prevention static |
| Akki Lavarunner // Tok-Tok, Volcano Born | deals damage to an opponent | none expected (damage +1 replacement for red sources) |
| Bushi Tenderfoot // Kenzo the Hardhearted | a creature it damaged this turn dies | "dealt damage by this creature this turn" tracking: check |
| Budoka Gardener // Dokai, Weaver of Life | ten or more lands, on activation | none expected (`MayPutALandFromHand`) |
| Budoka Pupil // Ichiga, Who Topples Oaks | ki counters, end step, "you may" | none expected (Quillmane Baku has the ki shape) |
| Callow Jushi // Jaraku the Interloper | ki counters | none expected |
| Cunning Bandit // Azamuki, Treachery Incarnate | ki counters | none expected |
| Faithful Squire // Kaiso, Memory of Loyalty | ki counters | none expected |
| Hired Muscle // Scarmaker | ki counters | none expected |
| Homura, Human Ascendant // Homura's Essence | returns flipped when it dies | decision 6 |
| Initiate of Blood // Goka the Unjust | "when that creature dies this turn" | a delayed dies trigger: check |
| Jushi Apprentice // Tomoya the Revealer | nine or more cards in hand, on activation | none expected |
| Kitsune Mystic // Autumn-Tail, Kitsune Sage | end step, two or more Auras | Autumn-Tail moves an Aura: check |
| Kuon, Ogre Ascendant // Kuon's Essence | end step, three or more creatures died this turn | none expected |
| Nezumi Graverobber // Nighteyes the Desecrator | opponent's graveyard empty | none expected |
| Nezumi Shortfang // Stabwhisker the Odious | opponent's hand empty after discard | none expected |
| Orochi Eggwatcher // Shidako, Broodmistress | ten or more creatures | none expected |
| Sasaya, Orochi Ascendant // Sasaya's Essence | seven or more lands revealed | the Essence's extra-mana trigger: check |
| Student of Elements // Tobita, Master of Winds | state trigger, "has flying" | none expected (`WhenState`) |

Four of them (Rune-Tail, Homura, Kuon and Sasaya; Erayo too, but it is banned) are legendary creatures on their top half and so can be commanders. A commander that flips is still the commander (CR 903.3; the designation is the card's), and one that flips into a Legendary Enchantment simply deals no combat damage. Nothing in the commander code reads the face.

## Testing

All in the ordinary packages, no new harness:

- `internal/game/flip_test.go`: the verb keeps `InstanceID`, epoch, counters, damage, attachments, tap state, timestamp and summoning sickness; the name, type line, P/T, keywords and catalog key change; colour and mana value do not (CR 710.1c); flipping twice does nothing and emits one event (CR 710.4); a face-down permanent, a transform card, a modal DFC and a single-faced card all refuse with no event; `TransformPermanentForEffect` on a flip card does nothing.
- Leaving and coming back: a flipped permanent that dies, is bounced, exiled, or blinked arrives unflipped on its top half (CR 710.4, 110.5b); a flipped commander sent to the command zone likewise.
- Copies (CR 707.2, 707.3, 110.5c): a Clone of a flipped flip card enters as the unflipped top half and can flip itself; a flipped permanent that becomes a copy of an unflipped flip card shows the copied card's flipped half; a flipped flip card that is a copy of a vanilla creature until end of turn comes back flipped.
- The layer cache: an anthem for legendary creatures affects Tok-Tok and not Akki Lavarunner across the flip with no other event in between (CR 710.2's own example).
- `EventFlipPermanent` is narrated (`TestEveryEventKindIsNarratedOrDeliberatelySilent`) and bumps the layer version.
- Snapshot: `Flipped` classified `carried` in `snapshot_drift_test.go`'s card plan with a probe in `snapshot_carried_test.go`, and `closure_fields_test.go` classifies it (a plain bool, no closure). Additive to schema v8; a snapshot written before has every card unflipped, which is true of every snapshot written before.
- Deck: `printedFaces` stamps per-face keywords on a flip card; `deck.Validate` no longer warns on one.
- Cards: each catalog card's own test, and Rune-Tail's end-to-end (gain to 30, the state trigger resolves, damage to your creatures is prevented). The catalog soak (`AISEAT_CATALOG_GAMES`) picks the cards up as they register.
- Client: `cardImage.test.ts` (a flip card asks for face 0 whatever its active face) and a `Card.svelte` test for the rotation class.

## Alternatives considered

1. **Reuse transform.** Add `flip` to `isDoubleFacedPermanent` and let `TransformPermanentForEffect` turn it over. Rejected: it would make a flip card a double-faced card to every reader of that predicate (`CanTurnFaceDown`'s CR 712.16 refusal, `IsDoubleFacedForEffect`'s targeting), fire "whenever a permanent transforms", let a second transform unflip it (CR 710.4), and change its colour and cost (CR 710.1c).
2. **`ActiveFace` as the status, no new field.** Cheaper, and right for every printed flip card on its own. Wrong for CR 110.5c and 707.3 (a flipped permanent that becomes a copy of a flip card) and for a face-down turn of a flipped permanent. This is owner question 1's option B.
3. **Stamp face 0's cost and colours onto face 1 at import.** Rejected in decision 3: it misstates the printed face on the wire and leaves every existing import and snapshot wrong.
4. **A separate printed image per half (cropping or a synthesised image).** Scryfall has none, and serving the one printed card turned over is what the physical card does.

## Out of scope

1. **Mutate (CR 730.2h).** A merged permanent with a flip card in it. The engine has no mutate (no merged permanents at all), so there is nothing to carry a flip component; whoever builds mutate reads `Flipped` off the merged permanent.
2. **Flipping a permanent that is not a flip card.** No printed effect does it; CR 110.5c's status-without-relevance arises only through copy effects, which decision 4 covers.
3. **Erayo, Soratami Ascendant**, banned in Commander; **Curse of the Fire Penguin**, an Unhinged card; and the **Role-token flip records**, which are tokens. The engine's Role tokens come from token templates and are unaffected.
4. **Naming a flip card's alternative name with characteristics (CR 201.4c's second sentence).** `CardNameMatches` already accepts either half's name; a name choice restricted by characteristics ("a nonland card name") judging the alternative half by its own characteristics is the same gap the modal DFC case has, and waits on the card that asks.

## Consequences

**Cards.** Up to 19 Commander-legal flip cards become writable; Rune-Tail, Kitsune Ascendant, the card the issue waits on, needs nothing else.

**Cost.** One `Card` field (snapshot-carried), one file in `internal/game` (`flip.go`), one in `cards/effects`, one event and log kind, two branches in `face.go` (`SetFace`, `FaceKeywordsApply`), two in `copy.go`, one line in `MoveCard`, one graveyard-return refactor, an image fallback in `cards/index.go`, and a rotation in the client. No protocol field, no schema bump.

**The thing most likely to be got wrong later.** Reading `ActiveFace != 0` as "transformed". For a flip card it means flipped, which is not transformed: it keeps its colour and cost, it can't be turned back, and it is not a double-faced card. Code that means "transformed" must ask the layout too, as `manaCostForValue` and `faceForClaimLocked` already do.

## Questions for the owner (answered 2026-10-13)

1. **How is "flipped" stored?**
   - **A.** No new field: `ActiveFace == 1` on a flip card is the status. Cheapest; every printed flip card plays right. A flipped permanent that becomes a copy of another flip card shows the unflipped half (CR 110.5c, 707.3 wrong), and a flipped permanent turned face down and back up comes back unflipped.
   - **B.** A `Card.Flipped` status bool (CR 110.5), with `ActiveFace` following it for flip cards (decision 1). One snapshot-carried field more; the copy and face-down cases are right.
   - **Recommendation: B.** The rules model flipped as a status like tapped, which is already a field, and the copy rules (CR 707.3) need the status to outlive the face. The cost is one bool and its snapshot plumbing, which the drift tests enforce.
   - **Decision:** (B) — a `Card.Flipped` status field, with `ActiveFace` following it for flip cards.

2. **How does a flipped permanent look on the board?**
   - **A.** The card image turned 180°, as in paper, so the flipped half reads right side up; tapped is 270°. The hover overlay shows it the same way, with the top half as the inset.
   - **B.** The image stays upright, with a "Flipped" badge and the flipped half's name as the tile label.
   - **C.** Both: rotate, and a small badge for anyone who does not know flip cards.
   - **Recommendation: A.** It is what the physical card does, it needs no new UI element, and the flipped half's name, type and P/T are then readable on the art itself. The hover panel's text already names the half that is up.
   - **Decision:** (A) — the image turned 180° (270° when tapped), on the board and in the hover overlay.

3. **What ships in the first implementation PR?**
   - **A.** The engine (decisions 1–8) and all 19 cards in one PR.
   - **B.** The engine with three proof cards that each exercise a different path: Rune-Tail (a state trigger, the card the issue waits on), Akki Lavarunner (a self trigger, and keywords that change on the flip), and Homura (entering flipped, decision 6). The other 16 follow as a card batch issue, the ones marked "check" in decision 8 audited there.
   - **C.** The engine and Rune-Tail only.
   - **Recommendation: B.** Three cards cover every engine path this ADR adds, keep the engine PR reviewable, and the card batch is mechanical once the verb exists.
   - **Decision:** (B) — the engine with Rune-Tail, Akki Lavarunner and Homura; the other 16 cards follow as a card-batch issue.
