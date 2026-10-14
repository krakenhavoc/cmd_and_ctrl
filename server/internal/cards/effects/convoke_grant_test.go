package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// convoke_grant_test.go — "<spells> you cast have convoke" (#2137,
// #3245): Hoarding Broodlord and Dazzling Theater.

const dazzlingTheaterOracle = "c46a02db-13d6-477f-9da0-822599470168"

// cgTwoGreenCreatures puts two untapped green 1/1s under `owner`.
func cgTwoGreenCreatures(g *game.Game, owner uuid.UUID) []uuid.UUID {
	return []uuid.UUID{
		pushTapCostPermanent(g, owner, "Elf A", "Creature — Elf", "G"),
		pushTapCostPermanent(g, owner, "Elf B", "Creature — Elf", "G"),
	}
}

// cgBear is a {1}{G} vanilla creature card owned by `owner`.
func cgBear(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: owner, Controller: owner}
}

func cgTapped(g *game.Game, ids []uuid.UUID) bool {
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); !ok || !c.Tapped {
			return false
		}
	}
	return true
}

func TestHoardingBroodlordGivesSpellsCastFromExileConvoke(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Hoarding Broodlord", "Creature — Dragon", mcBroodlordOracle, false)
	elves := cgTwoGreenCreatures(g, me.ID)

	bear := cgBear(me.ID)
	g.WithWriteLock(func() {
		g.Exile.PushTop(bear)
		g.GrantCastPermissionOverCardForEffect(bear.InstanceID, game.CastPermission{
			Player: me.ID, Zone: game.ZoneExile, Duration: game.WhileInZoneDuration()})
	})
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{FromZone: "exile", Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a spell cast from exile: %v", err)
	}
	if !cgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestHoardingBroodlordGivesNoConvokeFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Hoarding Broodlord", "Creature — Dragon", mcBroodlordOracle, false)
	elves := cgTwoGreenCreatures(g, me.ID)
	bear := cgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Error("a spell cast from hand was convoked under Hoarding Broodlord")
	}
}

func TestNoConvokeFromExileWithoutBroodlord(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	elves := cgTwoGreenCreatures(g, me.ID)
	bear := cgBear(me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	room := roomCardC(me.ID, dazzlingTheaterOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}", "W")

	// Prop Room cast first: the Theater door is still locked.
	castRoomC(t, g, me.ID, room, 1)
	passPriorityAroundTable(t, g)
	elves := cgTwoGreenCreatures(g, me.ID)
	locked := cgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(locked) })
	if err := g.CastSpell(me.ID, locked.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a locked Dazzling Theater gave convoke")
	}

	fillPoolColored(me, "W", 4)
	unlockDoorC(t, g, me.ID, room.InstanceID, game.DoorLeft)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, locked.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a creature spell under an unlocked Theater: %v", err)
	}
	if !cgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestDazzlingTheaterGivesNoncreatureSpellsNoConvoke(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	room := roomCardC(me.ID, dazzlingTheaterOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}", "W")
	castRoomC(t, g, me.ID, room, 0)
	passPriorityAroundTable(t, g)
	elves := cgTwoGreenCreatures(g, me.ID)
	rite := game.Card{InstanceID: uuid.New(), Name: "Rite", TypeLine: "Sorcery", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(rite) })
	if err := g.CastSpell(me.ID, rite.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Error("a sorcery was convoked under Dazzling Theater")
	}
}
