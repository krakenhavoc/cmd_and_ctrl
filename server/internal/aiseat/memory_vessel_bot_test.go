package aiseat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// memory_vessel_bot_test.go — #2559 through the real card. Under Memory
// Vessel a heuristic seat may play only the cards it exiled, and nothing
// from its hand: it is offered no hand play, and it still takes its turn
// out of exile rather than passing with a land drop and seven cards on
// offer.
//
// Not behind AISEAT_GAME_TESTS: one activation, one decision.

const botMemoryVesselOracle = "0179bc62-823e-46b9-b536-342904fedafc"

func TestHeuristicSeatPlaysFromExileUnderMemoryVessel(t *testing.T) {
	g := newSettledTable(t, 2559)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)

	forest := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(forest)
	vessel := game.Card{
		InstanceID: uuid.New(), Name: "Memory Vessel", TypeLine: "Artifact", OracleID: botMemoryVesselOracle,
		Owner: me.ID, Controller: me.ID,
	}
	vessel.AddKnowersAll(seatIDs(g))
	g.Battlefield.PushTop(vessel)
	if err := g.ActivateCatalogAbility(me.ID, vessel.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Memory Vessel: %v", err)
	}
	for i := 0; i < len(g.Seats) && !g.Exile.Contains(forest.InstanceID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Exile.Contains(forest.InstanceID) {
		t.Fatal("the Forest on top was not exiled")
	}

	moves := legal.EnumerateFor(g, me.ID)
	for _, m := range moves {
		if m.Type == legal.TypeCastSpell && me.Hand.Contains(m.Source) {
			t.Errorf("offered %q from the banned hand", m.Label)
		}
	}
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	d, err := heuristic.New().Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("decision %d out of range of %d moves", d.Index, len(moves))
	}
	chosen := moves[d.Index]
	var params struct {
		FromZone string `json:"from_zone"`
	}
	_ = json.Unmarshal(chosen.Params, &params)
	if chosen.Type != legal.TypeCastSpell || params.FromZone != "exile" {
		t.Errorf("the heuristic chose %q; with a land drop and its exiled cards on offer it should play from exile", chosen.Label)
	}
}
