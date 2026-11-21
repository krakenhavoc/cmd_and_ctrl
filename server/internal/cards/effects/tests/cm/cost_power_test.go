package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushCostPowerCreature puts a catalog creature with the given oracle ID
// on the battlefield through the path that stamps its statics.
func pushCostPowerCreature(g *game.Game, owner uuid.UUID, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
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

func TestCostPowerOneOneCrewsACrewThreeVehicleAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Cost Power Vehicle", effectstest.CostPowerVehicleOracle, 4, 4)
	pilot := pushCostPowerCreature(g, me.ID, effectstest.CostPowerCrewOnlyOracle, 1, 1)

	if got := effectstest.CountedPowerOf(t, g, pilot, game.CostPowerCrew); got != 3 {
		t.Fatalf("CountedPower(crew) = %d, want 3", got)
	}
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, pilot); err != nil {
		t.Fatalf("a 1/1 counting as 3 could not crew a crew-3 Vehicle alone: %v", err)
	}
	if c, _ := effectstest.BattlefieldCardByID(g, pilot); !c.Tapped {
		t.Error("the crewing creature did not tap")
	}
	// The creature's own power is untouched throughout (Dynamite
	// Diver's ruling): "where X is its power" still reads 1.
	if got := effectstest.RealPowerOf(t, g, pilot); got != 1 {
		t.Errorf("CurrentPower after crewing = %d, want 1", got)
	}
}

func TestCostPowerWithoutTheRuleOneOneCannotCrewThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Cost Power Vehicle", effectstest.CostPowerVehicleOracle, 4, 4)
	plain := effectstest.PushCrewerForTest(g, me.ID, "Plain", 1)

	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, plain); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("a plain 1/1 crewing a crew-3 Vehicle: err = %v, want ErrInsufficientCrew", err)
	}
}

// The rule names the costs it changes: this creature crews as though
// its power were 2 greater, and saddles for its plain power.
func TestCostPowerCrewOnlyRuleDoesNotHelpASaddle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Cost Power Mount", effectstest.CostPowerMountOracle, 2)
	pilot := pushCostPowerCreature(g, me.ID, effectstest.CostPowerCrewOnlyOracle, 1, 1)

	if got := effectstest.CountedPowerOf(t, g, pilot, game.CostPowerSaddle); got != 1 {
		t.Fatalf("CountedPower(saddle) = %d, want 1", got)
	}
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, pilot); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("a crew-only creature saddling a Saddle 3 Mount: err = %v, want ErrInsufficientCrew", err)
	}
	if c, _ := effectstest.BattlefieldCardByID(g, pilot); c.Tapped {
		t.Error("a refused saddle tapped the creature")
	}
}

func TestCostPowerBothRuleSaddlesAndCrews(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Cost Power Mount", effectstest.CostPowerMountOracle, 2)
	pilot := pushCostPowerCreature(g, me.ID, effectstest.CostPowerBothOracle, 1, 1)

	if err := effectstest.SaddleAbility(t, g, me.ID, mount, pilot); err != nil {
		t.Fatalf("a 1/1 counting as 3 could not saddle a Saddle 3 Mount: %v", err)
	}
}

func TestCostPowerUsingToughnessCrewsForToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Cost Power Vehicle", effectstest.CostPowerVehicleOracle, 4, 4)
	wall := pushCostPowerCreature(g, me.ID, effectstest.CostPowerToughnessOracle, 0, 4)

	if got := effectstest.CountedPowerOf(t, g, wall, game.CostPowerCrew); got != 4 {
		t.Fatalf("CountedPower(crew) of a 0/4 using toughness = %d, want 4", got)
	}
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, wall); err != nil {
		t.Fatalf("a 0/4 using its toughness could not crew a crew-3 Vehicle: %v", err)
	}
	if got := effectstest.RealPowerOf(t, g, wall); got != 0 {
		t.Errorf("CurrentPower = %d, want 0", got)
	}
}

// CR 613.1f: the rule is the creature's own ability, so a creature that
// loses all abilities loses it, and counts for its plain power again.
func TestCostPowerIsGoneAfterLosingAllAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Cost Power Vehicle", effectstest.CostPowerVehicleOracle, 4, 4)
	pilot := pushCostPowerCreature(g, me.ID, effectstest.CostPowerCrewOnlyOracle, 1, 1)
	if got := effectstest.CountedPowerOf(t, g, pilot, game.CostPowerCrew); got != 3 {
		t.Fatalf("setup: CountedPower = %d, want 3", got)
	}

	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(pilot),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})

	if got := effectstest.CountedPowerOf(t, g, pilot, game.CostPowerCrew); got != 1 {
		t.Errorf("CountedPower after losing all abilities = %d, want 1", got)
	}
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, pilot); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("crewing after losing all abilities: err = %v, want ErrInsufficientCrew", err)
	}
	if got := effectstest.RealPowerOf(t, g, pilot); got != 1 {
		t.Errorf("CurrentPower = %d, want 1", got)
	}
}

// Owner decision 2: the bonus is added to the unclamped power and the
// sum is clamped, so a -1/1 with +2 counts 1, not 2.
func TestCostPowerNegativePowerAddsBeforeItClamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pilot := pushCostPowerCreature(g, me.ID, effectstest.CostPowerCrewOnlyOracle, -1, 1)

	if got := effectstest.CountedPowerOf(t, g, pilot, game.CostPowerCrew); got != 1 {
		t.Errorf("CountedPower of a -1/1 with +2 = %d, want 1", got)
	}
	if got := effectstest.RealPowerOf(t, g, pilot); got != 0 {
		t.Errorf("CurrentPower of a -1/1 = %d, want 0 (clamped, unchanged by the rule)", got)
	}
}

func TestCostPowerTwoRulesAddUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Cost Power Vehicle", effectstest.CostPowerVehicleOracle, 4, 4)
	twice := pushCostPowerCreature(g, me.ID, effectstest.CostPowerTwiceOracle, 1, 1)

	if got := effectstest.CountedPowerOf(t, g, twice, game.CostPowerCrew); got != 5 {
		t.Fatalf("CountedPower with two +2 rules = %d, want 1 + 4 = 5", got)
	}
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, twice); err != nil {
		t.Fatalf("crew: %v", err)
	}
}

// A creature that counts for more still counts alongside others: the
// payment is a sum.
func TestCostPowerCountsAlongsideOtherCrewers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Cost Power Vehicle", effectstest.CostPowerVehicleOracle, 4, 4)
	pilot := pushCostPowerCreature(g, me.ID, effectstest.CostPowerCrewOnlyOracle, 1, 1) // counts 3
	plain := effectstest.PushCrewerForTest(g, me.ID, "Plain", 1)

	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, plain); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("one plain 1/1: err = %v, want ErrInsufficientCrew", err)
	}
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, pilot, plain); err != nil {
		t.Fatalf("a Pilot and a 1/1 (4 counted) could not crew a crew-3 Vehicle: %v", err)
	}
}
