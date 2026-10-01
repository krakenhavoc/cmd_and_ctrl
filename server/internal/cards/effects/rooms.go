package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms.go — the card-facing half of ADR 0103, Rooms (CR 709.5).
//
// A Room is ONE catalog entry for both halves, keyed on the bare
// oracle ID, whose abilities are told apart by the door gate
// (game.DoorUnlocked). Write it with Room, which stamps the gate on
// every ability of each door so none can be forgotten:
//
//	Register(Room(RoomSpec{
//	    OracleID: "…",
//	    Name:     "Roaring Furnace // Steaming Sauna",
//	    Left:  Door{Triggered: []game.TriggeredAbility{WhenYouUnlockThisDoor(game.DoorLeft, "Roaring Furnace — …", …)}},
//	    Right: Door{NoMaxHandSize: true, Triggered: []game.TriggeredAbility{AtYourEndStep("Steaming Sauna — draw a card", …)}},
//	}))
//
// Casting either half, entering with the cast door unlocked, the
// unlock special action and the characteristics of a locked door are
// all the engine's (game/rooms.go, game/split.go); a card file says
// only what each door does.

// Door is what one half of a Room prints. Every slot is one a door
// ability can live in and the designation gate reaches.
type Door struct {
	Static               []game.StaticAbility
	Triggered            []game.TriggeredAbility
	Activated            []ActivatedAbility
	CostModifiers        []game.CostModifier
	TriggerDoublers      []game.TriggerDoubler
	GatedCastPermissions []game.CastPermissionGate
	Replacements         []game.ReplacementEffect
	UntapStep            []game.UntapStepPermission
	NoMaxHandSize        bool
}

// RoomSpec is a Room's catalog declaration: its identity, its two
// doors, and the non-ability data a Room may carry (the ability
// bundles a door's static grants, ADR 0093).
type RoomSpec struct {
	OracleID     string
	Name         string
	Completeness Completeness
	Caveats      []string
	Left, Right  Door
	Grants       []AbilityGrant
}

// Room builds a Room's Spec: both doors' abilities in one entry, each
// gated on its own door (CR 709.5: a locked half has no rules text).
// Register refuses a Room Spec with an ungated ability.
func Room(r RoomSpec) Spec {
	s := Spec{
		OracleID:     r.OracleID,
		Name:         r.Name,
		Completeness: r.Completeness,
		Caveats:      r.Caveats,
		Grants:       r.Grants,
		room:         true,
	}
	for _, side := range []game.DoorSide{game.DoorLeft, game.DoorRight} {
		d := r.Left
		if side == game.DoorRight {
			d = r.Right
		}
		gate := game.DoorUnlocked(side)
		for _, a := range d.Static {
			a.ActiveWhen = gate
			s.Static = append(s.Static, a)
		}
		for _, t := range d.Triggered {
			t.ActiveWhen = gate
			s.Triggered = append(s.Triggered, t)
		}
		for _, a := range d.Activated {
			a.ActiveWhen = gate
			s.Activated = append(s.Activated, a)
		}
		for _, m := range d.CostModifiers {
			m.ActiveWhen = gate
			s.CostModifiers = append(s.CostModifiers, m)
		}
		for _, td := range d.TriggerDoublers {
			td.ActiveWhen = gate
			s.TriggerDoublers = append(s.TriggerDoublers, td)
		}
		for _, p := range d.GatedCastPermissions {
			p.ActiveWhen = gate
			s.GatedCastPermissions = append(s.GatedCastPermissions, p)
		}
		for _, rep := range d.Replacements {
			rep.ActiveWhen = gate
			s.Replacements = append(s.Replacements, rep)
		}
		for _, u := range d.UntapStep {
			u.ActiveWhen = gate
			s.UntapStep = append(s.UntapStep, u)
		}
		if d.NoMaxHandSize {
			// One door at most prints it; two would need two gates.
			s.NoMaxHandSize = true
			s.NoMaxHandSizeWhen = gate
		}
	}
	return s
}

// checkRoomSpec is Register's guard on door gates (ADR 0103).
//
// A Room Spec must gate EVERY ability on a door: an ungated one would
// be active while its door is locked, which is stronger than printed.
// And it must declare nothing in a slot the gate cannot reach — a mana
// ability, an "as enters" hook, a keyword — because the Room constructor
// has no way to lock it. A door gate on a Spec that is NOT a Room is a
// mistake too: nothing but a Room has doors.
func checkRoomSpec(spec Spec) string {
	gates := specDesignations(spec)
	if !spec.room {
		for _, d := range gates {
			if d.Kind == game.DesignationDoorUnlocked {
				return "gates an ability on a Room door but is not built with effects.Room"
			}
		}
		return ""
	}
	for _, d := range gates {
		if d.Kind != game.DesignationDoorUnlocked || d.Door == game.DoorNone {
			return "is a Room with an ability that is not gated on a door"
		}
	}
	if len(spec.ManaAbilities) > 0 || len(spec.ManaTriggers) > 0 || spec.AsEnters != nil ||
		len(spec.PrintedKeywords) > 0 || spec.OnResolve != nil || spec.Targets != nil ||
		spec.Modes != nil || len(spec.AlternativeCosts) > 0 || spec.AdditionalCost != nil ||
		len(spec.OptionalCosts) > 0 || len(spec.CastPermissions) > 0 {
		return "is a Room with an ability in a slot a door gate cannot reach"
	}
	return ""
}

// DoorUnlocked is the CR 709.5 gate, for a card file that builds a door
// ability outside Room (a token copy of a door, say). Prefer Room.
func DoorUnlocked(door game.DoorSide) game.Designation { return game.DoorUnlocked(door) }

// --- triggers ---------------------------------------------------------

// ThisDoorUnlocked matches "when you unlock this door" (CR 709.5h):
// this Room was given that door's unlocked designation — after it was
// on the battlefield, or as it entered with the door of the half that
// was cast.
func ThisDoorUnlocked(door game.DoorSide) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Kind == game.EventDoorUnlocked && ev.CardID == source.InstanceID &&
			game.DoorSide(ev.Amount) == door
	}
}

// WhenYouUnlockThisDoor is "When you unlock this door, …" printed on
// `door`. Gated on the same door, which is unlocked by the time the
// event fires, so the gate never keeps it from triggering.
func WhenYouUnlockThisDoor(door game.DoorSide, label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventDoorUnlocked, ThisDoorUnlocked(door), label, effect)
	t.ActiveWhen = game.DoorUnlocked(door)
	return t
}

// YouFullyUnlockedARoom matches "whenever you fully unlock a Room"
// (CR 709.5i): a Room got its second unlocked designation, unlocked by
// this ability's controller.
func YouFullyUnlockedARoom(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventRoomFullyUnlocked && ev.Actor == source.Controller
}

// WheneverYouFullyUnlockARoom is "Whenever you fully unlock a Room, …".
func WheneverYouFullyUnlockARoom(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventRoomFullyUnlocked, YouFullyUnlockedARoom, label, effect)
}

// EnchantmentEnteredUnderYourControl — "whenever an enchantment you
// control enters", the source itself included.
func EnchantmentEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsEnchantment()
}

// Eerie is the Duskmourn ability word's shared first line, one printed
// ability with two trigger conditions: "Whenever an enchantment you
// control enters and whenever you fully unlock a Room, …".
func Eerie(label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventETB, game.EventRoomFullyUnlocked},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventRoomFullyUnlocked {
				return YouFullyUnlockedARoom(ev, source, lki, g)
			}
			return EnchantmentEnteredUnderYourControl(ev, source, lki, g)
		}, label, effect)
}

// --- readers ----------------------------------------------------------

// UnlockedDoorsYouControl is the number of unlocked doors among Rooms
// `player` controls (Rampaging Soulrager, Misty Salon).
func UnlockedDoorsYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !game.HasSharedTypeLine(c) {
			continue
		}
		for _, door := range []game.DoorSide{game.DoorLeft, game.DoorRight} {
			if c.Unlocked.Has(door) {
				n++
			}
		}
	}
	return n
}

// UnlockedDoorNamesYouControl is the names of the unlocked doors among
// Rooms `player` controls, with repeats (Promising Stairs counts the
// different ones).
func UnlockedDoorNamesYouControl(g *game.Game, player uuid.UUID) []string {
	var out []string
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !game.HasSharedTypeLine(c) {
			continue
		}
		out = append(out, game.NamesOf(c)...)
	}
	return out
}

// IsFullyUnlocked reports whether the named permanent is a Room with
// both doors unlocked.
func IsFullyUnlocked(g *game.Game, id uuid.UUID) bool {
	return g.UnlockedDoorsForEffect(id).Full()
}

// --- instructions (CR 709.5f-g) ---------------------------------------

// UnlockADoor is "unlock a door of <Room>" (CR 709.5f): the player
// chooses a LOCKED door of it and it is unlocked. With one locked door
// there is nothing to choose; with two, an option_pick asks which. A
// Room with no locked door, or one that has left, does nothing. Then,
// if set, runs once the door is unlocked (or nothing happened).
type UnlockADoor struct {
	Room   uuid.UUID
	Player uuid.UUID
	Then   func(ctx *Context) error
}

func (u UnlockADoor) Apply(ctx *Context) error {
	return chooseDoorThen(ctx, u.Player, u.Room, true, false, u.Then)
}

// LockOrUnlockADoor is "lock or unlock a door of <Room>" (Keys to the
// House, Marina Vendrell): the player chooses either door; a locked one
// is unlocked (CR 709.5f) and an unlocked one locked (CR 709.5g).
type LockOrUnlockADoor struct {
	Room   uuid.UUID
	Player uuid.UUID
}

func (u LockOrUnlockADoor) Apply(ctx *Context) error {
	return chooseDoorThen(ctx, u.Player, u.Room, true, true, nil)
}

// UnlockALockedDoorOfARoomYouControl is "unlock a locked door of a Room
// you control" (Ghostly Dancers): not targeted, so every (Room, locked
// door) pair the player controls is an option.
type UnlockALockedDoorOfARoomYouControl struct {
	Player uuid.UUID
}

func (u UnlockALockedDoorOfARoomYouControl) Apply(ctx *Context) error {
	player := u.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	type pick struct {
		room uuid.UUID
		door game.DoorSide
	}
	var picks []pick
	var opts []game.ChoiceOption
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != player || !game.HasSharedTypeLine(c) {
			continue
		}
		for _, door := range []game.DoorSide{game.DoorLeft, game.DoorRight} {
			if c.Unlocked.Has(door) {
				continue
			}
			picks = append(picks, pick{c.InstanceID, door})
			opts = append(opts, game.ChoiceOption{
				Label: "Unlock " + c.Faces[game.FaceOfDoor(door)].Name,
				Cards: []uuid.UUID{c.InstanceID},
			})
		}
	}
	switch len(picks) {
	case 0:
		return nil
	case 1:
		return ctx.Game.UnlockDoorForEffect(picks[0].room, picks[0].door, player)
	}
	return PickOption{
		Player:   player,
		Question: "Unlock a locked door of a Room you control",
		Options:  opts,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(picks) {
				return nil
			}
			return ctx.Game.UnlockDoorForEffect(picks[index].room, picks[index].door, player)
		},
	}.Apply(ctx)
}

// chooseDoorThen asks `player` which door of `room` to act on — a
// locked one to unlock when `unlock`, an unlocked one to lock when
// `lock` — and does it. Asks only when there is more than one door to
// choose between (CR 709.5f-g: the player chooses).
func chooseDoorThen(ctx *Context, player, room uuid.UUID, unlock, lock bool, then func(*Context) error) error {
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	finish := func(ctx *Context) error {
		if then != nil {
			return then(ctx)
		}
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(room)
	if !ok || !game.HasSharedTypeLine(c) || !ctx.Game.IsUnlockableRoomForEffect(room) {
		return finish(ctx)
	}
	type act struct {
		door   game.DoorSide
		unlock bool
	}
	var acts []act
	var opts []game.ChoiceOption
	for _, door := range []game.DoorSide{game.DoorLeft, game.DoorRight} {
		name := c.Faces[game.FaceOfDoor(door)].Name
		switch {
		case !c.Unlocked.Has(door) && unlock:
			acts = append(acts, act{door, true})
			opts = append(opts, game.ChoiceOption{Label: "Unlock " + name})
		case c.Unlocked.Has(door) && lock:
			acts = append(acts, act{door, false})
			opts = append(opts, game.ChoiceOption{Label: "Lock " + name})
		}
	}
	do := func(ctx *Context, a act) error {
		var err error
		if a.unlock {
			err = ctx.Game.UnlockDoorForEffect(room, a.door, player)
		} else {
			err = ctx.Game.LockDoorForEffect(room, a.door, player)
		}
		if err != nil {
			return err
		}
		return finish(ctx)
	}
	switch len(acts) {
	case 0:
		return finish(ctx)
	case 1:
		return do(ctx, acts[0])
	}
	return PickOption{
		Player:   player,
		Question: "Choose a door of " + c.Faces[0].Name + " // " + c.Faces[1].Name,
		Options:  opts,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(acts) {
				return finish(ctx)
			}
			return do(ctx, acts[index])
		},
	}.Apply(ctx)
}
