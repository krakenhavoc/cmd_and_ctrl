package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// rooms_test.go — Rooms (CR 709.5), ADR 0103, against an uncatalogued
// fixture (so what is pinned is the engine's lifecycle, which every
// Room gets whether or not a card file exists) plus stub catalog
// entries where a door gate has to be seen switching.

const roomOracle = "7a0b1c2d-3e4f-5061-7283-94a5b6c7d8e9"

// roomFixture is a Room with a cheap left door and a dear right one,
// in two colours — Roaring Furnace // Steaming Sauna's shape.
func roomFixture(owner uuid.UUID) Card {
	return splitFixture(owner,
		Face{Name: "Front Hall", TypeLine: "Enchantment — Room", ManaCost: "{1}{R}", Colors: []string{"R"}, OracleText: "Front text."},
		Face{Name: "Back Room", TypeLine: "Enchantment — Room", ManaCost: "{3}{U}{U}", Colors: []string{"U"}, OracleText: "Back text."},
	)
}

func roomInHand(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	c := roomFixture(p.ID)
	c.OracleID = roomOracle
	return putInHand(t, g, p, c)
}

// castRoom casts `face` of a Room from hand and resolves it.
func castRoom(t *testing.T, g *Game, p *Player, face int) uuid.UUID {
	t.Helper()
	id := roomInHand(t, g, p)
	if err := g.CastSpell(p.ID, id, CastSpellParams{Face: face}); err != nil {
		t.Fatalf("cast face %d of the Room: %v", face, err)
	}
	passPriorityUntilResolvedForTest(t, g, id)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Room did not resolve onto the battlefield")
	}
	return id
}

func roomOnBattlefield(g *Game, id uuid.UUID) Card {
	var out Card
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, id); c != nil {
			out = *c
		}
	})
	return out
}

func eventIndex(g *Game, kind EventKind, card uuid.UUID, from int) int {
	for i := from; i < len(g.Events); i++ {
		if g.Events[i].Kind == kind && g.Events[i].CardID == card {
			return i
		}
	}
	return -1
}

func TestRoomIsASharedTypeLineSplitCard(t *testing.T) {
	r := roomFixture(uuid.New())
	if !HasSharedTypeLine(r) {
		t.Fatal("a Room fixture is not a shared-type-line card")
	}
	if HasSharedTypeLine(plainSplit(uuid.New())) {
		t.Error("an instant // sorcery split card read as a Room")
	}
	r.SetFaceDown(FaceDownManifested)
	if HasSharedTypeLine(r) {
		t.Error("a face-down Room still has halves (CR 708.2a)")
	}
}

// TestRoomEntersWithTheCastDoorUnlocked is CR 709.5d and 709.5h: the
// half that was cast is unlocked as the permanent enters, its door's
// characteristics are the permanent's, "when you unlock this door"
// fires right after the ETB, and entering is never a full unlock.
func TestRoomEntersWithTheCastDoorUnlocked(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	before := len(g.Events)
	id := castRoom(t, g, me, 1)

	room := roomOnBattlefield(g, id)
	if !room.Unlocked.Has(DoorRight) || room.Unlocked.Has(DoorLeft) {
		t.Fatalf("Unlocked = %b, want the right door only", room.Unlocked)
	}
	if room.Name != "Back Room" || room.ManaCost != "{3}{U}{U}" || room.ManaValue() != 5 {
		t.Errorf("on the battlefield: %q %q, want the right half's name and cost", room.Name, room.ManaCost)
	}
	if colors := printedColors(room); len(colors) != 1 || colors[0] != "U" {
		t.Errorf("colours = %v, want blue only (the locked half has no mana cost)", colors)
	}
	etb := eventIndex(g, EventETB, id, before)
	unlocked := eventIndex(g, EventDoorUnlocked, id, before)
	if etb < 0 || unlocked < 0 || unlocked < etb {
		t.Fatalf("ETB at %d, door unlocked at %d: want the unlock right after the ETB (CR 709.5h)", etb, unlocked)
	}
	if ev := g.Events[unlocked]; DoorSide(ev.Amount) != DoorRight || ev.Label != "Back Room" || ev.Actor != me.ID {
		t.Errorf("door event = %+v, want the right door, its name, and the caster", ev)
	}
	if eventIndex(g, EventRoomFullyUnlocked, id, before) >= 0 {
		t.Error("entering with one door was announced as a full unlock")
	}
}

// TestRoomPutOntoTheBattlefieldIsFullyLocked is CR 709.5d's last
// sentence, and CR 709.5's statics: no designation, so no name, no
// mana cost, mana value 0 and no colour.
func TestRoomPutOntoTheBattlefieldIsFullyLocked(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := roomInHand(t, g, me)
	g.WithWriteLock(func() {
		if _, err := MoveCard(me.Hand, g.Battlefield, id); err != nil {
			t.Fatal(err)
		}
	})
	room := roomOnBattlefield(g, id)
	if room.Unlocked != 0 || room.Name != "" || room.ManaCost != "" || room.ManaValue() != 0 {
		t.Fatalf("a Room put onto the battlefield: unlocked %b, name %q, cost %q", room.Unlocked, room.Name, room.ManaCost)
	}
	if colors := printedColors(room); len(colors) != 0 {
		t.Errorf("colours = %v, want none", colors)
	}
	if room.TypeLine != "Enchantment — Room" || !room.IsEnchantment() {
		t.Errorf("type line %q: the shared type line stays (CR 709.5a)", room.TypeLine)
	}
	if offers := UnlockOffers(room); len(offers) != 2 {
		t.Errorf("UnlockOffers = %+v, want both doors", offers)
	}
}

// TestUnlockSpecialAction is CR 709.5e / 116.2m: pay the locked door's
// cost, any time you have priority with the stack empty in your main
// phase; the second door is a full unlock (CR 709.5i).
func TestUnlockSpecialAction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := castRoom(t, g, me, 0)

	if err := g.PerformSpecialAction(me.ID, id, SpecialActionUnlock, SpecialActionParams{}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("unlock naming no door: err = %v, want ErrInvalidParam", err)
	}
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionUnlock, SpecialActionParams{Door: DoorLeft}); !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Errorf("unlock an unlocked door: err = %v, want ErrSpecialActionNotOffered", err)
	}
	before := len(g.Events)
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionUnlock, SpecialActionParams{Door: DoorRight}); err != nil {
		t.Fatalf("unlock the right door: %v", err)
	}
	room := roomOnBattlefield(g, id)
	if !room.Unlocked.Full() {
		t.Fatalf("Unlocked = %b, want both doors", room.Unlocked)
	}
	if room.Name != "Front Hall // Back Room" || room.ManaValue() != 7 {
		t.Errorf("fully unlocked: %q, mana value %d — want both names and 7", room.Name, room.ManaValue())
	}
	unlocked := eventIndex(g, EventDoorUnlocked, id, before)
	full := eventIndex(g, EventRoomFullyUnlocked, id, before)
	if unlocked < 0 || full < unlocked {
		t.Fatalf("door unlocked at %d, fully unlocked at %d: want both, in that order", unlocked, full)
	}
	if g.Events[full].Actor != me.ID {
		t.Errorf("fully unlocked by %v, want the player who unlocked it", g.Events[full].Actor)
	}
	if n := eventIndex(g, EventSpecialAction, id, before); n < 0 {
		t.Error("no EventSpecialAction for the unlock")
	}
	if offers := UnlockOffers(room); len(offers) != 0 {
		t.Errorf("a fully unlocked Room still offers %+v", offers)
	}
}

// TestUnlockTimingIsSorcerySpeed: not outside a main phase, not with
// the stack full, not on another player's turn (CR 709.5e).
func TestUnlockTimingIsSorcerySpeed(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := roomInHand(t, g, me)
	theirs := roomFixture(them.ID)
	g.WithWriteLock(func() {
		if _, err := MoveCard(me.Hand, g.Battlefield, id); err != nil {
			t.Fatal(err)
		}
		g.Battlefield.PushTop(theirs)
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].materialiseDoors()
		}
	})
	// Beginning phase: not a main phase.
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionUnlock, SpecialActionParams{Door: DoorLeft}); !errors.Is(err, ErrSpecialActionTiming) {
		t.Errorf("unlock in the beginning phase: err = %v, want ErrSpecialActionTiming", err)
	}
	toMainPhase(t, g)
	// The opponent's Room, on my turn.
	if err := g.PerformSpecialAction(them.ID, theirs.InstanceID, SpecialActionUnlock, SpecialActionParams{Door: DoorLeft}); !errors.Is(err, ErrSpecialActionTiming) {
		t.Errorf("unlock on another player's turn: err = %v, want ErrSpecialActionTiming", err)
	}
	// I cannot unlock their Room at all (CR 709.5e: its controller).
	if err := g.PerformSpecialAction(me.ID, theirs.InstanceID, SpecialActionUnlock, SpecialActionParams{Door: DoorLeft}); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("unlock somebody else's Room: err = %v, want ErrCardNotFound", err)
	}
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionUnlock, SpecialActionParams{Door: DoorLeft}); err != nil {
		t.Errorf("unlock in my main phase with an empty stack: %v", err)
	}
}

// TestDoorGateSwitchesTheDoorsAbilities: an ability printed on a door
// exists only while that door is unlocked (CR 709.5, ADR 0071 gate).
func TestDoorGateSwitchesTheDoorsAbilities(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	withCatalog(t, roomOracle, &CardDef{
		Static: []StaticAbility{{Layer: Layer6Ability, ActiveWhen: DoorUnlocked(DoorRight)}},
		Triggered: []TriggeredAbility{{
			Watches: []EventKind{EventDoorUnlocked}, Key: "right door trigger",
			ActiveWhen: DoorUnlocked(DoorRight),
		}},
		HandSize: []HandSizeStatic{{Kind: HandSizeNoMaximum, When: DoorUnlocked(DoorRight)}},
	})
	id := castRoom(t, g, me, 0)

	room := roomOnBattlefield(g, id)
	if n := len(StaticAbilitiesForCard(room)); n != 0 {
		t.Errorf("locked right door: %d statics, want 0", n)
	}
	if n := len(TriggersForCard(room)); n != 0 {
		t.Errorf("locked right door: %d triggers, want 0", n)
	}
	var max int
	g.WithWriteLock(func() { max = g.EffectiveMaxHandSizeLocked(me) })
	if max == NoMaxHandSize {
		t.Error("a locked door's \"no maximum hand size\" applied")
	}
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionUnlock, SpecialActionParams{Door: DoorRight}); err != nil {
		t.Fatal(err)
	}
	room = roomOnBattlefield(g, id)
	if n := len(StaticAbilitiesForCard(room)); n != 1 {
		t.Errorf("unlocked right door: %d statics, want 1", n)
	}
	if n := len(TriggersForCard(room)); n != 1 {
		t.Errorf("unlocked right door: %d triggers, want 1", n)
	}
	g.WithWriteLock(func() { max = g.EffectiveMaxHandSizeLocked(me) })
	if max != NoMaxHandSize {
		t.Error("an unlocked door's \"no maximum hand size\" did not apply")
	}
}

// TestLockAndLeave: CR 709.5g locks a door; CR 400.7 makes a Room that
// leaves and comes back a new object with both doors locked.
func TestLockAndLeave(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := castRoom(t, g, me, 1)

	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.LockDoorForEffect(id, DoorRight, me.ID); err != nil {
			t.Fatal(err)
		}
	})
	if room := roomOnBattlefield(g, id); room.Unlocked != 0 || room.Name != "" {
		t.Fatalf("after locking: unlocked %b, name %q", room.Unlocked, room.Name)
	}
	if eventIndex(g, EventDoorLocked, id, before) < 0 {
		t.Error("no EventDoorLocked")
	}
	g.WithWriteLock(func() {
		if err := g.UnlockDoorForEffect(id, DoorLeft, uuid.Nil); err != nil {
			t.Fatal(err)
		}
		if _, err := MoveCard(g.Battlefield, me.Graveyard, id); err != nil {
			t.Fatal(err)
		}
	})
	inYard := findCardForTest(me.Graveyard, id)
	if inYard.Unlocked != 0 || inYard.Name != "Front Hall // Back Room" || inYard.ManaValue() != 7 {
		t.Fatalf("in the graveyard: %+v, want the whole card with no doors", inYard)
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(me.Graveyard, g.Battlefield, id); err != nil {
			t.Fatal(err)
		}
	})
	if back := roomOnBattlefield(g, id); back.Unlocked != 0 || back.Name != "" {
		t.Errorf("back on the battlefield: unlocked %b, name %q — want fully locked (CR 400.7, 709.5d)", back.Unlocked, back.Name)
	}
}

// TestCopyOfARoomIsTheWholeCardWithNoDoors is CR 709.5 / 707.2: the
// halves are copiable, the designations are not.
func TestCopyOfARoomIsTheWholeCardWithNoDoors(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := castRoom(t, g, me, 1)
	v := CopiableValuesOf(roomOnBattlefield(g, id))
	if v.Name != "Front Hall // Back Room" || v.ManaCost != "{1}{R}{3}{U}{U}" || len(v.Faces) != 2 {
		t.Errorf("copiable values = %q %q (%d faces), want the whole card", v.Name, v.ManaCost, len(v.Faces))
	}
	var copyCard Card
	copyCard.setPrintedValues(v)
	copyCard.materialiseDoors()
	if copyCard.Unlocked != 0 || copyCard.Name != "" {
		t.Errorf("a copy materialised on the battlefield: %q, want no doors and no name", copyCard.Name)
	}
}

// TestFaceDownRoomOffersNoUnlock: a face-down Room is a 2/2 with no text.
func TestFaceDownRoomOffersNoUnlock(t *testing.T) {
	r := roomFixture(uuid.New())
	r.materialiseDoors()
	r.SetFaceDown(FaceDownManifested)
	if offers := SpecialActionsOfferedByCard(r); len(offers) != 0 {
		for _, o := range offers {
			if o.Kind == SpecialActionUnlock {
				t.Fatalf("a face-down Room offers %+v", o)
			}
		}
	}
}

// TestUnlockManaRestrictions: an unlock is paid with the unlock spend
// purpose, so mana restricted to casting does not pay it and mana
// restricted to unlocking does (Smoky Lounge, Creeping Peeper).
func TestUnlockManaRestrictions(t *testing.T) {
	unlock := specialActionSpendContext(SpecialActionUnlock)
	if unlock.allows([]string{ManaRestrictCast}) {
		t.Error("cast-only mana paid an unlock")
	}
	if !unlock.allows([]string{ManaRestrictUnlock}) {
		t.Error("unlock mana did not pay an unlock")
	}
	lounge := ManaRestrictAnyOf([]string{ManaRestrictCast, ManaRestrictSubtype("Room")}, []string{ManaRestrictUnlock})
	if !unlock.allows([]string{lounge}) {
		t.Error("Smoky Lounge mana did not pay an unlock")
	}
	roomSpell := ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Enchantment"}, Subtypes: []string{"Room"}}
	if !roomSpell.allows([]string{lounge}) {
		t.Error("Smoky Lounge mana did not pay for a Room spell")
	}
	other := ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Creature"}}
	if other.allows([]string{lounge}) {
		t.Error("Smoky Lounge mana paid for a creature spell")
	}
	if specialActionSpendContext(SpecialActionForetell).allows([]string{ManaRestrictUnlock}) {
		t.Error("unlock mana paid a foretell")
	}
}

// TestRoomSurvivesARestorePoint: the designations are carried (ADR
// 0103 Decision 12), so a restore brings the door abilities back.
func TestRoomSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := castRoom(t, g, me, 1)
	_, restored := roundTrip(t, g)
	room := roomOnBattlefield(restored, id)
	if !room.Unlocked.Has(DoorRight) || room.Unlocked.Has(DoorLeft) || room.Name != "Back Room" {
		t.Fatalf("restored Room: unlocked %b, name %q — want the right door and its name", room.Unlocked, room.Name)
	}
}

// TestCopyOfARoomSpellEntersFullyLocked is owner decision 3: a copy of
// a spell is not cast (CR 707.10), so CR 709.5d gives it no door.
func TestCopyOfARoomSpellEntersFullyLocked(t *testing.T) {
	spell := roomFixture(uuid.New())
	spell.SetFace(1)
	if got := castDoorsOf(spell, &StackItem{IsCopy: true}); got != 0 {
		t.Errorf("castDoorsOf(copy) = %b, want none", got)
	}
	if got := castDoorsOf(spell, &StackItem{}); got != DoorRightUnlocked {
		t.Errorf("castDoorsOf(cast right half) = %b, want the right door", got)
	}
}

// TestDoorGateReachesReplacementsAndUntapPermissions: the two slots
// ADR 0103 added the gate to (Torture Pit's replacement, Prop Room's
// untap permission) exist only while their door is unlocked.
func TestDoorGateReachesReplacementsAndUntapPermissions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	withCatalog(t, roomOracle, &CardDef{
		Replacements: []ReplacementEffect{{
			Watches:    []EventKind{EventCounterPlaced},
			AppliesTo:  func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventCounter },
			Replace:    func(ev *ReplacementEvent, _ *Game, _ *Card) error { ev.CounterDelta *= 2; return nil },
			Label:      "right door doubles counters",
			ActiveWhen: DoorUnlocked(DoorRight),
		}},
		UntapStep: []UntapStepPermission{{
			AppliesTo:  func(*Game, *Card, uuid.UUID) bool { return true },
			Untaps:     func(*Game, *Card, *Card) bool { return true },
			Label:      "right door untaps",
			ActiveWhen: DoorUnlocked(DoorRight),
		}},
	})
	id := castRoom(t, g, me, 0)
	bear := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	counters := func() int {
		var n int
		g.WithWriteLock(func() { n = findBattlefieldCard(g, bear).Counters[CounterPlusOne] })
		return n
	}
	perms := func() int {
		var n int
		g.WithWriteLock(func() { n = len(g.activeUntapStepPermissionsLocked(g.Seats[1].ID)) })
		return n
	}
	if err := g.AddCounter(bear, CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if got := counters(); got != 1 {
		t.Errorf("locked door: %d counters, want 1 (the replacement does not exist)", got)
	}
	if got := perms(); got != 0 {
		t.Errorf("locked door: %d untap permissions, want 0", got)
	}
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionUnlock, SpecialActionParams{Door: DoorRight}); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(bear, CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if got := counters(); got != 3 {
		t.Errorf("unlocked door: %d counters, want 3 (1 + doubled 2)", got)
	}
	if got := perms(); got != 1 {
		t.Errorf("unlocked door: %d untap permissions, want 1", got)
	}
}
