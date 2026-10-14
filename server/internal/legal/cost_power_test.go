package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cost_power_test.go — ADR 0186 Decision 4 and 6 (#3549, seam #2705): the
// enumerator's crew payment counts each creature for what the engine
// counts it for, and prefers the creature with the lower real power when
// two count the same. Teamwork reuses the walk and counts plain power.

const (
	oracleCostPowerVehicle = "test-legal-cost-power-vehicle"
	oracleCostPowerPilot   = "test-legal-cost-power-pilot"
)

func init() {
	effects.Register(effects.Spec{
		OracleID: oracleCostPowerVehicle,
		Name:     "Cost Power Vehicle",
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Effect:  effects.CrewEffect("Cost Power Vehicle"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
		}},
	})
	effects.Register(effects.Spec{
		OracleID: oracleCostPowerPilot,
		Name:     "Cost Power Pilot",
		Static:   []game.StaticAbility{effects.CrewsAsThoughPowerGreater(2, game.CostPowerCrew)},
	})
}

// pushPilot puts a 1/1 that crews as though its power were 2 greater on
// the battlefield, with its statics applied.
func pushPilot(g *game.Game, p *game.Player) uuid.UUID {
	id := battlefieldCard(g, p, game.Card{
		Name: "Pilot", TypeLine: "Creature — Pilot", OracleID: oracleCostPowerPilot, Power: 1, Toughness: 1,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		g.RecomputeLayersIfStaleLocked()
	})
	return id
}

func crewIDsOf(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		CrewIDs []string `json:"crew_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	return p.CrewIDs
}

// A 1/1 that counts 3 crews a Crew 3 Vehicle alone, and the move names it
// and nobody else — even with a 3/3 beside it that counts the same, which
// is the tie-break: the lower real power is tapped, so the 3/3 stays free
// to attack.
func TestCrewMoveNamesTheOneOneThatCountsThree(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	vehicle := battlefieldCard(g, active, game.Card{
		Name: "Cost Power Vehicle", TypeLine: "Artifact — Vehicle", OracleID: oracleCostPowerVehicle, Power: 4, Toughness: 4,
	})
	battlefieldCard(g, active, creature("Ogre", "{2}{R}", 3, 3))
	pilot := pushPilot(g, active)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, vehicle)
	if len(acts) != 1 {
		t.Fatalf("want one crew activation, got %d: %v", len(acts), labels(acts))
	}
	ids := crewIDsOf(t, acts[0])
	if len(ids) != 1 || ids[0] != pilot.String() {
		t.Fatalf("crew_ids = %v, want the Pilot alone (%s)", ids, pilot)
	}
	dispatchAll(t, g, active.ID, acts)
}

// Without anything else to tap, the Pilot is still offered: its counted
// power clears a number its real power could not.
func TestCrewIsOfferedToAPilotThatCountsEnough(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	vehicle := battlefieldCard(g, active, game.Card{
		Name: "Cost Power Vehicle", TypeLine: "Artifact — Vehicle", OracleID: oracleCostPowerVehicle, Power: 4, Toughness: 4,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), vehicle); len(acts) != 0 {
		t.Fatalf("a Vehicle with nothing to crew it was offered: %v", labels(acts))
	}

	pushPilot(g, active)
	acts := activationsOf(legal.EnumerateFor(g, active.ID), vehicle)
	if len(acts) != 1 {
		t.Fatalf("a 1/1 counting as 3 was not offered the crew: %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)
}

// Teamwork reuses the crew walk and counts plain power (CR 702.194a has
// no "as though its power were greater"): Repulsor Blast's teamwork 2 is
// paid by two plain 1/1s, and a Pilot's plain power of 1 is no more than
// that.
func TestTeamworkStillCountsPlainPower(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	blast := handCard(active, game.Card{
		Name: "Repulsor Blast", TypeLine: "Sorcery", OracleID: oracleRepulsorBlast, ManaCost: "{3}{R}",
	})
	lands(g, active, "Mountain", "Mountain", 4)
	pilot := pushPilot(g, active) // counts 3 to crew, 1 to teamwork
	elf := battlefieldCard(g, active, creature("Elf", "{G}", 1, 1))
	battlefieldCard(g, opp, creature("Ogre", "{2}{R}", 3, 3))
	advanceTo(t, g, game.StepPrecombatMain)

	paid := 0
	for _, m := range castMovesFor(legal.EnumerateFor(g, active.ID), blast) {
		p := teamworkBlightParams(t, m)
		if len(p.OptionalCosts) == 0 {
			continue
		}
		paid++
		if len(p.TeamworkIDs) != 2 {
			t.Errorf("%q pays teamwork 2 with %v, want both 1-power creatures (the Pilot counts 1 here, not 3)", m.Label, p.TeamworkIDs)
		}
		want := map[string]bool{pilot.String(): true, elf.String(): true}
		for _, id := range p.TeamworkIDs {
			if !want[id] {
				t.Errorf("%q names an unexpected creature %s", m.Label, id)
			}
		}
	}
	if paid == 0 {
		t.Fatal("the paid cast was not offered")
	}
}
