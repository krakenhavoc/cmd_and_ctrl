package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// convoke_grant_test.go — "<spells> you cast have convoke" (#2137,
// #3245): Hoarding Broodlord and Dazzling Theater.

const dazzlingTheaterOracle = "c46a02db-13d6-477f-9da0-822599470168"

func TestHoardingBroodlordGivesSpellsCastFromExileConvoke(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Hoarding Broodlord", "Creature — Dragon", effectstest.McBroodlordOracle, false)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)

	bear := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() {
		g.Exile.PushTop(bear)
		g.GrantCastPermissionOverCardForEffect(bear.InstanceID, game.CastPermission{
			Player: me.ID, Zone: game.ZoneExile, Duration: game.WhileInZoneDuration()})
	})
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{FromZone: "exile", Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a spell cast from exile: %v", err)
	}
	if !effectstest.CgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestHoardingBroodlordGivesNoConvokeFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Hoarding Broodlord", "Creature — Dragon", effectstest.McBroodlordOracle, false)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)
	bear := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Error("a spell cast from hand was convoked under Hoarding Broodlord")
	}
}

func TestNoConvokeFromExileWithoutBroodlord(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)
	bear := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() {
		g.Exile.PushTop(bear)
		g.GrantCastPermissionOverCardForEffect(bear.InstanceID, game.CastPermission{
			Player: me.ID, Zone: game.ZoneExile, Duration: game.WhileInZoneDuration()})
	})
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{FromZone: "exile", Strict: true, TapIDs: elves}); err == nil {
		t.Error("a spell was convoked with no grant")
	}
}

func TestDazzlingTheaterGivesCreatureSpellsConvokeOnceUnlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, dazzlingTheaterOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}", "W")

	// Prop Room cast first: the Theater door is still locked.
	effectstest.CastRoomC(t, g, me.ID, room, 1)
	effectstest.PassPriorityAroundTable(t, g)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)
	locked := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(locked) })
	if err := g.CastSpell(me.ID, locked.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a locked Dazzling Theater gave convoke")
	}

	effectstest.FillPoolColored(me, "W", 4)
	effectstest.UnlockDoorC(t, g, me.ID, room.InstanceID, game.DoorLeft)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, locked.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a creature spell under an unlocked Theater: %v", err)
	}
	if !effectstest.CgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestDazzlingTheaterGivesNoncreatureSpellsNoConvoke(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, dazzlingTheaterOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}", "W")
	effectstest.CastRoomC(t, g, me.ID, room, 0)
	effectstest.PassPriorityAroundTable(t, g)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)
	rite := game.Card{InstanceID: uuid.New(), Name: "Rite", TypeLine: "Sorcery", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(rite) })
	if err := g.CastSpell(me.ID, rite.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Error("a sorcery was convoked under Dazzling Theater")
	}
}
