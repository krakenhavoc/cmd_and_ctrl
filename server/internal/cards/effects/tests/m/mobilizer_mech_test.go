package testsm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mobilizerMechOracle = "3c1a2a2d-656f-4021-b046-59e14383a8c8"

func TestMobilizerMechIsFullWithNoCaveat(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, mobilizerMechOracle, "Mobilizer Mech")
}

// The trigger offers only another Vehicle you control, and the one
// chosen is an artifact creature until end of turn without having been
// crewed.
func TestMobilizerMechAnimatesAnotherVehicleYouControl(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	mech := s.PushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	other := s.PushVehicle("Other", effectstest.SmugglersCopterOracle, 3, 3)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(s.G, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", OracleID: effectstest.SmugglersCopterOracle, TypeLine: effectstest.CrewCardsVehicleTypeLine,
		Power: 3, Toughness: 3, Owner: s.Opp.ID, Controller: s.Opp.ID,
	})
	crewer := s.PushCreature("Crewer", "Creature — Human", 3, 3)
	seat := s.G.Turn.ActiveSeat

	s.CrewAndResolve(t, mech, 0, crewer)
	pick := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the becomes-crewed trigger")
	}
	if !slices.Equal(pick.PickTargetCards, []uuid.UUID{other}) {
		t.Fatalf("offered %v, want only the other Vehicle you control (%s); not the Mech (%s), their Vehicle (%s) or the crewer",
			pick.PickTargetCards, other, mech, theirs)
	}
	if err := s.G.ResolvePickTarget(pick.ID, s.Me.ID, game.TargetRef{Kind: game.TargetCard, ID: mech}); err == nil {
		t.Fatal("the Mech itself was accepted as the \"other\" Vehicle")
	}
	effectstest.PickTriggerTarget(t, s.G, s.Me.ID, other)
	if effectstest.IsCreatureNow(s.G, other) {
		t.Fatal("the other Vehicle is a creature before the trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	c := effectstest.GaLayered(t, s.G, other)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Vehicle") {
		t.Fatalf("the other Vehicle has subtypes %q, want an artifact creature that is still a Vehicle", c.Effective().Subtypes)
	}
	if n := len(effectstest.CrewEventsThisTurn(s.G, game.EventBecameCrewed)); n != 1 {
		t.Errorf("%d became-crewed events, want 1: animating a Vehicle is not crewing it", n)
	}
	if effectstest.IsCreatureNow(s.G, theirs) {
		t.Error("an opponent's Vehicle was animated")
	}

	effectstest.AdvancePastCleanupOf(t, s.G, seat)
	if effectstest.IsCreatureNow(s.G, other) || effectstest.IsCreatureNow(s.G, mech) {
		t.Error("a Vehicle is still a creature after the turn ended")
	}
}

// "Up to one": with no other Vehicle the trigger still goes on the stack
// and resolves doing nothing; and it triggers on every crew, not only
// the first of a turn.
func TestMobilizerMechTriggersEachTimeItBecomesCrewed(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	mech := s.PushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	a := s.PushCreature("A", "Creature — Human", 3, 3)
	b := s.PushCreature("B", "Creature — Human", 3, 3)

	s.CrewOnly(t, mech, a)
	if effectstest.TriggerOnStack(s.G, mech) != nil {
		t.Fatal("the trigger is on the stack before the crew ability resolved")
	}
	for i := 0; i < 8 && effectstest.TriggerOnStack(s.G, mech) == nil && !effectstest.StackFullyEmpty(s.G); i++ {
		if err := s.G.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.TriggerOnStack(s.G, mech) == nil {
		t.Fatal("no trigger on the stack with no other Vehicle; \"up to one\" needs no target")
	}
	effectstest.PassPriorityAroundTable(t, s.G)

	other := s.PushVehicle("Other", effectstest.SmugglersCopterOracle, 3, 3)
	s.CrewAndResolve(t, mech, 0, b)
	effectstest.PickTriggerTarget(t, s.G, s.Me.ID, other)
	effectstest.PassPriorityAroundTable(t, s.G)
	if !effectstest.IsCreatureNow(s.G, other) {
		t.Error("the second crew of the turn did not trigger the Mech")
	}
}

// Crew 3 is a real cost: a creature with power 2 cannot pay it.
func TestMobilizerMechCrewNeedsPowerThree(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	mech := s.PushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	weak := s.PushCreature("Weak", "Creature — Human", 2, 2)
	effectstest.ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, mech, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{weak}}); err == nil {
		t.Fatal("crew 3 was paid with power 2")
	}
	if !effectstest.HasEffectiveKeyword(t, s.G, mech, "flying") {
		t.Error("Mobilizer Mech has no flying")
	}
}

// A Vehicle that left before its crew ability resolved did not become
// crewed, and neither does the new object the card came back as (CR
// 400.7).
func TestMobilizerMechDoesNotTriggerForAVehicleThatLeftInResponse(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	mech := s.PushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	other := s.PushVehicle("Other", effectstest.SmugglersCopterOracle, 3, 3)
	crewer := s.PushCreature("Crewer", "Creature — Human", 3, 3)

	s.CrewOnly(t, mech, crewer)
	s.G.WithWriteLock(func() {
		if err := s.G.BounceToHandForEffect(mech); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if err := s.G.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: s.Me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, mech); err != nil {
		t.Fatalf("return the Mech: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if n := len(effectstest.CrewEventsThisTurn(s.G, game.EventBecameCrewed)); n != 0 {
		t.Errorf("%d became-crewed events for a Vehicle that left in response", n)
	}
	if effectstest.LatestPickTarget(s.G, s.Me.ID) != nil || effectstest.TriggerOnStack(s.G, mech) != nil {
		t.Error("the returned Mech's trigger fired for the old object's crew ability")
	}
	if effectstest.IsCreatureNow(s.G, other) {
		t.Error("the other Vehicle was animated")
	}
}
