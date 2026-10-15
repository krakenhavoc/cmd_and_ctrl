package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// crew_cards_test.go — ADR 0186 slice 11 (#3553): the Vehicles and Pilots
// that read the crew record or the counted-power rule, one proof per card
// for its crew-specific clause.

const (
	turtleVanOracle          = "aeb9a2e6-ad11-4e2c-81d0-caf279d31ece"
	getawayCarOracle         = "954b30b2-b88c-4558-b06c-4518b49d7b20"
	schoonerOracle           = "7f82bae3-feaa-4967-a843-32a4f72806d5"
	goldenArgosyOracle       = "b636a16d-f5ca-443d-8e6f-a2c240700c36"
	locomotiveOracle         = "c68c2a98-9a6d-4199-af3a-78eb92beda48"
	veteranMotoristOracle    = "3e6f51ae-7239-4fcc-919f-aa2437e8ef15"
	gearshiftAceOracle       = "dde61c4c-1815-45b3-9e68-3097c28c363c"
	speedwayFanaticOracle    = "5e7e1b23-ac4b-4f95-9388-feb099d6779f"
	hotshotMechanicOracle    = "8576201b-186d-4e28-97d4-deb344f14436"
	giantOxOracle            = "b42bdf54-9055-405c-87e6-0912dbe2bd68"
	shorikaiOracle           = "e2962883-6daf-4a84-af73-917afc73092c"
	prodigysPrototypeOracle  = "405a54b3-fb9f-4503-a4fc-3ba6a5e58205"
	bornToDriveOracle        = "7ddd84d2-1a6d-43f5-a42d-3c66eee5f7e5"
	crewCardsVehicleTypeLine = "Artifact — Vehicle"
)

// crewScene is the common table: the active player's crewed Vehicle
// attacking, with its attack trigger waiting.
type crewScene struct {
	g       *game.Game
	me, opp *game.Player
}

func newCrewScene(t *testing.T) crewScene {
	t.Helper()
	g := newCatalogGame(t)
	return crewScene{g: g, me: g.Seats[g.Turn.ActiveSeat], opp: g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]}
}

// pushVehicle is a catalog Vehicle with its printed P/T, past summoning
// sickness, carried through the timestamping path.
func (s crewScene) pushVehicle(name, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(s.g, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: crewCardsVehicleTypeLine,
		Power: power, Toughness: toughness, Owner: s.me.ID, Controller: s.me.ID,
	})
}

func (s crewScene) pushCreature(name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(s.g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: s.me.ID, Controller: s.me.ID,
	})
}

// crewAndResolve crews `vehicle` (ability index `idx`) with the creatures
// and lets the ability resolve.
func (s crewScene) crewAndResolve(t *testing.T, vehicle uuid.UUID, idx int, crewers ...uuid.UUID) {
	t.Helper()
	toMain(t, s.g)
	if err := s.g.ActivateCatalogAbility(s.me.ID, vehicle, idx, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, s.g)
	if c, ok := battlefieldCardByID(s.g, vehicle); !ok || !c.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}
}

func requireFull(t *testing.T, oracle, name string) {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("%s: want complete with no caveat, got %+v", name, spec)
	}
}

func TestCrewCardsAreFullWithNoCaveat(t *testing.T) {
	for name, oracle := range map[string]string{
		"Turtle Van": turtleVanOracle, "Getaway Car": getawayCarOracle, "Subterranean Schooner": schoonerOracle,
		"Golden Argosy": goldenArgosyOracle, "Luxurious Locomotive": locomotiveOracle,
		"Veteran Motorist": veteranMotoristOracle, "Gearshift Ace": gearshiftAceOracle,
		"Speedway Fanatic": speedwayFanaticOracle, "Hotshot Mechanic": hotshotMechanicOracle,
		"Giant Ox": giantOxOracle, "Shorikai, Genesis Engine": shorikaiOracle,
		"Prodigy's Prototype": prodigysPrototypeOracle, "Born to Drive": bornToDriveOracle,
	} {
		requireFull(t, oracle, name)
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
			s := newCrewScene(t)
			van := s.pushVehicle("Turtle Van", turtleVanOracle, 4, 4)
			other := s.pushVehicle("Other Van", smugglersCopterOracle, 3, 3)
			picked := s.pushCreature(tc.name, tc.typeLine, 1, 1)
			bystander := s.pushCreature("Other Crewer", "Creature — Mutant Ninja Turtle", 1, 1)
			s.crewAndResolve(t, other, 0, bystander)
			s.crewAndResolve(t, van, 0, picked)

			declareAttack(t, s.g, s.opp.ID, van)
			pick := latestPickTarget(s.g, s.me.ID)
			if pick == nil {
				t.Fatal("no target prompt for the attack trigger")
			}
			if !slices.Equal(pick.PickTargetCards, []uuid.UUID{picked}) {
				t.Fatalf("offered %v, want only the creature that crewed this Van (%s)", pick.PickTargetCards, picked)
			}
			pickTriggerTarget(t, s.g, s.me.ID, picked)
			passPriorityAroundTable(t, s.g)
			if got := countersOf(t, s.g, picked); got != tc.want {
				t.Errorf("%s has %d +1/+1 counters, want %d", tc.name, got, tc.want)
			}
			if got := countersOf(t, s.g, bystander); got != 0 {
				t.Errorf("the creature that crewed the other Vehicle got %d counters", got)
			}
		})
	}
}

// --- Getaway Car --------------------------------------------------

func TestGetawayCarReturnsACrewerWhenItAttacks(t *testing.T) {
	s := newCrewScene(t)
	car := s.pushVehicle("Getaway Car", getawayCarOracle, 4, 3)
	crewer := s.pushCreature("Crewer", "Creature — Human", 1, 1)
	idle := s.pushCreature("Idle", "Creature — Human", 1, 1)
	s.crewAndResolve(t, car, 0, crewer)

	declareAttack(t, s.g, s.opp.ID, car)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if slices.Contains(pick.PickTargetCards, idle) || !slices.Contains(pick.PickTargetCards, crewer) {
		t.Fatalf("offered %v, want only the crewer", pick.PickTargetCards)
	}
	pickTriggerTarget(t, s.g, s.me.ID, crewer)
	passPriorityAroundTable(t, s.g)
	if zoneKindOf(s.g, crewer) != game.ZoneHand {
		t.Errorf("the crewer is in %v, want the hand", zoneKindOf(s.g, crewer))
	}
	if zoneKindOf(s.g, idle) != game.ZoneBattlefield {
		t.Error("the idle creature left the battlefield")
	}
}

func TestGetawayCarReturnsACrewerWhenItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	attacker, blocker := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	raider := pushVanillaCreature(g, attacker.ID, "Raider", 2, 2)
	car := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Getaway Car", OracleID: getawayCarOracle, TypeLine: crewCardsVehicleTypeLine,
		Power: 4, Toughness: 3, Owner: blocker.ID, Controller: blocker.ID,
	})
	crewer := pushDiesCreatureForTest(g, blocker.ID, "Crewer", "", "Creature — Human", 1, 1)
	crewForTest(t, g, blocker.ID, car, crewer)

	declareAttack(t, g, blocker.ID, raider)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(car, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	pickTriggerTarget(t, g, blocker.ID, crewer)
	passPriorityAroundTable(t, g)
	if zoneKindOf(g, crewer) != game.ZoneHand {
		t.Errorf("the crewer is in %v, want the owner's hand", zoneKindOf(g, crewer))
	}
}

// "Up to one": with every crewer gone the trigger still goes on the stack
// (a target is not required, CR 601.2c) and resolves doing nothing.
func TestGetawayCarTriggersWithNoCrewerLeft(t *testing.T) {
	s := newCrewScene(t)
	car := s.pushVehicle("Getaway Car", getawayCarOracle, 4, 3)
	crewer := s.pushCreature("Crewer", "Creature — Human", 1, 1)
	s.crewAndResolve(t, car, 0, crewer)
	s.g.WithWriteLock(func() {
		if err := s.g.BounceToHandForEffect(crewer); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	declareAttack(t, s.g, s.opp.ID, car)
	if triggerOnStack(s.g, car) == nil {
		t.Fatal("the trigger was removed for want of a target; \"up to one\" needs none")
	}
	passPriorityAroundTable(t, s.g)
	if zoneKindOf(s.g, crewer) != game.ZoneHand {
		t.Error("the bounced crewer moved again")
	}
}

func TestGetawayCarRefusesAnIllegalTarget(t *testing.T) {
	s := newCrewScene(t)
	car := s.pushVehicle("Getaway Car", getawayCarOracle, 4, 3)
	crewer := s.pushCreature("Crewer", "Creature — Human", 1, 1)
	idle := s.pushCreature("Idle", "Creature — Human", 1, 1)
	s.crewAndResolve(t, car, 0, crewer)
	declareAttack(t, s.g, s.opp.ID, car)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := s.g.ResolvePickTarget(pick.ID, s.me.ID, game.TargetRef{Kind: game.TargetCard, ID: idle}); err == nil {
		t.Error("a creature that did not crew the Car was accepted")
	}
}

// --- Subterranean Schooner ---------------------------------------

func TestSubterraneanSchoonerMakesItsCrewerExplore(t *testing.T) {
	s := newCrewScene(t)
	ship := s.pushVehicle("Subterranean Schooner", schoonerOracle, 3, 4)
	crewer := s.pushCreature("Crewer", "Creature — Human", 1, 1)
	idle := s.pushCreature("Idle", "Creature — Human", 1, 1)
	s.crewAndResolve(t, ship, 0, crewer)
	// A land on top: it goes to hand.
	pushLibraryCard(s.me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	handBefore := s.me.Hand.Size()

	declareAttack(t, s.g, s.opp.ID, ship)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil || slices.Contains(pick.PickTargetCards, idle) {
		t.Fatalf("bad prompt %+v", pick)
	}
	pickTriggerTarget(t, s.g, s.me.ID, crewer)
	passPriorityAroundTable(t, s.g)
	if got := s.me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1 for the revealed land", handBefore, got)
	}
	if got := countersOf(t, s.g, crewer); got != 0 {
		t.Errorf("the crewer got %d counters for a land", got)
	}
}

func TestSubterraneanSchoonerNonlandPutsACounterOnTheCrewer(t *testing.T) {
	s := newCrewScene(t)
	ship := s.pushVehicle("Subterranean Schooner", schoonerOracle, 3, 4)
	crewer := s.pushCreature("Crewer", "Creature — Human", 1, 1)
	s.crewAndResolve(t, ship, 0, crewer)
	pushLibraryCard(s.me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})

	declareAttack(t, s.g, s.opp.ID, ship)
	pickTriggerTarget(t, s.g, s.me.ID, crewer)
	passPriorityAroundTable(t, s.g)
	if got := countersOf(t, s.g, crewer); got != 1 {
		t.Errorf("the crewer has %d counters after exploring a nonland, want 1", got)
	}
}

// --- Golden Argosy -------------------------------------------------

func TestGoldenArgosyExilesItsCrewersAndReturnsThemTapped(t *testing.T) {
	s := newCrewScene(t)
	argosy := s.pushVehicle("Golden Argosy", goldenArgosyOracle, 3, 6)
	other := s.pushVehicle("Other Vehicle", smugglersCopterOracle, 3, 3)
	a := s.pushCreature("Crewer A", "Creature — Human", 1, 1)
	b := s.pushCreature("Crewer B", "Creature — Human", 1, 1)
	elsewhere := s.pushCreature("Crewed Elsewhere", "Creature — Human", 1, 1)
	s.crewAndResolve(t, other, 0, elsewhere)
	s.crewAndResolve(t, argosy, 0, a, b)

	declareAttack(t, s.g, s.opp.ID, argosy)
	passPriorityAroundTable(t, s.g)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if got := zoneKindOf(s.g, id); got != game.ZoneExile {
			t.Errorf("crewer %s is in %v, want exile", name, got)
		}
	}
	if got := zoneKindOf(s.g, elsewhere); got != game.ZoneBattlefield {
		t.Errorf("a creature that crewed another Vehicle left the battlefield (%v)", got)
	}

	advanceTo(t, s.g, game.StepEnd)
	passPriorityAroundTable(t, s.g)
	for _, name := range []string{"Crewer A", "Crewer B"} {
		id := findBattlefieldByName(s.g, name)
		if id == uuid.Nil {
			t.Fatalf("%s did not return at the end step", name)
		}
		if c, _ := battlefieldCardByID(s.g, id); !c.Tapped {
			t.Errorf("%s returned untapped", name)
		}
	}
}

// --- Luxurious Locomotive ----------------------------------------

func TestLuxuriousLocomotiveMakesATreasurePerCrewer(t *testing.T) {
	s := newCrewScene(t)
	loco := s.pushVehicle("Luxurious Locomotive", locomotiveOracle, 6, 5)
	a := s.pushCreature("Crewer A", "Creature — Human", 1, 1)
	b := s.pushCreature("Crewer B", "Creature — Human", 1, 1)
	idle := s.pushCreature("Idle", "Creature — Human", 1, 1)
	s.crewAndResolve(t, loco, 0, a, b)

	declareAttack(t, s.g, s.opp.ID, loco)
	passPriorityAroundTable(t, s.g)
	if got := treasureCount(s.g); got != 2 {
		t.Errorf("%d Treasures, want one per crewer (2)", got)
	}

	// Crew is limited to once each turn.
	if err := s.g.ActivateCatalogAbility(s.me.ID, loco, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{idle}}); err == nil {
		t.Error("the Locomotive was crewed a second time in one turn")
	}
}

// A crewer that left before the attack still counts: it is a distinct
// creature object in the Locomotive's record.
func TestLuxuriousLocomotiveCountsDepartedCrewers(t *testing.T) {
	s := newCrewScene(t)
	loco := s.pushVehicle("Luxurious Locomotive", locomotiveOracle, 6, 5)
	a := s.pushCreature("Crewer A", "Creature — Human", 1, 1)
	b := s.pushCreature("Crewer B", "Creature — Human", 1, 1)
	s.crewAndResolve(t, loco, 0, a, b)
	s.g.WithWriteLock(func() {
		if err := s.g.BounceToHandForEffect(a); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})

	declareAttack(t, s.g, s.opp.ID, loco)
	passPriorityAroundTable(t, s.g)
	if got := treasureCount(s.g); got != 2 {
		t.Errorf("%d Treasures, want 2 (a departed crewer still crewed it)", got)
	}
}

// --- Veteran Motorist, Gearshift Ace, Speedway Fanatic ------------

func pushCrewPilot(s crewScene, name, oracle, typeLine string, power, toughness int) uuid.UUID {
	return pushDiesCreatureForTest(s.g, s.me.ID, name, oracle, typeLine, power, toughness)
}

// The trigger is on the stack above the crew ability, so it resolves first
// while the Vehicle is not yet a creature (CR 208.3a).
func TestVeteranMotoristPumpsTheVehicleItCrews(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	copter := pushVehicleForTest(s.g, s.me.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	motorist := pushCrewPilot(s, "Veteran Motorist", veteranMotoristOracle, "Creature — Dwarf Pilot", 3, 1)

	crewBy(t, s.g, s.me.ID, copter, motorist)
	if n := len(s.g.StackMeta); n != 2 {
		t.Fatalf("stack holds %d items, want the crew ability under the Motorist's trigger", n)
	}
	if triggerOnStack(s.g, motorist) == nil {
		t.Fatal("the Motorist's trigger is not on the stack")
	}
	if c, _ := battlefieldCardByID(s.g, copter); c.IsCreature() {
		t.Fatal("the Copter was a creature before the trigger resolved")
	}
	resolveTopOfStack(t, s.g)
	if got := effectivePower(t, s.g, copter); got != 4 {
		t.Errorf("Copter power %d after the trigger, want 4", got)
	}
	passPriorityAroundTable(t, s.g)
	if c, _ := battlefieldCardByID(s.g, copter); !c.IsCreature() {
		t.Error("the Copter did not become a creature")
	}
	if got, want := effectiveToughness(t, s.g, copter), 4; got != want {
		t.Errorf("Copter toughness %d, want %d", got, want)
	}
}

// Another creature crewing does not trigger it; neither does saddling.
func TestVeteranMotoristOnlyTriggersForItsOwnCrew(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	copter := pushVehicleForTest(s.g, s.me.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	motorist := pushCrewPilot(s, "Veteran Motorist", veteranMotoristOracle, "Creature — Dwarf Pilot", 3, 1)
	plain := pushCrewerForTest(s.g, s.me.ID, "Plain", 1)
	mount := pushMountForTest(s.g, s.me.ID, "Gilded Ghoda", gildedGhodaOracle, 2)

	crewBy(t, s.g, s.me.ID, copter, plain)
	if triggerOnStack(s.g, motorist) != nil {
		t.Error("the Motorist triggered for another creature's crew")
	}
	passPriorityAroundTable(t, s.g)
	if err := saddleAbility(t, s.g, s.me.ID, mount, motorist); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	if triggerOnStack(s.g, motorist) != nil {
		t.Error("the Motorist triggered for a saddle")
	}
}

func TestGearshiftAceGivesTheVehicleFirstStrike(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	copter := pushVehicleForTest(s.g, s.me.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	ace := pushCrewPilot(s, "Gearshift Ace", gearshiftAceOracle, "Creature — Dwarf Pilot", 2, 1)

	if containsString(effectiveAbilities(t, s.g, copter), "first strike") {
		t.Fatal("the Copter starts with first strike")
	}
	crewBy(t, s.g, s.me.ID, copter, ace)
	passPriorityAroundTable(t, s.g)
	if !containsString(effectiveAbilities(t, s.g, copter), "first strike") {
		t.Error("the crewed Vehicle did not gain first strike")
	}
	if !containsString(effectiveAbilities(t, s.g, ace), "first strike") {
		t.Error("the Ace lost its own first strike")
	}
}

func TestSpeedwayFanaticGivesTheVehicleHaste(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	copter := pushVehicleForTest(s.g, s.me.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	fan := pushCrewPilot(s, "Speedway Fanatic", speedwayFanaticOracle, "Creature — Human Pilot", 2, 1)

	crewBy(t, s.g, s.me.ID, copter, fan)
	passPriorityAroundTable(t, s.g)
	if !containsString(effectiveAbilities(t, s.g, copter), "haste") {
		t.Error("the crewed Vehicle did not gain haste")
	}
}

// A Vehicle that left in response is a new object and gains nothing.
func TestSpeedwayFanaticGivesNothingToAVehicleThatLeft(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	copter := pushVehicleForTest(s.g, s.me.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	fan := pushCrewPilot(s, "Speedway Fanatic", speedwayFanaticOracle, "Creature — Human Pilot", 2, 1)

	crewBy(t, s.g, s.me.ID, copter, fan)
	flickerForTest(t, s.g, s.me.ID, copter)
	passPriorityAroundTable(t, s.g)
	id := findBattlefieldByName(s.g, "Smuggler's Copter")
	if id == uuid.Nil {
		t.Skip("the Copter did not come back")
	}
	if containsString(effectiveAbilities(t, s.g, id), "haste") {
		t.Error("a Vehicle that left and came back gained haste")
	}
}

// --- Hotshot Mechanic, Giant Ox -----------------------------------

func TestHotshotMechanicCrewsAsThoughItsPowerWereTwoGreater(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	vehicle := pushTestVehicle(s.g, s.me.ID, 4, 4)
	mount := pushMountForTest(s.g, s.me.ID, "Saddle Test Mount", costPowerMountOracle, 3)
	mech := pushPilotCard(s.g, s.me.ID, "Hotshot Mechanic", "Artifact Creature — Fox Pilot", hotshotMechanicOracle, 2, 1)

	if got := countedPowerOf(t, s.g, mech, game.CostPowerCrew); got != 4 {
		t.Errorf("counts %d for crew, want 4", got)
	}
	if got := countedPowerOf(t, s.g, mech, game.CostPowerSaddle); got != 2 {
		t.Errorf("counts %d for saddle, want its plain 2", got)
	}
	if got := effectivePower(t, s.g, mech); got != 2 {
		t.Errorf("real power %d, want 2", got)
	}
	if err := crewCostPowerVehicle(t, s.g, s.me.ID, vehicle, mech); err != nil {
		t.Errorf("the Mechanic could not crew a crew-3 Vehicle alone: %v", err)
	}
	passPriorityAroundTable(t, s.g)
	if err := saddleAbility(t, s.g, s.me.ID, mount, pushCrewerForTest(s.g, s.me.ID, "Two", 2)); err == nil {
		t.Error("a plain 2-power creature saddled a Saddle 3 Mount")
	}
}

func TestGiantOxCrewsUsingToughness(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	vehicle := pushTestVehicle(s.g, s.me.ID, 4, 4)
	ox := pushPilotCard(s.g, s.me.ID, "Giant Ox", "Creature — Ox", giantOxOracle, 0, 6)

	if got := countedPowerOf(t, s.g, ox, game.CostPowerCrew); got != 6 {
		t.Errorf("counts %d for crew, want its toughness 6", got)
	}
	if got := countedPowerOf(t, s.g, ox, game.CostPowerSaddle); got != 0 {
		t.Errorf("counts %d for saddle, want its plain power 0", got)
	}
	if got := effectivePower(t, s.g, ox); got != 0 {
		t.Errorf("real power %d, want 0", got)
	}
	if err := crewCostPowerVehicle(t, s.g, s.me.ID, vehicle, ox); err != nil {
		t.Errorf("the Ox could not crew a crew-3 Vehicle alone: %v", err)
	}
}

// --- Shorikai, Genesis Engine ---------------------------------------

func TestShorikaiLootsAndMakesACrewOnlyPilot(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	shorikai := pushVehicleForTest(s.g, s.me.ID, "Shorikai, Genesis Engine", shorikaiOracle, 8, 8)
	if err := s.g.AddManaForEffect(s.me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	handBefore := s.me.Hand.Size()

	if err := s.g.ActivateCatalogAbility(s.me.ID, shorikai, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := battlefieldCardByID(s.g, shorikai); !c.Tapped {
		t.Error("Shorikai did not tap for its own cost")
	}
	passPriorityAroundTable(t, s.g)
	if got := s.me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d -> %d, want +2 before the discard", handBefore, got)
	}
	if got := discardOwed(s.g, s.me.ID); got != 1 {
		t.Errorf("discard owed %d, want 1", got)
	}

	var pilot game.Card
	s.g.ReadSnapshot(func() {
		for _, c := range s.g.Battlefield.Cards {
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
	if got := countedPowerOf(t, s.g, pilot.InstanceID, game.CostPowerCrew); got != 3 {
		t.Errorf("the Pilot counts %d for crew, want 3", got)
	}
	if got := countedPowerOf(t, s.g, pilot.InstanceID, game.CostPowerSaddle); got != 1 {
		t.Errorf("the Pilot counts %d for saddle, want its plain 1", got)
	}
}

// Crew 8 is the second ability; an 8-power creature crews it.
func TestShorikaiCrewsForEight(t *testing.T) {
	s := newCrewScene(t)
	shorikai := pushVehicleForTest(s.g, s.me.ID, "Shorikai, Genesis Engine", shorikaiOracle, 8, 8)
	big := s.pushCreature("Big", "Creature — Giant", 8, 8)
	small := s.pushCreature("Small", "Creature — Human", 7, 7)
	toMain(t, s.g)
	if err := s.g.ActivateCatalogAbility(s.me.ID, shorikai, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Error("7 power crewed a crew-8 Vehicle")
	}
	s.crewAndResolve(t, shorikai, 1, big)
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
			s := newCrewScene(t)
			s.pushVehicle("Prodigy's Prototype", prodigysPrototypeOracle, 3, 4)
			var attackers []uuid.UUID
			for i := 0; i < tc.vehicles; i++ {
				v := s.pushVehicle("Crewed Vehicle", costPowerVehicleOracle, 3, 3)
				c := s.pushCreature("Crewer", "Creature — Human", 3, 3)
				s.crewAndResolve(t, v, 0, c)
				attackers = append(attackers, v)
			}
			if tc.creature {
				attackers = append(attackers, s.pushCreature("Attacker", "Creature — Human", 2, 2))
			}
			toMain(t, s.g)
			declareAttack(t, s.g, s.opp.ID, attackers...)
			passPriorityAroundTable(t, s.g)

			var pilots []game.Card
			s.g.ReadSnapshot(func() {
				for _, c := range s.g.Battlefield.Cards {
					if c.Name == "Pilot" {
						pilots = append(pilots, c)
					}
				}
			})
			if len(pilots) != tc.want {
				t.Fatalf("%d Pilots, want %d", len(pilots), tc.want)
			}
			for _, p := range pilots {
				if got := countedPowerOf(t, s.g, p.InstanceID, game.CostPowerCrew); got != 3 {
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
	s := newCrewScene(t)
	host := s.pushCreature("Host", "Creature — Human", 1, 1)
	s.pushCreature("Friend", "Creature — Human", 1, 1)
	vehicle := s.pushVehicle("Idle Vehicle", smugglersCopterOracle, 3, 3)

	aura := castCatalogSpell(t, s.g, "Born to Drive", auraTypeLine, bornToDriveOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, s.g)
	if !s.g.Battlefield.Contains(aura) {
		t.Fatal("Born to Drive did not resolve to the battlefield")
	}
	// Host, Friend and the Vehicle: three permanents that are creatures
	// and/or Vehicles.
	if got := effectivePower(t, s.g, host); got != 4 {
		t.Errorf("enchanted creature power %d, want 1 + 3", got)
	}
	// Crewing the Vehicle does not count it twice.
	crewer := s.pushCreature("Crewer", "Creature — Human", 3, 3)
	if got := effectivePower(t, s.g, host); got != 5 {
		t.Errorf("after another creature arrived: power %d, want 1 + 4", got)
	}
	s.crewAndResolve(t, vehicle, 0, crewer)
	if got := effectivePower(t, s.g, host); got != 5 {
		t.Errorf("after the Vehicle was crewed: power %d, want it counted once (5)", got)
	}
	// An opponent's creature is not counted.
	pushDiesCreatureForTest(s.g, s.opp.ID, "Enemy", "", "Creature — Human", 2, 2)
	if got := effectivePower(t, s.g, host); got != 5 {
		t.Errorf("an opponent's creature changed the bonus: %d", got)
	}
}

// On an uncrewed Vehicle the bonus waits for it to become a creature.
func TestBornToDriveOnAnUncrewedVehicleWaitsForItToBeACreature(t *testing.T) {
	s := newCrewScene(t)
	vehicle := s.pushVehicle("Idle Vehicle", smugglersCopterOracle, 3, 3)
	crewer := s.pushCreature("Crewer", "Creature — Human", 3, 3)
	castCatalogSpell(t, s.g, "Born to Drive", auraTypeLine, bornToDriveOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: vehicle}})
	passPriorityAroundTable(t, s.g)
	if got := effectivePower(t, s.g, vehicle); got != 3 {
		t.Errorf("uncrewed Vehicle power %d, want its printed 3", got)
	}
	s.crewAndResolve(t, vehicle, 0, crewer)
	// Crewer and the Vehicle: 3 + 2.
	if got := effectivePower(t, s.g, vehicle); got != 5 {
		t.Errorf("crewed Vehicle power %d, want 3 + 2", got)
	}
}

func TestBornToDriveChannelMakesTwoCrewOnlyPilots(t *testing.T) {
	s := newCrewScene(t)
	toMain(t, s.g)
	id := pushCatalogHandCard(s.me, "Born to Drive", auraTypeLine, bornToDriveOracle)
	if err := s.g.AddManaForEffect(s.me.ID, uuid.Nil, "{C}{C}{W}"); err != nil {
		t.Fatal(err)
	}
	offered := false
	for _, m := range legal.EnumerateFor(s.g, s.me.ID) {
		if m.Source == id && m.Type == legal.TypeActivateAbility {
			offered = true
		}
	}
	if !offered {
		t.Error("channel is not offered from the hand")
	}
	if err := s.g.ActivateCatalogAbility(s.me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	passPriorityAroundTable(t, s.g)
	if got := zoneKindOf(s.g, id); got != game.ZoneGraveyard {
		t.Errorf("the channeled card is in %v, want the graveyard", got)
	}
	n := 0
	s.g.ReadSnapshot(func() {
		for _, c := range s.g.Battlefield.Cards {
			if c.Name == "Pilot" && c.TokenKey == game.TokenKey("pilot-crews") {
				n++
			}
		}
	})
	if n != 2 {
		t.Errorf("%d crew-only Pilots, want 2", n)
	}
}
