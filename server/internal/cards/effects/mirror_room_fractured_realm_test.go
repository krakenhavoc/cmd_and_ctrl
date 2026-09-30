package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mirrorRoomOracle = "e7a1c21b-5966-4ce0-a925-0b3cb3c3db00"

func mirrorRoomCard(owner uuid.UUID) game.Card {
	return roomsDCard(owner, mirrorRoomOracle, "Mirror Room", "{2}{U}", "Fractured Realm", "{5}{U}{U}", "U")
}

// reflections counts the Reflection tokens copying "Grizzly Bears".
func reflections(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Grizzly Bears" && IsToken(c) && c.Controller == owner && c.HasSubtype("Reflection") && c.HasSubtype("Bear") {
			n++
		}
	}
	return n
}

// Mirror Room's unlock trigger copies the target creature with the
// Reflection type added to its own.
func TestMirrorRoomCopiesACreatureAsAReflection(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b16Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2, "G")

	roomsDCast(t, g, me, mirrorRoomCard(me.ID), 0)

	if n := reflections(g, me.ID); n != 1 {
		t.Fatalf("Reflection copies = %d, want 1", n)
	}
}

// Fractured Realm doubles a triggered ability of a permanent you
// control: with its door unlocked, a second Mirror Room's unlock makes
// two copies. The door is gated, so the locked half does nothing.
func TestFracturedRealmDoublesTriggeredAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b16Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2, "G")

	// Fractured Realm (right half) on the battlefield, door unlocked.
	roomsDCast(t, g, me, mirrorRoomCard(me.ID), 1)
	if findBattlefieldByName(g, "Fractured Realm") == uuid.Nil {
		t.Fatal("Fractured Realm is not on the battlefield")
	}
	// A second Room cast on its Mirror Room half: its trigger is doubled.
	roomsDCast(t, g, me, mirrorRoomCard(me.ID), 0)
	if n := reflections(g, me.ID); n != 2 {
		t.Fatalf("Reflection copies with Fractured Realm unlocked = %d, want 2", n)
	}
}

// With only Mirror Room unlocked (Fractured Realm still locked), there is
// no doubling: one Room cast on each half is one copy from its own trigger.
func TestFracturedRealmLockedDoesNotDouble(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b16Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2, "G")

	roomsDCast(t, g, me, mirrorRoomCard(me.ID), 0)
	roomsDCast(t, g, me, mirrorRoomCard(me.ID), 0)
	if n := reflections(g, me.ID); n != 2 {
		t.Fatalf("two Mirror Room unlocks, no doubler: copies = %d, want 2", n)
	}
}
