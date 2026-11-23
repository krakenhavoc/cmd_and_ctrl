package testsr

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestRoomConstructorGatesEveryDoor: every ability Room builds carries
// its own door's gate, and the Spec passes the Register guard.
func TestRoomConstructorGatesEveryDoor(t *testing.T) {
	s := effects.Room(effects.RoomSpec{
		OracleID: "test-room-gates",
		Name:     "Gate // Test",
		Left:     effects.Door{Static: []game.StaticAbility{{Layer: game.Layer6Ability}}},
		Right: effects.Door{
			Triggered:    []game.TriggeredAbility{{Key: "right"}},
			Replacements: []game.ReplacementEffect{{Label: "right rep"}},
			UntapStep:    []game.UntapStepPermission{{Label: "right untap"}},
		},
	})
	if s.Static[0].ActiveWhen != game.DoorUnlocked(game.DoorLeft) {
		t.Errorf("left static gate = %+v", s.Static[0].ActiveWhen)
	}
	if s.Triggered[0].ActiveWhen != game.DoorUnlocked(game.DoorRight) ||
		s.Replacements[0].ActiveWhen != game.DoorUnlocked(game.DoorRight) ||
		s.UntapStep[0].ActiveWhen != game.DoorUnlocked(game.DoorRight) {
		t.Error("a right-door ability is not gated on the right door")
	}
	if problem := effects.CheckRoomSpec(s); problem != "" {
		t.Errorf("checkRoomSpec: %s", problem)
	}
}

// TestRegisterRefusesAnUngatedRoomAbilityAndAStrayDoorGate is the
// Register guard (ADR 0103): an ability on a Room that is not gated on
// a door, a slot a door cannot reach, and a door gate on a non-Room.
func TestRegisterRefusesAnUngatedRoomAbilityAndAStrayDoorGate(t *testing.T) {
	ungated := effects.Room(effects.RoomSpec{OracleID: "test-room-ungated", Name: "Ungated"})
	ungated.Triggered = append(ungated.Triggered, game.TriggeredAbility{Key: "free"})
	if p := effects.CheckRoomSpec(ungated); !strings.Contains(p, "not gated") {
		t.Errorf("ungated Room ability: %q", p)
	}
	mana := effects.Room(effects.RoomSpec{OracleID: "test-room-mana", Name: "Mana"})
	mana.ManaAbilities = []effects.ManaAbility{{Produced: "{W}", Label: "Add {W}"}}
	if p := effects.CheckRoomSpec(mana); !strings.Contains(p, "cannot reach") {
		t.Errorf("a Room mana ability: %q", p)
	}
	stray := effects.Spec{OracleID: "test-stray-door", Name: "Stray", Static: []game.StaticAbility{{ActiveWhen: game.DoorUnlocked(game.DoorLeft)}}}
	if p := effects.CheckRoomSpec(stray); !strings.Contains(p, "not built with effects.Room") {
		t.Errorf("a door gate outside a Room: %q", p)
	}
	defer func() {
		if recover() == nil {
			t.Error("Register accepted an ungated Room ability")
		}
	}()
	effects.Register(ungated)
}

// TestUnlockTriggersFireOnEntryAndOnUnlock is CR 709.5h and 709.5i
// through the real harvester.
func TestUnlockTriggersFireOnEntryAndOnUnlock(t *testing.T) {
	counts := map[string]int{}
	effectstest.RegisterTestRoom(t, counts)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.TestRoomCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatalf("cast Test Foyer: %v", err)
	}
	effectstest.SettleOrdering(t, g, me.ID)
	if counts["left"] != 1 || counts["right"] != 0 || counts["full"] != 0 {
		t.Fatalf("after casting the left half: %v, want the left door's trigger once (CR 709.5h)", counts)
	}
	if err := g.PerformSpecialAction(me.ID, c.InstanceID, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorRight}); err != nil {
		t.Fatalf("unlock the right door: %v", err)
	}
	effectstest.SettleOrdering(t, g, me.ID)
	if counts["left"] != 1 || counts["right"] != 1 || counts["full"] != 1 {
		t.Fatalf("after unlocking the right door: %v, want right and full once each", counts)
	}
}

// TestEerieFiresOnAnEnchantmentAndOnAFullUnlock is the Duskmourn eerie
// line's two trigger conditions.
func TestEerieFiresOnAnEnchantmentAndOnAFullUnlock(t *testing.T) {
	counts := map[string]int{}
	effectstest.RegisterTestRoom(t, map[string]int{})
	const watcher = "test-eerie-watcher"
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: watcher, Name: "Eerie Watcher",
		Triggered: []game.TriggeredAbility{effects.Eerie("Eerie Watcher — eerie", func(*game.Game, *game.StackItem) error {
			counts["eerie"]++
			return nil
		})},
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Eerie Watcher", TypeLine: "Creature — Spirit",
			OracleID: watcher, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	c := effectstest.TestRoomCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Test Vault: %v", err)
	}
	effectstest.SettleOrdering(t, g, me.ID)
	if counts["eerie"] != 1 {
		t.Fatalf("an enchantment entering: eerie fired %d times, want 1", counts["eerie"])
	}
	if err := g.PerformSpecialAction(me.ID, c.InstanceID, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorLeft}); err != nil {
		t.Fatal(err)
	}
	effectstest.SettleOrdering(t, g, me.ID)
	if counts["eerie"] != 2 {
		t.Fatalf("a full unlock: eerie fired %d times in all, want 2", counts["eerie"])
	}
	if n := effects.UnlockedDoorsYouControl(g, me.ID); n != 2 {
		t.Errorf("UnlockedDoorsYouControl = %d, want 2", n)
	}
	if names := effects.UnlockedDoorNamesYouControl(g, me.ID); len(names) != 2 {
		t.Errorf("UnlockedDoorNamesYouControl = %v, want both doors", names)
	}
	if !effects.IsFullyUnlocked(g, c.InstanceID) {
		t.Error("IsFullyUnlocked = false")
	}
}

// TestUnlockADoorInstruction is CR 709.5f with one locked door (no
// prompt) and LockOrUnlockADoor with both open (an option_pick).
func TestUnlockADoorInstruction(t *testing.T) {
	effectstest.RegisterTestRoom(t, map[string]int{})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.TestRoomCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatal(err)
	}
	effectstest.SettleOrdering(t, g, me.ID)
	item := &game.StackItem{Controller: me.ID, SourceCardID: c.InstanceID}
	g.WithWriteLock(func() {
		if err := (effects.UnlockADoor{Room: c.InstanceID}).Apply(effects.NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	if !effects.IsFullyUnlocked(g, c.InstanceID) {
		t.Fatal("UnlockADoor with one locked door did not unlock it")
	}
	g.WithWriteLock(func() {
		if err := (effects.LockOrUnlockADoor{Room: c.InstanceID}).Apply(effects.NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.PendingChoices) == 0 {
		t.Fatal("LockOrUnlockADoor with two doors to choose from asked nothing")
	}
}
