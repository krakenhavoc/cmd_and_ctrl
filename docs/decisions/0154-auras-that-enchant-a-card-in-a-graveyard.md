# ADR 0154 — Auras that enchant a card in a graveyard (Animate Dead)

**Status:** Accepted (owner decisions 2026-10-13); amended 2026-10-13 for #3158, 2026-10-14 for #3159 and 2026-10-14 for #3156 (see the end) · 2026-10-13 · S53 — Faces, zones, and shared payment (ADR first), tracker [#1910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1910)
**Issues:** [#1780](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1780) (seam: an Aura that enchants a card in a graveyard). Registry row: `aura-on-a-graveyard-card` (status `missing`; it changes with the implementation, not with this ADR).
**Owner decisions:** 2026-10-13. The owner accepted the ADR and chose the recommended option, (a), on all six questions. The answers are recorded under [Questions for the owner](#questions-for-the-owner-answered-2026-10-13), and no section above them changed.
**Follow-ups:** [#3156](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3156) (Necromancy), [#3157](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3157) (Spellweaver Volute), [#3158](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3158) (the CR 303.4f prompt for every Aura), [#3159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3159) (a flickered host keeps its attachments: bug, starts with a failing test), [#3160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3160) (CR 701.3b for every Aura), [#3161](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3161) (reanimation-family bot pricing).
**Numbering:** assigned **0154** by the S53 orchestrator, which gave 0150–0155 to six parallel ADRs. After `git fetch --all --prune` I listed `docs/decisions/` on every remote branch: `origin/develop` and `origin/main` stop at 0149, `origin/docs/1892-flip-cards` adds 0151 and `origin/docs/1861-opponent-caused-discard` adds 0153. No branch has 0154.
**Builds on:** [ADR 0036](0036-attachments.md) (the attachment relation `Card.AttachedTo`, the Aura attach on resolution, CR 704.5m/n; its "Out of scope" list names "Auras entering attached from anywhere but the stack (Animate Dead…)"), [ADR 0141](0141-bestow.md) (an Aura whose enchant clause is not its card's own: a flag on `Card`, read first by `attachmentLegalLocked`), [ADR 0115](0115-commanders-die.md) (CR 903.9a, a commander leaving a graveyard as a state-based action), #936 (`Card.ObjectEpoch`, the CR 400.7 identity of an object), #663 and #2169 (delayed triggers keyed on an event).

---

## Context

Animate Dead is the 223rd most-played card on EDHREC in the 2026-09-24 Scryfall dump, and it is not in the catalog. Slice 294-a of #294 left it out on purpose. Building it without its enchant clause would make it a Reanimate with no target and no -1/-0 that only costs two mana, which is stronger than printed.

### The cards

I searched the oracle text of every Commander-legal card in the 2026-09-24 dump (31,900 distinct names) for `Enchant … card in a graveyard`, `enchant creature put onto the battlefield with`, and `becomes an Aura`.

| Card | EDHREC rank | What it needs |
|---|---|---|
| Animate Dead | 223 | this seam only |
| Necromancy | 1017 | this seam's gained clause, plus two more seams: "becomes an Aura" as it enters, and "if you cast it any time a sorcery couldn't have been cast" |
| Dance of the Dead | 2284 | this seam, plus an upkeep trigger keyed on the **enchanted creature's controller**, who also pays (no card in the catalog has one) |
| Spellweaver Volute | 18634 | this seam's graveyard host, plus "copy the enchanted instant card. You may cast the copy", which the engine has no primitive for |

**Four cards.** The issue named "Spellweaver Helix-style" cards. Spellweaver Helix is an artifact that imprints, not an Aura, so it is not in this family. Spellweaver Volute is the card the issue meant. Last Voyage of the _____ matched "becomes an Aura", but it returns a creature from **your** graveyard with a plain "enchant creature" and needs stickers, so it is not in this family either. Cloudform, Lightform, Rageform, the twelve Licids and Takklemaggot also matched "becomes an Aura" and do not need a graveyard host.

### The rules

Read in the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026):

- **CR 303.4:** "An Aura enters the battlefield attached to an object or player. What an Aura can be attached to is defined by its enchant keyword ability." Nothing limits the object to a permanent.
- **CR 303.4a** and **CR 115.1b:** an Aura spell is always targeted, and its enchant ability defines the target. Animate Dead's target is "creature card in a graveyard".
- **CR 608.3c:** a resolving Aura spell "is put onto the battlefield … attached to the player or object it was targeting". **CR 608.3b** first re-checks that target as CR 608.2b describes.
- **CR 303.4c:** "If an Aura is enchanting an illegal object or player as defined by its enchant ability and other applicable effects, the object it was attached to no longer exists, or the player it was attached to has left the game, the Aura is put into its owner's graveyard." **CR 704.5m** is the state-based action.
- **CR 400.7:** a card that changes zones is a new object. A creature card that leaves the graveyard (to the battlefield, to exile, to the command zone) is no longer the object the Aura enchants, even when it comes back.
- **CR 702.5a:** enchant "restricts what an Aura spell can target and what an Aura can enchant". **CR 702.5c:** with several enchant abilities, all apply.
- **CR 603.4:** "When this Aura enters, if it's on the battlefield, …" is an intervening "if". It is checked when the trigger event happens and again on resolution.
- **CR 611.2c:** the "loses … and gains …" is a continuous effect from a resolved triggered ability. Its set of affected objects (this Aura) is fixed when it begins, and it lasts as long as the Aura is that object.
- **CR 603.7, 603.7b, 603.7c:** "When this Aura leaves the battlefield, that creature's controller sacrifices it" is created by the enters trigger as it resolves, so it is a **delayed** triggered ability. It triggers once, the next time the Aura leaves. It affects "that creature" only while it is the same object in the same zone.
- **CR 701.3a–c:** an Aura can't be attached to an object it couldn't enchant. An effect that tries "doesn't move" it (also **CR 303.4j**). Attaching to a different object gives the Aura a new timestamp.
- **CR 303.4f, 303.4g, 303.4i:** an Aura put onto the battlefield other than by resolving as an Aura spell has its host chosen as it enters. With no legal host it stays where it was.
- **CR 903.9a:** a commander in a graveyard may go to the command zone as a state-based action. That is a zone change, so an Aura on it falls off.

### The rulings

From Scryfall's rulings for the four cards (the Gatherer rulings, as Scryfall republishes them):

- **Animate Dead, 2016-06-08:** "You target a creature card in a graveyard when you cast it. It enters the battlefield attached to that card. Then it returns that card to the battlefield, and attaches itself to the card again (since the card is a new object on the battlefield). Animate Dead itself never moves into a graveyard during this process."
- **Animate Dead, 2016-06-08:** "If Animate Dead isn't on the battlefield as its triggered ability resolves, none of its effects happen. The creature card won't be returned to the battlefield."
- **Animate Dead, 2016-06-08:** "Once the creature is returned to the battlefield, Animate Dead can't be attached to anything other than it … Attempting to move Animate Dead to another creature won't work."
- **Animate Dead, 2016-06-08:** shroud and protection work only on the battlefield, so a creature card with shroud can be targeted.
- **Animate Dead, 2016-06-08:** "If the creature put onto the battlefield has protection from black—or if the creature can't legally be enchanted by Animate Dead for another reason—Animate Dead won't be able to attach to it. It will be put into the graveyard as a state-based action, causing its delayed triggered ability to trigger. When the trigger resolves, if the creature's still on the battlefield, its controller will sacrifice it." Dance of the Dead and Necromancy have the same ruling.
- **Spellweaver Volute, 2007-05-01:** "The Aura will be on the battlefield attached to a card in a different zone. … If the enchanted card changes zones … Spellweaver Volute 'falls off' and is put into its owner's graveyard as a state-based action."
- **Necromancy, 2005-08-01:** it "enters as an enchantment and then becomes an Enchant Creature Aura as a triggered ability upon entering."
- **Dance of the Dead, 2004-10-04:** with two Dance of the Dead on one creature, each gives +1/+1, and the controller may pay each upkeep cost.

### What the engine has

Each piece below is in `server/internal/game` unless named otherwise.

- **The relation:** `Card.AttachedTo TargetRef` on the attachment (ADR 0036 decision 1). A `TargetRef` is a kind and an instance ID. It has no epoch. An instance ID survives a zone change; the object's identity is `Card.ObjectEpoch`, bumped once per move in `MoveCard` (zone.go). `ObjectRef{ID, Epoch}` (permanent_lki.go) is the pair other code uses to name one object.
- **The attach:** `AttachForEffect` (attach.go) returns early with `EventAttachSkipped` "host is not on the battlefield" whenever a card host is not on the battlefield. `attachResolvedAuraLocked`, which `announceEntryLocked` (entry_choice.go) calls between the zone move and `EventETB`, attaches a resolving Aura through it. So an Aura whose target is a graveyard card **lands unattached**.
- **The check:** `attachmentLegalLocked` reads a catalogued Aura's enchant clause as its catalog `Targets` (`TargetSpecFor(catalogKeyOf(c))`), through `specMatchLocked`. `specMatchLocked` already finds a card in any zone and accepts it when the zone is in `TargetSpec.Zones`. So an Aura **attached** to a graveyard card, with a `ZoneGraveyard` clause, would already be judged legal. The problem is that nothing can attach it. Unattached, `attachmentSBALocked` puts it into the graveyard (CR 704.5m) before its enters trigger is put on the stack. Its enters trigger then fails its intervening "if".
- **The clause:** a graveyard clause already exists. `TargetCardInGraveyard` (cards/effects/targets.go) is `Mode: "card_in_graveyard"`, `Zones: [ZoneGraveyard]`. Reanimate uses it as `targetCreatureInAnyGraveyard()`, and the announce gate, the CR 608.2b re-check, the view's legal set and the bot enumerator all handle it.
- **The return:** `ReturnFromGraveyardUnderControlForEffect` and `ReturnFromGraveyardTappedForEffect` (effect_api.go). Both go through the CR 614 entry pipeline and keep the card's instance ID.
- **A precedent for a different enchant clause:** ADR 0141's `Card.Bestowed`, checked by `attachmentLegalLocked` before the catalog clause.
- **A precedent for "attach this Aura to that creature, if it is still that object":** Gift of Immortality (cards/effects/gift_of_immortality.go) carries an `ObjectRef` in `EffectParams` and compares the epoch.
- **Delayed triggers on an event:** `DelayedTrigger.On` and `Condition` (#663). `ScheduleDelayedTriggerForEffect` gives an event-keyed trigger with no duration `UntilEndOfTurn`, because every printed one so far says "this turn". `DurationKind` already has `Indefinite`.
- **The view:** `CardView.AttachedTo` (protocol/view.go) ships the host ref as it is. The client (`PlayerPanel.svelte` `attachmentsByHost` and `hostedCardIDs`, `seatSummary.ts` `attachmentHostIDs`) draws an attachment behind its host only when the host is on the battlefield. Otherwise the attachment stays in its own row, so a graveyard host already degrades safely.

## Decision

### 1. The enchant clause is the existing graveyard target clause

`effects.EnchantCardInGraveyard(preds ...CardPredicate)` returns `TargetCardInGraveyard("enchant creature card in a graveyard", …)`, with the label taken from the predicates (Spellweaver Volute's reads "enchant instant card in a graveyard"). An Aura's `Spec.Targets` is its enchant clause, as for every catalogued Aura, so the cast, the CR 608.2b re-check, the view and the enumerator need no change. A card in a graveyard has no shroud or hexproof that applies (CR 113.6: a permanent's abilities usually function only on the battlefield; CR 702.11b and 702.18a speak of a permanent), which `canBeTargetedByLocked` already answers per zone.

### 2. `AttachForEffect` may attach an Aura to a card off the battlefield

`AttachForEffect` keeps its battlefield-host path unchanged. A `TargetCard` host that is **not** on the battlefield is accepted only when all of these hold:

- the attachment is an Aura (`IsAura()`),
- its enchant clause (`TargetSpecFor(catalogKeyOf(c))`) lists the zone the host is in (`findCardZoneLocked`), and
- `specMatchLocked(SourceChooser(c.Controller), spec, host, false)` accepts the host.

Anything else stays a quiet `EventAttachSkipped` (CR 701.3b). `attachResolvedAuraLocked` and the copy path in spell_copy.go reach this through `AttachForEffect`, so a resolving Animate Dead, or a copy of one, lands attached to its target (CR 608.3c). `runAsAttachedLocked` is not run for an off-battlefield host: "As this becomes attached to a creature" reads a permanent.

### 3. The attachment remembers which object it is on (CR 400.7)

`Card.AttachedEpoch int` is the host's `ObjectEpoch` at the moment of the attach. `AttachForEffect` sets it; every place that clears `AttachedTo` clears it (`MoveCard`'s battlefield exit, `UnattachForEffect`, `attachmentSBALocked`, the spell-copy reset).

`attachmentLegalLocked` reads it for an **off-battlefield host only** (Question 2): if the card the ID names is in a graveyard with a different epoch, it left and came back, and the object the Aura enchanted "no longer exists" (CR 303.4c). That case is rare (a card returned to hand and discarded again within one resolution), but without it the Aura would silently follow the card.

`AttachForEffect`'s "did it move?" test (`moved := c.AttachedTo != host`) also compares the epoch. Re-attaching to the same instance ID as a **new object** is attaching to a different object (CR 701.3b–c): it gets a new timestamp, and the "As this becomes attached" clause runs.

Every other path is already right:

- The card leaves the graveyard for exile, a library or a hand: `specMatchLocked` finds it outside `ZoneGraveyard`, the Aura is illegal and goes to the graveyard (CR 704.5m). That is the Spellweaver Volute ruling.
- The card is reanimated by something else while Animate Dead's trigger waits (Reanimate in response): it is now on the battlefield, which the original clause does not allow, so the Aura goes to the graveyard. When the trigger resolves, its source is gone and it does nothing (CR 603.4).
- A commander in a graveyard goes to the command zone (CR 903.9a): the same, from the command zone.

### 4. The gained clause: `Card.EnchantSwapped` and `Card.EnchantReturned`

"It loses 'enchant creature card in a graveyard' and gains 'enchant creature put onto the battlefield with this Aura'" is modelled the way ADR 0141 modelled bestow's gained "enchant creature": as state on the Aura's `Card`, read first by `attachmentLegalLocked`, not as a layer-6 effect (Question 6).

- `EnchantSwapped bool` means the Aura's enchant clause is now "enchant creature put onto the battlefield with this Aura". It is set when the trigger resolves and before anything else in the trigger (the printed order), even when nothing comes back.
- `EnchantReturned ObjectRef` is that creature: the instance and epoch of the permanent the same resolution put onto the battlefield. Zero means none was.

With `EnchantSwapped` set, `attachmentLegalLocked` answers legal only for a host that is on the battlefield, is a creature (effective types), passes the protection check every attachment goes through, and is the object `EnchantReturned` names (ID and epoch). An Aura attached anywhere else, or to nothing, is illegal and goes to the graveyard (CR 704.5m).

Both fields are cleared by `MoveCard` on every zone change, beside `Bestowed` (CR 400.7: the new object never had the effect, CR 611.2c). They live on the Aura, so Necromancy's own "becomes an Aura with 'enchant creature put onto the battlefield with Necromancy'" can set them later from its own trigger.

`AttachForEffect` honours CR 701.3b for a swapped Aura: an attach to anything but `EnchantReturned` is skipped and the Aura stays where it is, as the ruling says ("Attempting to move Animate Dead to another creature won't work"). For other Auras, the general check is [#3160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3160).

### 5. The enters trigger: one catalog effect

`effects.AnimateEnchantedCard{Tapped bool}` is the body shared by Animate Dead (`Tapped: false`) and Dance of the Dead (`Tapped: true`). It is registered with `WhenThisEnters`, and resolves in this order:

1. **CR 603.4:** `AbilitySourceGoneForEffect(item)`, the "if it's on the battlefield" re-check. If the source is gone, nothing happens (the first ruling). The check at trigger time needs nothing: an `EventETB` is only seen for a permanent on the battlefield.
2. Set `EnchantSwapped` on the Aura (decision 4).
3. Read the host off `AttachedTo`. If it is not a card in a graveyard with the attached epoch, stop: there is no "enchanted creature card". The swapped clause then makes the Aura illegal, and the state-based action puts it into the graveyard.
4. `ReturnFromGraveyardUnderControlForEffect` (or `…TappedForEffect`) with the Aura's controller: "under your control", from any player's graveyard. A paused or redirected entry (an entry prompt, a replacement) is read back the way Gift of Immortality does: the returned card keeps its instance ID.
5. If it landed on the battlefield, set `EnchantReturned` to its `ObjectRef` and `AttachForEffect(aura, host)`. That is a new object, so it is a new attach with a new timestamp (CR 701.3c). If it did not land, leave `EnchantReturned` zero.
6. Schedule the delayed sacrifice trigger (decision 6), whether or not step 5 attached. The ruling has the Aura that could not attach (protection from black) go to the graveyard, "causing its delayed triggered ability to trigger".

Nothing in steps 2–6 lets a player act, and no state-based action runs inside a resolution, so the Aura is never judged half-way. After the trigger, the state-based check sees one of three boards: the Aura legally on the returned creature, the Aura attached to a creature it can't enchant (protection) and put into the graveyard, or the Aura attached to a card that never came back and put into the graveyard.

"Enchanted creature gets -1/-0" (Animate Dead) and "+1/+1 and doesn't untap" (Dance of the Dead) are the existing `PumpAttached` and `enchantedDoesntUntap` statics. Before the return they read nothing: the layer pass and `AttachedHostOf` only see battlefield hosts, so a creature **card** in a graveyard is never pumped.

### 6. The sacrifice is a delayed trigger with no expiry

"When this Aura leaves the battlefield, that creature's controller sacrifices it" is scheduled in step 6 as a `DelayedTrigger`:

- `On: [EventLTB]`, with a `Condition` that the leaving object is the Aura (its `ObjectRef`, taken in step 1);
- `Duration: Indefinite`, so `ScheduleDelayedTriggerForEffect` does not stamp `UntilEndOfTurn` on it. It fires once and is gone (CR 603.7b). Indefinite on an event-keyed trigger is new, and needs a test that it survives the cleanup step and a snapshot round trip;
- `Params.Object`: `EnchantReturned`, or zero when nothing came back.

When it resolves, if `Params.Object` is still that object on the battlefield (CR 603.7c), **its current controller** sacrifices it through `SacrificePermanentForEffect` (CR 701.21a). The controller of the delayed trigger is the Aura's controller (CR 603.7e). A creature that changed control is sacrificed by its new controller, which is what "that creature's controller" says.

Because it is a delayed trigger and not a printed leaves ability, an Animate Dead that leaves before its enters trigger resolves sacrifices nothing (there is no "that creature"), and an Aura that is moved by some other effect still sacrifices the creature it returned.

### 7. The snapshot

Three `Card` fields are carried in `cardSnapshot`, each additive within schema 8 and omitted at its zero value: `attachedEpoch`, `enchantSwapped`, `enchantReturned`. A file without them restores an Aura with no epoch pin and the printed clause, which is every game before this change, since no catalogued Aura could be attached off the battlefield before. The drift test (`snapshot_drift_test.go`) lists all three as carried. The delayed trigger is already carried by the delayed-trigger snapshot, and its `Indefinite` duration by `Duration`.

### 8. The wire and the client

No new wire field. `CardView.attached_to` already names the graveyard card's instance ID. The cast picker already offers graveyard targets: Reanimate uses the same `card_in_graveyard` mode.

What is missing is a way to **see** what an Animate Dead on the battlefield is about to return while its trigger is on the stack, which is when an opponent decides whether to exile that card. Recommended (Question 4): a badge on the Aura, "Enchanting {card name} in {owner}'s graveyard", derived in `PlayerPanel.svelte` the way `curseTargets` is, by looking the `attached_to` id up in the seats' graveyard zones. It is a pure client derivation, with a `client/src/lib` unit test beside `attachmentHostIDs`.

### 9. The bot

- **The enumerator** offers one cast per creature card in any graveyard, as it does for Reanimate. No change (`legal/` test only).
- **The heuristic** prices the target with `cardTargetValue` (aiseat/heuristic/moves.go). A card in the bot's own graveyard is worth its `cardValue`, and a card in an opponent's graveyard is worth `OwnPermanentTarget`. Animate Dead is priced exactly like Reanimate, which is the honest baseline. The pricing undervalues stealing an opponent's creature, and `cardValue`'s 0.6 "can't cast it for a while" discount on an expensive creature is backwards for reanimation. Both apply to Reanimate today. The owner chose to fix it for the whole family in [#3161](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3161).
- **Dance of the Dead's upkeep payment** is asked of the enchanted creature's controller, through the existing optional-cost prompt, which the heuristic already answers.

### 10. What this does not change

- An Aura attached to a battlefield host is judged exactly as before. The epoch pin is read only for an off-battlefield host (Question 2). (Amended 2026-10-14 for #3159, see the end: a pinned battlefield host is judged by its epoch too.)
- No catalog card listens for `EventAttach` (only the layer listener, which bumps its version). An attach to a graveyard card bumps the layer version for nothing, which is harmless.
- `AttachedHostOf`, `AttachmentsOf` and "enchanted creature" statics read the battlefield only, so a graveyard host is never "enchanted creature".

## Cards

| Card | Ships with the seam? | What it proves |
|---|---|---|
| Animate Dead | yes, **Caveats** (Question 1) | the whole seam: the graveyard host, the swapped clause, the return under your control, the -1/-0, the delayed sacrifice |
| Dance of the Dead | yes (Question 1) | the tapped return, `enchantedDoesntUntap`, the upkeep pay-to-untap asked of the enchanted creature's controller |
| Necromancy | follow-up, [#3156](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3156) (Question 1) | the same gained clause, set by a different trigger. It needs "becomes an Aura" as a layer-4 effect from a resolved trigger, and the flash-with-cleanup-sacrifice cast that only applies if you used its own permission (ruling 2022-12-08) |
| Spellweaver Volute | follow-up, [#3157](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3157) | a long-lived graveyard host. It needs "copy the enchanted card; you may cast the copy" and a CR 303.4f choose-a-new-host attach |

Animate Dead's caveat (Question 1): "Put onto the battlefield any other way than by casting it (Sun Titan, Replenish), it enchants nothing and goes to the graveyard." The engine has no CR 303.4f "choose what it enchants as it enters" for any Aura: Replenish and Brilliant Restoration already carry caveats for it. For most Auras that costs a card. For Animate Dead it also means Sun Titan's recursion, Worldgorger Dragon's loop and Replenish lines do not work, and players will try them.

## Tests

- `game/attach_graveyard_test.go`: an Aura with a graveyard clause, cast at a creature card in an opponent's graveyard, lands attached and survives the state-based check; the target exiled in response counters it (CR 608.2b); the card leaving the graveyard while the Aura is attached (exile, the command zone) puts the Aura into the graveyard; the card returned to hand and discarded within one resolution (a new object) does too (decision 3); `AttachForEffect` onto a graveyard card with a battlefield-only clause is skipped (CR 701.3b).
- `cards/effects/animate_dead_test.go`: Animate Dead returns an opponent's creature under its caster's control, attached, at -1/-0; Dance of the Dead returns it tapped, it stays tapped in the untap step, and its controller may pay {1}{B} in their upkeep to untap it; the Aura destroyed makes the creature's **current** controller sacrifice it (after a control change too); the creature leaving first leaves the Aura to the state-based action, and the delayed trigger then does nothing; the Aura removed in response to its enters trigger returns nothing and sacrifices nothing (ruling 1); a returned creature with protection from black is not enchanted, the Aura goes to the graveyard and the creature is sacrificed (ruling 5); an effect that tries to move a swapped Animate Dead to another creature leaves it where it is (ruling 3); the delayed trigger survives a cleanup step and an undo snapshot.
- `game/snapshot_drift_test.go` and the snapshot shape file: the three fields, carried. A restore mid-trigger (Animate Dead on the battlefield attached to a graveyard card, its trigger on the stack) round-trips.
- `legal/animate_dead_test.go`: one cast per creature card in every graveyard, each accepted.
- `protocol/`: `attached_to` on the Aura names the graveyard card while the trigger waits.
- `client/src/lib`: the graveyard-host badge (if Question 4 is (a)).
- The oracle fixtures for the new cards, the coverage mechanic, and the roadmap registry row moved to `implemented` with `ADR: "0154-…"` and the follow-ups as `Waiting`.

## Alternatives considered

- **Leave the Aura unattached and "remember" the target on the side**, exempting it from CR 704.5m until its trigger resolves. It would need a second legality exemption beside the attach prompt one (`attachPromptOpenForLocked`), and it is not what the rules say: Animate Dead *is* attached to the card (first ruling). A Spellweaver Volute could never work that way.
- **Model Animate Dead as a Reanimate that brings an Aura along**, with no graveyard host at all. It is simpler and wrong in each ruled case: no target to fizzle, nothing to respond to on the stack, and nothing that falls off when the card is exiled in response.
- **Put an epoch on `TargetRef`** instead of a field on the attachment. `TargetRef` is the target type for every spell and ability, on the wire and in every snapshot. Pinning targets to objects is a real gap (a spell's graveyard target that leaves and comes back is still found), but it is a much larger change than this seam needs. It is listed under Out of scope.
- **The gained clause as a real layer-6 ability effect.** The engine does not model "enchant" as an ability on the layer pass. It reads the catalog `Targets`, and ADR 0141 kept it that way. Making enchant a layer-6 ability would let an ability-removal effect strip it, which no card in the catalog does to an Aura. That is a cost with no card to show for it (Question 6).
- **Expire the sacrifice trigger at end of turn** like other event-keyed delayed triggers. Wrong: Animate Dead can sit on a creature for the whole game.

## Out of scope

- **CR 303.4f for Auras entering other than by resolving**, for every Aura (Sun Titan, Replenish, Brilliant Restoration, Worldgorger Dragon's loop). It is [#3158](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3158). Until then Animate Dead carries the caveat above.
- **An epoch pin on every attachment.** Reading `attach.go`, a host flickered within one resolution (Cloudshift, Ephemerate) appears to come back with its Auras still attached, because nothing compares the object. No test pins it either way. The owner chose to file it as a bug that starts with a failing test: [#3159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3159).
- **An epoch on spell targets** (`TargetRef`), above.
- **Necromancy and Spellweaver Volute**: [#3156](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3156) and [#3157](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3157).
- **CR 701.3b for every Aura**: [#3160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3160).
- **Reanimation-family bot pricing**: [#3161](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3161).
- **CR 303.4d for an Aura that becomes a creature.** Unchanged from ADR 0141.

## Questions for the owner (answered 2026-10-13)

Each lists the recommended option first. The owner chose (a), the recommended option, on every question. Each answer is recorded under its question and is binding.

1. **Q1. Which cards ship with the seam, and at what completeness?**
   - **(a) Recommended:** Animate Dead with the CR 303.4f caveat, and Dance of the Dead, in the seam PR. Necromancy, Spellweaver Volute and an engine-wide CR 303.4f prompt each get a follow-up issue.
   - **(b)** Animate Dead marked Full, with no caveat, the way other catalogued Auras are marked. It is consistent, but players will reach for Sun Titan and Worldgorger first.
   - **(c)** Animate Dead only. Dance of the Dead waits for an "upkeep of enchanted creature's controller" helper of its own.
   - **Decision:** (a). Animate Dead ships with the CR 303.4f caveat, and Dance of the Dead ships with the seam. Necromancy is [#3156](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3156), Spellweaver Volute is [#3157](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3157), and the CR 303.4f prompt is [#3158](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3158).
2. **Q2. How far does the epoch pin reach?**
   - **(a) Recommended:** off-battlefield hosts only, in this seam. File the battlefield case (a flickered host keeping its Auras) as a bug, starting with a failing test, so the fix is measured on its own.
   - **(b)** Every attachment, now. It closes the flicker gap at once, but it changes the legality of every Aura and Equipment in one PR whose subject is a graveyard host.
   - **Decision:** (a). The epoch check reads only hosts off the battlefield. The flicker case is [#3159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3159), which starts with a failing test.
3. **Q3. CR 701.3b for every Aura?** Decision 4 makes a swapped Animate Dead refuse to move. Today any other Aura that an effect attaches to something it can't enchant (Illusory Gains onto a creature with protection) moves and is then put into the graveyard, where the rules say it "doesn't move".
   - **(a) Recommended:** only the swapped clause, in this seam, and a separate issue for the general check in `AttachForEffect`. The general check changes existing cards' outcomes.
   - **(b)** The general check now: `AttachForEffect` refuses any Aura onto a host its enchant clause rejects.
   - **Decision:** (a). CR 701.3b is enforced only for the rewritten clause in this seam. The general check is [#3160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3160).
4. **Q4. Show the graveyard host?**
   - **(a) Recommended:** a client-derived badge on the Aura, "Enchanting {card} in {owner}'s graveyard", so an opponent can see what is coming back while the trigger waits. No wire change.
   - **(b)** No client change. The Aura sits in its own row and the board does not say what it enchants.
   - **Decision:** (a). A badge derived on the client, "Enchanting {card} in {owner}'s graveyard", with no wire change.
5. **Q5. The bot's price for the target.**
   - **(a) Recommended:** unchanged. Animate Dead is priced like Reanimate. Open an issue for reanimation pricing (an opponent's creature undervalued, the expensive creature discounted) that fixes the whole family with an arena measurement, as ADR 0126 asks of any weight change.
   - **(b)** Price a graveyard creature target of an Aura with a graveyard clause by `CreatureValue` whoever owns it, in this seam. It fixes Animate Dead and leaves Reanimate as it is.
   - **Decision:** (a). The price is unchanged: Animate Dead is priced like Reanimate. The family-wide fix, measured with an arena run, is [#3161](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3161).
6. **Q6. How the gained clause is represented.**
   - **(a) Recommended:** two `Card` fields (`EnchantSwapped`, `EnchantReturned`) read first by `attachmentLegalLocked`, as ADR 0141 did for bestow.
   - **(b)** A general layer-6 "enchant" ability that effects can grant and remove. It is more faithful to CR 613.1f, but no card in the catalog needs it, and it touches every Aura's legality.
   - **Decision:** (a). Two `Card` fields, `EnchantSwapped` and `EnchantReturned`, read first by `attachmentLegalLocked`.

## Amendment — 2026-10-13: a reanimation with counters and a tail (#3172, cav-h1)

Decision 5 gave the graveyard return a continuation, `ReturnFromGraveyardToBattlefieldThenForEffect`, so
that a "then" that must see the NEW permanent lands even when the creature first asks a question (a
Clone's copy pick, a shockland's payment). Three more cards were waiting on the same shape, one of them
with counters:

- **`ReturnFromGraveyardWithCountersThenForEffect(card, controller, tapped, counters, then)`** is that
  continuation beside the "enters with counters" clause. The counters ride the entry event (CR 614.1c),
  so a Doubling Season sees them; `then` runs once, with the entering permanent's ID or `uuid.Nil`.
  Makeshift Mannequin grants its sacrifice ability from the tail, so a returned Clone has it.
- **Persist** ("with a -1/-1 counter on it") uses `ReturnFromGraveyardWithCountersForEffect`, the form
  without a tail: the counter is part of the entry, not a second step after it.
- **Presumed Dead** suspects the returned creature from the tail, so a returned Clone is suspected.

Back-out: stop passing the counter map (Persist), skip the grant (Mannequin) or the suspect (Presumed
Dead) in the tail; each named test fails (`TestPersistCounterRidesTheEntry`,
`TestMakeshiftMannequinGrantLandsAfterACloneAnswersItsEntryQuestion`,
`TestPresumedDeadSuspectsACloneThatAsksAsItReturns`).

## Amendment 2026-10-13 ([#3158](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3158)): CR 303.4f for every Aura

Owner decision Q1 (a) shipped Animate Dead with a caveat and filed the host choice as #3158. This amendment is that follow-up. No section above changes; it removes the caveat Question 1 gave Animate Dead and the "Out of scope" item it names.

**The rule.** CR 303.4f: an Aura entering the battlefield other than by resolving as an Aura spell, when the effect putting it there doesn't specify what it enchants, has its controller choose a legal object or player for it as it enters. CR 303.4g: with no legal choice it stays in its current zone (a token isn't created). The named-host case, an effect that puts an Aura onto the battlefield attached to something it can't legally enchant, is cited here as CR 303.4i, the number this ADR's Context section uses for the pinned edition. The pinned TXT was not in the builder's checkout, so check the subrule letters against it.

**Decision 11. The choice is the last step of the entry's CR 614 window.** When the apply-loop has nothing left to apply, `settleEnteringAuraHostLocked` (`game/aura_host.go`) asks, before anything moves. That keeps "as it enters" observable: the Aura is never on the battlefield unattached, so CR 704.5m sees nothing, and no enters trigger fires for an Aura that then falls off. It also reuses the pause every entry question already has. The reanimation, the library search, the exile return, the hand / library / graveyard / exile put batch (#1322) and created tokens (#762) are all `entryResumable` (the "declared gap" about the search and reanimation paths predates #478), and the paused event resumes through `replacementResume`. A simultaneous entry walks one card's window at a time, so Replenish asks once per Aura.

It applies to an entry that is a move onto the battlefield (or a created token), is not a resolving spell (CR 303.4a covers those), is not face down, does not enter as a copy, and whose card is an Aura with a catalogued enchant clause. An uncatalogued Aura is still a manual object, as `attachmentLegalLocked` treats it.

**Decision 12. The candidates.** The Aura's catalogued `TargetSpec`, read through `specMatchesLocked` without the targeting gate (CR 115.1: choosing a host targets nothing, so hexproof and shroud don't apply), then CR 702.16c's protection test against the Aura, for a permanent or a player. A graveyard clause (decision 1) offers the creature cards in every graveyard. The Aura itself and every other card of the same simultaneous entry are excluded: they are not yet objects an Aura can enter attached to, and the Replenish ruling says an Aura it returns can't enchant another permanent it returns. One candidate is a forced choice and asks nothing. None is CR 303.4g, and the entry is cancelled. The chooser is the would-be controller on the event, which ADR 0102 keeps current.

**Decision 13. Where the answer goes.** `ReplacementEvent.auraHost` (with `auraHostSettled`) carries it to the one finisher, which attaches through `AttachForEffect` between the zone move and `EventETB` (`attachEnteringAuraLocked`, beside `attachResolvedAuraLocked`). So decision 2's off-battlefield attach and decision 3's epoch pin apply to a graveyard host exactly as they do for a cast Animate Dead, and decision 5's enters trigger reads the host off `AttachedTo` unchanged: Sun Titan returning Animate Dead returns the chosen creature. The fields are plain values on the event, copied by the undo clone of a paused frame. Nothing new is snapshotted, because a paused entry is already counted by `ContinuationCensus.ChoiceResumeFrames`.

**Decision 14. An effect that names the host is not asked.** `SearchLibrarySpec.EntersAttachedTo` (Light-Paws), `ZoneEntryOptions.AttachedTo` (Mantle of the Ancients' batch, and Gift of Immortality's single return through the batch door) and `TokenEntryOptions.AttachedTo` (a copy of an Aura spell, CR 707.10) seed `auraHostFixed`. The window checks the named host against the same predicate and cancels the entry when it is illegal (CR 303.4i), so a Light-Paws search that finds an "enchant land" Aura leaves it in the library. A non-Aura with a named host (Mantle's Equipment) is attached at the same point and left to CR 704.5n. Those three cards used to enter and then attach in a continuation; they now enter attached.

**Decision 15. The prompt.** `PendingChoiceAuraHost` (`aura_host`) carries the candidates in `PickTargetCards` / `PickTargetPlayers` with min and max 1, rides the `pick_target` wire projection and is answered with the `{target}` (or one-element `{targets}`) payload, so the client's board picker answers it, and a graveyard card through the zone browser. It blocks the table (entry_controller's reason) and is dropped, not reassigned, when its chooser leaves: the paused entry is abandoned and the Aura stays where it was. The enumerator offers one move per host. The heuristic prices a graveyard host by its creature value whoever owns it, since every Aura with a graveyard clause returns that card under its controller. A battlefield or player host is read like an Aura spell's target: an opponent's permanent as removed, the bot's own at `OwnPermanentTarget` with its best creature first. That inherits the cast path's blind spot (nothing on the wire says whether an Aura is a buff or a curse) and does not touch Question 5's cast price (#3161).

**What it does not do.** The sandbox `move_card` verb cannot pause and keeps its old behaviour. An Aura entering as a copy (Copy Enchantment's shape) is not asked; no catalogued card has it. Spellweaver Volute's "attach it to another instant card" moves an Aura already on the battlefield, which is not CR 303.4f, and stays with #3157.

**Cards.** Animate Dead is Full. So are Light-Paws, Emperor's Voice, Brilliant Restoration, Ondu Spiritdancer, Zur the Enchanter, Carmen, Cruel Skymarcher, Open the Vaults, Guardian Sunmare, Armored Skyhunter, It That Betrays and Replenish. Nazgûl Battle-Mace drops its Aura caveat and keeps its Rest in Peace one. The registry row `aura-chooses-its-host-as-it-enters` is implemented.

## Amendment 2026-10-14 ([#3159](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3159)): the epoch pin for a battlefield host

Question 2 kept decision 3's epoch check to off-battlefield hosts and filed the battlefield case as a bug that starts with a failing test. This amendment is that bug's fix. It changes decision 3 and the first bullet of decision 10; everything else stands.

**The failing case.** Chaos Warp on a creature with Pacifism and Bonesplitter attached, its owner's library empty. The creature is tucked, the shuffle leaves it as the only card, the reveal finds it and puts it onto the battlefield, all in one resolution and with the same instance ID (the library-to-battlefield put keeps it). No state-based check ran while the creature was gone, and `attachmentLegalLocked` found "the host" by ID, so Pacifism stayed and the Equipment stayed attached. CR 400.7 makes the returned creature a new object: the Aura goes to the graveyard (CR 303.4c, CR 704.5m) and the Equipment becomes unattached (CR 301.5c, CR 704.5n). Cloudshift and Ephemerate, the cards the issue names, were never affected: the exile return mints a new instance ID (`resetAsNewObjectLocked`), so the old link names nothing and the state-based action always swept it. They keep regression tests.

**Decision 3, amended.** `AttachForEffect` records the host's `ObjectEpoch` in `Card.AttachedEpoch` for every card host, on or off the battlefield, and sets a new `Card.AttachedPinned`. `attachmentLegalLocked` judges a battlefield host by the epoch when the link is pinned; an off-battlefield host is judged by it always, as before. A pinned link whose host's epoch has moved on is illegal, and `attachmentSBALocked` handles it through its existing branches: an Aura to the graveyard, an Equipment or Fortification unattached, a bestowed Aura unbestowed (CR 702.103f).

**Why a flag and not the zero value.** Epoch 0 is a real epoch, a card that has never changed zones, so a zero `AttachedEpoch` cannot also mean "not recorded". Two kinds of link were never recorded: one set by hand (test fixtures set `AttachedTo` directly, hundreds of them), and one restored from a file written before this amendment. Both restore unpinned and keep the old behaviour, rather than falling off on their first state-based check. `AttachedPinned` is carried in the snapshot (additive within v8, absent when false) and cleared wherever `AttachedTo` is.

**The move test.** `AttachForEffect`'s "did it move?" compares the epoch when the existing link is pinned or the new host is off the battlefield. Re-equipping a creature that left and came back with its ID is a move (CR 701.3c: a new timestamp, and the "As this becomes attached" clause runs). Re-equipping the same object is not. The CR 701.3b probe (`attachmentLegalOnLocked`) is unpinned, because it asks about the host as it is now.

**Decision 10, amended.** An Aura or Equipment attached to a battlefield host is judged as before, plus one question: is the host still the object it was attached to.

**What it does not do.** Between the host's return and the next state-based check, the rest of the same resolution still sees the old link: a static such as "enchanted creature can't attack" or "equipped creature gets +2/+0" applies to the returned creature through `IsAttachedTo`, which compares the instance ID only. Nothing in the catalog reads the returned creature's characteristics inside that window that we know of, and the state-based check runs before any player gets priority.

**Tests.** `TestChaosWarpedHostReturningAsTheSameIDLosesItsAura` and `TestChaosWarpedHostReturningAsTheSameIDLosesItsEquipment` (cards/effects), `TestHostReturningAsANewObjectKeepsNoAura` and `TestHostReturningAsANewObjectKeepsNoEquipment` (game, raw `MoveCard` out and back), `TestHandSetAttachmentIsNotPinned` (the compatibility half) and `TestReequipOntoTheSameObjectIsNotAMove`. Cloudshift keeps `TestCloudshiftedHostLosesItsAura` and `TestCloudshiftedHostLosesItsEquipment`.

**Back-out.** Delete the `else if c.AttachedPinned && …` branch in `attachmentLegalLocked`: the two Chaos Warp tests and the two game tests fail. Stop setting `AttachedPinned` in `AttachForEffect`: the same four fail, and `TestReequipOntoTheSameObjectIsNotAMove` too.

## Amendment 2026-10-14 ([#3156](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3156)): Necromancy

Question 1 left Necromancy for a follow-up because it needs two seams beyond decision 4's gained clause. This amendment adds them. Nothing above changes, except that an Aura attached to nothing with the swapped clause is now illegal even when the catalog has no enchant clause for it.

**Decision 16. "It becomes an Aura" is a layer-4 effect from the resolved trigger.** Necromancy is an enchantment spell, not an Aura spell: the cast has no target and it enters unattached. Its enters trigger targets "creature card from a graveyard" as it is put on the stack. `effects.BecomeAnAuraAndReanimateTarget` resolves it in printed order: the CR 603.4 re-check, then the Aura subtype as an `AddSubtypesMod("Aura")` scoped effect pinned to this object for as long as it is on the battlefield (CR 611.2c, CR 613.1d), then decision 4's swapped clause, then the return under the trigger controller's control, and decision 5's steps 5 and 6 unchanged (`animateAttachAndSchedule`): the clause names the new permanent, the Aura attaches to it and the delayed "that creature's controller sacrifices it" is scheduled. A target gone by resolution makes the trigger do nothing (CR 608.2b), so Necromancy stays a plain enchantment, which is how it plays on paper.

The enchant ability is decision 4's state on the `Card`, not a catalog clause, because Necromancy prints none. So `attachmentLegalLocked`'s unattached branch now answers false for a swapped Aura before it reaches the uncatalogued-Aura exemption: one whose creature never came back is attached to nothing and goes to the graveyard (CR 704.5m).

**Decision 17. "You may cast this spell as though it had flash. If you cast it any time a sorcery couldn't have been cast, …".** `Spec.FlashWithCleanupSacrifice` (`CardDef.FlashWithCleanupSacrifice`) is step 1b of `castTimingOpenLocked`: the card opens the instant-speed window for itself. It is not the flash keyword, and a per-player restriction (Teferi, Time Raveler) still closes it (CR 101.2). `CastSpell` stamps `StackItem.OwnFlashCast` at announce, through `castUsesOwnFlashLocked`: true only when a sorcery couldn't be cast now and the same timing gate with the card's own permission switched off would have refused. The 2022-12-08 ruling limits the sacrifice to a cast that used the card's own permission. A player with another one (Vedalken Orrery) would always choose it, so the engine chooses it for them and asks nothing. As the spell becomes a permanent (`runEntryHooksLocked`, beside the alternative-cost entry hooks), `scheduleOwnFlashCleanupSacrificeLocked` schedules a CR 603.7 delayed trigger `At: StepCleanup` for that object (its CR 400.7 epoch in `Params.Object`). When it resolves, the permanent's current controller sacrifices it if it is still that object. The trigger fires as the cleanup step begins, so the active player gets priority and another cleanup step follows (CR 514.3a, cleanup.go). A copy is not cast and never carries the mark. `ownFlashCast` is carried in the stack item's snapshot, additive within v9.

The same rider is printed on Lightning Reflexes, Ward of Lights, Mystic Veil, Spider Climb, Parapet, Soar, Armor of Thorns, Relic Ward and Grave Servitude. None is catalogued yet; each needs only the flag.

**Cards.** Necromancy is Full. The registry row `becomes-an-aura-as-it-enters` is implemented.

**Tests.** `cards/effects/necromancy_test.go`: the sorcery-speed cast (an enchantment that becomes an Aura on the returned creature, no cleanup sacrifice), the end-step cast sacrificed at cleanup with its creature, a cast another flash permission allowed and that is kept, a sorcery-only restriction closing the window, the target exiled before the trigger resolves, the Aura losing its creature, the Aura destroyed, and a restore point. `game/attach_graveyard_test.go`: `TestSwappedAuraWithNoPrintedClauseAttachedToNothingIsIllegal`.
