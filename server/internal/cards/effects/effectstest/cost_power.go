package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_power_test.go — ADR 0186 Decision 4 (#3549, seam #2705): a
// creature that "crews Vehicles as though its power were 2 greater" or
// "using its toughness rather than its power". The cards that print the
// rule arrive in later slices, because the client does not count the
// bonus until the slice after this one; every card here is built in
// this file.

const (
	CostPowerVehicleOracle   = "test-cost-power-vehicle"
	CostPowerMountOracle     = "test-cost-power-mount"
	CostPowerCrewOnlyOracle  = "test-cost-power-crew-only"
	CostPowerBothOracle      = "test-cost-power-both"
	CostPowerToughnessOracle = "test-cost-power-toughness"
	CostPowerTwiceOracle     = "test-cost-power-twice"
)

func init() {
	// A Vehicle that says Crew 3 and a Mount that says Saddle 3.
	effects.Register(effects.Spec{
		OracleID: CostPowerVehicleOracle,
		Name:     "Cost Power Vehicle",
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Effect:  effects.CrewEffect("Cost Power Vehicle"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:  CostPowerMountOracle,
		Name:      "Cost Power Mount",
		Activated: []effects.ActivatedAbility{effects.Saddle(3)},
	})
	// "This creature crews Vehicles as though its power were 2 greater."
	effects.Register(effects.Spec{
		OracleID: CostPowerCrewOnlyOracle,
		Name:     "Cost Power Crew Only",
		Static:   []game.StaticAbility{effects.CrewsAsThoughPowerGreater(2, game.CostPowerCrew)},
	})
	// "This creature saddles Mounts and crews Vehicles as though its
	// power were 2 greater."
	effects.Register(effects.Spec{
		OracleID: CostPowerBothOracle,
		Name:     "Cost Power Both",
		Static:   []game.StaticAbility{effects.CrewsAsThoughPowerGreater(2, game.CostPowerCrew|game.CostPowerSaddle)},
	})
	// "This creature crews Vehicles using its toughness rather than its power."
	effects.Register(effects.Spec{
		OracleID: CostPowerToughnessOracle,
		Name:     "Cost Power Toughness",
		Static:   []game.StaticAbility{effects.CrewsUsingToughness(game.CostPowerCrew)},
	})
	// Two copies of the rule on one creature.
	effects.Register(effects.Spec{
		OracleID: CostPowerTwiceOracle,
		Name:     "Cost Power Twice",
		Static: []game.StaticAbility{
			effects.CrewsAsThoughPowerGreater(2, game.CostPowerCrew),
			effects.CrewsAsThoughPowerGreater(2, game.CostPowerCrew),
		},
	})
}

// countedPowerOf is Card.CountedPower for the object as the engine reads
// it: layers fresh, under the lock.
func CountedPowerOf(t *testing.T, g *game.Game, id uuid.UUID, use game.CostPowerUse) int {
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

func RealPowerOf(t *testing.T, g *game.Game, id uuid.UUID) int {
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

func CrewCostPowerVehicle(t *testing.T, g *game.Game, owner, vehicle uuid.UUID, crewers ...uuid.UUID) error {
	t.Helper()
	return g.ActivateCatalogAbility(owner, vehicle, 0, game.ActivateAbilityParams{CrewIDs: crewers})
}
