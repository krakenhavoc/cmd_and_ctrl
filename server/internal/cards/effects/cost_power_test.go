package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_power_test.go — ADR 0186 Decision 4 (#3549, seam #2705): a
// creature that "crews Vehicles as though its power were 2 greater" or
// "using its toughness rather than its power". The cards that print the
// rule arrive in later slices, because the client does not count the
// bonus until the slice after this one; every card here is built in
// this file.

const (
	costPowerVehicleOracle   = "test-cost-power-vehicle"
	costPowerMountOracle     = "test-cost-power-mount"
	costPowerCrewOnlyOracle  = "test-cost-power-crew-only"
	costPowerBothOracle      = "test-cost-power-both"
	costPowerToughnessOracle = "test-cost-power-toughness"
	costPowerTwiceOracle     = "test-cost-power-twice"
)

func init() {
	// A Vehicle that says Crew 3 and a Mount that says Saddle 3.
	Register(Spec{
		OracleID: costPowerVehicleOracle,
		Name:     "Cost Power Vehicle",
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Effect:  CrewEffect("Cost Power Vehicle"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
		}},
	})
	Register(Spec{
		OracleID:  costPowerMountOracle,
		Name:      "Cost Power Mount",
		Activated: []ActivatedAbility{Saddle(3)},
	})
	// "This creature crews Vehicles as though its power were 2 greater."
	Register(Spec{
		OracleID: costPowerCrewOnlyOracle,
		Name:     "Cost Power Crew Only",
		Static:   []game.StaticAbility{CrewsAsThoughPowerGreater(2, game.CostPowerCrew)},
	})
	// "This creature saddles Mounts and crews Vehicles as though its
	// power were 2 greater."
	Register(Spec{
		OracleID: costPowerBothOracle,
		Name:     "Cost Power Both",
		Static:   []game.StaticAbility{CrewsAsThoughPowerGreater(2, game.CostPowerCrew|game.CostPowerSaddle)},
	})
	// "This creature crews Vehicles using its toughness rather than its power."
	Register(Spec{
		OracleID: costPowerToughnessOracle,
		Name:     "Cost Power Toughness",
		Static:   []game.StaticAbility{CrewsUsingToughness(game.CostPowerCrew)},
	})
	// Two copies of the rule on one creature.
	Register(Spec{
		OracleID: costPowerTwiceOracle,
		Name:     "Cost Power Twice",
		Static: []game.StaticAbility{
			CrewsAsThoughPowerGreater(2, game.CostPowerCrew),
			CrewsAsThoughPowerGreater(2, game.CostPowerCrew),
		},
	})
}

// pushCostPowerCreature puts a catalog creature with the given oracle ID
// on the battlefield through the path that stamps its statics.
func pushCostPowerCreature(g *game.Game, owner uuid.UUID, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Pilot " + oracle,
		OracleID:   oracle,
		TypeLine:   "Creature — Pilot",
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
}

// countedPowerOf is Card.CountedPower for the object as the engine reads
// it: layers fresh, under the lock.
func countedPowerOf(t *testing.T, g *game.Game, id uuid.UUID, use game.CostPowerUse) int {
	t.Helper()
	var n int
	var found bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n, found = c.CountedPower(use), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return n
}

func realPowerOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var n int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.CurrentPower()
			}
		}
	})
	return n
}

func crewCostPowerVehicle(t *testing.T, g *game.Game, owner, vehicle uuid.UUID, crewers ...uuid.UUID) error {
	t.Helper()
	return g.ActivateCatalogAbility(owner, vehicle, 0, game.ActivateAbilityParams{CrewIDs: crewers})
}

func TestCostPowerOneOneCrewsACrewThreeVehicleAlone(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	pilot := pushCostPowerCreature(g, me.ID, costPowerCrewOnlyOracle, 1, 1)

	if got := countedPowerOf(t, g, pilot, game.CostPowerCrew); got != 3 {
		t.Fatalf("CountedPower(crew) = %d, want 3", got)
	}
	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, pilot); err != nil {
		t.Fatalf("a 1/1 counting as 3 could not crew a crew-3 Vehicle alone: %v", err)
	}
	if c, _ := battlefieldCardByID(g, pilot); !c.Tapped {
		t.Error("the crewing creature did not tap")
	}
	// The creature's own power is untouched throughout (Dynamite
	// Diver's ruling): "where X is its power" still reads 1.
	if got := realPowerOf(t, g, pilot); got != 1 {
		t.Errorf("CurrentPower after crewing = %d, want 1", got)
	}
}

func TestCostPowerWithoutTheRuleOneOneCannotCrewThree(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	plain := pushCrewerForTest(g, me.ID, "Plain", 1)

	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, plain); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("a plain 1/1 crewing a crew-3 Vehicle: err = %v, want ErrInsufficientCrew", err)
	}
}

// The rule names the costs it changes: this creature crews as though
// its power were 2 greater, and saddles for its plain power.
func TestCostPowerCrewOnlyRuleDoesNotHelpASaddle(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := pushMountForTest(g, me.ID, "Cost Power Mount", costPowerMountOracle, 2)
	pilot := pushCostPowerCreature(g, me.ID, costPowerCrewOnlyOracle, 1, 1)

	if got := countedPowerOf(t, g, pilot, game.CostPowerSaddle); got != 1 {
		t.Fatalf("CountedPower(saddle) = %d, want 1", got)
	}
	if err := saddleAbility(t, g, me.ID, mount, pilot); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("a crew-only creature saddling a Saddle 3 Mount: err = %v, want ErrInsufficientCrew", err)
	}
	if c, _ := battlefieldCardByID(g, pilot); c.Tapped {
		t.Error("a refused saddle tapped the creature")
	}
}

func TestCostPowerBothRuleSaddlesAndCrews(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := pushMountForTest(g, me.ID, "Cost Power Mount", costPowerMountOracle, 2)
	pilot := pushCostPowerCreature(g, me.ID, costPowerBothOracle, 1, 1)

	if err := saddleAbility(t, g, me.ID, mount, pilot); err != nil {
		t.Fatalf("a 1/1 counting as 3 could not saddle a Saddle 3 Mount: %v", err)
	}
}

func TestCostPowerUsingToughnessCrewsForToughness(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	wall := pushCostPowerCreature(g, me.ID, costPowerToughnessOracle, 0, 4)

	if got := countedPowerOf(t, g, wall, game.CostPowerCrew); got != 4 {
		t.Fatalf("CountedPower(crew) of a 0/4 using toughness = %d, want 4", got)
	}
	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, wall); err != nil {
		t.Fatalf("a 0/4 using its toughness could not crew a crew-3 Vehicle: %v", err)
	}
	if got := realPowerOf(t, g, wall); got != 0 {
		t.Errorf("CurrentPower = %d, want 0", got)
	}
}

// CR 613.1f: the rule is the creature's own ability, so a creature that
// loses all abilities loses it, and counts for its plain power again.
func TestCostPowerIsGoneAfterLosingAllAbilities(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	pilot := pushCostPowerCreature(g, me.ID, costPowerCrewOnlyOracle, 1, 1)
	if got := countedPowerOf(t, g, pilot, game.CostPowerCrew); got != 3 {
		t.Fatalf("setup: CountedPower = %d, want 3", got)
	}

	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(pilot),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})

	if got := countedPowerOf(t, g, pilot, game.CostPowerCrew); got != 1 {
		t.Errorf("CountedPower after losing all abilities = %d, want 1", got)
	}
	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, pilot); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("crewing after losing all abilities: err = %v, want ErrInsufficientCrew", err)
	}
	if got := realPowerOf(t, g, pilot); got != 1 {
		t.Errorf("CurrentPower = %d, want 1", got)
	}
}

// Owner decision 2: the bonus is added to the unclamped power and the
// sum is clamped, so a -1/1 with +2 counts 1, not 2.
func TestCostPowerNegativePowerAddsBeforeItClamps(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pilot := pushCostPowerCreature(g, me.ID, costPowerCrewOnlyOracle, -1, 1)

	if got := countedPowerOf(t, g, pilot, game.CostPowerCrew); got != 1 {
		t.Errorf("CountedPower of a -1/1 with +2 = %d, want 1", got)
	}
	if got := realPowerOf(t, g, pilot); got != 0 {
		t.Errorf("CurrentPower of a -1/1 = %d, want 0 (clamped, unchanged by the rule)", got)
	}
}

func TestCostPowerTwoRulesAddUp(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	twice := pushCostPowerCreature(g, me.ID, costPowerTwiceOracle, 1, 1)

	if got := countedPowerOf(t, g, twice, game.CostPowerCrew); got != 5 {
		t.Fatalf("CountedPower with two +2 rules = %d, want 1 + 4 = 5", got)
	}
	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, twice); err != nil {
		t.Fatalf("crew: %v", err)
	}
}

// A creature that counts for more still counts alongside others: the
// payment is a sum.
func TestCostPowerCountsAlongsideOtherCrewers(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	pilot := pushCostPowerCreature(g, me.ID, costPowerCrewOnlyOracle, 1, 1) // counts 3
	plain := pushCrewerForTest(g, me.ID, "Plain", 1)

	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, plain); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("one plain 1/1: err = %v, want ErrInsufficientCrew", err)
	}
	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, pilot, plain); err != nil {
		t.Fatalf("a Pilot and a 1/1 (4 counted) could not crew a crew-3 Vehicle: %v", err)
	}
}
