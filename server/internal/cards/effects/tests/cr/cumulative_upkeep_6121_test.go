package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cumulative_upkeep_6121_test.go — #6121: Jötun Grunt's "Cumulative
// upkeep—Put two cards from a single graveyard on the bottom of their
// owner's library". CR 702.24a: "If [cost] has choices associated with
// it, each choice is made separately for each age counter", so with two
// age counters the four cards may come two from one graveyard and two
// from another, but never three and one.

func cu6121Card(z *game.Zone, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	z.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: owner, Controller: owner})
	return id
}

func TestGraveyardPairsCumulativeUpkeep(t *testing.T) {
	const oid = "test-6121-grunt-upkeep"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	grunt := cu6189Spec(t, g, oid, "Test Grunt",
		effects.CumulativeUpkeepPaying("Test Grunt — cumulative upkeep: put two cards from a single graveyard on the bottom of their owner's library",
			effects.GraveyardToLibraryBottomPayment(2, 2), "card", "cards"))
	var mine, theirs []uuid.UUID
	g.WithWriteLock(func() {
		for _, n := range []string{"M1", "M2", "M3"} {
			mine = append(mine, cu6121Card(me.Graveyard, me.ID, n))
		}
		for _, n := range []string{"O1", "O2"} {
			theirs = append(theirs, cu6121Card(opp.Graveyard, opp.ID, n))
		}
	})

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	a := prompt.PayAction()
	if a == nil || a.Kind != game.PayActionGraveyardToLibraryBottom || a.Count != 2 || a.GroupSize != 2 {
		t.Fatalf("first upkeep's payment = %+v, want two cards from a single graveyard", a)
	}
	// One from each graveyard is not "two from a single graveyard".
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{mine[0], theirs[0]}); err == nil {
		t.Fatal("a split pair paid the upkeep")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, theirs); err != nil {
		t.Fatalf("pay with the opponent's two: %v", err)
	}
	if opp.Graveyard.Size() != 0 {
		t.Errorf("opponent's graveyard holds %d, want both put away", opp.Graveyard.Size())
	}
	for i, id := range []uuid.UUID{theirs[1], theirs[0]} {
		if opp.Library.Cards[i].InstanceID != id {
			t.Errorf("opponent's library bottom %d is %s, want the named card", i, opp.Library.Cards[i].Name)
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, grunt); !ok {
		t.Fatal("a paid upkeep sacrificed the permanent")
	}

	// Second upkeep: four cards, two per graveyard — but only three
	// graveyard cards are left in whole pairs' reach (one pair), so the
	// payment can't be made (CR 118.3) and the Grunt is sacrificed.
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if a := prompt.PayAction(); a == nil || a.Count != 4 || a.GroupSize != 2 {
		t.Fatalf("second upkeep's payment = %+v, want four cards in pairs (CR 702.24a)", a)
	}
	if g.PayActionPayableForEffect(me.ID, prompt.PayAction()) {
		t.Error("three graveyard cards pay four in pairs")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, mine); err == nil {
		t.Error("three cards paid four")
	}
}
