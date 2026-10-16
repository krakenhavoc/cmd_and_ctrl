package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_x_cast_cost_view_test.go — #1948: a hand card whose additional
// cost is "exile X creature cards from your graveyard" ships the X prompt
// flag, the picker's options and the largest X the graveyard pays, off
// the walk the validator reads; Soaring Stoneglider's exile branch ships
// the fixed picker and is payable only with two cards.

func TestChillHauntingCarriesItsExileXPicker(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	var creatures []string
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Chill Haunting", TypeLine: "Instant", OracleID: "ebe61471-e236-4ff8-9416-151970b5efd6",
			ManaCost: "{1}{B}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
		for range 2 {
			c := game.Card{InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie", Owner: me.ID, Controller: me.ID}
			creatures = append(creatures, c.InstanceID.String())
			me.Graveyard.PushTop(c)
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	})

	ac := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String()).AdditionalCost
	if ac == nil || !ac.ExileCostX || !ac.DemandsX || ac.ExileCostN != 0 {
		t.Fatalf("additional_cost = %+v, want an exile X cost that demands X", ac)
	}
	if ac.ExileCostXMax != 2 || len(ac.ExileCostOptions) != 2 {
		t.Errorf("x max %d with options %v, want 2 (the two creature cards)", ac.ExileCostXMax, ac.ExileCostOptions)
	}
	for i, s := range creatures {
		if i >= len(ac.ExileCostOptions) || ac.ExileCostOptions[i] != s {
			t.Errorf("options %v, want the creature cards %v", ac.ExileCostOptions, creatures)
			break
		}
	}
	if ac.ExileCostZone != "graveyard" || ac.ExileCostLabel != "X creature cards" {
		t.Errorf("zone %q label %q", ac.ExileCostZone, ac.ExileCostLabel)
	}
}

func TestSoaringStonegliderShipsItsExileBranch(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Soaring Stoneglider", TypeLine: "Creature — Elephant Cleric", OracleID: "62d18d2b-fa68-49cd-b62d-803728ddfa8b",
			ManaCost: "{2}{W}", Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
		me.Graveyard.PushTop(game.NewCard("Fuel", me.ID))
	})
	branches := func() []AdditionalCostView {
		ac := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String()).AdditionalCost
		if ac == nil || len(ac.Branches) != 2 {
			t.Fatalf("additional_cost = %+v, want two branches", ac)
		}
		return ac.Branches
	}
	b := branches()
	if b[0].Key != "exile" || b[0].ExileCostN != 2 || len(b[0].ExileCostOptions) != 1 || b[0].Payable {
		t.Errorf("exile branch = %+v, want two cards, one option, not payable", b[0])
	}
	if b[1].Key != "mana" || b[1].ManaCost != "{1}{W}" || !b[1].Payable {
		t.Errorf("mana branch = %+v", b[1])
	}
	g.WithWriteLock(func() { me.Graveyard.PushTop(game.NewCard("Fuel", me.ID)) })
	if b = branches(); !b[0].Payable || len(b[0].ExileCostOptions) != 2 {
		t.Errorf("with two cards the exile branch = %+v, want payable", b[0])
	}
}
