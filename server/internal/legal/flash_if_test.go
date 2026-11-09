package legal_test

import (
	"encoding/json"
	"testing"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Timely Ward, Asinine Antics
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// flash_if_test.go — #5207, the enumerator half of "as though it had
// flash if …": a cast open only because of the card's own permission is
// offered only with the choices that make it apply (CR 601.3b), and a
// cast open only because an optional cost is announced is offered only
// announcing it (CR 601.3c). Every move is dispatched against a clone, so
// the engine accepts each one (#544).

func flashIfMoveParams(t *testing.T, m legal.Move) (targets []string, optional []int) {
	t.Helper()
	var p struct {
		Targets []struct {
			ID string `json:"id"`
		} `json:"targets"`
		OptionalCosts []int `json:"optional_costs"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	for _, r := range p.Targets {
		targets = append(targets, r.ID)
	}
	return targets, p.OptionalCosts
}

func TestFlashIfTargetsEnumeratesOnlyACommanderAtInstantSpeed(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	ward := handCard(active, game.Card{Name: "Timely Ward", TypeLine: "Enchantment — Aura", ManaCost: "{2}{W}",
		OracleID: "8e589dca-d31e-4d2f-81b1-18e5f9198795"})
	commander := battlefieldCard(g, active, game.Card{Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: true})
	bear := battlefieldCard(g, active, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	lands(g, active, "Plains", "Plains", 3)
	advanceTo(t, g, game.StepBeginCombat)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, ward)
	if len(casts) == 0 {
		t.Fatal("no instant-speed cast onto the commander")
	}
	for _, m := range casts {
		targets, _ := flashIfMoveParams(t, m)
		for _, id := range targets {
			if id == bear.String() {
				t.Errorf("move %q targets the plain creature at instant speed", m.Label)
			}
		}
		if len(targets) != 1 || targets[0] != commander.String() {
			t.Errorf("move %q targets %v, want the commander", m.Label, targets)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestFlashForMoreEnumeratesOnlyTheSurchargedCastAtInstantSpeed(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	antics := handCard(active, game.Card{Name: "Asinine Antics", TypeLine: "Sorcery", ManaCost: "{2}{U}{U}",
		OracleID: "0a0a051c-59ed-4286-b842-a34da7ef15d5"})
	lands(g, active, "Island", "Island", 6)
	advanceTo(t, g, game.StepBeginCombat)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, antics)
	if len(casts) == 0 {
		t.Fatal("no instant-speed cast paying {2} more")
	}
	for _, m := range casts {
		if _, optional := flashIfMoveParams(t, m); len(optional) != 1 || optional[0] != 0 {
			t.Errorf("move %q announces optional costs %v at instant speed, want [0]", m.Label, optional)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}
