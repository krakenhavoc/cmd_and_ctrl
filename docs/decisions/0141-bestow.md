# ADR 0141 — Bestow: a creature cast as an Aura

**Status:** Accepted · 2026-10-09 · S58 — Deck requests, October batch
**Issues:** [#2862](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2862) (seam: bestow). Deck request: Nighthowler in Mill Me Mad ([#2065](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2065)). Registry row: `bestow`.
**Owner decisions:** none needed. The rules fix every behaviour below. The modelling choices (a bit on the alternative cost, a flag on the card, a layer-4 effect on the battlefield) have no product effect and are explained where they are made.
**Numbering:** after `git fetch --all`, no branch on the remote and no local branch or worktree has an ADR above 0140, and no open pull request adds one. This ADR takes **0141**.
**Builds on:** [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) and S22's `AlternativeCost` (an offer claimed at announce, read back off the stack item), [ADR 0036](0036-attachments.md) (the attachment relation, the Aura attach on resolution, the CR 704.5m/n state-based action), [ADR 0082](0082-casting-face-down-and-turning-face-up.md) and [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) §4 (an offer that changes what the spell IS: `FaceDown`, `CastsFace`), ADR 0036 decision 21 (reconfigure's "not a creature while attached", a layer-4 effect from the board), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic prices what a cast does).

The seam lands in one PR: the rules, the snapshot field, the bot's price and six cards. The client needed no code; a test pins that it offers the cast.

---

## Context

Nighthowler shipped with a caveat: "Bestow is not available — the Nighthowler can only be cast as a creature." It blocks deck request #2065. Bestow is on 43 Commander-legal cards in the 2026-09-24 Scryfall dump, and the engine had no way to cast a creature card as an Aura spell.

### The rules

Read in the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026):

- **CR 702.103a:** "Bestow [cost]" means "As you cast this spell, you may choose to cast it bestowed. If you do, you pay [cost] rather than its mana cost." It follows the rules for alternative costs (CR 601.2b, 601.2f–h).
- **CR 702.103b:** "As a spell cast bestowed is put onto the stack, it becomes an Aura enchantment and gains enchant creature. It is a bestowed Aura spell, and the permanent it becomes as it resolves will be a bestowed Aura. These effects last until the spell or the permanent it becomes ceases to be bestowed." Being an Aura spell, it needs a legal target (CR 601.2c, 303.4a).
- **CR 702.103c:** a copy of a bestowed Aura spell is a bestowed Aura spell too.
- **CR 702.103d:** "When casting a spell bestowed, only its characteristics as modified by the bestow ability are evaluated to determine if it can be cast." "Creature spells can't be cast" does not stop it; "you may cast creature spells from the top of your library" does not allow it.
- **CR 702.103e:** "As a bestowed Aura spell begins resolving, if its target is illegal, it ceases to be bestowed and the effect making it an Aura spell ends. It continues resolving as a creature spell." **CR 608.3b** says the same from the resolution side: a bestowed Aura spell with an illegal target becomes a creature spell and resolves as CR 608.3a describes, where any other spell with an illegal target does not resolve.
- **CR 702.103f:** "If a bestowed Aura becomes unattached, it ceases to be bestowed. If a bestowed Aura is attached to an illegal object or player, it becomes unattached and ceases to be bestowed. This is an exception to rule 704.5m."
- **CR 702.103g:** a bestowed Aura that phases in unattached ceases to be bestowed.
- **CR 205.1a** (an effect that makes an object a new card type replaces its card types) and **CR 205.3d** (an object can't have a subtype that doesn't match one of its types): a bestowed Nighthowler is an "Enchantment — Aura", not a Horror.

### What the engine had

- Alternative costs (`game.AlternativeCost`): a price, plus clauses a keyword staples to it. `Targets` rewrites the target clause (cleave, awaken), and the announce gate, the CR 608.2b re-check, the view's per-offer legal set and the bot enumerator all read the rewrite through `TargetSpecUnderAlternativeCost`. `FaceDown` and `CastsFace` change what the spell is, and `CastFaceOf` gives the view and the enumerator the object CastSpell will put on the stack.
- Auras: an Aura is cast targeting, and `attachResolvedAuraLocked` attaches the permanent to the target as it lands, before `EventETB`. `attachmentLegalLocked` reads the Aura's own enchant clause (its catalog `Targets`), and `attachmentSBALocked` puts an illegal or unattached Aura into its owner's graveyard (CR 704.5m).
- CR 608.2b: `spellAllTargetsIllegalLocked` counters a spell whose every target is illegal, permanent spells included. Nothing let a spell continue past it.
- Layers: the pass runs over the battlefield only. Off the battlefield the type readers (`HasCardType`, `HasSubtype`, `printedShared`) take a cold path off the printed fields, with a branch for a face-down object. Reconfigure (#2639) added a layer-4 "not a creature" effect derived from the board.

## Decision

### 1. Bestow is an alternative cost with one more bit

`AlternativeCost.Bestow bool`, built by `effects.Bestow(cost)` with the key `bestow` (`game.BestowKey`), the label "Bestow {cost}", the cost, and `Targets: EnchantCreature()`. The enchant creature clause is an ordinary target clause, so the announce gate demands a creature target, the resolution re-check judges it, the view ships it as the offer's `target_mode` / `legal_targets`, and the enumerator makes one move per creature. Nothing in those four paths changed.

It is the third clause after `FaceDown` and `CastsFace` that changes what the spell is, and it follows them: CastSpell sets the state right after the claim is resolved, before the cast-path check, the timing gate and the cast restrictions read the card (CR 702.103d), and `CastFaceOf` sets it on the copy the view's timing stamp and the bot enumerator judge.

### 2. `Card.Bestowed`: the spell and the permanent it becomes

One bool on `Card`. CastSpell sets it on its working copy and on the card on the stack. It means "this object is a bestowed Aura spell or a bestowed Aura".

- `MoveCard` clears it on **every** zone change, beside `stackGranted`, because a new object was never cast bestowed (CR 400.7). A countered bestowed spell is an ordinary creature card in its owner's graveyard, and a bounced bestowed Aura an ordinary card in hand. The CR 400.7 reset in `resetAsNewObjectLocked` clears it too.
- The battlefield entry of a resolving spell reads it off the stack card before the move (`landEntryLocked`) and sets it on the permanent (`seedBestowedEntryLocked`), after the CR 400.7 reset and before `announceEntryLocked`'s Aura attach and `EventETB`. So the attach finds an Aura, and "whenever a creature enters" never sees one.
- It is not a copiable value (`CopiableValuesOf` reads the printed fields). A Clone of a bestowed Aura is an enchantment creature.

### 3. What a bestowed object is (CR 702.103b)

`Characteristic.becomeBestowedAura`: lose Creature and every creature subtype (`loseCreatureType`, reconfigure's), gain Enchantment and Aura. Supertypes are kept (Kestia stays legendary). P/T and abilities are the card's.

- **On the stack** there is no layer pass, so the cold path applies it: `printedShared`, `hasCardType`, `HasSubtype` and `HasAllCreatureTypes` read `bestowBaseline` for a bestowed card, as they read the CR 708.2 body for a face-down one. A bestowed spell is not a creature spell to a counterspell's target clause, a cast trigger or a cast restriction, and the stack view's type line reads "Enchantment — Aura".
- **On the battlefield** it is a layer-4 effect derived from the board, `bestowContinuousEffectsLocked`, one more source list in `activeStaticAbilitiesLocked` beside reconfigure's. Its timestamp is the permanent's entry: the effect began as the spell went on the stack, and the entry is the earliest the permanent has, so a later type-changing effect still applies on top. The layer pass's own baseline (`printedFresh`) does not apply it.

"Gains enchant creature" is not a keyword the engine lists. It is the offer's target clause at announce, and `bestowedAttachmentLegalLocked` once attached: legal attached to a creature on the battlefield, illegal attached to anything else or to nothing. Protection is asked first, by the code every attachment already goes through.

### 4. Resolution: CR 702.103e and 608.3b

`resolveTopOfStackLocked` asks `bestowTargetIllegalLocked` before the CR 608.2b fizzle: a bestowed spell whose target is illegal (the same `spellAllTargetsIllegalLocked` re-check) is not countered. `endBestowOnStackLocked` clears the flag on the stack card and the resolver's copy and drops the item's targets, so the fizzle finds nothing to judge, the entry seeds nothing, and the attach finds no target. The spell resolves as a creature spell and enters unattached (CR 608.3a). Every other spell is judged exactly as before.

### 5. Unattached: CR 702.103f and 702.103g

- `attachmentLegalLocked` answers false for a bestowed Aura attached to an illegal object or to nothing. `attachmentSBALocked` records the doomed attachment as bestowed, clears the flag (`unbestowLocked`, which bumps the layer version), unattaches it and does **not** route it to the graveyard: the exception to CR 704.5m. The flag is cleared before `EventUnattach` goes out, so a listener sees the enchantment creature it now is. A 0/0 then dies to CR 704.5f on the next pass of the same state-check loop.
- Between the host leaving and that check, the link still points at the host (ADR 0036's rule: nothing sweeps the reverse direction). `bestowedNowLocked` already treats the Aura as unattached when its host is gone (CR 701.3d), as reconfigure does, so it is a creature from that moment.
- `UnattachForEffect` clears it too.
- CR 702.103g needs nothing more: a phased-out permanent is off the battlefield slice, and one that phases in with no host is attached to nothing, which the same state-based action handles.

### 6. Copies (CR 702.103c)

A copy of a spell copies the card, flag included, so a copied bestowed spell is bestowed on the stack and CR 702.103e applies to it. `resolvePermanentSpellCopyLocked` seeds the flag on the token the copy becomes before the Aura attach it already performs.

### 7. The snapshot

`cardSnapshot.bestowed`, additive within schema 7 and omitted when false, so every frozen fixture re-encodes byte for byte. A file without it restores no object as bestowed, which is every game before this change. The drift test classifies `Card.Bestowed` as carried. No stack-item field was added: the stack card carries the flag.

### 8. The wire and the client

No new field. The offer reaches `CardView.alternative_costs` like cleave's, with the enchant creature clause as its `target_mode` / `legal_targets`, so the client's cast picker lists "Bestow {cost}" beside the printed cost, and `castTargetOverride` opens the target picker on the creature. The stack view shows the Aura's type line, and the battlefield shows the Aura attached. `client/src/lib/bestow.test.ts` pins the chain.

### 9. The bot

The enumerator offers the creature cast and one bestowed cast per creature (`legal/bestow_test.go`). The heuristic prices a bestowed cast's target in `bestowTargetsValue` instead of `targetsValue`, which reads an opposing creature as removed:

- onto its own creature: the card's value is already counted (the Aura half of every bestow card gives about its own body), plus `Config.BestowShare` (0.25) of it for surviving the host as a creature, plus the SickCreature discount back when the host can attack now;
- onto an opponent's creature: twice the card's value off, below passing.

So the bot bestows onto its own creature when it has one and the mana, and never onto an opponent's. The baseline config prices bestow at 0, as before.

## Cards

| Card | What it proves |
|---|---|
| Nighthowler | the deck card: "this creature and enchanted creature each get +X/+X" (`PumpSelfCreatureOrAttachedPer`), the host's death leaving a creature whose X counts the host, Full with the caveat gone |
| Eidolon of Countless Battles | the same shape counting creatures and Auras, the bestowed Aura counting itself as an Aura and then as a creature |
| Boon Satyr | flash on a bestow card, and CR 702.103e: its target gone, it resolves as a 4/2 creature |
| Celestial Archon | `PumpAttached` and `GrantToAttached` on a bestowed host |
| Hopeful Eidolon | a keyword grant, and a bestowed Aura that is not a creature |
| Nyxborn Rollicker | the plainest one, and the enumerator's two offers |

The other 37 Commander-legal bestow cards are card work. Several need their own seam: Detective's Phoenix (collect evidence, and a bestow cast from the graveyard), Hypnotic Siren (control of the host), Nyxborn Hydra (an X bestow cost), Springheart Nantuko (a token copy of the host).

## Tests

- `cards/effects/bestow_test.go`: every card is Full with the offer; Nighthowler bestowed is an Aura spell on the stack with no creature types, attaches, is not a creature and pumps the host by X; its host destroyed, it stays as a creature whose X counts the host; Eidolon of Countless Battles counts itself as an Aura and then as a creature; Boon Satyr whose target is destroyed in response resolves as an unattached 4/2; the creature cast is untouched; the bestowed cast needs a creature target; Celestial Archon's and Hopeful Eidolon's grants reach the host.
- `legal/bestow_test.go`: the creature cast and a bestowed cast onto each creature are enumerated, and every one is accepted.
- `protocol/bestow_view_test.go`: the hand card's bestow offer carries the enchant creature clause with the creature in it; a bestowed spell's stack type line is "Enchantment — Aura".
- `aiseat/heuristic/bestow_test.go`: bestowing onto its own creature beats the creature cast, onto an opponent's is below passing, and the baseline is unchanged.
- `client/src/lib/bestow.test.ts`: the picker offers both casts, the bestow cast targets with the offer's clause and the creature cast with none, and the payload names the key.
- The snapshot shape file, the drift test, the oracle fixtures for the five new cards, the coverage mechanic `bestow` and the roadmap tables carry the rest.

## Consequences

- A creature card can be cast as an Aura spell and come back as a creature, which no card could do before.
- The layer pass gathers one more source list. It walks the battlefield once and returns nothing at a table with no bestowed Aura.
- Every cold-path type read checks one more bool, after the face-down check it sits beside.
- `attachmentSBALocked`'s Aura branch has one exception, keyed on the flag.

## Out of scope

- **The other bestow cards** (Cards above).
- **CR 303.4d for an animated bestowed Aura** ("an Aura that's also a creature can't enchant anything"). The engine does not enforce that rule for any Aura, and no card in the catalog animates an Aura.
- **Showing "enchant creature" as an ability** on a bestowed Aura. The type line and the attachment say it; no effect in the catalog reads the enchant keyword itself.
