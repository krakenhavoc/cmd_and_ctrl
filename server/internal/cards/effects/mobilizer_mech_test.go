package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mobilizerMechOracle = "3c1a2a2d-656f-4021-b046-59e14383a8c8"

func TestMobilizerMechIsFullWithNoCaveat(t *testing.T) {
	requireFullNoCaveats(t, mobilizerMechOracle, "Mobilizer Mech")
}

// The trigger offers only another Vehicle you control, and the one
// chosen is an artifact creature until end of turn without having been
// crewed.
func TestMobilizerMechAnimatesAnotherVehicleYouControl(t *testing.T) {
	s := newCrewScene(t)
	mech := s.pushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	other := s.pushVehicle("Other", smugglersCopterOracle, 3, 3)
	theirs := pushBattlefieldCardWithTimestamp(s.g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", OracleID: smugglersCopterOracle, TypeLine: crewCardsVehicleTypeLine,
		Power: 3, Toughness: 3, Owner: s.opp.ID, Controller: s.opp.ID,
	})
	crewer := s.pushCreature("Crewer", "Creature — Human", 3, 3)
	seat := s.g.Turn.ActiveSeat

	s.crewAndResolve(t, mech, 0, crewer)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the becomes-crewed trigger")
	}
	if !slices.Equal(pick.PickTargetCards, []uuid.UUID{other}) {
		t.Fatalf("offered %v, want only the other Vehicle you control (%s); not the Mech (%s), their Vehicle (%s) or the crewer",
			pick.PickTargetCards, other, mech, theirs)
	}
	if err := s.g.ResolvePickTarget(pick.ID, s.me.ID, game.TargetRef{Kind: game.TargetCard, ID: mech}); err == nil {
		t.Fatal("the Mech itself was accepted as the \"other\" Vehicle")
	}
	pickTriggerTarget(t, s.g, s.me.ID, other)
	if isCreatureNow(s.g, other) {
		t.Fatal("the other Vehicle is a creature before the trigger resolved")
	}
	passPriorityAroundTable(t, s.g)
	c := gaLayered(t, s.g, other)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Vehicle") {
		t.Fatalf("the other Vehicle has subtypes %q, want an artifact creature that is still a Vehicle", c.Effective().Subtypes)
	}
	if n := len(crewEventsThisTurn(s.g, game.EventBecameCrewed)); n != 1 {
		t.Errorf("%d became-crewed events, want 1: animating a Vehicle is not crewing it", n)
	}
	if isCreatureNow(s.g, theirs) {
		t.Error("an opponent's Vehicle was animated")
	}

	advancePastCleanupOf(t, s.g, seat)
	if isCreatureNow(s.g, other) || isCreatureNow(s.g, mech) {
		t.Error("a Vehicle is still a creature after the turn ended")
	}
}

// "Up to one": with no other Vehicle the trigger still goes on the stack
// and resolves doing nothing; and it triggers on every crew, not only
// the first of a turn.
func TestMobilizerMechTriggersEachTimeItBecomesCrewed(t *testing.T) {
	s := newCrewScene(t)
	mech := s.pushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	a := s.pushCreature("A", "Creature — Human", 3, 3)
	b := s.pushCreature("B", "Creature — Human", 3, 3)

	s.crewOnly(t, mech, a)
	if triggerOnStack(s.g, mech) != nil {
		t.Fatal("the trigger is on the stack before the crew ability resolved")
	}
	for i := 0; i < 8 && triggerOnStack(s.g, mech) == nil && !stackFullyEmpty(s.g); i++ {
		if err := s.g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if triggerOnStack(s.g, mech) == nil {
		t.Fatal("no trigger on the stack with no other Vehicle; \"up to one\" needs no target")
	}
	passPriorityAroundTable(t, s.g)

	other := s.pushVehicle("Other", smugglersCopterOracle, 3, 3)
	s.crewAndResolve(t, mech, 0, b)
	pickTriggerTarget(t, s.g, s.me.ID, other)
	passPriorityAroundTable(t, s.g)
	if !isCreatureNow(s.g, other) {
		t.Error("the second crew of the turn did not trigger the Mech")
	}
}

// Crew 3 is a real cost: a creature with power 2 cannot pay it.
func TestMobilizerMechCrewNeedsPowerThree(t *testing.T) {
	s := newCrewScene(t)
	mech := s.pushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	weak := s.pushCreature("Weak", "Creature — Human", 2, 2)
	toMain(t, s.g)
	if err := s.g.ActivateCatalogAbility(s.me.ID, mech, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{weak}}); err == nil {
		t.Fatal("crew 3 was paid with power 2")
	}
	if !hasEffectiveKeyword(t, s.g, mech, "flying") {
		t.Error("Mobilizer Mech has no flying")
	}
}

// A Vehicle that left before its crew ability resolved did not become
// crewed, and neither does the new object the card came back as (CR
// 400.7).
func TestMobilizerMechDoesNotTriggerForAVehicleThatLeftInResponse(t *testing.T) {
	s := newCrewScene(t)
	mech := s.pushVehicle("Mobilizer Mech", mobilizerMechOracle, 3, 4)
	other := s.pushVehicle("Other", smugglersCopterOracle, 3, 3)
	crewer := s.pushCreature("Crewer", "Creature — Human", 3, 3)

	s.crewOnly(t, mech, crewer)
	s.g.WithWriteLock(func() {
		if err := s.g.BounceToHandForEffect(mech); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if err := s.g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: s.me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, mech); err != nil {
		t.Fatalf("return the Mech: %v", err)
	}
	passPriorityAroundTable(t, s.g)
	if n := len(crewEventsThisTurn(s.g, game.EventBecameCrewed)); n != 0 {
		t.Errorf("%d became-crewed events for a Vehicle that left in response", n)
	}
	if latestPickTarget(s.g, s.me.ID) != nil || triggerOnStack(s.g, mech) != nil {
		t.Error("the returned Mech's trigger fired for the old object's crew ability")
	}
	if isCreatureNow(s.g, other) {
		t.Error("the other Vehicle was animated")
	}
}
