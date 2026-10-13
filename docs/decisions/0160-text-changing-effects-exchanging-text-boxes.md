# ADR 0160 — Text-changing effects (CR 612): exchanging text boxes is a layer-3 substitution of the ability key

**Status:** Accepted · 2026-10-13 · Backlog (Kuja deck, #3213)
**Issue:** [#2779](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2779) (seam: text-changing effects).
Registry row: `text-changing-effects` in `server/internal/roadmap/registry.go`.
**Numbering:** reserved in the #2779 claim comment. Swept `origin/develop` and every remote head
(30 refs after `git fetch --prune`): the highest file was `0159-per-table-discord-voice-rooms.md`, and no
open issue or PR names 0160.
**Builds on:** [ADR 0012](0012-layer-system.md) (the layer system; its §5 left layer 3 a stub),
[ADR 0043](0043-copy-effects.md) (copiable values, the copy-choice prompt, and its 2026-09-28
amendment's duration copies, materialised before the pass), [ADR 0093](0093-abilities-granted-to-other-permanents.md)
(the composed ability key: own + layered grants), [ADR 0041](0041-game-persistence.md) phase 3
(`ScopedEffect` mods are data), [ADR 0063](0063-durations-and-control.md) (durations and pins).

---

## Context

### The rules

- **CR 612.1:** a text-changing effect changes an object's text, generally its rules text and type line.
- **CR 612.3:** effects that add or remove abilities do not change text, so a GRANTED ability is never
  touched by a text-changing effect.
- **CR 612.5 / CR 701.12h:** Exchange of Words exchanges the text boxes of two creatures. "This replaces
  all of the rules text of each object with the rules text of the other object" and "the rules text of
  each permanent becomes the PREVIOUS rules text of the other."
- **CR 701.12a:** "if the entire exchange can't be completed, no part of the exchange occurs."
- **CR 613.1c:** text-changing effects apply in layer 3: after copy effects (layer 1, including the
  face-down characteristics of 1b) and control (layer 2), before types, colours, abilities and P/T.
- **CR 707.2:** a copy takes the copiable values. "Other effects (including type-changing and
  text-changing effects) … are not copied."
- **CR 603.10a / 603.6a:** a leaves-the-battlefield trigger looks back at the permanent as it last
  existed; an enters trigger looks at the permanent as it exists after the event.

### The cards

- **Exchange of Words** {1}{U}{U} Enchantment: "When this enchantment enters, choose two target
  creatures. For as long as this enchantment remains on the battlefield, exchange the text boxes of
  those creatures."
- **Deadpool, Trading Card** {2}{B}{R} 5/3: "As Deadpool enters, you may exchange his text box and
  another creature's. At the beginning of your upkeep, you lose 3 life. {3}, Sacrifice this creature:
  Each other player draws a card." Shipped in #3252 with the exchange as a caveat.
- The other cards on the row change WORDS, not whole text boxes: a colour word (Glamerdye, Sleight of
  Mind, Alter Reality, Balduvian Shaman, Whim of Volrath), a basic land type (Magical Hack), either
  (Mind Bend, Crystal Spray, Spectral Shift, Trait Doctoring) or a creature type (Artificial Evolution,
  New Blood). That half is §9, and does not ship here.

### What the engine has

A card's rules text is not text in this engine. It is the catalog entry its **key** resolves to: the
triggers, activated and mana abilities, statics, replacements, cost modifiers and printed keywords a
`Spec` declares, all reached through `catalogKeyOf` (`game/effect_hooks.go`), which every reader goes
through. ADR 0093 composes that key with the layer-6 grants into `CatalogAbilityKey`. The deck
importer's keywords (`Card.Keywords`, Scryfall's `keywords` array) are the second road for printed
keywords, merged into the layer-0 baseline by `printedInputsOf`.

Layer 3 has one occupant: `ModSetName` (#2562), a layer-pass mod that writes `Characteristic.Name`.

## Decisions

### 1. A text box is a catalog key plus the importer's keywords

The rules text of an object is exactly two things here: the catalog key its abilities resolve through,
and its import-road printed keywords. A text-box exchange therefore swaps **the key and the keywords**,
and nothing else. Name, types, power, toughness, loyalty, mana cost and colour are read off the flat
printed fields (`Card.Name`, `TypeLine`, `Power`, `ManaCost`, `Colors`), which the exchange does not
touch, so they stay where they were, as CR 612.5 says.

The value is `game.TextBoxValues{Key, Keywords, From}`: `Key` is the other object's `catalogKeyOf`
(face suffix and CR 707.9a copy grants included, since both are part of its rules text), `Keywords` its
import-road keywords, `From` the name of the object the text was printed on, for the wire. An empty
`Key` is a real value: the rules text of a face-down creature, or of a creature with no catalog entry,
is "nothing the engine automates".

### 2. The record is a scoped effect; the result is materialised before the layer pass

An exchange is two `ScopedEffect` records (ADR 0041 phase 3), one per creature, each carrying one
`ModExchangeText` mod (`"textBox"`, layer 3, reader `readerText`) with the text box it now has, BY
VALUE. Values, not a reference to the other creature: CR 701.12h says each gets the other's PREVIOUS
text, so it is read once, when the exchange happens, and the other creature leaving later changes
nothing. Each record is pinned to its creature (ADR 0063), so a creature that leaves takes its record
with it (CR 400.7) and the other keeps its new text for the rest of the effect's duration: while
Exchange of Words remains, or indefinitely for Deadpool.

The record is **not** applied in the layer pass. The pass gathers each permanent's statics from its key
before layer 1 runs, so a text box that changes which statics exist has to be decided first. This is
the same reason ADR 0043's duration copies are materialised onto the flat fields before the pass, and
it is done the same way: `materialiseTextBoxesLocked` (`game/text_box.go`) runs at the top of every
recompute, right after `materialiseDurationCopiesLocked`, when a record is registered, and when the
sweep drops one. It writes the winning value onto `Card.TextBox`, a derived field: rebuilt, never
persisted (the drift table says so), cleared on every battlefield exit.

`catalogKeyOf` answers `TextBox.Key` when the field is set, and `printedInputsOf` (layer 0) and
`printsDevoid` read `TextBox.Keywords` instead of `Card.Keywords`. That is the whole of the substitution:
every catalog reader already goes through `catalogKeyOf`, so the triggers that fire, the activated and
mana rows offered (to the host's controller, since the host is "this creature", ADR 0093 Decision 4),
the statics gathered, the replacements, the cost modifiers, the legal-move enumerator and the wire all
follow with no change of their own. The two instance-carried ability lists (`Card.ManaAbilities`,
`Card.ActivatedAbilities`, a legacy token's) are skipped under a record, as they are under ability
removal.

**Timestamps within layer 3 (CR 613.7).** Each record carries a whole text box, so applying them in
timestamp order leaves exactly the latest one: `textBoxFor` takes the record with the greatest
timestamp, ties to registration order, as `durationCopyFor` does. A second exchange of an already
exchanged creature reads its CURRENT text box (the first exchange's value), which is CR 701.12h's
"previous rules text".

### 3. Ordering against the other layers

- **Layer 1 first.** A duration copy materialises onto the flat fields before the text box is
  decided, and the record wins over it: a creature that becomes a copy of something after its text was
  exchanged keeps the exchanged text (layer 3 after layer 1). An entry copy is the same.
- **A copy OF an exchanged object copies its printed text.** `CopiableValuesOf` reads the flat fields,
  never `Card.TextBox`, so a Clone of an exchanged creature gets that creature's own (or copied) text
  box, not the exchanged one (CR 707.2).
- **Face-down (layer 1b) before layer 3.** A face-down creature has no rules text, so the creature it
  is exchanged with gets none, and it gets that creature's text. The record wins over CR 708.2a's
  silence in `catalogKeyOf` for exactly that reason.
- **Layer 6 after.** `CatalogAbilityKey` is still own + layered grants, with "own" now the exchanged
  key. A grant stays on the object it was granted to (CR 612.3); an ability-removing effect removes
  the exchanged abilities, as it removed the printed ones (#3253's zone grants compose the same way).

### 4. When the exchange happens, and when it does not

`Game.ExchangeTextBoxesForEffect(source, a, b, d, label)` needs both creatures on the battlefield and
distinct, or does nothing (CR 701.12a). It reads both current text boxes, registers both records with
one timestamp, and materialises. It emits no event: nothing in the catalog triggers on a text change.

### 5. Last-known information

`snapshotLKILocked` records `Card.TextBox` in `triggerIdentityLKI`, and the LTB harvest restores it onto
the identity it reads triggers from, so an exchanged creature that dies with "When this creature dies"
in its gained text triggers (CR 603.10a), and one that dies having given its own dies-trigger away does
not.

### 6. Deadpool: an entry choice that exchanges instead of copying

"As Deadpool enters, you may exchange his text box and another creature's" is CR 614.1c's shape, the
one ADR 0043 Decision 3 built for Clone: a replacement with a decision inside it, answered before the
permanent lands. It reuses that prompt: `CopySelector.ExchangeText` makes the selector's pick the
creature to exchange with instead of the object to copy. The prompt is the same `copy_target` kind
with the same `{card_ids}` answer (empty declines), the same enumerator and the same resume, plus one
wire flag, `exchanges_text`, so the client words it as an exchange. On an answer the event carries the
pick, and `settleEntryTextExchangeLocked` registers the exchange once the zone move has stamped
Deadpool's object identity and **before** `EventETB`, exactly where `settleTimedEntryCopyLocked` puts
Cursed Mirror's record. So Deadpool enters with the other creature's text, and an enters trigger in it
triggers for him (CR 603.6a), while the other creature now has Deadpool's upkeep loss and sacrifice
ability. The catalog's `OnETB` hook is read off the live permanent's key for the same reason.

The enumerator offers the prompt's answers unchanged. The heuristic reads `exchanges_text` and scores
an opponent's creature by its value and one of its own at just above declining
(`textExchangeValue`): Deadpool takes the best text an opponent has, and hands his upkeep drawback to
that opponent's creature rather than to one of the bot's.

### 7. Wire

`CardView.text_box_from` names the object whose text box a permanent has, when it is not its own. The
abilities, keyword chips, activated and mana rows, AUTO badge and token text already follow the key.
The client shows the line in the card inspector beside the granted-abilities list.

### 8. Snapshots, undo and Clone

The records are plain data (`Mod.TextBox []TextBoxValues`, strings only) carried by the scoped-effect
registry; `Card.TextBox` is rebuilt from them by the recompute a restore forces. `Clone` shares the
immutable `*TextBoxValues` pointer, as it shares `Card.printed`. No closure is added, so the closure
ceiling is unchanged. A binary that does not know the `textBox` kind refuses the restore point
(`ErrUnknownEffectKey`), as for any other kind.

### 9. The word-changing half (not in this ADR's PR)

Glamerdye's family changes a colour word, basic land type or creature type "everywhere it appears in
the text" of one object. In this engine the words are not text either: they are arguments to
predicates in card files (`NonBlack()`, `Color("black")`, `OfCreatureType("Goblin")`), to keywords
(`protection from black`, `swampwalk`) and to statics' `AppliesTo` closures. The plan:

1. A second layer-3 mod, `changeText`, carrying `{Kind: color | landType | creatureType, From, To}`,
   on a record pinned to the changed object (or to a spell on the stack, ADR 0104's stack pin),
   materialised into a per-object word map `Card.TextWords` beside `Card.TextBox`, in timestamp order.
2. The words become data where they are printed. A predicate that names a colour or type word reads
   it through `ctx.Word(kind, word)`, which consults the SOURCE object's map (the object whose text
   the predicate is part of): targets clauses through `TargetSpec` evaluation, statics through
   `StaticAbility` evaluation with the source in hand. Keywords that carry a word (protection,
   landwalk, the "-walk" family) translate at `HasKeyword`.
3. Every card file that writes a colour or type word as a Go literal inside a closure is invisible to
   it, so the coverage package gets a probe that lists them, and the first slice converts only the
   predicates the waiting cards' likely victims use (Doom Blade's nonblack, the protection and
   landwalk keywords, lords' tribes).

That is a card-file migration, not a seam, and is why it is left out: the colour-word cards stay on the
row as waiting.

## Alternatives considered

- **Apply the exchange in the layer pass.** It would write a field of `Characteristic`, but the pass's
  static gather runs before it, so an exchanged-in anthem would never be gathered. Rejected for the
  reason ADR 0043 materialises copies.
- **Rewrite `Card.OracleID`.** Then name lookups by oracle ID, `CopiableValuesOf` and the printed
  baseline would all move with the text, which is wrong three ways (CR 612.5, CR 707.2).
- **A per-object record that is not a scoped effect.** It has no duration, so Exchange of Words could
  not end, and no snapshot support. The scoped registry already has both.

## Consequences

- Exchange of Words ships **Full**; Deadpool, Trading Card's caveat is gone and he ships **Full**.
- The row narrows to the word-changing cards (§9).
- `catalogKeyOf` grows one branch, taken only by a permanent under a record.

## Out of scope

- Volrath's Shapeshifter ("full text", CR 612.6) and the un-set text-box cards (Phoebe, Animate
  Graveyard, Do-It-Yourself Seraph), none Commander-legal.
- An instance-carried token's abilities (a legacy token whose abilities are not registered under its
  token key) do not travel with its text; every token template in the catalog registers its key since
  #521.
- A copy exception's dropped devoid (`Card.ColorCDADropped`, #2322) is not carried with the text.
