package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crew_power_view_test.go — ADR 0186 Decision 6 (#3549, seam #2705):
// `crew_power` is what each offered creature counts for on this ability.
// The client sums it instead of `power`; a client that ignores it
// under-counts a Pilot and never over-counts.

const viewCostPowerPilotOracle = "test-view-cost-power-pilot"

func init() {
	// "This creature crews Vehicles as though its power were 2 greater."
	// Crew only, so the same creature shows 3 on a Vehicle and 1 on a
	// Mount, which is the per-cost reading the field exists to carry.
	effects.Register(effects.Spec{
		OracleID: viewCostPowerPilotOracle,
		Name:     "View Cost Power Pilot",
		Static:   []game.StaticAbility{effects.CrewsAsThoughPowerGreater(2, game.CostPowerCrew)},
	})
}

func seatCostPowerPilot(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "Pilot", TypeLine: "Creature — Pilot", OracleID: viewCostPowerPilotOracle,
			Power: 1, Toughness: 1, Owner: owner, Controller: owner,
		})
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		g.RecomputeLayersIfStaleLocked()
	})
	return id
}

func TestCrewPowerCarriesTheCountedPowerOfEachOffer(t *testing.T) {
	g := buildActiveGame(t)
	vehicle, owner := seatCrewVehicle(g)
	pilot := seatCostPowerPilot(g, owner)
	plain := seatCreature(g, owner, "Plain", 2, false, false)

	ab := vehicleView(t, g, vehicle).ActivatedAbilities[0]
	if got := ab.CrewPower[pilot.String()]; got != 3 {
		t.Errorf("crew_power[pilot] = %d, want 3 (a 1/1 that crews as though 2 greater)", got)
	}
	if got := ab.CrewPower[plain.String()]; got != 2 {
		t.Errorf("crew_power[plain] = %d, want 2 (plain power)", got)
	}
	if len(ab.CrewPower) != len(ab.CrewOptions.Cards) {
		t.Errorf("crew_power has %d entries, crew_options %d; they must name the same creatures", len(ab.CrewPower), len(ab.CrewOptions.Cards))
	}
	// The creature's own power on the wire is untouched.
	for _, c := range ViewOfGame(g).Battlefield.Cards {
		if c.InstanceID == pilot.String() && c.Power != 1 {
			t.Errorf("the Pilot's wire power = %d, want 1: the rule never changes the creature's power", c.Power)
		}
	}
}

// A Mount's saddle ability rides the same fields and reads the saddle
// cost's counted power: a crew-only creature is 1 there.
func TestCrewPowerOnASaddleAbilityReadsTheSaddleCost(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	mount := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: mount, Name: "Test Mount", TypeLine: "Creature — Horse Mount",
		OracleID: "00000000-0000-0000-0000-0000000000dd", Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "Saddle 3", Cost: game.AbilityCost{Saddle: 3}},
		},
	})
	pilot := seatCostPowerPilot(g, owner)

	ab := vehicleView(t, g, mount).ActivatedAbilities[0]
	if !ab.Saddle {
		t.Fatal("setup: the ability is not a saddle")
	}
	if got, ok := ab.CrewPower[pilot.String()]; !ok || got != 1 {
		t.Errorf("crew_power[pilot] on a saddle = %d (present %v), want 1", got, ok)
	}
	if _, ok := ab.CrewPower[mount.String()]; ok {
		t.Error("crew_power names the Mount, which may not pay for its own saddle")
	}
}

func TestNonCrewAbilityCarriesNoCrewPower(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Ordinary Artifact", TypeLine: "Artifact",
		OracleID: "00000000-0000-0000-0000-0000000000cc", Owner: owner, Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{Label: "{T}: do a thing", Cost: game.AbilityCost{Tap: true}}},
	})
	seatCostPowerPilot(g, owner)
	if ab := vehicleView(t, g, id).ActivatedAbilities[0]; ab.CrewPower != nil {
		t.Errorf("a non-crew ability carried crew_power: %v", ab.CrewPower)
	}
}

// Teamwork counts plain power (CR 702.194a), so its prompt carries no
// crew_power, whoever is on the board.
func TestTeamworkPromptCarriesNoCrewPower(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	stubTeamworkBlightCard(t)
	seatCostPowerPilot(g, me.ID)
	pushViewCreature(g, me.ID, 3, false) // enough plain power for teamwork 3
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Team Thing", TypeLine: "Sorcery", OracleID: viewTeamworkBlightOracle,
			ManaCost: "{1}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if len(c.OptionalCosts) == 0 || c.OptionalCosts[0].TeamworkOptions == nil || len(c.OptionalCosts[0].TeamworkOptions.Cards) == 0 {
		t.Fatalf("setup: no teamwork options: %+v", c.OptionalCosts)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "crew_power") {
		t.Errorf("a teamwork prompt carries crew_power: %s", raw)
	}
}
