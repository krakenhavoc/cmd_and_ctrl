# ADR 0202 — Splice

**Status:** Accepted · 2026-10-16 · S76 — Cards: incubate, face-down and outside the game
**Issue:** [#1989](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1989) (found landing ADR 0108 PR 2, #1890). Follow-up: [#4055](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4055) (a splice cost that isn't mana).
**Numbering:** every remote head was listed (58 refs, `git for-each-ref refs/remotes/origin/` then `git ls-tree -r --name-only <ref> docs/decisions/`). The highest in use is **0201**, so this is **0202**.
**Builds on:** [ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md) (an additional cost announced at CR 601.2b and priced by the one pricer at CR 601.2f), [ADR 0065](0065-modal-and-multi-target-clauses.md) (a flat target list whose refs name their clause; the bullet walk of its 2026-10-09 amendment) and [ADR 0108](0108-turn-scoped-effects-object-history-and-damage-shields.md) §3 (Overblaze's text).

---

## Context

CR 702.47a: "Splice is a static ability that functions while a card is in your hand. 'Splice onto [quality] [cost]' means 'You may reveal this card from your hand as you cast a [quality] spell. If you do, that spell gains the text of this card's rules text and you pay [cost] as an additional cost to cast that spell.'" The card stays in the hand, and may even be discarded to pay a "discard a card" cost of the spell it is spliced onto.

The rest of CR 702.47 says what the gained text is:

- **702.47b**: a card can't be spliced if its text's required choices (targets) can't be made; a card is spliced onto a spell at most once; several cards are revealed at once, the caster orders their effects, and the main spell's effects happen first.
- **702.47c**: the spell keeps the main spell's characteristics and gains only the rules text. Text that names the spliced card refers to the spell (Glacial Ray spliced onto a blue spell is the blue spell dealing the damage, so it can target a creature with protection from red).
- **702.47d**: targets for the added text are chosen normally (CR 601.2c), and a spell with targets won't resolve if all of them are illegal on resolution.
- **702.47e**: the spell loses the splice changes once it leaves the stack.
- **CR 707.10**: a copy of the spell copies its text and every decision made for it, the spliced text and its targets included.

Nothing in the cast path revealed a second card from the hand, added its cost, or added its text to the spell. Overblaze and Candles' Glow were catalogued as waiting on it; every other splice card was absent.

## Decisions

### 1. A splice card declares the ability; its text is the rest of its row

`CardDef.Splice *SpliceAbility` (`Onto`, `ManaCost`), declared on the card with `effects.SpliceOntoArcane(cost)` or `effects.SpliceOntoInstantOrSorcery(cost)` (Everdream, Splicer's Skill). The card's own `Targets` and `OnResolve` **are** its rules text: the body written for casting the card alone is the body a spell gains. Nothing is written twice.

`effects.Register` (`checkSplice`) refuses a splice with no mana cost (a non-mana splice cost is not built, #4055, and a card without its cost would be stronger than printed), an unknown quality, no body, and a text a spliced spell could not carry whole: modes, an announced X or division, or a cost of its own besides the splice cost. No printed splice card has any of those.

### 2. The announcement names the revealed cards, in effect order

`CastSpellParams.Splices []uuid.UUID` (`cast_spell.splices`) names the hand cards, in the order their effects happen (CR 702.47b). It is checked with the rest of CR 601.2b, after the modes and before the targets, by `announcedSplicesLocked`: each must be a card in the caster's hand other than the spell, named once, whose splice quality the spell being cast has, read off the face-materialised announce copy (so a face-down spell is spliced onto by nothing).

Rejected: an index into `OptionalCosts`. Splice is not the spell's own optional cost but another card's ability, which hand card is revealed is the choice, and two copies of Glacial Ray may both be spliced. A list of IDs says that; an index cannot.

### 3. A spliced text's targets ride the flat list, marked with their text

`TargetRef.Splice` (1-based, `json:"splice,omitempty"`) names the spliced text a ref answers; its `Slot` and `Mode` index that text's own clause list. Zero is the spell's own text, which is every ref that existed before. The wire keeps ADR 0065 §7's one flat `targets` list (`targets[].splice`).

The cast takes those refs out of `params.Targets` before anything reads it, so every check the main text has always had runs exactly as before. Each spliced text's refs are validated against **its card's** clause list, judged from the spell being cast (CR 702.47c). A text with a required target that cannot be given one is refused, which is CR 702.47b's "can't choose to use a splice ability". A "becomes the target" trigger and a spell-targeting tax see the whole list.

### 4. The splice costs join the one price

`spliceManaLocked` adds each revealed card's splice cost in `printedCostLocked`, after the optional costs and Spree's mode costs and before the cost modifiers: CR 601.2f makes it an additional cost, so it stands where ADR 0073 §3 puts kicker. The cast, the tap budget, the auto-tapper, the auto-tap price preview (its `splices` parameter) and the view all read that one function.

### 5. The stack item carries the texts; resolution views each on its own

`StackItem.Spliced []SplicedText` holds each text's card ID (for the view), catalog key (its identity on disk: the body and clause are read from the catalog by it), name, and targets. `StackItem.Targets` stays the main text's, so the ~750 card bodies that read `item.Targets` read what they always did. `StackItem.AllTargets()` is the whole list, read by the CR 608.2b check, the stack view and "a spell with a single target".

At resolution the main text runs, then each spliced text in order (CR 702.47b), against `splicedView`: a shallow copy of the item whose `Targets` and clause list are the text's own. Its source is still the spell (CR 702.47c), so the damage, the shield and the multiplier are the spell's. The spliced texts are parked behind a prompt an earlier instruction opened, with the remaining bullets, in the same walk (`modeWalk.splices`), so Vital Surge spliced onto Psychic Puppetry gains its life after Puppetry's tap-or-untap answer.

**CR 608.2b.** The spell is countered by the rules only when it had targets and every target of every text is illegal. A text whose own targets are all illegal does nothing while another text resolves: no part of it may act on an illegal target, and every instruction of every printed splice card is about its target. A clause that says it resolves anyway (#2182) keeps its text running.

### 6. Copies, undo and snapshots

`createSpellCopyLocked` copies `Spliced` (CR 707.10). `cloneStackItem` deep-copies it; the snapshot carries it (`stackItems[].spliced`, additive within v9); the parked walk carries the texts still to run (`pausedModeWalk.splices`, additive). Nothing new holds a closure.

### 7. The view and the client

The hand card's cast surface gains `splice_offers`: each card in the viewer's hand that could be spliced onto it (`SpliceOffersLocked`: a splice card other than the spell whose quality the spell has and whose required targets can all be chosen), with its label, splice cost and its text's clause judged from the spell. It is per viewer and never public: it names other cards in the hand. A stack item shows `spliced` (the names, revealed as it was cast) and its targets carry `splice`.

The client offers the splice cards as toggles in the cast modal that already asks for the optional costs (the same CR 601.2b question, "what am I paying for this?"), in the order the player ticks them, and adds each ticked text's clauses to the target walk after the spell's own. The revealed cards are shown to the table with the spell on the stack.

### 8. Bots never splice

The enumerator offers no splice. A splice is a second card's whole effect and a price on top; the bot's per-cast budget is already modes × targets × optional costs, and a bot that holds both an Arcane spell and a splice card can still cast each on its own. Written in `docs/bot.md`.

## Not covered

- **A splice cost that isn't mana** (#4055): Veil of Secrecy, Horobi's Whisper, Roar of Jukai, Hundred-Talon Strike and Torrent of Stone stay out of the catalog.
- **New targets for a copy's spliced text.** A copy keeps the spliced text's targets; CR 707.10c's "may choose new targets" prompt walks the main text's steps only. A retarget ("change the target of target spell") changes the main text's targets only. Both are fewer choices than the rules allow, never more.
- **Splice cards that aren't catalogued yet** for other reasons (Through the Breach, Reweave, Shifting Borders, Fell Beast's Shriek): each is an ordinary card PR now.

## Consequences

- Overblaze, Glacial Ray, Desperate Ritual, Goryo's Vengeance, Candles' Glow, Kodama's Might, Wear Away, Consuming Vortex, Spiritual Visit, Vital Surge, Dampen Thought, Soulless Revival, Lifted by Clouds, Hideous Laughter, Strange Inversion, Psychic Puppetry, Blessed Breath, Into the Fray, Evermind, Everdream and Splicer's Skill join the catalog, all Full.
- The "Splice onto Arcane" seam closes; "A splice cost that isn't mana" opens (#4055).
- `TestSpliceMatchesThePrintedSplice` (`cards/coverage`) holds every declared splice to the printed quality and cost, and fails a catalogued card that prints splice and declares none.
