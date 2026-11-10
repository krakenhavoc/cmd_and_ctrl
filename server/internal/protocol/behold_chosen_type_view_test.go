package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Celestial Reunion
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// behold_chosen_type_view_test.go — #5406: a "choose a creature type and
// behold two creatures of that type" offer ships the count, the types the
// viewer could choose (each with its objects) and every object of every
// choosable type, off the engine's own walk (#544).

const viewCelestialReunionOracle = "ba0a815c-1dfe-47c1-890f-496c670124f3"

func reunionOffer(t *testing.T, g *game.Game, me *game.Player) OptionalCostView {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Celestial Reunion", TypeLine: "Sorcery", OracleID: viewCelestialReunionOracle,
			ManaCost: "{X}{G}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if len(c.OptionalCosts) != 1 {
		t.Fatalf("optional_costs = %+v, want the one behold offer", c.OptionalCosts)
	}
	return c.OptionalCosts[0]
}

func TestChosenTypeBeholdOfferListsTypesAndCreatures(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elfHand, elfTable := uuid.New(), uuid.New()
	known := map[uuid.UUID]bool{me.ID: true}
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: elfHand, Name: "Elf Friend", TypeLine: "Creature — Elf Druid", Owner: me.ID, Controller: me.ID, KnownBy: known})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin", Owner: me.ID, Controller: me.ID, KnownBy: known})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Elf Rite", TypeLine: "Kindred Instant — Elf", Owner: me.ID, Controller: me.ID, KnownBy: known})
		g.Battlefield.PushTop(game.Card{InstanceID: elfTable, Name: "Elf Scout", TypeLine: "Creature — Elf Scout", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	})
	o := reunionOffer(t, g, me)
	if o.Key != game.BeholdKey || !o.Reveal || !o.Behold || o.BeholdCount != 2 {
		t.Fatalf("offer = %+v, want a behold of two", o)
	}
	if len(o.BeholdTypes) != 1 || o.BeholdTypes[0].Type != "Elf" {
		t.Fatalf("behold_types = %+v, want only Elf", o.BeholdTypes)
	}
	elves := o.BeholdTypes[0].Cards
	if len(elves) != 2 || !slices.Contains(elves, elfHand.String()) || !slices.Contains(elves, elfTable.String()) {
		t.Errorf("Elf options = %v, want my Elf card and my Elf permanent", elves)
	}
	if o.RevealOptions == nil || o.RevealOptions.Min != 2 || o.RevealOptions.Max != 2 || len(o.RevealOptions.Cards) != 2 {
		t.Errorf("reveal_options = %+v, want the two Elves, exactly two", o.RevealOptions)
	}
}

// One creature of each type: present-and-empty options and no types.
func TestChosenTypeBeholdOfferWithNoPairIsEmpty(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	})
	o := reunionOffer(t, g, me)
	if o.RevealOptions == nil || len(o.RevealOptions.Cards) != 0 || len(o.BeholdTypes) != 0 {
		t.Errorf("offer = %+v, want present-and-empty options and no types", o)
	}
}
