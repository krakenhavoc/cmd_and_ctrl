# ADR 0219 — An ability a resolved effect gives an exiled card

**Status:** Accepted · 2026-11-05 · S79 — Card coverage: every Standard set, newest first
**Issue:** [#4973](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4973) (Star Trek). Q, Stacking the Deck was left out of batch 17 (#5003) for this clause.
**Numbering:** checked on 2026-11-05 after `git fetch origin '+refs/heads/*:refs/remotes/origin/*'`. `git ls-tree` of `docs/decisions/` on every remote head (48 heads; the loop in [README.md](README.md#picking-a-number)) shows 0217 and 0218 taken and nothing above 0218. This is **0219**.
**Builds on:** [ADR 0093](0093-abilities-granted-to-other-permanents.md) (the grant bundle, its refs, and the 2026-10-13 amendment that gives a static's grant to cards in a hand, graveyard or exile), [ADR 0066](0066-granted-cast-and-play-permissions.md) (the card-copy amendment: "copy that card, and you may cast the copy"), [ADR 0041](0041-game-persistence.md) (additive snapshots, the closure ratchet).

---

## Context

Q, Stacking the Deck: "When Q enters, shuffle your library, then exile cards from the top of your library until you exile a nonland card. That nonland card gains 'Exile a card from your hand: Create a copy of this card. You may cast the copy. Activate only during your turn and only if this card is exiled.'"

The engine had two ways for an object to have an ability it does not print, and neither fits:

- **A layer-6 grant from a static** (ADR 0093), and since 2026-10-13 a static's grant to cards in a hand, graveyard or exile (`zone_grants.go`). Both are re-derived on every layer pass from a source that is still on the battlefield. Q's grant comes from a trigger that has resolved. Nothing is left to re-derive it from, and Q leaving the battlefield must not end it.
- **`Card.GrantedAbilities`**, the CR 707.9a "except it has '…'" name. It is a copiable value (CR 707.2), so a copy of the card would have it. A gained ability is not a copiable value, and the copy this very ability makes must not have it.

The pinned rules (20260925):

- **CR 113.10:** "Effects can add or remove abilities of objects. An effect that adds an ability will state that the object 'gains' or 'has' that ability, or similar."
- **CR 611.2a:** a continuous effect from a resolution with no stated duration "lasts until the end of the game."
- **CR 400.7:** "An object that moves from one zone to another becomes a new object with no memory of, or relation to, its previous existence." So the effect stops applying when the card leaves exile. A card that comes back to exile is a new object without the ability.
- **CR 113.6b:** "An ability that states which zones it functions in functions only from those zones." "Only if this card is exiled" is that.
- **CR 707.12** and **CR 704.5e:** the copy is created in exile and cast during the resolution; a copy of a card that is not cast ceases to exist.

## Decision

1. **`Card.ExileGrants []GrantedAbility` holds the grant on the exiled object.** It is the same plain data a zone grant writes (bundle key, grantor, activator). There is no closure and no pointer, so nothing new is reachable from `Game` and the closure census does not move. The one writer is `Game.GrantAbilityToExiledCardForEffect(card, source, activator, key)`, which refuses a card that is not in exile (CR 608.2b) and an empty key.

2. **One reader, no new path.** `abilityGrantsOf` returns a card's zone grants followed by its `ExileGrants`. `activatedAbilityRows` already reads it, so the activation path, the legal-move enumerator, the wire's `zone_abilities` row (with `granted_by`) and a restored stack item's `grant:<bundle>:<i>:<n>` ref all see the row with no change of their own. The bundle is an ordinary exile-zone `AbilityGrant` (`Zone: ZoneExile`, its activated row `Zones: [exile]`), so `checkZoneGrantBundle` holds it to activated abilities that function from exile.

3. **Keyed to the object.** `MoveCard` clears it on every zone change, beside `SuspendGranted` (the #4455 precedent for an effect that gives an exiled card something). A card that leaves exile loses the ability, and coming back does not restore it. It is not a copiable value, so `newCardCopyLocked` never copies it.

4. **Persisted, additively within v9.** `cardSnapshot.exileGrants` is `[{key, source?, activator?}]`, omitted when empty. A file written before it restores as "has only its printed abilities", which is the state before the field existed. A binary from before it that reads a newer file drops the key, and the exiled card loses the ability. That is the same trade #4455 made for `suspendGranted`. The drift census classifies the field as carried.

5. **No wire change.** An exiled card's activatable rows already reach its owner on `zone_abilities` (#1221), and the client already offers them from the zone browser. The row's `granted_by` names Q. `docs/protocol.md` says so in one sentence.

## Consequences

- Q, Stacking the Deck ships Full with no caveats.
- Any later "that card gains '<activated ability>'" on an exiled card follows the same recipe: an exile-zone bundle in `Spec.Grants` and a call to `GrantAbilityToExiledCardForEffect` from the effect's continuation. Its bundle key is an on-disk identity: never renamed, never reused.
- Only exile is supported. A resolved effect that gives a card in a hand or graveyard an ability would widen the writer's zone check and the doc comment, nothing else. No printed card needs it yet.
- A triggered or static ability gained by an exiled card is still not modelled. `checkZoneGrantBundle` refuses it at boot.
