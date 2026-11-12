package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mindlinkMechOracle = "009a4073-5946-4131-928c-c5c373745344"

func TestMindlinkMechIsFullWithNoCaveat(t *testing.T) {
	requireFullNoCaveats(t, mindlinkMechOracle, "Mindlink Mech")
}

// The Mech becomes a copy of the crewer: its name and abilities, but
// 4/3, a Vehicle artifact as well, and flying. It ends with the turn.
func TestMindlinkMechBecomesACopyOfACrewer(t *testing.T) {
	s := newCrewScene(t)
	mech := s.pushVehicle("Mindlink Mech", mindlinkMechOracle, 4, 3)
	elves := pushBattlefieldCardWithTimestamp(s.g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", OracleID: llanowarElvesOracle, TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Owner: s.me.ID, Controller: s.me.ID,
	})
	idle := s.pushCreature("Idle", "Creature — Human", 1, 1)
	seat := s.g.Turn.ActiveSeat

	s.crewAndResolve(t, mech, 0, elves)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the becomes-crewed trigger")
	}
	if !slices.Equal(pick.PickTargetCards, []uuid.UUID{elves}) {
		t.Fatalf("offered %v, want only the creature that crewed it (%s), not %s", pick.PickTargetCards, elves, idle)
	}
	if err := s.g.ResolvePickTarget(pick.ID, s.me.ID, game.TargetRef{Kind: game.TargetCard, ID: idle}); err == nil {
		t.Fatal("a creature that did not crew the Mech was accepted")
	}
	pickTriggerTarget(t, s.g, s.me.ID, elves)
	passPriorityAroundTable(t, s.g)

	c := gaLayered(t, s.g, mech)
	eff := c.Effective()
	if eff.Name != "Llanowar Elves" {
		t.Fatalf("the Mech is named %q, want Llanowar Elves", eff.Name)
	}
	if eff.Power != 4 || eff.Toughness != 3 {
		t.Errorf("the copy is %d/%d, want 4/3", eff.Power, eff.Toughness)
	}
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Vehicle") || !c.HasSubtype("Elf") || !c.HasSubtype("Druid") {
		t.Errorf("the copy has subtypes %q, want an artifact creature Elf Druid Vehicle", eff.Subtypes)
	}
	if !hasEffectiveKeyword(t, s.g, mech, "flying") {
		t.Error("the copy has no flying")
	}
	if c.OracleID != llanowarElvesOracle {
		t.Errorf("the copy's oracle id is %q, want the Elves' (its abilities come with it)", c.OracleID)
	}

	advancePastCleanupOf(t, s.g, seat)
	c = gaLayered(t, s.g, mech)
	if c.Effective().Name != "Mindlink Mech" || c.IsCreature() || c.OracleID != mindlinkMechOracle {
		t.Errorf("after the turn the Mech is %q (%q), want Mindlink Mech, not a creature", c.Effective().Name, c.Effective().Subtypes)
	}
}

// "For the first time each turn" is about the Mech becoming crewed: a
// first crew by a legendary creature leaves the trigger with no legal
// target, and a second crew that turn does not trigger. The count is
// carried across a restore. Next turn it triggers again.
func TestMindlinkMechOnlyTheFirstCrewOfATurnTriggers(t *testing.T) {
	s := newCrewScene(t)
	mech := s.pushVehicle("Mindlink Mech", mindlinkMechOracle, 4, 3)
	legend := s.pushCreature("Legend", "Legendary Creature — Human", 1, 1)
	plain := s.pushCreature("Plain", "Creature — Human", 1, 1)
	seat := s.g.Turn.ActiveSeat

	s.crewAndResolve(t, mech, 0, legend)
	if latestPickTarget(s.g, s.me.ID) != nil || triggerOnStack(s.g, mech) != nil {
		t.Fatal("the trigger found a target: a legendary creature is not one")
	}

	g := restoreRoundTrip(t, s.g, true)
	if err := g.ActivateCatalogAbility(s.me.ID, mech, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{plain}}); err != nil {
		t.Fatalf("second crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, s.me.ID) != nil || triggerOnStack(g, mech) != nil {
		t.Fatal("the second crew of the turn triggered Mindlink Mech")
	}
	if c := gaLayered(t, g, mech); c.Effective().Name != "Mindlink Mech" {
		t.Fatalf("the Mech became %q", c.Effective().Name)
	}

	// The next turn of the same player: the first crew triggers again.
	advancePastCleanupOf(t, g, seat)
	advanceToStepOf(t, g, seat, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(s.me.ID, mech, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{plain}}); err != nil {
		t.Fatalf("next turn's crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, s.me.ID, plain)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, mech); c.Effective().Name != "Plain" || c.Effective().Power != 4 || c.Effective().Toughness != 3 {
		t.Errorf("next turn the Mech is %q %d/%d, want a 4/3 copy of Plain", c.Effective().Name, c.Effective().Power, c.Effective().Toughness)
	}
}

// A copy in progress is a restore point: the restored Mech is still the
// copy, and still stops being one at end of turn.
func TestMindlinkMechCopySurvivesARestore(t *testing.T) {
	s := newCrewScene(t)
	mech := s.pushVehicle("Mindlink Mech", mindlinkMechOracle, 4, 3)
	plain := s.pushCreature("Plain", "Creature — Human", 1, 1)
	seat := s.g.Turn.ActiveSeat
	s.crewAndResolve(t, mech, 0, plain)
	pickTriggerTarget(t, s.g, s.me.ID, plain)
	passPriorityAroundTable(t, s.g)

	g := restoreRoundTrip(t, s.g, true)
	c := gaLayered(t, g, mech)
	if c.Effective().Name != "Plain" || c.Effective().Power != 4 || !hasEffectiveKeyword(t, g, mech, "flying") || !c.HasSubtype("Vehicle") {
		t.Fatalf("restored Mech is %q %q, want the 4/3 flying Vehicle copy of Plain", c.Effective().Name, c.Effective().Subtypes)
	}
	advancePastCleanupOf(t, g, seat)
	if c := gaLayered(t, g, mech); c.Effective().Name != "Mindlink Mech" {
		t.Errorf("after the turn the restored Mech is still %q", c.Effective().Name)
	}
}
