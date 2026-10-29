# ADR 0210 — Putting a card onto the battlefield as a different type

**Status:** Accepted · 2026-10-29 · S76 — Cards: incubate, face-down and outside the game.
**Issue:** [#4506](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4506) (found during The Hobbit, #4454).
**Numbering:** checked on 2026-10-29. `develop` ends at 0209. `git ls-tree` over every remote head's `docs/decisions/` shows no file above 0209; the heads checked are listed in the PR. Reserved on #4506.
**Builds on:** [ADR 0041](0041-game-persistence.md) phase 3 (`ScopedEffect` records), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (granted bundles), [ADR 0104](0104-gaining-control-of-a-spell.md) (the epoch pin taken as a permanent lands), [ADR 0109](0109-rule-gates-land-types-mana-and-cost-components.md) §10 (the CR 614.12 look-ahead), [ADR 0154](0154-auras-that-enchant-a-card-in-a-graveyard.md)'s 2026-10-23 amendment (a creature that returns as an Aura).

---

## Context

Two Hobbit cards put a card onto the battlefield with its type line set by the effect that puts it there:

- **Supper for Spiders** {1}{B} instant: "Put onto the battlefield under your control all creature cards in your opponents' graveyards that were put there from the battlefield this turn. They are Food artifacts with '{2}, {T}, Sacrifice this artifact: You gain 3 life.' (They lose all other types and subtypes.)"
- **Tom, Bert, and William**: "When Tom, Bert, and William die, if they were a creature, return them to the battlefield. They're an artifact. (They're no longer a creature.)"

The rules (Comprehensive Rules, 2026-09-25):

- CR 614.1c: "Effects that read '[This permanent] enters with . . . ,' 'As [this permanent] enters . . . ,' or '[This permanent] enters as . . . ' are replacement effects."
- CR 614.12: to decide which replacement effects apply to a permanent entering, "check the characteristics of the permanent as it would exist on the battlefield, taking into account replacement effects that have already modified how it enters the battlefield".
- CR 205.1a: an effect that sets an object's card type replaces its card types, and the subtypes correlated with a removed card type go with it. CR 205.4b: supertypes are independent and stay.
- CR 707.2: a copy takes copiable values. Type-changing effects are not copied.
- CR 400.7: an object that changes zones is a new object.

What the engine had:

- **The type change itself.** `SetTypesMod` (layer 4, CR 205.1a's set; The Irencrag, #2562) and `GrantAbilitiesMod` (layer 6, ADR 0093), both `ScopedEffect` mods.
- **"Put there from the battlefield this turn".** `putIntoGraveyardFromBattlefieldThisTurn`, read by Twilight Shepherd, Gerrard's Hourglass Pendant and Cry of the Carnarium.
- **A type change registered after the return.** The Enduring cycle (`glimmer_return.go`) registers `RemoveTypesMod("Creature")` once the card has landed. Between the landing and that registration the card is a creature, and `EventETB` goes out in that window, so "whenever a creature enters" sees it.
- **A type change set as the card lands.** Harold and Bob's Aura form (`Card.ReturnedAura`) is set before `EventETB`. It is read through the card's own catalogued statics, so it serves only a card that prints its own other form.

Supper for Spiders can return any creature card, catalogued or not. The type it gets comes from the spell, not from the card.

## Decision

### 1. `ZoneEntryOptions.EntersAs`

The putting effect says what the card enters as:

```go
type EntersAs struct {
    Types    []string // the card types it has; at least one
    Subtypes []string // each one of Types' (CR 205.3d)
    Grants   []string // ADR 0093 bundle keys
    Label    string
}
```

It is plain data. The batch door (`PutOntoBattlefieldTogetherKeyedForEffect`, and `effects.ReturnFromGraveyardTogether` above it) seeds it onto every card's entry event, as it seeds `Tapped`. A batch with an unsound value panics when it begins, so the first test that uses it fails.

### 2. Registered as the permanent lands

The entry landing (`entry_choice.go`) registers one `ScopedEffect` with `SetTypesMod(Types, Subtypes)` and, when there are grants, `GrantAbilitiesMod(Grants...)`. It does this after the CR 400.7 reset and before the zone move and `EventETB`, beside `seedReturnedAuraEntryLocked`. The two mods share one record and one timestamp, so they are one effect (CR 613.7).

The record's affected object and duration are both pinned to the permanent by its `ObjectEpoch` (`PinObjectByEpoch`, `PinnedToEpoch`). That is the pin ADR 0104 takes for a stolen permanent spell as it lands, before the zone move stamps the entry. There is no stated duration (CR 611.2a). The record lasts as long as that object is on the battlefield, and the sweep collects it once the object leaves. A flicker or a reanimation is a new object without it, and a copy copies the printed card.

The card keeps its own abilities. Neither card says it loses them.

### 3. The entry window sees it (CR 614.12)

- **The dry run.** `entryCharacteristicsLocked` adds the overwrite as a record pinned to the dry permanent, at the dry permanent's entry stamp. So the look-ahead's keyword reads see the permanent as it will be.
- **Replacements that read the entering card.** While a window for an event that carries an overwrite gathers its replacements (`gatherReplacementsLocked`) or re-checks one (`stillAppliesLocked`), `Game.entryView` holds the dry run's characteristics for the entering card. `LookupCardForEffect` answers for that card, in its source zone, with them. An opponent's Authority of the Consuls ("creatures your opponents control enter tapped") therefore does not tap a Food.
- **Clearing.** The view is set and cleared around the one call, like `counterPlacerScope`. It is never set between actions, and it is neither snapshotted nor cloned.

### 4. The snapshot

`entersAs` is written on a paused entry's event (`replacementResume.event.entersAs`) and on its batch's options (`entryTail.batch.opts.entersAs`), both additive in v9 and absent when nil. A table paused inside such an entry, such as a Clone that Supper returns asking what to copy, is a restore point.

An older binary drops the key, and the card enters as itself. That is the same rollback cost `entersAsAura` accepted in v9. The record that lands is an ordinary `ScopedEffect` made of existing mod kinds, so a table with a Food on it needs nothing new.

No route from `Game` to a func or an interface is added. `entryView` holds a `Characteristic` value and is classified `dropped` in `snapshot_drift_test.go`.

### 5. Tom, Bert, and William

The dies trigger reads "if they were a creature" from the last-known `Types` it is handed (CR 603.4), as the Enduring cycle does. It returns the card only if the card in the graveyard is still the object that died (`sameGraveyardObject`). The return is `EntersAs{Types: ["Artifact"]}`, so Troll goes and Legendary stays. When the artifact dies, it was not a creature, the trigger does not fire, and there is no loop.

## Cards

- Supper for Spiders: Full.
- Tom, Bert, and William: Full.

## Tests

`server/internal/cards/effects/enter_as_different_type_test.go` drives each card through its catalog entry:

- **What Supper takes.** Only opponents' creature cards put into a graveyard from the battlefield this turn: not a milled card, not last turn's, not yours, not a noncreature card.
- **What comes back.** Noncreature Food artifacts under the caster's control. Soul Warden does not trigger. A legend stays legendary. A 0/0 survives state-based actions. The Food ability gains 3 life.
- **The entry window.** Authority of the Consuls neither taps the Food nor gains life from it.
- **A new object.** A token copy is the printed creature. A flicker returns a creature, which Soul Warden sees, and the old record is swept.
- **Restores.** A restore while a Food is on the table. A paused Clone entry, restored and answered, gives a copy that is a Food, and so does a second card of the same batch.
- **Tom, Bert, and William.** They return once, as a legendary noncreature artifact that Wrath of God does not touch. Destroyed as an artifact, they stay in the graveyard and the record is gone. Reanimated by something else, they are the Troll creature again. The sacrifice ability works on the artifact.

## Alternatives considered

- **Register the record after the return**, as the Enduring cycle does. Rejected: `EventETB` goes out while the card is still a creature.
- **A `Card` flag like `ReturnedAura`.** Rejected: it would need a catalogued "other form" on every card that can be returned. Supper's types come from the spell.
- **Rewrite the catalog's entry predicates to read a look-ahead.** Rejected: there are about 44 of them. Answering through `LookupCardForEffect` for the span of the gather reaches all of them.

## Out of scope

- **The Enduring cycle** (`glimmer_return.go`) still registers its `RemoveTypesMod` after the return, so "whenever a creature enters" sees a returning Glimmer. Moving it to `EntersAs{Types: ["Enchantment"]}` is a follow-up.
- **"Can't enter" checks** (CR 614.17d) are unchanged. They are not part of the replacement gather.
