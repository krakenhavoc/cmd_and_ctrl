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

### 9. The word-changing half (not in this ADR's PR; shipped by the 2026-10-15 amendment below)

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

---

## Amendment 2026-10-15 — changing one word: colour words, basic land types and creature types (#2779)

**Status:** Accepted · 2026-10-15 · Backlog. Ships §9 as a mechanism, with a list of what it does not reach.
No new ADR number: this is the half §9 planned.

### The rules

- **CR 612.2:** a text-changing effect changes only words used in the right way: a colour word as a colour
  word, a land type as a land type. It never changes a name ("you can't Sleight proper nouns … a Black
  Vise"; Magical Hack can't change "Island Fish Jasconius"), and the cards' rulings add that it "alters all
  occurrences of the chosen word in the text box and the type line", that a word can't be changed to
  itself, that the words are chosen on resolution, and that a spell that becomes a permanent keeps the
  change.
- **CR 612.3:** effects that add or remove abilities do not change text, so a granted ability is never
  translated, and "it only changes what is printed on the card … it will not change any effects that are on
  the permanent."
- **CR 305.6:** a land with a basic land type has that type's mana ability, so a Forest changed to an Island
  taps for {U} (Magical Hack's ruling says exactly that). Its supertype stays: it is still basic.
- **CR 400.7 / 400.7a:** the change is the object's; a permanent spell hands it to the permanent.
- **CR 608.2b:** a spell whose targets are now illegal under its changed text does nothing.

### Decisions

**A1. The record.** A second layer-3 mod, `game.ModChangeText` (`"changeText"`, reader `readerText`),
carrying exactly one `TextWordChange{Kind, From, To}` in `Mod.Words`, where `Kind` is `color` (stored as the
WUBRG letter), `basicLandType` (CR 305.6's five) or `creatureType` (CR 205.3m). `Normalised` refuses a word
of the wrong class and a word changed to itself, at registration and at restore (`textWordsModProblem`).
The record is pinned to the changed object: a permanent by its entry stamp, or a spell on the stack by
ADR 0104's stack pin, which `repinSpellControlLocked` moves onto the permanent the spell becomes (CR
400.7a). `stackPinProblem` admits the kind on a spell pin. The duration is the card's: indefinite, or "until
end of turn" (Crystal Spray, Whim of Volrath). Plain data, so the closure ceiling is unchanged and a
restore point holding one is a restore point.

**A2. The result is materialised.** `materialiseTextWordsLocked` composes every record on an object in
timestamp order (CR 613.7) into the derived `Card.TextWords`, for battlefield permanents and stack spells,
at the top of every recompute (after the text boxes) and when the sweep drops a record. Composition is a
substitution: after "black → blue" then "blue → red", the printed black AND the printed blue read red; a
word that comes back to itself drops out. `MoveCard` clears `TextWords` on every zone change (CR 400.7);
the snapshot does not carry it (the drift table says rebuilt).

**A3. Who reads a word, and how.** The words are arguments, not text, so each reader that holds a word
consults the object whose text the word is in:

1. **The type line and printed keywords, in the layer pass.** `layerPassWithLocked` runs
   `TextWords.applyToBaseline` on each permanent's fresh baseline, before any layer-4 or layer-6 effect:
   basic land types and creature types in the subtypes (an Island Swamp under "Swamp → Island" is one
   Island), and the keyword tokens "protection from <colour>", "hexproof from <colour>" and basic landwalk.
   The catalog's printed-keyword row (layer 6, `Spec.PrintedKeywords`) reads through
   `PrintedKeywordsHonouringCopy`, which now translates too, so a catalog card's protection follows. Name and
   colour are not words in the text and do not move. A face-down permanent has no text and is skipped.
2. **Target clauses, through the source.** A target predicate is handed the caster and the candidate, never
   the object whose clause it is, and changing `CardOK`'s signature would touch hundreds of card files. So
   the targeting walk lends the CANDIDATE COPY its source's words (`TargetSource.lend`, `game.ThroughTextOf`
   setting the unexported `Card.wordLens`), and the effects package's word predicates read through it:
   `OfColor` (and so `NonBlack`) with `Card.HasColorWord`, and `Subtype`, `AnySubtype`, `OfSubtype`,
   `HasSubtype` and `OfCreatureType` with `Card.HasSubtypeWord`. The legal set, the announce check (CR
   601.2c) and the resolution re-check (CR 608.2b) are one walk, so they agree. The lens lives only on a
   value copy handed to a predicate; it is never stored, so it needs no snapshot, clone or race story (the
   view reads under the read lock concurrently and never writes a card).
3. **Effects that judge cards against their own clause.** `MatchingBattlefield` (every "destroy / exile /
   bounce all <predicate>") and `CounterAllMatching` lend the resolving spell's or ability source's words
   (`Context.textReader`).
4. **Statics that hold their source.** `Card.TextColor`, `TextSubtype` and `TextKeyword` answer a word as the
   source's text now reads it. `TribeFilter.Matches` reads a lord's printed tribes through it (a CHOSEN tribe
   is not text and stays), `KeywordGrant` the keyword it grants (Goblin King's mountainwalk grant becomes an
   islandwalk grant under Magical Hack), and Sylvan Anthem, Shield of the Oversoul and Tidal Influence their
   colour.

**A4. The choice.** No new prompt kind. The colour and land-type words are two ordinary `option_pick`
prompts: the word to replace (Mind Bend and its kin offer the five colours and the five land types
together), then the other words of the same class. A creature type is the `choose_creature_type` prompt
twice; for "the new creature type can't be Wall" the prompt gained `PendingChoice.ExcludedTypes` (snapshot
`excludedTypes`, additive within v9): the resolver refuses an excluded type, the wire's `type_options`
leaves it out, and the enumerator drops it and never offers nothing (it falls back to the first allowed
type). The replaced word is always excluded from the second prompt. The continuations are closures, so a
table waiting on one is not a restore point, as for a pile split; the answer comes within the resolution.
Crystal Spray's draw runs after the words are chosen, in printed order.

**A5. Wire.** `CardView.text_changes` lists the changed words ("black → blue", colours first), on a
permanent and on a spell; the inspector shows it beside the printed oracle text, which still uses the old
words. Cleared with the ability rows for a viewer who cannot see the card.

### What the substitution does not reach (honest coverage)

A word is translated only where it passes through one of the doors in A3. These do not:

- **Colours hard-coded in a closure.** About 30 card files call `Card.HasColor` with a literal colour (or a
  colour inside an `OfColor` evaluated where no lens is lent): triggers and conditions such as Ayara, First
  of Locthwain, Heroism, Inquisitor's Snare, Kefnet's Monument, Marang River Prowler, Merchant Scroll,
  Nightscape Familiar, Oran-Rief, Pyroblast, Soot Imp, Sorin of House Markov, Trusted Advisor, Veil of
  Summer and Goblin Anarchomancer, the cost-modifier helper `ColoredSpell`, and every `WheneverYouCast(OfColor
  …)` trigger (its helper judges the cast spell without lending it the source's words). `grep -l 'HasColor("'
  server/internal/cards/effects/*.go` lists them. A file that reads a CHOSEN colour (Lurebound Scarecrow,
  Case of the Shattered Pact) is correctly untranslated: a choice is not text.
- **Land types hard-coded in a closure.** About 30 files call `HasSubtype("Island")` (or another basic land
  type) directly rather than through a predicate: "if you control an Island" conditions, Seasinger, and
  the land-count readers. The predicates in A3.2 cover target clauses and sweeps only.
- **A creature type inside a protection quality** ("protection from Goblins"): the plural has no reliable
  singular round trip, so `TextWords.keyword` leaves it.
- **A departed source's text.** A target re-check against a source that has left reads its last-known
  characteristics, which do not record `TextWords`; the printed words stand.
- **A spell's own characteristics on the stack.** The baseline is translated in the battlefield layer pass
  only, so a creature spell's type line on the stack keeps its printed creature type until it resolves (its
  target clauses do follow, through A3.2).
- **Mana symbols** are not words (CR 612.2), so "{T}: Add {B}" never changes, which is correct.

Each of these is a card-file migration onto the A3 doors, not a new seam: a card that should change reads
`source.TextColor("W")` instead of `"W"`. They are left as they are rather than migrated blind, because each
needs its own oracle check that the literal really is a colour word of that card's text.

### Cards

Glamerdye (retrace), Sleight of Mind, Alter Reality (flashback), Magical Hack, Mind Bend, Crystal Spray,
Whim of Volrath (buyback) and Artificial Evolution ship **Full**. The rest of the original list wait on other
seams, each filed with its own row: Spectral Shift (entwine, #3866), Trait Doctoring (cipher, #3867), New
Blood (tapping a creature as an additional cost, #3868) and Balduvian Shaman (testing for and granting
cumulative upkeep, #3869). The `text-changing-effects` row is implemented.

### Alternatives considered

- **Pass the source to every predicate** (`CardPredicate` gaining a parameter). Correct, and the cleanest
  read, but it touches every predicate and every card file that writes one, for a family of a dozen cards.
  The lens reaches the same predicates with one change each.
- **Pull the candidate's colours back through the substitution** instead of translating the word. It works
  for "is it black?", but breaks every colour reader that is not a colour word ("colorless",
  "multicolored", a colour count): a blue creature under "blue → black" would read as colourless.
- **A goroutine-free "current source" field on `Game`** set during the walk. The view enumerates targets
  under the read lock concurrently, so a shared field is a data race.
