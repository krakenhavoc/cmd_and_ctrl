package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const grandEntrywayOracle = "f7574379-d10e-4bbf-b0b1-13feee753d8b"

// TestGrandEntrywayMakesAGlimmerOnUnlock casts the left half (the door
// unlocks as the Room enters, CR 709.5h) and then unlocks the right
// door for the counters.
func TestGrandEntrywayMakesAGlimmerAndElegantRotundaPutsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B16Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	c := effectstest.B16Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	room := effectstest.RoomCardC(me.ID, grandEntrywayOracle, "Grand Entryway", "{1}{W}", "Elegant Rotunda", "{2}{W}", "W")
	effectstest.CastRoomC(t, g, me.ID, room, 0)
	effectstest.SettleOrdering(t, g, me.ID)
	id := effectstest.FindBattlefieldByName(g, "Glimmer")
	if id == uuid.Nil {
		t.Fatal("unlocking Grand Entryway made no Glimmer")
	}
	glimmer, _ := effectstest.BattlefieldCardByID(g, id)
	if glimmer.Power != 1 || glimmer.Toughness != 1 || !glimmer.HasColor("W") ||
		!glimmer.IsCreature() || !glimmer.IsEnchantment() {
		t.Errorf("the token is %d/%d %v %q, want a 1/1 white enchantment creature", glimmer.Power, glimmer.Toughness, glimmer.Colors, glimmer.TypeLine)
	}
	if p, _ := effectstest.SizeOfC(t, g, a); p != 2 {
		t.Fatalf("the locked Elegant Rotunda put a counter on a Bear (%d power)", p)
	}

	effectstest.UnlockDoorC(t, g, me.ID, room.InstanceID, game.DoorRight)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.SettleOrdering(t, g, me.ID)
	if p, _ := effectstest.SizeOfC(t, g, a); p != 3 {
		t.Errorf("first target: %d power, want 3", p)
	}
	if p, _ := effectstest.SizeOfC(t, g, b); p != 3 {
		t.Errorf("second target: %d power, want 3", p)
	}
	if p, _ := effectstest.SizeOfC(t, g, c); p != 3 {
		t.Errorf("an unchosen creature: %d power, want 3 (unchanged)", p)
	}
	if !effects.IsFullyUnlocked(g, room.InstanceID) {
		t.Error("the Room is not fully unlocked")
	}
}
