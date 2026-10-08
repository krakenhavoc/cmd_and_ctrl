package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_lady_of_otaria_test.go — #2664: the end-step trigger behind the
// land-to-graveyard tally cell.

const theLadyOfOtariaOracle = "1ea4f882-6872-4d9e-9532-5ba2bf848d00"

func sacrificeALandFor(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	land := pushCatalogPermanent(g, owner, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(land); err != nil {
			t.Fatalf("setup sacrifice: %v", err)
		}
	})
}

func inHand(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestTheLadyOfOtariaTakesDwarvesAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "The Lady of Otaria", "Legendary Creature — Avatar", theLadyOfOtariaOracle, false)
	mk := func(name, typeLine string) uuid.UUID {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID}
		me.Library.PushTop(c)
		return c.InstanceID
	}
	bear := mk("Bear", "Creature — Bear")
	dwarfB := mk("Dwarf B", "Creature — Dwarf")
	bolt := mk("Bolt", "Instant")
	dwarfA := mk("Dwarf A", "Creature — Dwarf")
	sacrificeALandFor(t, g, me.ID)

	libBefore, handBefore := me.Library.Size(), me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, dwarfA, dwarfB)

	if !inHand(me, dwarfA) || !inHand(me, dwarfB) || inHand(me, bolt) || inHand(me, bear) {
		t.Fatalf("hand after the pick is wrong: %+v", me.Hand.Cards)
	}
	if me.Hand.Size() != handBefore+2 || me.Library.Size() != libBefore-2 {
		t.Errorf("hand %d -> %d, library %d -> %d, want +2 / -2", handBefore, me.Hand.Size(), libBefore, me.Library.Size())
	}
	// The two cards left behind are the bottom two (index 0 is the bottom), whatever their order.
	tail := me.Library.Cards[:2]
	got := map[uuid.UUID]bool{tail[0].InstanceID: true, tail[1].InstanceID: true}
	if !got[bolt] || !got[bear] {
		t.Errorf("the rest are not on the bottom: %v", tail)
	}
}

func TestTheLadyOfOtariaNeedsALandOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Lady of Otaria", "Legendary Creature — Avatar", theLadyOfOtariaOracle, false)
	sacrificeALandFor(t, g, opp.ID)
	advanceTo(t, g, game.StepEnd)
	if triggerOnStackFrom(g, "The Lady of Otaria") || chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's land in the graveyard triggered her")
	}

	g2 := newCatalogGame(t)
	pushCatalogPermanent(g2, g2.Seats[0].ID, "The Lady of Otaria", "Legendary Creature — Avatar", theLadyOfOtariaOracle, false)
	advanceTo(t, g2, game.StepEnd)
	if triggerOnStackFrom(g2, "The Lady of Otaria") {
		t.Fatal("triggered with no land lost")
	}
}
