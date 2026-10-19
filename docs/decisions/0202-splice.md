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

- **A splice cost that isn't mana** (#4055): Veil of Secrecy, Horobi's Whisper, Roar of Jukai, Hundred-Talon Strike and Torrent of Stone stayed out of the catalog until the 2026-10-19 amendment below built it.
- **New targets for a copy's spliced text.** A copy keeps the spliced text's targets; CR 707.10c's "may choose new targets" prompt walks the main text's steps only. A retarget ("change the target of target spell") changes the main text's targets only. Both are fewer choices than the rules allow, never more.
- **Splice cards that aren't catalogued yet** for other reasons (Through the Breach, Reweave, Shifting Borders, Fell Beast's Shriek): each is an ordinary card PR now.

## Consequences

- Overblaze, Glacial Ray, Desperate Ritual, Goryo's Vengeance, Candles' Glow, Kodama's Might, Wear Away, Consuming Vortex, Spiritual Visit, Vital Surge, Dampen Thought, Soulless Revival, Lifted by Clouds, Hideous Laughter, Strange Inversion, Psychic Puppetry, Blessed Breath, Into the Fray, Evermind, Everdream and Splicer's Skill join the catalog, all Full.
- The "Splice onto Arcane" seam closes; "A splice cost that isn't mana" opens (#4055).
- `TestSpliceMatchesThePrintedSplice` (`cards/coverage`) holds every declared splice to the printed quality and cost, and fails a catalogued card that prints splice and declares none.

## Amendment (2026-10-19, [#4055](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4055)): a splice cost that isn't mana

"Not covered" named five Kamigawa cards whose splice cost is a verb: Veil of Secrecy (return a blue creature you control to its
owner's hand), Horobi's Whisper (exile four cards from your graveyard), Roar of Jukai (an opponent gains 5 life),
Hundred-Talon Strike (tap an untapped white creature you control) and Torrent of Stone (sacrifice two Mountains). CR 702.47a:
"Paying a card's splice cost follows the rules for paying additional costs in rules 601.2b and 601.2f-h."

**Decisions.**

1. **`SpliceAbility.Cost *SpliceCost`** is the non-mana half, beside `ManaCost` (never both on a printed card). `SpliceCost` is
   built out of the components the activated abilities and the other additional costs already pay, so their walks, validators and
   payers are reused and the view, the bot enumerator and the engine cannot disagree about which object pays (#544): a
   sacrifice clause (`Sacrifice`), an exile-from-graveyard cost (`ExileCards`), a tap-others cost (`TapOthers`), a
   return-to-hand cost (`ReturnToHand`), and the one new component, `OpponentGainsLife`. `SpliceAbility.Label` prints the
   cost after the dash with its full stop ("Splice onto Arcane—Sacrifice two Mountains."), which is what
   `TestSpliceMatchesThePrintedSplice` holds the card to. `effects.Register` (`checkSplice`, `checkSpliceCost`) refuses a cost
   that is empty, both mana and not, unlabelled, of more than one component (every printed card has exactly one), a variable
   count, or a graveyard cost from anywhere but a fixed count in the graveyard.
2. **The payment rides its own list, not the cast's flat ones.** `CastSpellParams.SplicePayments []SplicePayment`, on the wire
   `splice_payments` (entry k pays `splices[k]`: `sacrifice_ids`, `exile_ids`, `tap_ids`, `return_ids`, `opponent`). The cast's own
   plan (ADR 0073) walks `discard_ids`, `sacrifice_ids` and `teamwork_ids` in plan order; folding another card's cost into those
   would make the client rebuild that order to say which card an object pays for, and would let a mistake in one card's cost
   change which object pays the spell's. Separate lists make the announcement say what it means. **Rejected:** appending the
   splice costs to `costPlan` as escalate's are (`escalatePayments`): the right shape when the cost belongs to the spell, the wrong
   one for a cost that belongs to a card in the hand.
3. **Validate all, then pay (CR 601.2b, 601.2h).** `validateSplicePaymentsLocked` runs with the cast's other announce checks:
   each payment is judged by the component's own validator (`validateSacrificeCostLocked`, `validateExileCardsCostLocked`,
   `validateTapOthersCostLocked`, `validateReturnToHandCostLocked`), a payment for a component the card lacks is refused rather
   than ignored, and **no object may be named twice**, by two splice cards or by a splice card and the cast's own costs
   (`castSpentObjects`; CR 118.3). That reads fewer choices than the rules allow (a creature could be tapped for one cost and
   then returned for another), never more. `paySplicePaymentsLocked` pays them with the spell on the stack, after the cast's own
   costs, in the order the cards were revealed, so a dies trigger lands above the spell. A commander the cost returns to a hand
   is asked about first (CR 903.9b), and a card an effect has paused on its way out cannot pay, like every other cost card.
4. **An unpayable cost is not offered (CR 118.3).** `SpliceOffersLocked` drops a card whose cost the caster cannot pay
   (`spliceCostPayableLocked`), so a hand never offers Torrent of Stone to a player with one Mountain; the announcement refuses it
   regardless.
5. **"An opponent gains N life."** The caster names the opponent as gift's recipient is named, from the opponents still in the
   game **who can gain life** (`SpliceOpponentsLocked`; CR 614.17b: if an event can't happen, a player can't choose to pay a cost
   that includes it). The gain is `gainLifeAsCostLocked`, the twin of `payLifeAsCostLocked`: it runs the CR 614 window without
   pausing (a cost is one indivisible step), so a doubler changes the amount and still pays, and a gain the window cancels
   refuses the cast.
6. **Nothing is recorded on `PaidCost`.** No printed splice text reads what its own cost paid, and the spell's own "the creature
   sacrificed" and "tapped to pay" readers must not see another card's. The auto-tapper keeps the named objects out of the mana
   it plans (`CastAutoTapExclusions`), a sacrificed Mountain still tapping for mana first as a Village Rites sacrifice may.

**View and client.** `splice_offers[]` gains `cost_label` and the picker for the component the cost has, in the shape of the
additional cost's: `sacrifice_label` / `sacrifice_options`, `exile_cost_n` / `_label` / `_options`, `tap_others_label` /
`_options`, `return_label` / `return_options`, `chooses_opponent` / `opponent_options`. The cost picker shows the cost as printed
where the mana symbol was and asks the opponent with the toggle; the object picks open one card at a time in the order ticked
(the sacrifice, return and tap pickers over the board, the exile picker over the graveyard), then the chain continues. Bots never
splice, so the enumerator is unchanged ([docs/bot.md](../bot.md)).

**Snapshot.** None: the payments are announce-time, and `StackItem.Spliced` carries the same texts as before.

**Cards** (5, all Full): Veil of Secrecy, Horobi's Whisper, Roar of Jukai, Hundred-Talon Strike, Torrent of Stone.

**Tests.** `cards/effects/splice_cost_test.go`: each card spliced onto Consuming Vortex pays its cost and runs its text, with the
refusals paired with an acceptance on the same board; a card is offered only when its cost can be paid; two splice cards never
share an object; payments are aligned with the splices; Register refuses an unbuildable cost. The client's
`spliceCost.render.test.ts` covers the cost label, the opponent picker and the wire.

Back-out: drop the `paySplicePaymentsLocked` call and the Mountains stay on the battlefield; drop the `used` check and one
creature pays two costs; drop `spliceCostPayableLocked` from `SpliceOffersLocked` and Torrent of Stone is offered with one
Mountain; drop the `SpliceOpponentsLocked` test and an opponent who can't gain life can be chosen.
