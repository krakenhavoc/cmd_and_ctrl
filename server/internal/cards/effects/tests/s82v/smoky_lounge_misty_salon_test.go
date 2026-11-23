package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const smokyLoungeOracle = "4f70b0ee-c24c-45fa-b878-9ba69266344f"

func smokyLoungeCard(owner uuid.UUID) game.Card {
	return effectstest.RoomsDCard(owner, smokyLoungeOracle, "Smoky Lounge", "{2}{R}", "Misty Salon", "{3}{U}", "R")
}

// Smoky Lounge adds {R}{R} at the beginning of its controller's first
// main phase, restricted to Room spells and unlocking doors. With only
// the other door unlocked it adds nothing.
func TestSmokyLoungeAddsRestrictedManaInTheFirstMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RoomsDOnBattlefield(g, me.ID, smokyLoungeOracle, "Smoky Lounge", "Misty Salon", game.DoorLeftUnlocked)
	effectstest.AdvanceToMain(t, g)
	effectstest.RoomsDSettle(t, g, me.ID)

	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %+v, want two mana", me.ManaPool)
	}
	want := game.ManaRestrictAnyOf(
		[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Room")},
		[]string{game.ManaRestrictUnlock},
	)
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "R" || len(tok.Restrictions) != 1 || tok.Restrictions[0] != want {
			t.Errorf("token = %+v, want {R} restricted to Room spells and unlocking", tok)
		}
	}
}

func TestSmokyLoungeLockedAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RoomsDOnBattlefield(g, me.ID, smokyLoungeOracle, "Smoky Lounge", "Misty Salon", game.DoorRightUnlocked)
	effectstest.AdvanceToMain(t, g)
	effectstest.RoomsDSettle(t, g, me.ID)
	if me.ManaPool.Total() != 0 {
		t.Fatalf("pool = %+v, want empty: the Smoky Lounge door is locked", me.ManaPool)
	}
}

// Misty Salon makes an X/X blue flying Spirit, X counting the unlocked
// doors among Rooms you control as the ability resolves: the door just
// unlocked, plus two on another Room.
func TestMistySalonSizesTheSpiritByUnlockedDoors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.RoomsDOnBattlefield(g, me.ID, effectstest.RoomsDLibraryOracle, "Other One", "Other Two", game.DoorLeftUnlocked|game.DoorRightUnlocked)

	effectstest.RoomsDCast(t, g, me, smokyLoungeCard(me.ID), 1)

	id := effectstest.FindBattlefieldByName(g, "Spirit")
	if id == uuid.Nil {
		t.Fatal("no Spirit token")
	}
	c, _ := effectstest.BattlefieldCard(g, id)
	if c.Power != 3 || c.Toughness != 3 || !effectstest.Contains(c.Keywords, "flying") || !effects.IsToken(c) || !effectstest.Contains(c.Colors, "U") {
		t.Errorf("Spirit = %+v, want a blue 3/3 flying token", c)
	}
	// Unlocking the other door is a full unlock and makes nothing more.
	effectstest.RoomsDUnlock(t, g, me, effectstest.FindBattlefieldByName(g, "Misty Salon"), game.DoorLeft)
	if n := effectstest.RoomsDCountNamed(g, "Spirit"); n != 1 {
		t.Errorf("Spirit tokens = %d after unlocking Smoky Lounge, want 1", n)
	}
}
