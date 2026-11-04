# ADR 0218 — A spell copy that gains a triggered ability

**Status:** Accepted · 2026-11-03 · S79 — Card coverage: every Standard set, newest first
**Issue:** [#4688](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4688) (this seam), found during [#4456](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4456) (Secrets of Strixhaven)
**Numbering:** checked on 2026-11-03 after `git fetch origin`. `develop` ends at 0215. `git ls-tree` over every remote head's `docs/decisions/` (the 44 heads are listed in the PR) shows 0216 on `docs/4870-planechase-adr` (#4872) and 0217 on `docs/4880-advanced-stats-adr` (#4882), and nothing above 0217. The files of every open PR show nothing higher either. This is **0218**.
**Builds on:** [ADR 0043](0043-copy-effects.md) (copy effects, its except clause, and its #666 amendment: a copy of a permanent spell becomes a token), [ADR 0083](0083-token-abilities.md) (abilities on a token, and why they are keys rather than closures), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (the layer-6 grant, which is the one this is not).

---

## Context

Choreographed Sparks' second mode reads "Copy target creature spell you control. The copy gains haste and 'At the beginning of the end step, sacrifice this token.'"

Issue #4688 described the gap this way: `CopySpell.Except` edits `game.PrintedValues`, which can add a keyword such as haste but has nowhere to put a triggered ability, so the token the copy becomes can't have the sacrifice trigger. The issue also left one question open: are the haste and the trigger copiable values of the copy, so that a Clone of the token copies them?

A delayed trigger ("sacrifice it at the beginning of the next end step") would be the wrong shape. The printed rider is an ability of the token. A token that loses all its abilities loses the rider, and a copy of the token gets the rider only if it is a copiable value.

## The rule

The pinned Comprehensive Rules (20260925) settle the question:

- **CR 707.10:** a copy of a spell copies the spell's characteristics and the choices made for it.
- **CR 707.9 and 707.9a:** a copy effect may modify the copying process, and "Some copy effects cause the copy to gain an ability as part of the copying process. This ability becomes part of the copiable values for the copy, along with any other abilities that were copied."
- **CR 707.2:** the copiable values are the printed values "as modified by other copy effects".
- **CR 608.3f:** a resolving copy of a permanent spell becomes a token. The token is not "created" for the purposes of any rule or effect that refers to creating a token.

"The copy gains haste and '…'" is one copy effect that gains two abilities as part of the copying process. So CR 707.9a makes both of them copiable values of the copy. The token the copy becomes has the copy's copiable values, and a later copy effect copies them (CR 707.2).

## Decision

1. **Haste and the trigger are copiable values of the copy.** The token keeps them. A Clone of the token, or a token copy of it, copies them too: the Clone has haste and is sacrificed at the beginning of the next end step. A token that loses all its abilities loses both, so it is not sacrificed. Spider-Verse's "If the copy is a permanent spell, it gains haste" (#4460, `copyGrant`) is not like this. That haste is an effect given to the copy once the copy exists (CR 611.2), not part of the copying process, so it is not copied. The difference is in the wording: "the copy gains" inside the copy instruction is part of the copying process, and a separate sentence after the copy is made is an effect.

2. **No new engine seam is needed. The existing CR 707.9a grant already does this.** `PrintedValues.GrantedAbilities` (#665, `game/copy_grants.go`) was built for "except it has '<ability>'" on a permanent that enters as a copy (Phantasmal Image, Sakashima the Impostor, Lazav, Dimir Mastermind). It reaches spell copies with no change:
   - `copySpellFromLocked` applies `Except` to `CopiableValuesOf(src)` and writes the result back with `setPrintedValues`, so the copy on the stack carries the grant's name.
   - `tokenCopyOfSpell` reads the copy's `CopiableValuesOf`, so the token carries the grant.
   - `CatalogKey` makes the token's key composite (`<oracle id>|grant:<name>`), so the trigger harvester finds the granted trigger like any other.
   - `CatalogAbilityKey`'s ability-removal rule (CR 613.1f) drops the grant along with every other ability of the object's own.
   - `CopiableValuesOf` copies `GrantedAbilities`, so a Clone gets the grant.
   - The grant is a name in the snapshot, so it survives a restore whether the copy is on the stack or is already a token.

   The issue's premise was out of date: `PrintedValues` has had a place for a granted triggered ability since #665. Nothing in the engine needed to change. The decision in point 1, the tests, and the card are what close #4688. A card file writes the clause as `Except: func(v *game.PrintedValues) { v.AddKeyword("haste"); v.GrantAbility("<card>/<what>") }`, and declares the bundle in `Spec.Grants`.

3. **"The end step" is the next end step that comes, on any player's turn.** The card's 2026-03-20 ruling says the word "next" was removed, so the trigger is `AtEachStep(StepEnd)`, as Lightning Serpent's is. A token made during an end step after that step's beginning stays until the next end step.

4. **The grant is data.** The bundle is static catalog data, registered by the card that grants it (`Spec.Grants`, `choreographed-sparks/sacrifice-this-token`). The copy and the token carry only its name. No func field is reachable from `Game`, the snapshot shape does not change, and the closure census does not move.

## Consequences

- Choreographed Sparks ships as Full with no caveats.
- Any later "copy target spell. The copy gains '<triggered ability>'" card has a recipe to follow. Its grant key is an on-disk identity, the same as every other grant key: never renamed, never reused.
- A copy effect whose gained ability is not part of the copying process (Spider-Verse's shape) keeps using `copyGrant`. A card file chooses between the two from the printed wording, as point 1 describes.
