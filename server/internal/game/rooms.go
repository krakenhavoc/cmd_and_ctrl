package game

import "github.com/google/uuid"

// rooms.go — Rooms (CR 709.5), ADR 0103.
//
// A Room is a split permanent card whose halves share one type line
// (HasSharedTypeLine). Each half is a DOOR (CR 709.5j). Two
// designations, "left half unlocked" and "right half unlocked"
// (CR 709.5c), decide which halves the permanent has: a locked half
// has no name, no mana cost and no rules text (CR 709.5).
//
// The pieces, each in one place:
//
//	Card.Unlocked            the two designations (card.go)
//	materialiseDoors         a Room's flat fields on the battlefield
//	Designation.Active       the door gate on printed abilities
//	                         (designations.go)
//	UnlockDoorForEffect /    the two writers, and the only places the
//	LockDoorForEffect        designations change outside the entry
//	applyEntersUnlocked…     CR 709.5d: the cast door, as it enters
//	UnlockOffers             the CR 116.2m special action's offers
//	                         (special_action.go carries it out)

// DoorMask is a Room's unlocked designations (CR 709.5c): one bit per
// door. The zero value is both doors locked.
type DoorMask uint8

const (
	// DoorLeftUnlocked is CR 709.5c's "left half unlocked".
	DoorLeftUnlocked DoorMask = 1 << iota
	// DoorRightUnlocked is CR 709.5c's "right half unlocked".
	DoorRightUnlocked
)

// DoorBit is the designation bit for a door. Zero for DoorNone.
func DoorBit(door DoorSide) DoorMask {
	switch door {
	case DoorLeft:
		return DoorLeftUnlocked
	case DoorRight:
		return DoorRightUnlocked
	}
	return 0
}

// Has reports whether the door is unlocked.
func (m DoorMask) Has(door DoorSide) bool {
	b := DoorBit(door)
	return b != 0 && m&b != 0
}

// Full reports whether both doors are unlocked.
func (m DoorMask) Full() bool {
	return m&(DoorLeftUnlocked|DoorRightUnlocked) == DoorLeftUnlocked|DoorRightUnlocked
}

// DoorOfFace is the door a face index names: face 0 is the left door,
// face 1 the right.
func DoorOfFace(face int) DoorSide {
	switch face {
	case 0:
		return DoorLeft
	case 1:
		return DoorRight
	}
	return DoorNone
}

// FaceOfDoor is the inverse of DoorOfFace; -1 for DoorNone.
func FaceOfDoor(door DoorSide) int {
	switch door {
	case DoorLeft:
		return 0
	case DoorRight:
		return 1
	}
	return -1
}

// DoorName is the wire spelling of a door: "left" / "right".
func DoorName(door DoorSide) string {
	switch door {
	case DoorLeft:
		return "left"
	case DoorRight:
		return "right"
	}
	return ""
}

// ParseDoorName is the inverse of DoorName. Anything else is DoorNone.
func ParseDoorName(s string) DoorSide {
	switch s {
	case "left":
		return DoorLeft
	case "right":
		return DoorRight
	}
	return DoorNone
}

// materialiseDoors writes a Room's battlefield characteristics into the
// flat fields (CR 709.5): the names, costs and colours of the halves
// whose door is unlocked, and the shared type line (CR 709.5a). Both
// locked is a permanent with no name, no mana cost, mana value 0 and
// no colour (CR 105.2). A no-op for anything that is not a Room.
func (c *Card) materialiseDoors() {
	if !HasSharedTypeLine(*c) {
		return
	}
	c.materialiseHalves(c.Unlocked.Has(DoorLeft), c.Unlocked.Has(DoorRight))
	c.TypeLine = c.Faces[0].TypeLine
}

// IsUnlockedForEffect reports whether the named battlefield Room has
// that door unlocked. False for anything not on the battlefield
// (CR 400.7) and for a card that is not a Room.
//
// Caller must hold g.mu.
func (g *Game) IsUnlockedForEffect(roomID uuid.UUID, door DoorSide) bool {
	c := findBattlefieldCard(g, roomID)
	return c != nil && HasSharedTypeLine(*c) && c.Unlocked.Has(door)
}

// UnlockedDoorsForEffect is the named battlefield Room's unlocked
// designations, or zero when it is not a Room on the battlefield.
//
// Caller must hold g.mu.
func (g *Game) UnlockedDoorsForEffect(roomID uuid.UUID) DoorMask {
	c := findBattlefieldCard(g, roomID)
	if c == nil || !HasSharedTypeLine(*c) {
		return 0
	}
	return c.Unlocked
}

// UnlockDoorForEffect gives a Room one of its unlocked designations
// (CR 709.5f) and announces it. One of the two writers of
// Card.Unlocked outside the entry and the snapshot restore.
//
// Idempotent, like SolveCaseForEffect: unlocking a door that is
// already unlocked does nothing and emits nothing, so "when you unlock
// this door" never fires for an unlock that did not happen.
//
// `actor` is the player who unlocks it — the special action's taker,
// or the player an effect instructs. uuid.Nil means the Room's
// controller.
//
// Caller must hold g.mu in write mode.
func (g *Game) UnlockDoorForEffect(roomID uuid.UUID, door DoorSide, actor uuid.UUID) error {
	card := findBattlefieldCard(g, roomID)
	if card == nil {
		return ErrCardNotFound
	}
	if !HasSharedTypeLine(*card) || DoorBit(door) == 0 {
		return ErrInvalidParam
	}
	if card.Unlocked.Has(door) {
		return nil
	}
	if actor == uuid.Nil {
		actor = card.Controller
	}
	card.Unlocked |= DoorBit(door)
	card.materialiseDoors()
	g.announceDoorsUnlockedLocked(*card, DoorBit(door), actor, false)
	return nil
}

// LockDoorForEffect removes one of a Room's unlocked designations
// (CR 709.5g) and announces it. The other writer. Idempotent the same
// way: locking a locked door does nothing.
//
// Caller must hold g.mu in write mode.
func (g *Game) LockDoorForEffect(roomID uuid.UUID, door DoorSide, actor uuid.UUID) error {
	card := findBattlefieldCard(g, roomID)
	if card == nil {
		return ErrCardNotFound
	}
	if !HasSharedTypeLine(*card) || DoorBit(door) == 0 {
		return ErrInvalidParam
	}
	if !card.Unlocked.Has(door) {
		return nil
	}
	if actor == uuid.Nil {
		actor = card.Controller
	}
	card.Unlocked &^= DoorBit(door)
	card.materialiseDoors()
	g.EmitEvent(Event{
		Kind:   EventDoorLocked,
		Actor:  actor,
		Source: card.InstanceID,
		CardID: card.InstanceID,
		Target: card.InstanceID,
		Amount: int(door),
		Label:  card.Faces[FaceOfDoor(door)].Name,
	})
	return nil
}

// announceDoorsUnlockedLocked emits one EventDoorUnlocked per door in
// `added`, left first, and EventRoomFullyUnlocked when those doors
// completed the pair (CR 709.5i: it had one and got the other, or had
// neither and got both). `room` is the permanent AFTER the change.
//
// `entering` is the CR 709.5d case, where the designation was given as
// the permanent entered: a door unlocked that way still triggers "when
// you unlock this door" (CR 709.5h), and it can never complete the
// pair, because one spell casts one half.
func (g *Game) announceDoorsUnlockedLocked(room Card, added DoorMask, actor uuid.UUID, entering bool) {
	for _, door := range []DoorSide{DoorLeft, DoorRight} {
		if added&DoorBit(door) == 0 {
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventDoorUnlocked,
			Actor:  actor,
			Source: room.InstanceID,
			CardID: room.InstanceID,
			Target: room.InstanceID,
			Amount: int(door),
			Label:  room.Faces[FaceOfDoor(door)].Name,
		})
	}
	if !entering && room.Unlocked.Full() && added != 0 {
		g.EmitEvent(Event{
			Kind:   EventRoomFullyUnlocked,
			Actor:  actor,
			Source: room.InstanceID,
			CardID: room.InstanceID,
			Target: room.InstanceID,
		})
	}
}

// applyEntersUnlockedLocked is the battlefield landing's half of
// ReplacementEvent.EntersUnlocked (CR 709.5d): the permanent that has
// just landed gets the designation of the half that was cast, before
// any event about its arrival fires, so the unlocked door's statics
// apply as it enters. Returns the doors it set, for the announcement
// after EventETB.
//
// Caller must hold g.mu.
func (g *Game) applyEntersUnlockedLocked(cardID uuid.UUID, doors DoorMask) DoorMask {
	c := findBattlefieldCard(g, cardID)
	if c == nil || !HasSharedTypeLine(*c) {
		return 0
	}
	added := doors &^ c.Unlocked
	c.Unlocked |= doors
	c.materialiseDoors()
	return added
}

// castDoorsOf is the designation a resolving Room spell gives the
// permanent it becomes (CR 709.5d): the door of the half that was
// cast. Zero for a copy — a copy of a spell is not cast (CR 707.10),
// so a copy of a Room spell enters with both doors locked (ADR 0103
// owner decision 3) — and for anything that is not a Room.
func castDoorsOf(spell Card, item *StackItem) DoorMask {
	if item == nil || item.IsCopy || !HasSharedTypeLine(spell) {
		return 0
	}
	return DoorBit(DoorOfFace(spell.ActiveFace))
}

// IsUnlockableRoomForEffect reports whether `id` is a face-up Room on
// the battlefield — the only object a door instruction can act on.
//
// Caller must hold g.mu.
func (g *Game) IsUnlockableRoomForEffect(id uuid.UUID) bool {
	c := findBattlefieldCard(g, id)
	return c != nil && HasSharedTypeLine(*c)
}

// UnlockOffers is the CR 116.2m / 709.5e special action a Room offers:
// one per locked door, priced at that half's mana cost. Derived from
// the card, never declared (ADR 0103 Option 3B), so an uncatalogued
// Room can be unlocked too. Nil for anything that is not a face-up
// Room with a locked door.
func UnlockOffers(c Card) []SpecialAction {
	if !HasSharedTypeLine(c) {
		return nil
	}
	var out []SpecialAction
	for _, door := range []DoorSide{DoorLeft, DoorRight} {
		if c.Unlocked.Has(door) {
			continue
		}
		f := c.Faces[FaceOfDoor(door)]
		label := "Unlock " + f.Name
		if f.ManaCost != "" {
			label += " " + f.ManaCost
		}
		out = append(out, SpecialAction{
			Kind:  SpecialActionUnlock,
			Cost:  f.ManaCost,
			Label: label,
			Door:  door,
		})
	}
	return out
}
