# ADR 0137 — Craft: materials from two zones, the return from exile transformed, and the link to what was exiled

**Status:** Accepted · 2026-10-08 · S58 — Deck requests, October batch
**Issues:** [#2124](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2124) (the seam). Deck requests: Tithing Blade for the Mishra deck ([#2033](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2033)); Clay-Fired Bricks and Unstable Glyphbridge for the Sami Whammy deck ([#2190](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2190)). Tracker [#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077).
**Numbering:** the AGENTS.md §4 sweep on 2026-10-08 (every remote head, `git log --all -- docs/decisions`) found 0135 as the highest number on any branch; 0136 is claimed by a parallel builder. This ADR takes **0137**, reserved for it in the #2124 claim comment.
**Builds on:** [ADR 0020](0020-activated-abilities.md) (the exile-a-permanent cost component, its 2026-10-03 amendment, Decisions 52–54, and exile-this from the battlefield, Decisions 45–46), [ADR 0079](0079-transforming-a-permanent.md) (the exile-and-return verb, Decision 5, and the graveyard return, Decision 10), and [ADR 0100](0100-delve-either-or-and-variable-sacrifice-costs.md)'s delve link (`CastProvenance.Delved`), whose shape the craft link copies.

**Why a new ADR rather than an amendment.** Craft is one keyword with three parts, and they belong to three owners: the material cost is ADR 0020's component, the return is ADR 0079's verb, and the link to the materials is new per-object state that neither ADR owns. An amendment to either would describe a third of the keyword and point at the other two. The decisions below each name the ADR whose machinery they extend.

---

## Context

Craft (CR 702.167a) reads:

> Craft with [materials] [cost] means "[Cost], Exile this permanent, Exile [materials] from among permanents you control and/or cards in your graveyard: Return this card to the battlefield transformed under its owner's control. Activate only as a sorcery."

Three facts follow, and the engine had none of them:

1. **The materials span two zones (CR 702.167b).** A material named without the word "card" ("Craft with artifact", "Craft with two creatures", "Craft with Island") may be a permanent you control or a card in your graveyard, and one payment may mix them. `game.ExilePermanentsCost` (#1600) exiled permanents only; `game.ExileCost` exiles cards from one pile only.
2. **The return is from exile, to the back face, as a new object.** `ExileAndReturnTransformedForEffect` (ADR 0079 Decision 5) does both halves to a permanent in one effect, and #1900 added the graveyard form. Nothing returned a card the COST had already exiled.
3. **The crafted permanent may refer to its materials (CR 702.167c):** "the exiled cards used to craft it" are the cards in exile that were exiled to pay the craft cost. Jadeheart Attendant gains life equal to one's mana value; Wretched Bonemass, Sunbird Effigy, Locus of Enlightenment and The Grim Captain read them too.

About 24 Commander-legal cards have craft. Their material clauses come in five shapes:

| Shape | Cards |
|---|---|
| One material of a card type ("Craft with artifact / creature") | Tithing Blade, Braided Net, Clay-Fired Bricks, Unstable Glyphbridge, Jade Seedstones, Idol of the Deep King, Oteclan Landmark, Inverted Iceberg, Spring-Loaded Sawblades, Lodestone Needle, Master's Guide-Mural, Dire Flail |
| One material of a subtype ("Craft with Island / Cave") | Waterlogged Hulk, Kaslem's Stonetree |
| N of a card type ("two creatures", "six artifacts") | Visage of Dread, Tetzin |
| "One or more", an announced count | Altar of the Wretched, Sunbird Standard, Saheeli's Lattice, Paleontologist's Pick-Axe |
| A rule over the set, or graveyard only | Eye of Ojer Taq ("two that share a card type"), Throne of the Grim Captain ("a Dinosaur, a Merfolk, a Pirate, and a Vampire"), The Enigma Jewel ("four or more nonlands with activated abilities"), Ore-Rich Stalactite ("four or more red instant and/or sorcery cards", graveyard only) |

## Decisions

### 1. The materials are the exile-a-permanent component with a graveyard half (extends ADR 0020 Decision 52)

`game.ExilePermanentsCost` gains two fields:

- **`FromGraveyard bool`** — CR 702.167b's second zone. The candidate walk (`ExilePermanentsOptionsForEffect`) lists the activator's matching permanents and then the matching cards in the activator's OWN graveyard; the validator accepts either (`graveyardMaterialLocked`); the payer moves either through the one exit primitive with `MustSettleNow`, exactly as before. A card in another player's graveyard never pays, whoever controlled it last.
- **`Subtype string`** — a material named by a subtype ("Island", "Cave"), read through `Card.HasSubtype`, so a changeling is every creature type and a land an effect made an Island is one.

A graveyard card is judged on its characteristics there, which for a double-faced card are its front face's (CR 712.8a): a Visage of Dread in a graveyard is an artifact card, not a creature card, though its back face is a creature. A test pins that.

Not a new component. The view (`exile_permanent_options`), the enumerator, the validator, the payer and `PaidCost.Exiled` are the #1600 ones, so the picker, the bot and the engine still read one list (#544). The wire grows no field: the options list may now hold graveyard card IDs, after the permanents, and the client resolves them against the battlefield and the viewer's own graveyard (`exilePermanentOptions`, `client/src/lib/sacrificeCost.ts`).

**Refused at boot** (`effects.Register`): `FromGraveyard` on a mana ability (no printed mana ability has it, and the auto-tapper already refuses the component), and `FromGraveyard` beside an `ExileCards` component on one cost (both could then name the same graveyard card; no printed cost has both).

**The source is never a material.** It pays "Exile this permanent" (`ExileSelf`, ADR 0020 Decisions 45–46), and one object pays one component (CR 118.3), so every craft clause sets `ExcludeSource`. The reminder text's "another artifact" says the same thing for the artifact clauses.

**The bot** prices a graveyard material as fuel (`fuelValue`, the price the exile-N-cards cost charges) and a permanent material as before (`permanentValue`). Before this, an id in `exile_permanent_ids` that was not on the battlefield was priced at zero.

### 2. The return is a new verb, `ReturnCraftedFromExileForEffect` (extends ADR 0079 Decisions 5 and 10)

The card the cost exiled keeps its instance ID in exile (ADR 0020 Decision 45), so the ability reads it back by `item.SourceCardID`. The verb:

1. does nothing unless that card is still in exile, is not a token (CR 111.8), and can transform. A copy of a craft card that is not itself double-faced stays in exile, as ADR 0079 Decision 5 leaves a permanent it cannot transform;
2. sets the back face **while the card is in exile**, as ADR 0079 Decisions 5 and 10 do and for their reason (the CR 614 pipeline finds the entering card by ID in its source zone, so a back-face self-replacement must already be readable there);
3. returns it through the ordinary exile entry (`enterBattlefieldThroughPipelineLocked` with `entryTail.newObject`), so what enters is a NEW object (CR 400.7) — summoning sick, no counters, a fresh timestamp, every enters ability on the back face fires — under its **owner's** control (CR 702.167a), whoever activated;
4. puts the front face back if the entry was cancelled or redirected, so no card sits in exile showing its back. A paused entry keeps the back face, because the resume needs it.

`effects.Craft(label, mana, materials)` is the whole keyword: `Plus(ManaCost(mana), ExileThis(), materials)`, `SorcerySpeed`, and the return as a package-level effect. `CraftWith(cardType)`, `CraftWithN(n, cardType)` and `CraftWithSubtype(subtype)` build the materials.

A commander exiled by the craft cost, as the source or as a material, is offered the command zone by the CR 903.9a state-based action before the ability resolves (ADR 0115). If its owner takes it, the source does not return, or that material is not linked, which is what "the card is no longer in exile" means.

### 3. The link to the materials is `Card.CraftedWith []ObjectRef` (CR 702.167c)

The materials are linked as the objects they are in exile — `ObjectRef{ID, Epoch}`, delve's shape (`CastProvenance.Delved`) — and resolved with `Game.CraftMaterialsForEffect`, which is `DelvedCardsForEffect` under the craft name: the cards still in exile as those objects, in the order named. A material that left exile and came back is a new object nothing links to (CR 400.7).

- **Written** by the entry finisher from `entryTail.craftedWith`, after the CR 400.7 reset and before `EventETB`, so the crafted permanent's own enters trigger (Jadeheart Attendant) finds it on the inline path and the resumed one alike.
- **Only what is still in exile is linked.** At resolution the verb drops a material that has left exile and a token; CR 702.167c names "cards in exile that were exiled to pay the cost".
- **Cleared** by `MoveCard` on the way off the battlefield and by the new-object reset, like `Provenance`.
- **Carried** by clone, the snapshot (`craftedWith`, additive within schema v7), and the CR 608.2h record (`PermanentInfo.CraftedWith`), so a crafted permanent's trigger that resolves after it has left reads its last-known link. `effects.CraftMaterials(ctx)` reads it through the source object.
- **Not on `Provenance`.** Crafting is not casting, and `Provenance` is "how was the spell that became this permanent cast" (CR 400.7d).

Known edge, accepted: the link is taken from `PaidCost.Exiled` at resolution rather than at payment, so a material that left exile and was exiled again under the same instance ID between activation and resolution would count. Nothing in the catalog moves a card out of exile and back at instant speed in one window, and recording refs at payment would add a `PaidCost` field for that case alone.

## Cards

Eleven ship `full`: **Tithing Blade**, **Visage of Dread**, **Clay-Fired Bricks**, **Braided Net**, **Jade Seedstones** (the CR 702.167c reader), **Idol of the Deep King**, **Oteclan Landmark**, **Inverted Iceberg**, **Kaslem's Stonetree** and **Waterlogged Hulk** (the subtype clause), and **Spring-Loaded Sawblades**.

## Out of scope, stated

On the craft row's Waiting list, each with its blocker:

- **"One or more"** — Altar of the Wretched, Sunbird Standard, Saheeli's Lattice, Paleontologist's Pick-Axe. An announced count, the #1213 variable-sacrifice question one component over; three of them also read the materials' power, colours or copiable values.
- **A rule over the set** — Eye of Ojer Taq ("two that share a card type"), Throne of the Grim Captain (four different subtypes, then puts an exiled creature card onto the battlefield attacking), The Enigma Jewel ("four or more", and the back face gains the materials' activated abilities). `TargetSpec.EachOf` (ADR 0020 Decision 55) is the likely shape for the Throne.
- **Graveyard only, variable** — Ore-Rich Stalactite.
- **A different blocker on another face** — Unstable Glyphbridge (its back face's "they can't attack you … this turn" and "each opponent who attacked you … can't cast spells"; `GrantCantAttackPlayerForEffect` covers only the attacker's next turn), Lodestone Needle (the explore keyword action is not implemented), Master's Guide-Mural ("activate only if an artifact entered the battlefield under your control this turn" needs a per-turn tally by card type), Dire Flail (an Equipment granting a triggered ability with a reflexive "when you do"), Tetzin, Gnome Champion (six artifacts is expressible; the back face transforms another double-faced artifact and the front reads "double-faced artifact" entries).
