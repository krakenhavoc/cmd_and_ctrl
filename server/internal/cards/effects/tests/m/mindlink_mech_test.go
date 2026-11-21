package testsm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mindlinkMechOracle = "009a4073-5946-4131-928c-c5c373745344"

func TestMindlinkMechIsFullWithNoCaveat(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, mindlinkMechOracle, "Mindlink Mech")
}

// The Mech becomes a copy of the crewer: its name and abilities, but
// 4/3, a Vehicle artifact as well, and flying. It ends with the turn.
func TestMindlinkMechBecomesACopyOfACrewer(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	mech := s.PushVehicle("Mindlink Mech", mindlinkMechOracle, 4, 3)
	elves := effectstest.PushBattlefieldCardWithTimestamp(s.G, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", OracleID: effectstest.LlanowarElvesOracle, TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Owner: s.Me.ID, Controller: s.Me.ID,
	})
	idle := s.PushCreature("Idle", "Creature — Human", 1, 1)
	seat := s.G.Turn.ActiveSeat

	s.CrewAndResolve(t, mech, 0, elves)
	pick := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the becomes-crewed trigger")
	}
	if !slices.Equal(pick.PickTargetCards, []uuid.UUID{elves}) {
		t.Fatalf("offered %v, want only the creature that crewed it (%s), not %s", pick.PickTargetCards, elves, idle)
	}
	if err := s.G.ResolvePickTarget(pick.ID, s.Me.ID, game.TargetRef{Kind: game.TargetCard, ID: idle}); err == nil {
		t.Fatal("a creature that did not crew the Mech was accepted")
	}
	effectstest.PickTriggerTarget(t, s.G, s.Me.ID, elves)
	effectstest.PassPriorityAroundTable(t, s.G)

	c := effectstest.GaLayered(t, s.G, mech)
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
	if !effectstest.HasEffectiveKeyword(t, s.G, mech, "flying") {
		t.Error("the copy has no flying")
	}
	if c.OracleID != effectstest.LlanowarElvesOracle {
		t.Errorf("the copy's oracle id is %q, want the Elves' (its abilities come with it)", c.OracleID)
	}

	effectstest.AdvancePastCleanupOf(t, s.G, seat)
	c = effectstest.GaLayered(t, s.G, mech)
	if c.Effective().Name != "Mindlink Mech" || c.IsCreature() || c.OracleID != mindlinkMechOracle {
		t.Errorf("after the turn the Mech is %q (%q), want Mindlink Mech, not a creature", c.Effective().Name, c.Effective().Subtypes)
	}
}

// "For the first time each turn" is about the Mech becoming crewed: a
// first crew by a legendary creature leaves the trigger with no legal
// target, and a second crew that turn does not trigger. The count is
// carried across a restore. Next turn it triggers again.
func TestMindlinkMechOnlyTheFirstCrewOfATurnTriggers(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	mech := s.PushVehicle("Mindlink Mech", mindlinkMechOracle, 4, 3)
	legend := s.PushCreature("Legend", "Legendary Creature — Human", 1, 1)
	plain := s.PushCreature("Plain", "Creature — Human", 1, 1)
	seat := s.G.Turn.ActiveSeat

	s.CrewAndResolve(t, mech, 0, legend)
	if effectstest.LatestPickTarget(s.G, s.Me.ID) != nil || effectstest.TriggerOnStack(s.G, mech) != nil {
		t.Fatal("the trigger found a target: a legendary creature is not one")
	}

	g := effectstest.RestoreRoundTrip(t, s.G, true)
	if err := g.ActivateCatalogAbility(s.Me.ID, mech, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{plain}}); err != nil {
		t.Fatalf("second crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, s.Me.ID) != nil || effectstest.TriggerOnStack(g, mech) != nil {
		t.Fatal("the second crew of the turn triggered Mindlink Mech")
	}
	if c := effectstest.GaLayered(t, g, mech); c.Effective().Name != "Mindlink Mech" {
		t.Fatalf("the Mech became %q", c.Effective().Name)
	}

	// The next turn of the same player: the first crew triggers again.
	effectstest.AdvancePastCleanupOf(t, g, seat)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(s.Me.ID, mech, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{plain}}); err != nil {
		t.Fatalf("next turn's crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, s.Me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, mech); c.Effective().Name != "Plain" || c.Effective().Power != 4 || c.Effective().Toughness != 3 {
		t.Errorf("next turn the Mech is %q %d/%d, want a 4/3 copy of Plain", c.Effective().Name, c.Effective().Power, c.Effective().Toughness)
	}
}

// A copy in progress is a restore point: the restored Mech is still the
// copy, and still stops being one at end of turn.
func TestMindlinkMechCopySurvivesARestore(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	mech := s.PushVehicle("Mindlink Mech", mindlinkMechOracle, 4, 3)
	plain := s.PushCreature("Plain", "Creature — Human", 1, 1)
	seat := s.G.Turn.ActiveSeat
	s.CrewAndResolve(t, mech, 0, plain)
	effectstest.PickTriggerTarget(t, s.G, s.Me.ID, plain)
	effectstest.PassPriorityAroundTable(t, s.G)

	g := effectstest.RestoreRoundTrip(t, s.G, true)
	c := effectstest.GaLayered(t, g, mech)
	if c.Effective().Name != "Plain" || c.Effective().Power != 4 || !effectstest.HasEffectiveKeyword(t, g, mech, "flying") || !c.HasSubtype("Vehicle") {
		t.Fatalf("restored Mech is %q %q, want the 4/3 flying Vehicle copy of Plain", c.Effective().Name, c.Effective().Subtypes)
	}
	effectstest.AdvancePastCleanupOf(t, g, seat)
	if c := effectstest.GaLayered(t, g, mech); c.Effective().Name != "Mindlink Mech" {
		t.Errorf("after the turn the restored Mech is still %q", c.Effective().Name)
	}
}
