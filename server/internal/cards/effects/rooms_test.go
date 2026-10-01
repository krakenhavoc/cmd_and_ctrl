package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_test.go — the card-facing half of ADR 0103: the Room
// constructor stamps a door gate on every ability, Register refuses an
// ungated one, and the trigger shapes fire off the engine's events —
// "when you unlock this door" as the Room enters with the cast door
// (CR 709.5h) and after an unlock, "whenever you fully unlock a Room"
// (CR 709.5i), and the eerie pair.

const testRoomOracle = "test-room-fixture-oracle"

func testRoomCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   testRoomOracle,
		Owner:      owner,
		Controller: owner,
		Layout:     game.LayoutSplit,
		Faces: []game.Face{
			{Name: "Test Foyer", TypeLine: "Enchantment — Room", ManaCost: "{1}{W}", Colors: []string{"W"}},
			{Name: "Test Vault", TypeLine: "Enchantment — Room", ManaCost: "{3}{W}", Colors: []string{"W"}},
		},
	}
	c.SettleImported()
	return c
}

func registerTestRoom(t *testing.T, counts map[string]int) {
	t.Helper()
	bump := func(key string) Effect {
		return func(*game.Game, *game.StackItem) error {
			counts[key]++
			return nil
		}
	}
	registerForTest(t, Room(RoomSpec{
		OracleID:     testRoomOracle,
		Name:         "Test Foyer // Test Vault",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Test Foyer — left unlocked", bump("left")),
		}},
		Right: Door{
			Triggered: []game.TriggeredAbility{
				WhenYouUnlockThisDoor(game.DoorRight, "Test Vault — right unlocked", bump("right")),
				WheneverYouFullyUnlockARoom("Test Vault — fully unlocked", bump("full")),
			},
			NoMaxHandSize: true,
		},
	}))
}

// TestRoomConstructorGatesEveryDoor: every ability Room builds carries
// its own door's gate, and the Spec passes the Register guard.
func TestRoomConstructorGatesEveryDoor(t *testing.T) {
	s := Room(RoomSpec{
		OracleID: "test-room-gates",
		Name:     "Gate // Test",
		Left:     Door{Static: []game.StaticAbility{{Layer: game.Layer6Ability}}},
		Right: Door{
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
	if problem := checkRoomSpec(s); problem != "" {
		t.Errorf("checkRoomSpec: %s", problem)
	}
}

// TestRegisterRefusesAnUngatedRoomAbilityAndAStrayDoorGate is the
// Register guard (ADR 0103): an ability on a Room that is not gated on
// a door, a slot a door cannot reach, and a door gate on a non-Room.
func TestRegisterRefusesAnUngatedRoomAbilityAndAStrayDoorGate(t *testing.T) {
	ungated := Room(RoomSpec{OracleID: "test-room-ungated", Name: "Ungated"})
	ungated.Triggered = append(ungated.Triggered, game.TriggeredAbility{Key: "free"})
	if p := checkRoomSpec(ungated); !strings.Contains(p, "not gated") {
		t.Errorf("ungated Room ability: %q", p)
	}
	mana := Room(RoomSpec{OracleID: "test-room-mana", Name: "Mana"})
	mana.ManaAbilities = []ManaAbility{{Produced: "{W}", Label: "Add {W}"}}
	if p := checkRoomSpec(mana); !strings.Contains(p, "cannot reach") {
		t.Errorf("a Room mana ability: %q", p)
	}
	stray := Spec{OracleID: "test-stray-door", Name: "Stray", Static: []game.StaticAbility{{ActiveWhen: game.DoorUnlocked(game.DoorLeft)}}}
	if p := checkRoomSpec(stray); !strings.Contains(p, "not built with effects.Room") {
		t.Errorf("a door gate outside a Room: %q", p)
	}
	defer func() {
		if recover() == nil {
			t.Error("Register accepted an ungated Room ability")
		}
	}()
	Register(ungated)
}

// TestUnlockTriggersFireOnEntryAndOnUnlock is CR 709.5h and 709.5i
// through the real harvester.
func TestUnlockTriggersFireOnEntryAndOnUnlock(t *testing.T) {
	counts := map[string]int{}
	registerTestRoom(t, counts)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := testRoomCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatalf("cast Test Foyer: %v", err)
	}
	settleOrdering(t, g, me.ID)
	if counts["left"] != 1 || counts["right"] != 0 || counts["full"] != 0 {
		t.Fatalf("after casting the left half: %v, want the left door's trigger once (CR 709.5h)", counts)
	}
	if err := g.PerformSpecialAction(me.ID, c.InstanceID, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorRight}); err != nil {
		t.Fatalf("unlock the right door: %v", err)
	}
	settleOrdering(t, g, me.ID)
	if counts["left"] != 1 || counts["right"] != 1 || counts["full"] != 1 {
		t.Fatalf("after unlocking the right door: %v, want right and full once each", counts)
	}
}

// TestEerieFiresOnAnEnchantmentAndOnAFullUnlock is the Duskmourn eerie
// line's two trigger conditions.
func TestEerieFiresOnAnEnchantmentAndOnAFullUnlock(t *testing.T) {
	counts := map[string]int{}
	registerTestRoom(t, map[string]int{})
	const watcher = "test-eerie-watcher"
	registerForTest(t, Spec{
		OracleID: watcher, Name: "Eerie Watcher",
		Triggered: []game.TriggeredAbility{Eerie("Eerie Watcher — eerie", func(*game.Game, *game.StackItem) error {
			counts["eerie"]++
			return nil
		})},
	})
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Eerie Watcher", TypeLine: "Creature — Spirit",
			OracleID: watcher, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	c := testRoomCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Test Vault: %v", err)
	}
	settleOrdering(t, g, me.ID)
	if counts["eerie"] != 1 {
		t.Fatalf("an enchantment entering: eerie fired %d times, want 1", counts["eerie"])
	}
	if err := g.PerformSpecialAction(me.ID, c.InstanceID, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorLeft}); err != nil {
		t.Fatal(err)
	}
	settleOrdering(t, g, me.ID)
	if counts["eerie"] != 2 {
		t.Fatalf("a full unlock: eerie fired %d times in all, want 2", counts["eerie"])
	}
	if n := UnlockedDoorsYouControl(g, me.ID); n != 2 {
		t.Errorf("UnlockedDoorsYouControl = %d, want 2", n)
	}
	if names := UnlockedDoorNamesYouControl(g, me.ID); len(names) != 2 {
		t.Errorf("UnlockedDoorNamesYouControl = %v, want both doors", names)
	}
	if !IsFullyUnlocked(g, c.InstanceID) {
		t.Error("IsFullyUnlocked = false")
	}
}

// TestUnlockADoorInstruction is CR 709.5f with one locked door (no
// prompt) and LockOrUnlockADoor with both open (an option_pick).
func TestUnlockADoorInstruction(t *testing.T) {
	registerTestRoom(t, map[string]int{})
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := testRoomCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatal(err)
	}
	settleOrdering(t, g, me.ID)
	item := &game.StackItem{Controller: me.ID, SourceCardID: c.InstanceID}
	g.WithWriteLock(func() {
		if err := (UnlockADoor{Room: c.InstanceID}).Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	if !IsFullyUnlocked(g, c.InstanceID) {
		t.Fatal("UnlockADoor with one locked door did not unlock it")
	}
	g.WithWriteLock(func() {
		if err := (LockOrUnlockADoor{Room: c.InstanceID}).Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.PendingChoices) == 0 {
		t.Fatal("LockOrUnlockADoor with two doors to choose from asked nothing")
	}
}

// settleOrdering resolves everything, answering each CR 603.3b
// trigger_order prompt `chooser` gets in the order offered — a door
// trigger and a full-unlock trigger (or an eerie trigger) go on the
// stack together.
func settleOrdering(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
				if err := g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}
