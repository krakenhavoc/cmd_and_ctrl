package testscr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestCrewCardsAreFullWithNoCaveat(t *testing.T) {
	for name, oracle := range map[string]string{
		"Turtle Van": effectstest.TurtleVanOracle, "Getaway Car": effectstest.GetawayCarOracle, "Subterranean Schooner": effectstest.SchoonerOracle,
		"Golden Argosy": effectstest.GoldenArgosyOracle, "Luxurious Locomotive": effectstest.LocomotiveOracle,
		"Veteran Motorist": effectstest.VeteranMotoristOracle, "Gearshift Ace": effectstest.GearshiftAceOracle,
		"Speedway Fanatic": effectstest.SpeedwayFanaticOracle, "Hotshot Mechanic": effectstest.HotshotMechanicOracle,
		"Giant Ox": effectstest.GiantOxOracle, "Shorikai, Genesis Engine": effectstest.ShorikaiOracle,
		"Prodigy's Prototype": effectstest.ProdigysPrototypeOracle, "Born to Drive": effectstest.BornToDriveOracle,
	} {
		effectstest.RequireFull(t, oracle, name)
	}
}

// --- Turtle Van --------------------------------------------------

// Only creatures that crewed THIS Van are offered; a Turtle ends on two
// counters (one, then doubled), anything else on one.
func TestTurtleVanCountersAndDoublesForATurtle(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           int
	}{
		{"Plain Crewer", "Creature — Human", 1},
		{"Turtle Crewer", "Creature — Mutant Ninja Turtle", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := effectstest.NewCrewScene(t)
			van := s.PushVehicle("Turtle Van", effectstest.TurtleVanOracle, 4, 4)
			other := s.PushVehicle("Other Van", effectstest.SmugglersCopterOracle, 3, 3)
			picked := s.PushCreature(tc.name, tc.typeLine, 1, 1)
			bystander := s.PushCreature("Other Crewer", "Creature — Mutant Ninja Turtle", 1, 1)
			s.CrewAndResolve(t, other, 0, bystander)
			s.CrewAndResolve(t, van, 0, picked)

			effectstest.DeclareAttack(t, s.G, s.Opp.ID, van)
			pick := effectstest.LatestPickTarget(s.G, s.Me.ID)
			if pick == nil {
				t.Fatal("no target prompt for the attack trigger")
			}
			if !slices.Equal(pick.PickTargetCards, []uuid.UUID{picked}) {
				t.Fatalf("offered %v, want only the creature that crewed this Van (%s)", pick.PickTargetCards, picked)
			}
			effectstest.PickTriggerTarget(t, s.G, s.Me.ID, picked)
			effectstest.PassPriorityAroundTable(t, s.G)
			if got := effectstest.CountersOf(t, s.G, picked); got != tc.want {
				t.Errorf("%s has %d +1/+1 counters, want %d", tc.name, got, tc.want)
			}
			if got := effectstest.CountersOf(t, s.G, bystander); got != 0 {
				t.Errorf("the creature that crewed the other Vehicle got %d counters", got)
			}
		})
	}
}

// --- Getaway Car --------------------------------------------------

func TestGetawayCarReturnsACrewerWhenItAttacks(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	car := s.PushVehicle("Getaway Car", effectstest.GetawayCarOracle, 4, 3)
	crewer := s.PushCreature("Crewer", "Creature — Human", 1, 1)
	idle := s.PushCreature("Idle", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, car, 0, crewer)

	effectstest.DeclareAttack(t, s.G, s.Opp.ID, car)
	pick := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if slices.Contains(pick.PickTargetCards, idle) || !slices.Contains(pick.PickTargetCards, crewer) {
		t.Fatalf("offered %v, want only the crewer", pick.PickTargetCards)
	}
	effectstest.PickTriggerTarget(t, s.G, s.Me.ID, crewer)
	effectstest.PassPriorityAroundTable(t, s.G)
	if effectstest.ZoneKindOf(s.G, crewer) != game.ZoneHand {
		t.Errorf("the crewer is in %v, want the hand", effectstest.ZoneKindOf(s.G, crewer))
	}
	if effectstest.ZoneKindOf(s.G, idle) != game.ZoneBattlefield {
		t.Error("the idle creature left the battlefield")
	}
}

func TestGetawayCarReturnsACrewerWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, blocker := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	raider := effectstest.PushVanillaCreature(g, attacker.ID, "Raider", 2, 2)
	car := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Getaway Car", OracleID: effectstest.GetawayCarOracle, TypeLine: effectstest.CrewCardsVehicleTypeLine,
		Power: 4, Toughness: 3, Owner: blocker.ID, Controller: blocker.ID,
	})
	crewer := effectstest.PushDiesCreatureForTest(g, blocker.ID, "Crewer", "", "Creature — Human", 1, 1)
	effectstest.CrewForTest(t, g, blocker.ID, car, crewer)

	effectstest.DeclareAttack(t, g, blocker.ID, raider)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(car, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PickTriggerTarget(t, g, blocker.ID, crewer)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneKindOf(g, crewer) != game.ZoneHand {
		t.Errorf("the crewer is in %v, want the owner's hand", effectstest.ZoneKindOf(g, crewer))
	}
}

// "Up to one": with every crewer gone the trigger still goes on the stack
// (a target is not required, CR 601.2c) and resolves doing nothing.
func TestGetawayCarTriggersWithNoCrewerLeft(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	car := s.PushVehicle("Getaway Car", effectstest.GetawayCarOracle, 4, 3)
	crewer := s.PushCreature("Crewer", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, car, 0, crewer)
	s.G.WithWriteLock(func() {
		if err := s.G.BounceToHandForEffect(crewer); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, car)
	if effectstest.TriggerOnStack(s.G, car) == nil {
		t.Fatal("the trigger was removed for want of a target; \"up to one\" needs none")
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if effectstest.ZoneKindOf(s.G, crewer) != game.ZoneHand {
		t.Error("the bounced crewer moved again")
	}
}

func TestGetawayCarRefusesAnIllegalTarget(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	car := s.PushVehicle("Getaway Car", effectstest.GetawayCarOracle, 4, 3)
	crewer := s.PushCreature("Crewer", "Creature — Human", 1, 1)
	idle := s.PushCreature("Idle", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, car, 0, crewer)
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, car)
	pick := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := s.G.ResolvePickTarget(pick.ID, s.Me.ID, game.TargetRef{Kind: game.TargetCard, ID: idle}); err == nil {
		t.Error("a creature that did not crew the Car was accepted")
	}
}

// --- Subterranean Schooner ---------------------------------------

func TestSubterraneanSchoonerMakesItsCrewerExplore(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	ship := s.PushVehicle("Subterranean Schooner", effectstest.SchoonerOracle, 3, 4)
	crewer := s.PushCreature("Crewer", "Creature — Human", 1, 1)
	idle := s.PushCreature("Idle", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, ship, 0, crewer)
	// A land on top: it goes to hand.
	effectstest.PushLibraryCard(s.Me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	handBefore := s.Me.Hand.Size()

	effectstest.DeclareAttack(t, s.G, s.Opp.ID, ship)
	pick := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if pick == nil || slices.Contains(pick.PickTargetCards, idle) {
		t.Fatalf("bad prompt %+v", pick)
	}
	effectstest.PickTriggerTarget(t, s.G, s.Me.ID, crewer)
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := s.Me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1 for the revealed land", handBefore, got)
	}
	if got := effectstest.CountersOf(t, s.G, crewer); got != 0 {
		t.Errorf("the crewer got %d counters for a land", got)
	}
}

func TestSubterraneanSchoonerNonlandPutsACounterOnTheCrewer(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	ship := s.PushVehicle("Subterranean Schooner", effectstest.SchoonerOracle, 3, 4)
	crewer := s.PushCreature("Crewer", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, ship, 0, crewer)
	effectstest.PushLibraryCard(s.Me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})

	effectstest.DeclareAttack(t, s.G, s.Opp.ID, ship)
	effectstest.PickTriggerTarget(t, s.G, s.Me.ID, crewer)
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := effectstest.CountersOf(t, s.G, crewer); got != 1 {
		t.Errorf("the crewer has %d counters after exploring a nonland, want 1", got)
	}
}

// --- Golden Argosy -------------------------------------------------

func TestGoldenArgosyExilesItsCrewersAndReturnsThemTapped(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	argosy := s.PushVehicle("Golden Argosy", effectstest.GoldenArgosyOracle, 3, 6)
	other := s.PushVehicle("Other Vehicle", effectstest.SmugglersCopterOracle, 3, 3)
	a := s.PushCreature("Crewer A", "Creature — Human", 1, 1)
	b := s.PushCreature("Crewer B", "Creature — Human", 1, 1)
	elsewhere := s.PushCreature("Crewed Elsewhere", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, other, 0, elsewhere)
	s.CrewAndResolve(t, argosy, 0, a, b)

	effectstest.DeclareAttack(t, s.G, s.Opp.ID, argosy)
	effectstest.PassPriorityAroundTable(t, s.G)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if got := effectstest.ZoneKindOf(s.G, id); got != game.ZoneExile {
			t.Errorf("crewer %s is in %v, want exile", name, got)
		}
	}
	if got := effectstest.ZoneKindOf(s.G, elsewhere); got != game.ZoneBattlefield {
		t.Errorf("a creature that crewed another Vehicle left the battlefield (%v)", got)
	}

	effectstest.AdvanceTo(t, s.G, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, s.G)
	for _, name := range []string{"Crewer A", "Crewer B"} {
		id := effectstest.FindBattlefieldByName(s.G, name)
		if id == uuid.Nil {
			t.Fatalf("%s did not return at the end step", name)
		}
		if c, _ := effectstest.BattlefieldCardByID(s.G, id); !c.Tapped {
			t.Errorf("%s returned untapped", name)
		}
	}
}

// --- Luxurious Locomotive ----------------------------------------

func TestLuxuriousLocomotiveMakesATreasurePerCrewer(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	loco := s.PushVehicle("Luxurious Locomotive", effectstest.LocomotiveOracle, 6, 5)
	a := s.PushCreature("Crewer A", "Creature — Human", 1, 1)
	b := s.PushCreature("Crewer B", "Creature — Human", 1, 1)
	idle := s.PushCreature("Idle", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, loco, 0, a, b)

	effectstest.DeclareAttack(t, s.G, s.Opp.ID, loco)
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := effectstest.TreasureCount(s.G); got != 2 {
		t.Errorf("%d Treasures, want one per crewer (2)", got)
	}

	// Crew is limited to once each turn.
	if err := s.G.ActivateCatalogAbility(s.Me.ID, loco, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{idle}}); err == nil {
		t.Error("the Locomotive was crewed a second time in one turn")
	}
}

// A crewer that left before the attack still counts: it is a distinct
// creature object in the Locomotive's record.
func TestLuxuriousLocomotiveCountsDepartedCrewers(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	loco := s.PushVehicle("Luxurious Locomotive", effectstest.LocomotiveOracle, 6, 5)
	a := s.PushCreature("Crewer A", "Creature — Human", 1, 1)
	b := s.PushCreature("Crewer B", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, loco, 0, a, b)
	s.G.WithWriteLock(func() {
		if err := s.G.BounceToHandForEffect(a); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})

	effectstest.DeclareAttack(t, s.G, s.Opp.ID, loco)
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := effectstest.TreasureCount(s.G); got != 2 {
		t.Errorf("%d Treasures, want 2 (a departed crewer still crewed it)", got)
	}
}

// --- Veteran Motorist, Gearshift Ace, Speedway Fanatic ------------

func pushCrewPilot(s effectstest.CrewScene, name, oracle, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.PushDiesCreatureForTest(s.G, s.Me.ID, name, oracle, typeLine, power, toughness)
}

// The trigger is on the stack above the crew ability, so it resolves first
// while the Vehicle is not yet a creature (CR 208.3a).
func TestVeteranMotoristPumpsTheVehicleItCrews(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	copter := effectstest.PushVehicleForTest(s.G, s.Me.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	motorist := pushCrewPilot(s, "Veteran Motorist", effectstest.VeteranMotoristOracle, "Creature — Dwarf Pilot", 3, 1)

	effectstest.CrewBy(t, s.G, s.Me.ID, copter, motorist)
	if n := len(s.G.StackMeta); n != 2 {
		t.Fatalf("stack holds %d items, want the crew ability under the Motorist's trigger", n)
	}
	if effectstest.TriggerOnStack(s.G, motorist) == nil {
		t.Fatal("the Motorist's trigger is not on the stack")
	}
	if c, _ := effectstest.BattlefieldCardByID(s.G, copter); c.IsCreature() {
		t.Fatal("the Copter was a creature before the trigger resolved")
	}
	effectstest.ResolveTopOfStack(t, s.G)
	if got := effectstest.EffectivePower(t, s.G, copter); got != 4 {
		t.Errorf("Copter power %d after the trigger, want 4", got)
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if c, _ := effectstest.BattlefieldCardByID(s.G, copter); !c.IsCreature() {
		t.Error("the Copter did not become a creature")
	}
	if got, want := effectstest.EffectiveToughness(t, s.G, copter), 4; got != want {
		t.Errorf("Copter toughness %d, want %d", got, want)
	}
}

// Another creature crewing does not trigger it; neither does saddling.
func TestVeteranMotoristOnlyTriggersForItsOwnCrew(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	copter := effectstest.PushVehicleForTest(s.G, s.Me.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	motorist := pushCrewPilot(s, "Veteran Motorist", effectstest.VeteranMotoristOracle, "Creature — Dwarf Pilot", 3, 1)
	plain := effectstest.PushCrewerForTest(s.G, s.Me.ID, "Plain", 1)
	mount := effectstest.PushMountForTest(s.G, s.Me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)

	effectstest.CrewBy(t, s.G, s.Me.ID, copter, plain)
	if effectstest.TriggerOnStack(s.G, motorist) != nil {
		t.Error("the Motorist triggered for another creature's crew")
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if err := effectstest.SaddleAbility(t, s.G, s.Me.ID, mount, motorist); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	if effectstest.TriggerOnStack(s.G, motorist) != nil {
		t.Error("the Motorist triggered for a saddle")
	}
}

func TestGearshiftAceGivesTheVehicleFirstStrike(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	copter := effectstest.PushVehicleForTest(s.G, s.Me.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	ace := pushCrewPilot(s, "Gearshift Ace", effectstest.GearshiftAceOracle, "Creature — Dwarf Pilot", 2, 1)

	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, s.G, copter), "first strike") {
		t.Fatal("the Copter starts with first strike")
	}
	effectstest.CrewBy(t, s.G, s.Me.ID, copter, ace)
	effectstest.PassPriorityAroundTable(t, s.G)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, s.G, copter), "first strike") {
		t.Error("the crewed Vehicle did not gain first strike")
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, s.G, ace), "first strike") {
		t.Error("the Ace lost its own first strike")
	}
}

func TestSpeedwayFanaticGivesTheVehicleHaste(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	copter := effectstest.PushVehicleForTest(s.G, s.Me.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	fan := pushCrewPilot(s, "Speedway Fanatic", effectstest.SpeedwayFanaticOracle, "Creature — Human Pilot", 2, 1)

	effectstest.CrewBy(t, s.G, s.Me.ID, copter, fan)
	effectstest.PassPriorityAroundTable(t, s.G)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, s.G, copter), "haste") {
		t.Error("the crewed Vehicle did not gain haste")
	}
}

// A Vehicle that left in response is a new object and gains nothing.
func TestSpeedwayFanaticGivesNothingToAVehicleThatLeft(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	copter := effectstest.PushVehicleForTest(s.G, s.Me.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	fan := pushCrewPilot(s, "Speedway Fanatic", effectstest.SpeedwayFanaticOracle, "Creature — Human Pilot", 2, 1)

	effectstest.CrewBy(t, s.G, s.Me.ID, copter, fan)
	effectstest.FlickerForTest(t, s.G, s.Me.ID, copter)
	effectstest.PassPriorityAroundTable(t, s.G)
	id := effectstest.FindBattlefieldByName(s.G, "Smuggler's Copter")
	if id == uuid.Nil {
		t.Skip("the Copter did not come back")
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, s.G, id), "haste") {
		t.Error("a Vehicle that left and came back gained haste")
	}
}

// --- Hotshot Mechanic, Giant Ox -----------------------------------

func TestHotshotMechanicCrewsAsThoughItsPowerWereTwoGreater(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	vehicle := effectstest.PushTestVehicle(s.G, s.Me.ID, 4, 4)
	mount := effectstest.PushMountForTest(s.G, s.Me.ID, "Saddle Test Mount", effectstest.CostPowerMountOracle, 3)
	mech := effectstest.PushPilotCard(s.G, s.Me.ID, "Hotshot Mechanic", "Artifact Creature — Fox Pilot", effectstest.HotshotMechanicOracle, 2, 1)

	if got := effectstest.CountedPowerOf(t, s.G, mech, game.CostPowerCrew); got != 4 {
		t.Errorf("counts %d for crew, want 4", got)
	}
	if got := effectstest.CountedPowerOf(t, s.G, mech, game.CostPowerSaddle); got != 2 {
		t.Errorf("counts %d for saddle, want its plain 2", got)
	}
	if got := effectstest.EffectivePower(t, s.G, mech); got != 2 {
		t.Errorf("real power %d, want 2", got)
	}
	if err := effectstest.CrewCostPowerVehicle(t, s.G, s.Me.ID, vehicle, mech); err != nil {
		t.Errorf("the Mechanic could not crew a crew-3 Vehicle alone: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if err := effectstest.SaddleAbility(t, s.G, s.Me.ID, mount, effectstest.PushCrewerForTest(s.G, s.Me.ID, "Two", 2)); err == nil {
		t.Error("a plain 2-power creature saddled a Saddle 3 Mount")
	}
}

func TestGiantOxCrewsUsingToughness(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	vehicle := effectstest.PushTestVehicle(s.G, s.Me.ID, 4, 4)
	ox := effectstest.PushPilotCard(s.G, s.Me.ID, "Giant Ox", "Creature — Ox", effectstest.GiantOxOracle, 0, 6)

	if got := effectstest.CountedPowerOf(t, s.G, ox, game.CostPowerCrew); got != 6 {
		t.Errorf("counts %d for crew, want its toughness 6", got)
	}
	if got := effectstest.CountedPowerOf(t, s.G, ox, game.CostPowerSaddle); got != 0 {
		t.Errorf("counts %d for saddle, want its plain power 0", got)
	}
	if got := effectstest.EffectivePower(t, s.G, ox); got != 0 {
		t.Errorf("real power %d, want 0", got)
	}
	if err := effectstest.CrewCostPowerVehicle(t, s.G, s.Me.ID, vehicle, ox); err != nil {
		t.Errorf("the Ox could not crew a crew-3 Vehicle alone: %v", err)
	}
}

// --- Shorikai, Genesis Engine ---------------------------------------

func TestShorikaiLootsAndMakesACrewOnlyPilot(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	shorikai := effectstest.PushVehicleForTest(s.G, s.Me.ID, "Shorikai, Genesis Engine", effectstest.ShorikaiOracle, 8, 8)
	if err := s.G.AddManaForEffect(s.Me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	handBefore := s.Me.Hand.Size()

	if err := s.G.ActivateCatalogAbility(s.Me.ID, shorikai, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := effectstest.BattlefieldCardByID(s.G, shorikai); !c.Tapped {
		t.Error("Shorikai did not tap for its own cost")
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := s.Me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d -> %d, want +2 before the discard", handBefore, got)
	}
	if got := effectstest.DiscardOwed(s.G, s.Me.ID); got != 1 {
		t.Errorf("discard owed %d, want 1", got)
	}

	var pilot game.Card
	s.G.ReadSnapshot(func() {
		for _, c := range s.G.Battlefield.Cards {
			if c.Name == "Pilot" {
				pilot = c
			}
		}
	})
	if pilot.InstanceID == uuid.Nil {
		t.Fatal("no Pilot token")
	}
	if pilot.TokenKey != game.TokenKey("pilot-crews") {
		t.Errorf("token %q, want the crew-only Pilot", pilot.TokenKey)
	}
	if got := effectstest.CountedPowerOf(t, s.G, pilot.InstanceID, game.CostPowerCrew); got != 3 {
		t.Errorf("the Pilot counts %d for crew, want 3", got)
	}
	if got := effectstest.CountedPowerOf(t, s.G, pilot.InstanceID, game.CostPowerSaddle); got != 1 {
		t.Errorf("the Pilot counts %d for saddle, want its plain 1", got)
	}
}

// Crew 8 is the second ability; an 8-power creature crews it.
func TestShorikaiCrewsForEight(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	shorikai := effectstest.PushVehicleForTest(s.G, s.Me.ID, "Shorikai, Genesis Engine", effectstest.ShorikaiOracle, 8, 8)
	big := s.PushCreature("Big", "Creature — Giant", 8, 8)
	small := s.PushCreature("Small", "Creature — Human", 7, 7)
	effectstest.ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, shorikai, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Error("7 power crewed a crew-8 Vehicle")
	}
	s.CrewAndResolve(t, shorikai, 1, big)
}

// --- Prodigy's Prototype ---------------------------------------------

func TestProdigysPrototypeMakesOnePilotPerAttackDeclaration(t *testing.T) {
	for _, tc := range []struct {
		name     string
		vehicles int
		creature bool
		want     int
	}{
		{"one Vehicle", 1, false, 1},
		{"two Vehicles", 2, false, 1},
		{"only a creature", 0, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := effectstest.NewCrewScene(t)
			s.PushVehicle("Prodigy's Prototype", effectstest.ProdigysPrototypeOracle, 3, 4)
			var attackers []uuid.UUID
			for i := 0; i < tc.vehicles; i++ {
				v := s.PushVehicle("Crewed Vehicle", effectstest.CostPowerVehicleOracle, 3, 3)
				c := s.PushCreature("Crewer", "Creature — Human", 3, 3)
				s.CrewAndResolve(t, v, 0, c)
				attackers = append(attackers, v)
			}
			if tc.creature {
				attackers = append(attackers, s.PushCreature("Attacker", "Creature — Human", 2, 2))
			}
			effectstest.ToMain(t, s.G)
			effectstest.DeclareAttack(t, s.G, s.Opp.ID, attackers...)
			effectstest.PassPriorityAroundTable(t, s.G)

			var pilots []game.Card
			s.G.ReadSnapshot(func() {
				for _, c := range s.G.Battlefield.Cards {
					if c.Name == "Pilot" {
						pilots = append(pilots, c)
					}
				}
			})
			if len(pilots) != tc.want {
				t.Fatalf("%d Pilots, want %d", len(pilots), tc.want)
			}
			for _, p := range pilots {
				if got := effectstest.CountedPowerOf(t, s.G, p.InstanceID, game.CostPowerCrew); got != 3 {
					t.Errorf("the Pilot counts %d for crew, want 3", got)
				}
				if p.TokenKey != game.TokenKey("pilot-crews") {
					t.Errorf("token %q, want the crew-only Pilot", p.TokenKey)
				}
			}
		})
	}
}

// --- Born to Drive ------------------------------------------------

func TestBornToDriveGrowsWithCreaturesAndVehicles(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	host := s.PushCreature("Host", "Creature — Human", 1, 1)
	s.PushCreature("Friend", "Creature — Human", 1, 1)
	vehicle := s.PushVehicle("Idle Vehicle", effectstest.SmugglersCopterOracle, 3, 3)

	aura := effectstest.CastCatalogSpell(t, s.G, "Born to Drive", effectstest.AuraTypeLine, effectstest.BornToDriveOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, s.G)
	if !s.G.Battlefield.Contains(aura) {
		t.Fatal("Born to Drive did not resolve to the battlefield")
	}
	// Host, Friend and the Vehicle: three permanents that are creatures
	// and/or Vehicles.
	if got := effectstest.EffectivePower(t, s.G, host); got != 4 {
		t.Errorf("enchanted creature power %d, want 1 + 3", got)
	}
	// Crewing the Vehicle does not count it twice.
	crewer := s.PushCreature("Crewer", "Creature — Human", 3, 3)
	if got := effectstest.EffectivePower(t, s.G, host); got != 5 {
		t.Errorf("after another creature arrived: power %d, want 1 + 4", got)
	}
	s.CrewAndResolve(t, vehicle, 0, crewer)
	if got := effectstest.EffectivePower(t, s.G, host); got != 5 {
		t.Errorf("after the Vehicle was crewed: power %d, want it counted once (5)", got)
	}
	// An opponent's creature is not counted.
	effectstest.PushDiesCreatureForTest(s.G, s.Opp.ID, "Enemy", "", "Creature — Human", 2, 2)
	if got := effectstest.EffectivePower(t, s.G, host); got != 5 {
		t.Errorf("an opponent's creature changed the bonus: %d", got)
	}
}

// On an uncrewed Vehicle the bonus waits for it to become a creature.
func TestBornToDriveOnAnUncrewedVehicleWaitsForItToBeACreature(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	vehicle := s.PushVehicle("Idle Vehicle", effectstest.SmugglersCopterOracle, 3, 3)
	crewer := s.PushCreature("Crewer", "Creature — Human", 3, 3)
	effectstest.CastCatalogSpell(t, s.G, "Born to Drive", effectstest.AuraTypeLine, effectstest.BornToDriveOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: vehicle}})
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := effectstest.EffectivePower(t, s.G, vehicle); got != 3 {
		t.Errorf("uncrewed Vehicle power %d, want its printed 3", got)
	}
	s.CrewAndResolve(t, vehicle, 0, crewer)
	// Crewer and the Vehicle: 3 + 2.
	if got := effectstest.EffectivePower(t, s.G, vehicle); got != 5 {
		t.Errorf("crewed Vehicle power %d, want 3 + 2", got)
	}
}

func TestBornToDriveChannelMakesTwoCrewOnlyPilots(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	effectstest.ToMain(t, s.G)
	id := effectstest.PushCatalogHandCard(s.Me, "Born to Drive", effectstest.AuraTypeLine, effectstest.BornToDriveOracle)
	if err := s.G.AddManaForEffect(s.Me.ID, uuid.Nil, "{C}{C}{W}"); err != nil {
		t.Fatal(err)
	}
	offered := false
	for _, m := range legal.EnumerateFor(s.G, s.Me.ID) {
		if m.Source == id && m.Type == legal.TypeActivateAbility {
			offered = true
		}
	}
	if !offered {
		t.Error("channel is not offered from the hand")
	}
	if err := s.G.ActivateCatalogAbility(s.Me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := effectstest.ZoneKindOf(s.G, id); got != game.ZoneGraveyard {
		t.Errorf("the channeled card is in %v, want the graveyard", got)
	}
	n := 0
	s.G.ReadSnapshot(func() {
		for _, c := range s.G.Battlefield.Cards {
			if c.Name == "Pilot" && c.TokenKey == game.TokenKey("pilot-crews") {
				n++
			}
		}
	})
	if n != 2 {
		t.Errorf("%d crew-only Pilots, want 2", n)
	}
}
