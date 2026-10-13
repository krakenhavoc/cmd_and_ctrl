# ADR 0154 — Auras that enchant a card in a graveyard (Animate Dead)

**Status:** Accepted (owner decisions 2026-10-13) · 2026-10-13 · S53 — Faces, zones, and shared payment (ADR first), tracker [#1910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1910)
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

- An Aura attached to a battlefield host is judged exactly as before. The epoch pin is read only for an off-battlefield host (Question 2).
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

## Amendment (2026-10-13, #3159 and #3160): the epoch pin reaches every attachment, and `AttachForEffect` refuses an illegal attach

Both follow-ups land together, in `attach.go`, ahead of the graveyard-host seam. Neither changes a decision above; they replace what Q2 and Q3 deferred, and they fix the storage convention the seam will build on.

### 1. `Card.AttachedEpoch` exists, and it pins every attachment (#3159, Q2)

The field decision 3 describes was not in the tree when this amendment was written, so it is introduced here, with one change of convention: it holds the host's `ObjectEpoch` **plus one**, and zero means "not pinned". A fixture that sets `AttachedTo` by hand, or a game saved before the field existed, is unpinned and keeps the old behaviour. Storing the bare epoch would read every old save as "pinned to epoch 0" and bury every Aura on the first state-based check after a deploy. The graveyard-host seam must use the same convention.

- `AttachForEffect` stamps it for a card host (a player host has no epoch). `MoveCard`'s battlefield exit, `UnattachForEffect`, `attachmentSBALocked` and the spell-copy reset clear it with `AttachedTo`.
- `attachmentLegalLocked` reads it for a host **on the battlefield** too, which is the part Q2 deferred: a pinned attachment whose host is a different object now is illegal. An Aura goes to the graveyard (CR 303.4c, 704.5m), an Equipment is unattached and stays (CR 301.5c, 704.5n). Decision 3's reading for an off-battlefield host is unchanged and still belongs to the graveyard seam.
- `moved` in `AttachForEffect` compares the pin as well, so re-attaching to the same ID as a new object is a move: new timestamp, and "As this becomes attached" runs again.
- The snapshot carries it (`attachedEpoch`, additive, recorded in `v8.txt`); `Clone` copies it as a plain field.

What the bug needs to be reachable: the host must come back with the **same instance ID**, inside one resolution. `Flicker` and `ReturnFromExileTogether` mint a new ID (`resetAsNewObjectLocked`), so Cloudshift and Ephemerate were already right, and the tests in `attachments_flicker_test.go` pin that. A return through the graveyard or hand keeps the ID (`ReturnFromGraveyardForEffect` and the other `ReturnFrom…ForEffect` helpers), and that is the path `attach_new_object_test.go` drives.

### 2. An attach that can't be made does nothing (#3160, Q3)

`AttachForEffect` now asks `attachRefusalLocked` before it moves anything. A refusal is an `EventAttachSkipped` with the reason on `Label` (in player terms, so the game log can show it), and the attachment stays on its current host.

| Attachment | Refused when |
|---|---|
| any | the host has protection from it (CR 702.16c-d, and a player's protection from everything, CR 702.16i) |
| Aura with a catalog clause | the clause's `TargetSpec` rejects the host (CR 303.4j), evaluated without the targeting gate, the same predicate the sweep asks |
| bestowed Aura | the host is not a creature (CR 702.103b) |
| Equipment | the host is a player or a noncreature (CR 301.5c), or the Equipment is itself a creature without reconfigure |
| Fortification | the host is not a land (CR 301.6) |

Two scoping rules keep this from changing more than the issue asked. An Aura with no catalog clause (a sandbox Aura, a Role token) is refused only for protection, as before. An Equipment that carries a layer 4 static (Luxior, Giada's Gift: "equipped permanent is a creature") is not refused for a noncreature host, because the question cannot be answered before the link exists; the sweep decides, as it did. The protection test is one function shared by the refusal and the sweep, so the two cannot disagree.

The sweep gained the Fortification half it was missing: a Fortification now needs a land host, where it used to be held to the creature rule.

**Which catalog cards change.** The refusal matters only when an attachment is **moved** off a host it is already on, because a refused attach of a fresh permanent leaves it unattached, which the CR 704.5m sweep buries exactly as the old attach-then-undo did. Those cards: the equip family (`attachments.go`), Magnetic Theft and Brass Squire (an Equipment aimed at a creature protected from it), and Maddening Hex (a Curse aimed at a protected player). The three cards named in #3160, Illusory Gains, Prison Term and Shielded by Faith, are not in the catalog; they are the cases this makes right when they arrive. Light-Paws, Mantle of the Ancients, Gift of Immortality, Hammer of Nazahn, Rohirrim Chargers, Armored Skyhunter and the Role tokens put an attachment onto the battlefield and are unchanged in outcome.

### Back-out

Revert the `attachRefusalLocked` call in `AttachForEffect` and the refusal tests fail (`attach_refusal_test.go`, `attachments_refusal_test.go`); revert the epoch comparison in `attachmentLegalLocked` and `TestFlickeredHostKeepsNoAura` and `…NoEquipment` fail.
