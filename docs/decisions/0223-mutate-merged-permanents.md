# ADR 0223 — Mutate: merged permanents, over or under (CR 702.140, CR 730)

**Status:** Accepted · 2026-11-18 · S82 — Card coverage: sets outside Standard, newest first
**Issues:** [#5910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5910) (seam: mutate). Card: Everquill Phoenix, skipped in RPR batch 5 ([#5908](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5908)). Mutate is Ikoria's main mechanic ([#5512](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5512)), so this seam unblocks that set.
**Owner decisions:** none needed. Every choice below is fixed by the rules or reuses an existing ADR's seam.
**Numbering:** after `git fetch --all --prune`, `git ls-tree --name-only <ref> docs/decisions/` over all 222 remote branches. The highest number anywhere was **0222**, so this is **0223**.
**Builds on:** [ADR 0145](0145-meld.md) (meld: one `Card` carrying the cards it is made of, and the two exits that split it), [ADR 0141](0141-bestow.md) (bestow: an alternative cost that targets, and a spell that resolves as a creature when the target is illegal), the copy model of `game/copy.go` (S16.5, ADR 0043: copiable values materialised onto the flat printed fields, `PrintedSelf` restored on leaving), and `game/copy_grants.go` / `game/triggers_of_key.go` (#665, ADR 0093, #5683: a composite catalog key whose parts lend another card's rows).

---

## Context

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, `data/cr/cr.txt`):

- **CR 702.140a:** "Mutate [cost]" means "You may pay [cost] rather than pay this spell's mana cost. If you do, it becomes a mutating creature spell and targets a non-Human creature with the same owner as this spell." It follows the rules for alternative costs.
- **CR 702.140b:** as a mutating creature spell begins resolving, if its target is illegal, it ceases to be a mutating creature spell and continues resolving as a creature spell, put onto the battlefield under its controller's control. (Also CR 608.3b.)
- **CR 702.140c:** if the target is legal, the spell doesn't enter the battlefield. It merges with the target and becomes one object represented by more than one card or token (CR 730). The spell's controller chooses whether it goes on top or on the bottom. The result is a mutated permanent. (Also CR 608.3d.)
- **CR 702.140d:** "whenever a creature mutates" triggers when a spell merges with a creature as a result of a resolving mutating creature spell.
- **CR 702.140e:** a mutated permanent has all abilities of each card and token that represents it. Its other characteristics come from the topmost card or token.
- **CR 702.140f:** an effect that refers to or modifies the mutating creature spell refers to the mutated permanent it merges with.
- **CR 730.2:** merging places the object on top of or under the permanent. **730.2a:** the merged permanent has only its topmost component's characteristics; this is a copiable effect with the timestamp of the merge. **730.2b:** the merging object leaves its zone but the permanent has not just entered the battlefield. **730.2c:** it is the same object as before: not newly controlled, and continuous effects keep applying. **730.2d:** it is a token only if its topmost component is a token.
- **CR 730.3:** when a merged permanent leaves the battlefield, one permanent leaves and each component goes to the new zone. **730.3a:** in a graveyard or library its owner may arrange them in any order. **730.3c:** an effect that finds the new object finds all of them.
- **CR 903.3c:** if a player's commander is a component of a merged permanent, the merged permanent is that player's commander. **903.9c:** a merged commander sent to the command zone takes only its commander card there.

### What the engine had

Every piece of mutate already exists for another keyword, but not together:

- **Meld (ADR 0145)** taught the engine that one permanent can be several cards: the melded object is one `Card` that carries its cards in `MeldedFrom`, and `MoveCard` splits it wherever it lands but the battlefield or the stack, with the other cards landed by the two exits (`landMeldPassengersLocked`).
- **Bestow (ADR 0141)** is an alternative cost whose `Targets` is a clause, re-checked as the spell begins resolving, and a spell that resolves on without it when the target is illegal.
- **Copy effects** materialise copiable values onto the flat printed fields (`setPrintedValues`), stash the object's own in `PrintedSelf`, and put them back when the permanent leaves (`restorePrintedSelf`). `CopiableValuesOf` reads the flat fields, so a copy of a copy copies what it became.
- **Composite catalog keys** (`copy_grants.go`) let an object's text be its own card's entry plus named parts. A `triggers-of:<key>` part (#5683) lends another card's triggered rows, stamped with the lender's own catalog identity, so a harvested trigger resolves and restores under the lender's key.

## Decisions

### 1. A mutated permanent is the target's own `Card`, carrying every component in `MergedFrom`

The permanent the mutating spell merges with keeps its `InstanceID`, epoch, controller, counters, damage, attachments, timestamps and every continuous effect on it (CR 730.2c). It gains `Card.MergedFrom []Card`: **every** component, topmost first. The entry whose `InstanceID` is the permanent's own stands for the card the permanent was before it first merged (its copiable values at that moment, so a Clone that was copying something counts as what it copied); every other entry is a card or token that merged in, as it sat on the stack.

`MergedFrom` is a separate field from `MeldedFrom` because the two derive their characteristics differently (the melded back face against the topmost component) and their mana value differently (the sum of front faces against the topmost component's). The exits share one code path (decision 5).

### 2. The characteristics are a copy effect: the topmost component plus every other component's abilities

CR 730.2a calls the merge a copiable effect, and the engine already has the shape for one. After each merge `applyMergedValues` writes onto the permanent, through `setPrintedValues` and with `PrintedSelf` stashed exactly as `applyCopy` does:

- the topmost component's copiable values (CR 702.140e, 730.2a);
- its token status: the permanent is a token only if the topmost component is one (CR 730.2d), unlike an ordinary copy, which keeps the copier's;
- for each other component, one composite-key part `grant:merged:<that component's catalog key>` appended to `GrantedAbilities`, and its printed `Keywords` unioned into the permanent's (CR 702.140e: "all abilities of each card").

The part is a derived catalog key like #5683's `triggers-of:` part, but it lends the whole definition's abilities: triggered, activated, mana, static (including the keyword static), replacement, and the other ability slots `mergeCatalogParts` already merges. Triggered rows keep the lender's catalog stamp, so a trigger from a card underneath resolves and restores under that card's own key with no new effect key. `mergeCatalogParts` merges replacement effects for this kind of part only (the `MergedComponent` mark on the definition), so no existing grant changes.

Consequences that need no code of their own:

- **Copies (CR 707.2):** `CopiableValuesOf` reads the flat fields, so a Clone of a mutated permanent copies the topmost component's characteristics and every component's abilities, and is not itself merged (`MergedFrom` is not a copiable value).
- **Ability removal (CR 613.1f):** the parts are the object's own text, so Humility removes them with the rest.
- **Leaving (CR 400.7):** the battlefield-leave path's `restorePrintedSelf` puts the permanent's own card back, which is the carrier of decision 5.

### 3. Casting: an ordinary alternative cost that targets

`effects.Mutate(cost)` is an `AlternativeCost` keyed `mutate` (`game.MutateKey`) whose `Targets` is "target non-Human creature" with a new data gate, `TargetSpec.OwnedBySourceOwner`: the candidate's owner must be the spell's owner (CR 702.140a's "with the same owner as this spell", which is not always the caster). It is data judged in the shared target walk (`sourceAdmits`), so the announce check, the CR 608.2b re-check, the client's picker and the bot agree.

A stack item is a mutating creature spell while `AltCost == "mutate"` and it still has its target.

### 4. Resolution: illegal target, the over/under choice, the merge

- **CR 702.140b:** before the CR 608.2b fizzle check, a mutating spell whose target is illegal drops its target and resolves on as an ordinary creature spell (bestow's shape: `mutateTargetIllegalLocked`, `endMutateOnStackLocked`).
- **CR 702.140c's choice:** after the spell's effect, the controller is asked "over or under" with an ordinary option pick (`mutate/over-under`). The resolution tail (`resolution_tail.go`, #3613) holds the rest of the resolution behind it, and the answer is written onto the held stack item (`StackItem.MutateUnder`). An unanswered prompt (the chooser left) leaves it on top. Option pick prompts are already a restore point and already rendered by the client, so the choice needs no new protocol.
- **The merge (CR 730.2):** `finishSpellResolutionStageLocked` takes the card off the stack and merges it instead of entering it. No zone-move, enter or leaves-the-stack-for-the-battlefield event fires (CR 730.2b), so "enters" and "whenever a creature enters" do not trigger. `EventMutate` fires with the permanent as `CardID` and the spell's card as `Source`, which is what "whenever this creature mutates" reads (CR 702.140d); the trigger is harvested after the merge, so a trigger printed on the card that just went under (Everquill Phoenix) triggers.
- **CR 903.3c:** the permanent is a commander while any component is one.

### 5. Leaving: every component goes, one permanent left

The meld exits generalise to "a permanent made of several cards" (`componentsOf`): `MoveCard` splits a merged permanent wherever it lands but the battlefield and the stack. The carrier is the permanent's own card (the component with its `InstanceID`), which keeps the object's ID and gets the one zone-move and leaves-the-battlefield event (one creature died); every other component keeps its own ID, starts a new object (CR 400.7) and gets its own zone-move event from `landMeldPassengersLocked` (CR 730.3c). Pushed bottom first, so the topmost component ends on top. A merged permanent of exactly two components asks CR 730.3a's order question with meld's prompt; with more, the order is the merged order. A token component that lands off the battlefield ceases to exist through the ordinary state-based action. A merged commander's command-zone redirect sends the other components on to the zone the move was going to (CR 903.9c), as for meld; a commander that went under is a passenger and is owed CR 903.9a's question through its own `CommanderReturnDue`, and the carrier gets its own commander status back.

### 6. Snapshot and wire shape (additive)

- **Snapshot:** `cardSnapshot.MergedFrom []cardSnapshot` (`mergedFrom`, omitted when empty) and `stackItemSnapshot.MutateUnder bool` (`mutateUnder`, omitted when false). Both are additive within the current `SnapshotSchemaVersion`; no fixture changes. The composite key parts ride in `GrantedAbilities`, which is already carried.
- **Closure ratchet:** `Card.MergedFrom` is classified `rebuilt` (`closure_fields.txt`), as `Card.MeldedFrom` is; components carry no card-instance ability slices (they are cleared as they merge, as meld clears them), so no census count rises.
- **Wire:** `CardView.merged_from` (omitted when empty): each component as a `CardFaceView` (name, type line, cost, P/T, image), topmost first, hidden on a face-down card like `melded_from`. `LogMutate` (`mutate`) is the log line "<creature> mutated: <card> went on top" / "… went under it" (`Label` the merging card's name, `Amount` 1 for under). Old clients ignore both fields.
- **Client:** the hover zoom names the components ("mutated: A over B"), as it does for a melded permanent. The over/under choice is an option pick, which the client already renders.

## Out of scope (follow-ups, none needed by Everquill Phoenix)

- A **copy of a mutating creature spell** (CR 707.10, 702.140f) resolves as an ordinary token creature spell copy rather than merging.
- **Face-down components** (CR 730.2e–g): merging over a face-down permanent leaves it face down.
- A **duration copy effect** already on the target (Mirage Mirror) is re-derived over the merged values on the next layer pass.
- Mutating onto a **melded** permanent, and **stickers** on components (CR 123.5b–c).
- CR 730.3a's order question for **three or more** components; CR 730.3b's exile timestamp order.
