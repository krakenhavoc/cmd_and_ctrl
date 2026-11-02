package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// unattach_cost_test.go — #4799: the enumerator's half of "Unattach an
// Equipment from <this>" (Captain America, First Avenger). #544's
// invariant: one row per Equipment attached to him, the division sized
// from THAT Equipment's mana value, none at all with nothing attached —
// and every move offered is one the dispatcher accepts.

const oracleCapFirstAvenger = "7d038055-035d-4698-b6b6-e2859bf2f1d1"

func capFirstAvenger() game.Card {
	return game.Card{
		Name: "Captain America, First Avenger", TypeLine: "Legendary Creature — Human Soldier Hero",
		OracleID: oracleCapFirstAvenger, ManaCost: "{R}{W}{U}", Power: 4, Toughness: 4,
	}
}

func equipmentOn(g *game.Game, p *game.Player, name, cost string, host uuid.UUID) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name: name, TypeLine: "Artifact — Equipment", ManaCost: cost,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func unattachedIDsOf(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		UnattachIDs []string `json:"unattach_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.UnattachIDs
}

func TestUnattachCostIsOneRowPerEquipmentSizedByItsManaValue(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lands(g, active, "Mountain", "Mountain", 3)
	cap := battlefieldCard(g, active, capFirstAvenger())
	dagger := equipmentOn(g, active, "Dagger", "{1}", cap)
	axe := equipmentOn(g, active, "Axe", "{2}", cap)
	other := battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	equipmentOn(g, active, "Elsewhere", "{4}", other)
	for _, name := range []string{"A", "B", "C"} {
		battlefieldCard(g, opp, creature(name, "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), cap)
	if len(acts) == 0 {
		t.Fatal("the throw is not offered")
	}
	mv := map[string]int{dagger.String(): 1, axe.String(): 2}
	widest := map[string]int{}
	for _, m := range acts {
		ids := unattachedIDsOf(t, m)
		if len(ids) != 1 {
			t.Fatalf("move %q unattaches %v, want exactly one Equipment", m.Label, ids)
		}
		total, ok := mv[ids[0]]
		if !ok {
			t.Fatalf("move %q unattaches %s, which is not attached to Captain America", m.Label, ids[0])
		}
		d := decodeDivided(t, m)
		if len(d.Targets) > total {
			t.Errorf("%q: %d targets cannot divide %d", m.Label, len(d.Targets), total)
		}
		checkEvenDivision(t, m.Label, d, total)
		if n := len(d.Targets); n > widest[ids[0]] {
			widest[ids[0]] = n
		}
	}
	if widest[dagger.String()] != 1 || widest[axe.String()] != 2 {
		t.Errorf("widest throw per Equipment = %v, want the Dagger 1 target and the Axe 2", widest)
	}
	dispatchAll(t, g, active.ID, acts)
}

func TestUnattachCostIsNotEnumeratedWithNothingAttached(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lands(g, active, "Mountain", "Mountain", 3)
	cap := battlefieldCard(g, active, capFirstAvenger())
	// An Equipment on the battlefield, attached to nothing, pays nothing.
	battlefieldCard(g, active, game.Card{Name: "Loose Axe", TypeLine: "Artifact — Equipment", ManaCost: "{2}"})
	battlefieldCard(g, opp, creature("A", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), cap); len(acts) != 0 {
		t.Errorf("the throw is offered with no Equipment attached: %v", labels(acts))
	}
}
