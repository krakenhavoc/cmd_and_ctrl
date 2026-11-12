package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_crew_view_test.go — #5673, CR 702.122d: a creature that "can't
// crew Vehicles" is left out of a crew row's crew_options and
// crew_power, carries the "cant_crew" restriction token, and is still
// offered for a saddle cost.

const viewCantCrewOracle = "test-view-cant-crew"

func init() {
	// "This creature can't crew Vehicles."
	effects.Register(effects.Spec{
		OracleID: viewCantCrewOracle,
		Name:     "View Cant Crew",
		Static:   []game.StaticAbility{effects.RestrictSelf(game.CantCrew)},
	})
}

func seatCantCrewCreature(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "Held", TypeLine: "Creature — Test", OracleID: viewCantCrewOracle,
			Power: 5, Toughness: 5, Owner: owner, Controller: owner,
		})
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		g.RecomputeLayersIfStaleLocked()
	})
	return id
}

func TestCrewOptionsLeaveOutACreatureThatCantCrew(t *testing.T) {
	g := buildActiveGame(t)
	vehicle, owner := seatCrewVehicle(g)
	held := seatCantCrewCreature(g, owner)
	plain := seatCreature(g, owner, "Plain", 3, false, false)

	ab := vehicleView(t, g, vehicle).ActivatedAbilities[0]
	if ab.CrewOptions == nil || !slices.Contains(ab.CrewOptions.Cards, plain.String()) {
		t.Fatalf("crew_options = %+v, want the plain creature offered", ab.CrewOptions)
	}
	if slices.Contains(ab.CrewOptions.Cards, held.String()) {
		t.Error("crew_options offers a creature that can't crew Vehicles")
	}
	if _, ok := ab.CrewPower[held.String()]; ok {
		t.Error("crew_power counts a creature that can't crew Vehicles")
	}
	if got := vehicleView(t, g, held).Restrictions; !slices.Contains(got, "cant_crew") {
		t.Errorf("restrictions = %v, want cant_crew", got)
	}
}

// A saddle cost is a different cost: the same creature is offered.
func TestSaddleOptionsStillOfferACreatureThatCantCrew(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	mount := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: mount, Name: "Test Mount", TypeLine: "Creature — Horse Mount",
		OracleID: "00000000-0000-0000-0000-0000000000de", Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "Saddle 3", Cost: game.AbilityCost{Saddle: 3}},
		},
	})
	held := seatCantCrewCreature(g, owner)

	ab := vehicleView(t, g, mount).ActivatedAbilities[0]
	if ab.CrewOptions == nil || !slices.Contains(ab.CrewOptions.Cards, held.String()) {
		t.Errorf("a saddle row's crew_options = %+v, want the creature that can't crew Vehicles offered", ab.CrewOptions)
	}
}
