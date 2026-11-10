package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const fireglassMentorOracle = "b173245e-2009-48ff-ac42-5f28d12aec79"

func TestFireglassMentorIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(fireglassMentorOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Fireglass Mentor: %+v %v", spec, ok)
	}
}

func fireglassLands(me *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := n - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: ids[i], Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	}
	return ids
}

func TestFireglassMentorExilesTwoAndLetsYouPlayTheChosenOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dskB1Creature(g, me.ID, "Fireglass Mentor", "Creature — Lizard Warlock", fireglassMentorOracle, 2, 1)
	ids := fireglassLands(me, 3)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("pick %+v, want a choice between the two exiled cards", pick)
	}
	if z := zoneOfCard(g, ids[2]); z != game.ZoneLibrary {
		t.Errorf("a third card left the library (%v)", z)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("played the card that was not chosen")
	}
	if err := g.CastSpell(me.ID, ids[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("could not play the chosen card: %v", err)
	}
}

func TestFireglassMentorNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dskB1Creature(g, me.ID, "Fireglass Mentor", "Creature — Lizard Warlock", fireglassMentorOracle, 2, 1)
	ids := fireglassLands(me, 3)
	// Losing life yourself does not count.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil || zoneOfCard(g, ids[0]) != game.ZoneLibrary {
		t.Error("the Mentor triggered with no opponent life loss")
	}
}
