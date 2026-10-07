# ADR 0103 — Rooms: casting either half, locked doors, and unlocking them

**Status:** Accepted · 2026-09-30 · S50 — Seams from the deck re-checks
**Owner decisions:** 2026-09-30. All eight open questions are answered; see [Owner decisions](#owner-decisions-2026-09-30) at the end. Two answers changed the design: every split card now casts either half, with fuse and aftermath (Decision 3), and "fully unlocked" gets its own log line (Decision 9).
**Issue:** [#1756](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1756) (the Rooms seam). [#1640](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1640) (the ViviVoltron deck request) waits on it for Roaring Furnace // Steaming Sauna.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I listed `docs/decisions/` on every remote branch (37, then 39 on the repeat just before the push) and also collected every `docs/decisions/010*` file name on any ref with `git log --all`. Numbers up to 0102 are taken. Numbers 0104 and 0105 are reserved for ADRs being written in parallel, and both had appeared on branches by the repeat run. No branch has 0103, so this one takes **0103**.
**Builds on:** [ADR 0071](0071-designations-that-switch-abilities-on.md) (the designation gate; its Decision 3 sketched Rooms and reserved `DesignationDoorUnlocked`), [ADR 0034](0034-multi-face-cards.md) (`Faces`, `ActiveFace`, `SetFace`, `CastableFaces`, the `#1` catalog keys), [ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) Decision 4 (one `special_action` verb, one timing table), [ADR 0082](0082-casting-face-down-and-turning-face-up.md) Decision 6 (a special action taken on a battlefield permanent), [ADR 0069](0069-face-down-objects.md) (face-down objects), [ADR 0043](0043-copy-effects.md) (copiable values), [ADR 0013](0013-replacement-effects.md) (the entry pipeline), [ADR 0041](0041-game-persistence.md) (restore points), [ADR 0033](0033-ai-bot-seat.md) (the bot).
**Supersedes:** ADR 0071 Decision 3 ("Rooms: designed, not built"). **Amends:** ADR 0034 §4's `split` row, for every split card.

---

## Context

### What a Room is

A Room is a split card whose two halves are both enchantments and share one type line, "Enchantment — Room". Each half is a **door**. You cast one half. The permanent enters with that door **unlocked** and the other **locked**. A locked door has no name, no mana cost and no rules text. As a sorcery, you may pay a locked door's mana cost to unlock it.

The rules are CR 709, checked against `MagicCompRules 20260819.txt`:

| Rule | What it says |
|---|---|
| 709.2 | A split card is one card, however many halves it has. |
| 709.3, 709.3a | The player chooses which half to cast before putting it on the stack. Only that half is checked to see if it can be cast. |
| 709.3b | On the stack, only the cast half's characteristics exist. |
| 709.4, 709.4a–c | In every other zone, the card has both halves' characteristics combined: two names, the combined mana cost (so combined colours and mana value), every type and every ability. |
| 709.5 | A split **permanent** card with a shared type line has two static abilities: "as long as this permanent doesn't have the 'left half unlocked' designation, it doesn't have the name, mana cost, or rules text of this object's left half", and the same for the right half. These abilities, and which half each characteristic is in, are copiable values. |
| 709.5a | Both halves share the types and subtypes on the shared type line. |
| 709.5b | The existence of each half is a copiable value even on the stack (an exception to 709.3b). |
| 709.5c | "Left half unlocked" and "right half unlocked" are designations. A half without its designation is "locked". |
| 709.5d | The permanent gets the designation of the half that was **cast** as it enters. A Room that enters without being cast has neither. |
| 709.5e, 116.2m | Paying a locked half's mana cost to unlock it is a special action. You may take it any time you have priority and the stack is empty during a main phase of your turn. |
| 709.5f, 709.5g | "Unlock" and "lock" as instructions: the player chooses a locked (or unlocked) half, and the designation is given (or removed). |
| 709.5h | "When you unlock this door" triggers when the designation is given, **including as the permanent enters**. |
| 709.5i | "Fully unlock" triggers when the permanent has one designation and gets the other, or has neither and gets both. |
| 709.5j | A "door" is a half of a Room permanent. |

Two consequences are easy to miss. A Room with both doors locked has no name, no mana cost, mana value 0 and no colour (CR 105.2, 202.2: colour comes from the mana cost). And a Room in hand, the library or the graveyard has mana value equal to both halves added together (CR 202.3d): Funeral Room // Awakening Hall is mana value 11 there, 3 on the stack as Funeral Room, and 3 on the battlefield with only Funeral Room unlocked.

### What exists

Checked on `develop` at `4a5da9a6`:

- **The gate is reserved.** `game.DesignationDoorUnlocked` and `game.DoorSide` (`DoorNone`, `DoorLeft`, `DoorRight`) exist in `game/designations.go`. `Designation.Active` returns `false` for the door kind, and `effects.Register` (`cards/effects/registry.go`) panics on any Spec that declares a door gate.
- **The gate covers six catalog slots**: `Static`, `Triggered`, `Activated`, `CostModifiers`, `TriggerDoublers` and `GatedCastPermissions` (`specDesignations` in `cards/effects/designations.go`). A few engine types outside the catalog walk also carry `ActiveWhen` (attack taxes, cast and activation restrictions, exhaust permissions, mana triggers). `Replacements`, `UntapStep` and `NoMaxHandSize` do not.
- **Rooms import as `split` cards.** Scryfall has no `room` layout. A Room is `layout: "split"` with two `card_faces` whose `type_line` is the same string. That is why ADR 0034 found "zero `room` printings": its census was by layout. `toGameCard` builds `Faces` and calls `SetFace(0)`, so a Room is its left half in every zone today.
- **Only the left half can be cast.** `Card.CastableFaces` (`game/face.go`) returns `[0]` for `split`, and `deck/validate.go` warns "casts as its left half only".
- **Special actions** are one verb (`game.PerformSpecialAction`, `game/special_action.go`) with three per-kind tables: the zone (`specialActionZone`), the timing (`SpecialActionTimingOKLocked`) and the performer (`specialActionPerformer`). Four kinds are built: `foretell`, `suspend`, `plot` and `turn_face_up`. The last one acts on a battlefield permanent and its offer is **derived** from the card rather than declared in the catalog (`TurnFaceUpOffer`, ADR 0082 Decision 5). `SpecialActionParams` has a `Cost` field to pick between two offers of one kind, but no way to name a door.
- **Mana for a special action** is paid with the zero `ManaSpendContext`, so restricted mana can't pay for one. `ManaSpendPurpose` has `cast` and `activate` only.
- **Designations clear on a zone change.** `MoveCard`'s battlefield-exit block (`game/zone.go`) and `entry_tail.go` reset `ClassLevel`, `Solved`, `Harnessed`, `Monstrous` and `Prepared` (CR 400.7).
- **The wire** carries `faces`, `active_face`, `layout`, `class_level`, `solved` and `special_actions` on `CardView`. The client shows special actions as a section of the card's right-click menu (`client/src/lib/contextMenu.logic.ts`).
- **The bot** gets special actions from `legal/special_actions.go` and scores them all at one flat `SpecialActionValue` (`aiseat/heuristic`).

### What is missing

Everything in the table above except the gate's field: the unlocked state, casting the right half, the locked-half characteristics, entering unlocked, the unlock special action, the lock and unlock instructions, the two trigger shapes, the wire and client, and the bot move.

### The cards

The Scryfall dump (`data/scryfall/default-cards.json`, refreshed 2026-09-24) has **30 Room oracle IDs**:

- 23 from Duskmourn (`dsk`);
- 5 from the Duskmourn Commander decks (`dsc`);
- 2 Alchemy-only cards (`ydsk`): Crude Abattoir // Unsavory Kitchen and Solitary Study // Endless Corridor. They use "perpetually" and "conjure", which are digital-only, and are out of scope.

That leaves **28 paper Rooms**. [Cards covered](#cards-covered) lists them all.

Another 24 non-Room cards in the dump read Room state, not counting one Alchemy card. 16 are the Duskmourn **eerie** cards ("Whenever an enchantment you control enters and whenever you fully unlock a Room, …"). The rest are Keys to the House, Marina Vendrell, Ghostly Keybearer, Creeping Peeper, Rampaging Soulrager, Anthropede, Intruding Soulrager and Blu, Mansion Prince. None of them is catalogued.

---

## Options

There are four real choices. Each has a recommendation, and the Decisions section builds the recommended one.

### 1. How a Room's characteristics change with its doors

**A. Fork each reader.** Teach `printedCharacteristic`, `manaCostForValue`, `printedColors`, the name readers and the view about Rooms, one at a time. *Rejected.* This is the shape ADR 0046 and #622 removed: every reader is a place to forget, and the next reader added forgets by default.

**B. Materialise the flat fields (recommended).** Keep ADR 0034's rule that the flat printed fields (`Name`, `ManaCost`, `Colors`, …) hold what the object *is right now*. For a Room, "right now" depends on the zone and the doors, so one function, `materialiseSplit`, rewrites the flat fields at the few moments that change them: entering the battlefield, unlocking, locking, and leaving the battlefield or the stack. Every existing reader is then correct without being edited. It is the same bet ADR 0034 Decision 1 made, and that one paid off.

**C. A third "permanent" face.** Append a synthetic face to `Faces` at import that stands for the battlefield Room. *Rejected.* The face picker, the snapshot, copies (`PrintedValues` copies `Faces`) and every `len(Faces)` check would all see a face that isn't printed.

### 2. How the catalog keys a Room

**A. Two Specs, like adventure and prepare.** The left half under the bare oracle ID, the right under `"<oracle>#1"`. *Rejected.* An adventure's two halves are different objects with different effects. A Room's two halves are **one** permanent at once, so a two-key model would need the ability readers to merge two keys for one object, which is exactly what `CatalogKey` exists to avoid.

**B. One Spec, told apart by the door gate (recommended).** Both halves' abilities go on one Spec under the bare oracle ID, each with `ActiveWhen: DoorUnlocked(side)`. This is ADR 0071 Decision 3's design. `CatalogKey` returns the bare ID for a Room even while its right half is on the stack. Nothing on a Room functions on the stack (CR 113.6), so the one key costs nothing there.

### 3. What unlocking is

**A. An activated ability** on each door. *Rejected.* CR 709.5e and 116.2m make it a special action: it doesn't use the stack, can't be responded to and isn't stopped by effects that stop activations. An ability would get all three wrong.

**B. A fifth special action kind, `unlock` (recommended).** It uses ADR 0062's one verb and gets one row in each of its three tables, as `turn_face_up` did. The offer is derived from the card, not declared, so an **uncatalogued** Room can still be cast from either half and unlocked. Only its door abilities stay manual.

### 4. How a card file gates a door's abilities

**A. `ActiveWhen` on every ability by hand.** *Rejected as the only guard.* An ability that forgets its gate is active while its door is locked, which is stronger than printed.

**B. A `Room(left, right)` constructor plus a boot guard (recommended).** The constructor stamps each door's abilities with that door's gate. `Register` refuses a Room Spec with any ungated ability, and refuses one with abilities in a slot the gate can't reach.

---

## Decisions

### 1. What makes a card a Room, and where the doors live

```go
// game/rooms.go
func HasSharedTypeLine(c Card) bool // layout split, two faces, identical permanent type lines (CR 709.5)

type DoorMask uint8
const (
	DoorLeftUnlocked  DoorMask = 1 << iota // CR 709.5c "left half unlocked"
	DoorRightUnlocked                      // CR 709.5c "right half unlocked"
)

// on Card, in the existing bool block (TestCardHasNoInteriorPadding holds it there)
Unlocked DoorMask
```

`HasSharedTypeLine` is the structural test CR 709.5 gives ("split cards … with a single shared type line"), not a check for the Room subtype. All 30 Rooms pass it and no other card in the dump does. It is false for a face-down object.

`Unlocked` is battlefield state. It is cleared in the same two places as `Solved` and `Monstrous` (CR 400.7). It is not copiable, because nothing in the copy path reads it. `Designation.Active` reads it:

```go
case DesignationDoorUnlocked:
	return d.Door != DoorNone && c.Unlocked&doorBit(d.Door) != 0
```

### 2. What a Room is in each zone

`materialiseSplit(c *Card, state)` is the one writer of a split card's flat fields besides `SetFace`. It runs for every `split` card (owner decision 1). The door rows apply only when `HasSharedTypeLine` is true. The table below uses a Room as the example; for any other split card only the first two rows apply, plus a fused spell (Decision 3), which is the whole card on the stack.

| Where | Name | Mana cost / value | Colours | Rule |
|---|---|---|---|---|
| Hand, library, graveyard, exile, command zone | "Left // Right" | both costs concatenated, e.g. `{2}{B}{6}{B}{B}`, MV 11 | union of both halves | 709.4, 202.3d |
| Stack | the cast half's | the cast half's | the cast half's | 709.3b (this is today's `SetFace(i)`, unchanged) |
| Battlefield, one door unlocked | that half's | that half's | that half's | 709.5 |
| Battlefield, both unlocked | "Left // Right" | both costs | union | 709.5 |
| Battlefield, both locked | empty | empty, MV 0 | none | 709.5, 105.2 |

For a Room the type line is the shared line in every row (709.5a). For any other split card the whole-card type line is the union of both halves' types and subtypes (709.4c): Commit // Memory is "Instant Sorcery" in hand. Rules text is not a flat field: for a Room it is the door gate (Decision 7).

It runs at four moments:

1. **Import.** `toGameCard` calls it instead of `SetFace(0)` for a split card, so a split card in a decklist is the whole card.
2. **`MoveCard`.** It is the one place every zone change passes through. A split card moving anywhere except the stack or the battlefield is rematerialised as the whole card. `CastSpell` still calls `SetFace(i)` after the move to the stack.
3. **Entering the battlefield.** After the entry's doors are settled (Decision 4).
4. **The two door writers** (Decision 6).

Two readers need one small change each:

- **Names.** A flat `Name` can't hold two names, and CR 709.4a and 201.4b say an object "has the chosen name if one of its names is the chosen name". A new `game.NamesOf(c) []string` returns each present half's name. Central Elevator and Promising Stairs need it. Readers that compare `c.Name ==` are not moved in this change; that gap is listed under Consequences.
- **Copiable values.** `CopiableValuesOf` must return the Room's **whole-card** values plus its `Faces`, never the door-materialised ones. CR 709.5 makes the two static abilities and the halves copiable, and the designations not. A copy then materialises against its own `Unlocked`, which is zero.

### 3. Casting either half

`CastableFaces` returns `[0, 1]` for a Room, as for every split card (CR 709.3; see the end of this decision). The existing face picker, the per-face announce surface from #992, the enumerator's `CastableFacesUnder` walk and the timing gate's union over faces all follow without changes.

On the stack, `SetFace(i)` already gives the cast half's name, cost, colour and mana value (709.3b, 202.3d). `Faces` stays on the card, which is 709.5b: a copy of the spell still has both halves.

`faceOnResolve` already returns 0 for `split`, so the permanent resolves as face 0 and `materialiseSplit` takes over. `CatalogKey` gets one rule: a card with a shared type line keys on the bare oracle ID on every face (Option 2B).

**Every split card casts either half** (owner decision 1, CR 709.3). `CastableFaces` returns `[0, 1]` for every `split` card, narrowed by the zone the cast comes from, so that no half is castable where the rules forbid it:

- **Aftermath (CR 702.127a).** The half that prints aftermath is found from the faces' own oracle text (`Face.OracleText` starts with "Aftermath"), so no catalog entry is needed. In the dump it is always the right half of the 27 aftermath cards. That half may be cast **only from a graveyard**, from any zone and under any grant. The engine derives the graveyard permission for it from the card, the same way a printed flashback opens the graveyard, and exiles it instead of letting it leave the stack anywhere else when it was cast from a graveyard. The other half is cast like any card: from hand, and from a graveyard only if some other permission opens it (a Snapcaster Mage flashback grant does). The aftermath permission never opens the first half.
- **Fuse (CR 702.102).** A card with fuse cast **from hand** may be cast with both halves. The announcement gains `fuse: true` (`CastSpellParams.Fuse`). The spell on the stack is marked `Card.Fused` and materialised as the whole card: both names, the combined cost as its total cost (702.102c) and mana value (202.3d), and both halves' colours and types (702.102b, 709.4d). Its target clauses are the left half's clauses followed by the right half's, one announcement in ADR 0065's clause order. It resolves the left half's instructions, then the right half's (702.102d), each half reading only its own targets. The catalog gives a fused spell a key of its own, derived from the two halves' registrations; if either half is uncatalogued the fused spell resolves by hand, like any uncatalogued spell. A fused cast is refused from any zone but the hand, for a card without fuse, and for a card either of whose halves declares modes, an alternative cost or an additional or optional cost. None of the 22 fuse cards prints one, so the refusal only stops a cast the engine could not price correctly.

The deck validator's `split` warning is removed: the layout is played.

### 4. Entering with the cast door unlocked (709.5d)

The resolving spell seeds the entry, in the same place and the same way `EntersWithCountersFromCast` does (`applyCastEntryCountersLocked`, `game/entry_counters.go`). The entry event gains `EntersUnlocked DoorMask`, set from the stack item's cast face. When the card lands, `Unlocked` is written, the Room is materialised, and only then is `EventETB` emitted. So the unlocked door's statics apply as the permanent enters, and anything watching the entry sees the door.

Right after `EventETB`, in the same event batch, the engine emits `EventDoorUnlocked` for that door, because 709.5h says "when you unlock this door" triggers on entering too. Entering with one door is never a full unlock.

Every other entry gets no doors: reanimation, "put onto the battlefield", a flicker's return, a token copy, a copy of a Room spell (a copy is not cast, CR 707.10) and a Room that enters as a copy of something else. Each of these is the seeding site simply not being reached.

### 5. The unlock special action (709.5e, 116.2m)

A fifth kind, `SpecialActionUnlock = "unlock"`, with one row in each ADR 0062 table:

| Table | Row |
|---|---|
| Zone | `ZoneBattlefield`. The actor is the Room's **controller** (709.5e: "a player who controls"), as for `turn_face_up`. |
| Timing | `g.SorcerySpeedOpenLocked(player)`: priority, empty stack, the player's own main phase. The same predicate `plot` uses, so the two can't drift. Split second never matters: it needs a spell on the stack. |
| Performer | `unlockLocked`, which calls `UnlockDoorForEffect` (Decision 6). |

The **offer is derived**, not declared. `SpecialActionsOfferedByCard` adds one offer per locked door of a face-up Room, beside `TurnFaceUpOffer`. Each offer carries the door and its cost:

```go
// on SpecialAction
Door DoorSide // unlock only
// Cost = that half's printed mana cost; Label = "Unlock Awakening Hall {6}{B}{B}"
```

`SpecialActionParams` and the wire payload gain `door` (`"left"` / `"right"`). The existing `cost` field can't pick a door, because some Rooms print the same cost on both halves (Spiked Corridor // Torture Pit is `{3}{R}` twice).

**The mana** is paid with a new purpose, `SpendPurposeUnlock`. A new restriction tag, `ManaRestrictUnlock`, matches it. Smoky Lounge ("spend this mana only to cast Room spells and unlock doors") and Creeping Peeper need both. Mana restricted to casting still can't pay for an unlock. The price goes through `SpecialActionManaCostForEffect` like every other special action (#1319), so a future "unlock costs {1} less" card is one cost modifier.

A special action is neither a cast nor an activation, so effects that stop those (Grand Abolisher, Trickbind) do not stop an unlock. That needs no code; it is how the verb already works.

### 6. Unlock and lock as instructions (709.5f–g), and the events

The two writers, the only places `Unlocked` changes outside restore:

```go
func (g *Game) UnlockDoorForEffect(roomID uuid.UUID, door DoorSide, actor uuid.UUID) error
func (g *Game) LockDoorForEffect(roomID uuid.UUID, door DoorSide, actor uuid.UUID) error
```

Each is idempotent (unlocking an unlocked door does nothing and emits nothing, like `SolveCaseForEffect`). Each sets or clears the bit, calls `materialiseSplit` and then emits:

- `EventDoorUnlocked` (`CardID` = the Room, `Actor` = the unlocking player, `Amount` = the door side), and then, if both doors are now unlocked, `EventRoomFullyUnlocked` (709.5i);
- or `EventDoorLocked`.

All three events bump the layer version in `layer_listener.go`, beside `EventCaseSolved`, because a door changes a name, a cost, colours and which statics exist.

Card-side primitives, in a new `cards/effects/rooms.go`:

- `UnlockADoor{Room, Player}` — 709.5f. The player chooses a locked door. With one locked door there is no prompt. With two, it asks through the existing `option_pick` kind (`PickOption`), one option per door name.
- `LockOrUnlockADoor{Room, Player}` — Keys to the House and Marina Vendrell. One `option_pick` over "Lock X" / "Unlock Y".
- `UnlockALockedDoorOfARoomYouControl{Player}` — Ghostly Dancers. This is not targeted, so the options are every (Room, locked door) pair the player controls.

No new `PendingChoiceKind` is added. The choice gate, the enumerator's choice moves and the departure table are unchanged.

### 7. The door gate in the catalog, and the triggers

`Designation.Active` answers for doors (Decision 1). `effects.Register` stops panicking on the door gate. That guard is replaced by a stricter one: a Spec built by `Room(...)` must gate **every** ability on a door, and must declare nothing in a slot the gate can't reach.

```go
Register(Room(RoomSpec{
	OracleID: "d5f31713-d380-42ba-8052-4b8d9beb3958",
	Name:     "Roaring Furnace // Steaming Sauna",
	Left:  Door{Triggered: []game.TriggeredAbility{WhenYouUnlockThisDoor(game.DoorLeft, "Roaring Furnace — …", …)}},
	Right: Door{NoMaxHandSize: true, Triggered: []game.TriggeredAbility{AtYourEndStep("Steaming Sauna — draw a card", …)}},
}))
```

`Room` stamps `ActiveWhen: DoorUnlocked(side)` on each door's abilities and flattens both doors into one Spec. **Three more slots get the gate** in the engine PR, because 28 paper Rooms need them: `Replacements` (Torture Pit), `UntapStep` (Prop Room) and `NoMaxHandSize` (Steaming Sauna). Each is read through a per-object accessor with `activeOnly`, as ADR 0071 did for the first four.

Trigger constructors, in `triggers_common.go`:

- `WhenYouUnlockThisDoor(side, label, effect)`: watches `EventDoorUnlocked`, matches this Room and this door, and is gated on the same door. The door is unlocked by the time the event fires, so the gate is already satisfied. This is how 709.5h works on entry and after.
- `WheneverYouFullyUnlockARoom(label, effect)`: watches `EventRoomFullyUnlocked` with `ev.Actor == source.Controller`.
- `Eerie(label, effect)`: `OnAny` over an enchantment entering under your control and `EventRoomFullyUnlocked` by you. That is the 16 eerie cards' shared first line.

Board readers, in `cards/effects/rooms.go`: `UnlockedDoorsYouControl(g, player) int` (Rampaging Soulrager, Misty Salon), `UnlockedDoorNamesYouControl(g, player) []string` (Promising Stairs) and `IsFullyUnlocked(g, id) bool`.

### 8. Copies, face-down Rooms, and zone changes

- **Zone changes.** A Room that leaves the battlefield is a new object with both doors locked (CR 400.7). It is materialised as the whole card in its new zone. A flickered Room comes back fully locked, because it wasn't cast. A bounced Room can be recast from either half.
- **Phasing** is not a zone change (CR 702.26d), so the doors stay as they were, like counters.
- **Copies.** A permanent that enters as, or becomes, a copy of a Room gets both halves and the two static abilities (709.5), and no designations. It is a fully locked Room that its controller may unlock. This comes from Decision 2's `CopiableValuesOf` rule and needs no copy-path code.
- **Face down.** A face-down Room is a 2/2 with no text (CR 708.2a). `CatalogKey` already returns empty for it, so no door ability exists, and it offers no unlock. Its `Unlocked` bits are left alone. A Room manifested from the library entered face down and was not cast, so it has no doors. If it is later turned face up (Staff Room can do this; CR 701.40g only keeps an instant or sorcery face down), it is a fully locked Room. CR 708.8 says nothing relating to entering applies, so no door unlocks and nothing triggers.
- **Leaving the game.** CR 800.4a removes the owner's Rooms like any other object. Nothing new is needed.
- **Ability removal.** If something removed all of a Room's abilities (CR 613.1f), the two 709.5 static abilities would arguably go too and the locked half's text would come back. No catalogued card can do this to an enchantment that isn't a creature. It is out of scope and listed below.

### 9. The wire and the log

`CardView` gains one public field, present only for a face-up Room on the battlefield:

```go
Doors *RoomDoorsView `json:"doors,omitempty"`
type RoomDoorsView struct {
	Left  bool `json:"left"`  // unlocked?
	Right bool `json:"right"`
}
```

Door names, costs and images already ride on `faces`. `SpecialActionView` gains `door` (`"left"` / `"right"`) and `CardView.name` stays the rules name, so a fully locked Room sends an empty name. The client displays a nameless Room from `faces` (Decision 10). Door state is public, as a Class level is.

The log (#984 gate): `EventDoorUnlocked` and `EventDoorLocked` get narrated lines: "Alice unlocks Awakening Hall (Funeral Room // Awakening Hall)". `EventRoomFullyUnlocked` gets its own line too (owner decision 7): "Funeral Room // Awakening Hall is fully unlocked". It gives the eerie triggers something visible to point to. The log names the card by its whole name, because the card's identity is public and a log line is display, not a rules read.

### 10. The client

- **Two door buttons.** A Room on the battlefield shows a two-part door strip under the card, one segment per half. Each segment shows the half's name and a lock state. A locked door whose `unlock` offer is in `special_actions` becomes a button, "Unlock {cost}", which sends `special_action {card_id, kind: "unlock", door}`. Unavailable offers show greyed with the reason the menu already gives. The same rows also appear in the right-click menu's special-actions section, which needs no change beyond the door field.
- **Locked-half overlay.** The card art dims the locked half. The hover overlay shows both halves' text from `faces`, with the locked one marked "locked".
- **Casting.** Nothing new: `CastableFaces` now offers both halves, so the existing face picker opens for a Room in hand.
- **Name display.** A Room with an empty `name` is labelled from `faces` ("Funeral Room // Awakening Hall — locked") wherever the client prints a card name.

### 11. The bot

`legal/special_actions.go` walks the seat's own face-up Rooms on the battlefield (the same filtered walk it does for face-down permanents) and emits one `special_action` move per locked door when `SpecialActionTimingOKLocked` says yes and the door is affordable. Like every other enumerated payment it sends `strict` and `auto_tap`, and it prices with `SpendPurposeUnlock`, so a bot is never offered an unlock the engine refuses (#544).

Casting either half needs no enumerator change: it already walks `CastableFacesUnder`.

Scoring: the heuristic prices every special action at one flat `SpecialActionValue`. An unlock is closer to casting a spell of the door's mana value, and Awakening Hall at 8 mana is the whole point of the card. The recommendation is to score an `unlock` move like a cast of that mana value. See open question 5.

### 12. Snapshots and undo

- `CardSnapshot` gains `Unlocked DoorMask` (`json:"unlocked,omitempty"`). This is an additive change within schema version 7. It is recorded with `-update-shape`, with no bump. The flat fields are already captured, so a restored Room comes back materialised as it was.
- One new fixture file joins `testdata/snapshots/v7/`: a Room with one door unlocked and a Room spell on the stack cast as its right half. `-write-corpus` adds new boards without touching existing files, so no bump is needed.
- **Rollback.** A binary from before this change ignores `unlocked`. A restored Room keeps its name and cost but loses its door abilities, which is weaker than printed and a legal-looking state. This is the same trade `Solved` and `Monstrous` made.
- **Undo** restores the card, and `Unlocked` is a value field, so `cloneCard` needs nothing.
- The door choice uses `option_pick`, so there is nothing new to classify for restore points or the closure ratchet. Nothing here holds a closure.

### 13. Delivery

Four PRs, each green on its own:

1. **Engine.** Casting either half of every split card, aftermath and fuse (Decision 3), `HasSharedTypeLine`, `Unlocked`, `materialiseSplit` and its four call sites, `NamesOf`, the `CopiableValuesOf` rule, `CastableFaces` and `CatalogKey` for Rooms, the 709.5d seed, the two writers and three events with their layer bumps and log lines, `Designation.Active` for doors, the gate on `Replacements` / `UntapStep` / `NoMaxHandSize`, the new Register guard, the `unlock` special action with its derived offer, `door` param and spend purpose, the enumerator case, the wire field, the snapshot field and fixture, and the validator note. Tests use an uncatalogued Room imported from a fixture, which is enough to prove casting, entering, unlocking, triggers on the event, the view and the bot move.
2. **Client.** The door strip, buttons, overlay and name fallback.
3. **Cards.** `Room(...)`, the trigger constructors and readers, and the 20 Rooms marked "PR 3" below. This includes the three the registry lists as waiting.
4. **Room-adjacent cards.** The 16 eerie cards, Keys to the House, Marina Vendrell, Ghostly Keybearer, Creeping Peeper and Rampaging Soulrager. Anthropede and Intruding Soulrager only name the Room subtype, so they can go in any batch.

After PR 1 the `rooms` row in `server/internal/roadmap/registry.go` flips to implemented, and a closed-seam fragment is added for #1756.

---

## Cards covered

Every paper Room in the dump. "PR 3" means the card is expected to need nothing beyond PR 1 and existing primitives. Each one is still re-checked against the code when the batch is written, because the seam list has run stale before. "Waits" names what is missing.

| Room | Status | Notes |
|---|---|---|
| Bottomless Pool // Locker Room | PR 3 | |
| Central Elevator // Promising Stairs | PR 3 | uses `NamesOf` and the unlocked-door names reader |
| Charred Foyer // Warped Space | waits | Warped Space: a standing "pay {0} rather than the mana cost" for spells cast from exile, once each turn. No primitive. |
| Cramped Vents // Access Maze | waits | Access Maze: a standing "pay life equal to its mana value rather than its mana cost", once during each of your turns. No primitive. |
| Dazzling Theater // Prop Room | waits | Dazzling Theater: "creature spells you cast have convoke". No primitive grants convoke to spells. Prop Room needs the `UntapStep` gate from PR 1. |
| Defiled Crypt // Cadaver Lab | PR 3 | |
| Derelict Attic // Widow's Walk | PR 3 | |
| Dollmaker's Shop // Porcelain Gallery | PR 3 | needs a Toy token row |
| Experimental Lab // Staff Room | waits | Experimental Lab: manifest dread (CR 701.62a) is not built. |
| Funeral Room // Awakening Hall | PR 3 | on the registry's waiting list |
| Glassworks // Shattered Yard | PR 3 | |
| Grand Entryway // Elegant Rotunda | PR 3 | needs a Glimmer token row |
| Greenhouse // Rickety Gazebo | PR 3 | |
| Meat Locker // Drowned Diner | PR 3 | |
| Mirror Room // Fractured Realm | PR 3 | |
| Moldering Gym // Weight Room | waits | Weight Room: manifest dread. |
| Painter's Studio // Defaced Gallery | PR 3 | |
| Polluted Cistern // Dim Oubliette | PR 3 | |
| Restricted Office // Lecture Hall | PR 3 | |
| Roaring Furnace // Steaming Sauna | PR 3 | on the registry's waiting list; unblocks #1640. Needs the `NoMaxHandSize` gate from PR 1. |
| Secret Arcade // Dusty Parlor | waits | Secret Arcade: "permanent spells you control are enchantments" is a type change on spells on the stack. To be verified; likely no primitive. |
| Smoky Lounge // Misty Salon | PR 3 | needs `ManaRestrictUnlock` from PR 1 |
| Spiked Corridor // Torture Pit | PR 3 | Torture Pit needs the `Replacements` gate from PR 1 |
| Surgical Suite // Hospital Room | PR 3 | |
| Ticket Booth // Tunnel of Hate | waits | Ticket Booth: manifest dread. |
| Underwater Tunnel // Slimy Aquarium | waits | Slimy Aquarium: manifest dread. |
| Unholy Annex // Ritual Chamber | PR 3 | on the registry's waiting list |
| Walk-In Closet // Forgotten Cellar | PR 3 | Walk-In Closet uses `GatedCastPermissions` |

**20 ship in PR 3. 8 wait**: four on manifest dread, two on standing alternative-cost grants, one on granting convoke to spells, and one to be verified. The 8 are still playable after PR 1, because the unlock offer is derived: they can be cast from either half and unlocked, and only their door abilities are manual.

*Update 2026-10-07 (#2570):* the four manifest-dread doors shipped with manifest dread itself ([ADR 0082's amendment of the same date](0082-casting-face-down-and-turning-face-up.md)). Moldering Gym, Ticket Booth and Underwater Tunnel are complete; Experimental Lab keeps only its Staff Room caveat.

---

## Consequences

**Good.** A Room is one flag pair, one materialiser and one special action kind. Every reader of name, mana value and colour is right without being touched. The designation gate from ADR 0071 gets its fourth kind with no new filter. The unlock offer is derived, so the engine plays every Room's lifecycle, catalogued or not. The special action verb gets its fifth kind with no new verb.

**Costs, stated.**

- `MoveCard` gains one branch on `HasSharedTypeLine`. It is one layout string compare for every other card.
- `Card` gains one byte, inside the existing bool block.
- **Name matching stays partial.** Readers that compare `c.Name ==` won't match one half's name on a Room outside the battlefield. Only the readers the Room cards need move to `NamesOf`. The rest move when a card needs them.
- **A fused split spell has a synthetic catalog key.** It is the one catalog key that is not a printed face. Decision 3 says why it is derived rather than declared.

---

## Out of scope

- The two Alchemy Rooms (perpetual, conjure).
- Copying a fused split spell (CR 707.10 copies the fused status). No catalogued card copies an instant or sorcery that has fuse.
- Manifest dread (CR 701.62), which four Rooms wait on. It needs its own seam.
- Granting convoke to spells, and standing alternative-cost grants.
- A Room that loses all its abilities (Decision 8).
- Blu, Mansion Prince, a token copy of one half of a random Room. It is from an unknown set in the dump.

---

## Open questions for the owner

1. **Scope of right-half casting.** Should PR 1 open the right half for Rooms only (recommended), or for every split card? The second means handling fuse and aftermath now, because otherwise an aftermath half could be cast from hand.
2. **Rooms with a door we can't build.** For the 8 Rooms that wait, should they stay uncatalogued until the missing primitive exists (recommended: they still play through the derived unlock, with their door abilities manual)? Or should they ship with a caveat saying one door does nothing when unlocked?
3. **A copy of a Room spell.** Read literally, CR 709.5d gives the designation only when a half "was cast", and a copy of a spell is not cast (CR 707.10). So a token copy of a Room spell would enter fully locked. Is that the reading you want (recommended, because it is the literal text), or should a copy enter with the copied half unlocked?
4. **The client's door buttons.** Should unlocking be a button on the card's door strip, as well as a row in the right-click menu (recommended)? Or only the menu row, as foretell and plot are today?
5. **The bot's value for unlocking.** Should the heuristic score an unlock like casting a spell of the door's mana value (recommended)? Or keep the flat `SpecialActionValue` every special action has now?
6. **A fully locked Room's name on the wire.** Should `CardView.name` be empty for a fully locked Room, with the client falling back to `faces` (recommended: it keeps the rules name and the display name separate)? Or should the server send a display name such as "Funeral Room // Awakening Hall (locked)"?
7. **The log.** Should "fully unlocked" be silent because the unlock line just before it covers it (recommended), or get its own line so the eerie triggers have something visible to point to?
8. **Delivery.** Is four PRs right (engine, client, 20 Rooms, Room-adjacent cards)? Or should the three waiting-list Rooms ride in PR 1 so #1640 unblocks with the engine change?

---

## Owner decisions (2026-09-30)

1. **Every split card, not Rooms only.** The owner favours following CR 709.3 strictly. PR 1 opens the right half of every split card and handles aftermath and fuse now. An aftermath half can be cast only from a graveyard, and the aftermath permission never opens the first half. Fuse is cast from hand. If fuse had not fitted, it could have gone on a seam row, but no half may become castable where the rules forbid it. Decision 3 is rewritten to match.
2. **The 8 Rooms with a door we can't build stay uncatalogued.** They still play through the derived unlock, with their door abilities manual.
3. **A copy of a Room spell enters fully locked**, the literal reading of CR 709.5d and 707.10.
4. **Door-strip "Unlock {cost}" buttons, plus the right-click menu row.**
5. **The bot scores an unlock like casting a spell of the door's mana value.**
6. **A fully locked Room's `CardView.name` is empty**, and the client falls back to `faces`.
7. **"Fully unlocked" gets its own log line.** This is not the recommendation; Decision 9 is changed to match.
8. **Four PRs, as proposed.** Engine, client, the 20 Rooms, then the Room-adjacent cards.
