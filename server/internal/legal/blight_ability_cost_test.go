package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// blight_ability_cost_test.go — #4684: the enumerator's half of "Blight N"
// as part of an activated ability's cost (Gristle Glutton, Evershrike's
// Gift). #544's invariant: one move per creature the seat could blight,
// a survivor first, none with no creature, the price on MoveCost.Blight,
// and every move offered is one the dispatcher accepts.

const (
	oracleGristleGlutton  = "b7163800-41fc-410a-9650-50ae90842170"
	oracleEvershrikesGift = "b5bc4adf-9105-4901-989c-63612314af5c"
)

func blightedIDsOf(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		BlightIDs []string `json:"blight_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.BlightIDs
}

func TestBlightCostIsOneMovePerCreatureASurvivorFirst(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	glutton := battlefieldCard(g, active, game.Card{
		Name: "Gristle Glutton", TypeLine: "Creature — Goblin Scout", OracleID: oracleGristleGlutton,
		ManaCost: "{1}{R}", Power: 1, Toughness: 3,
	})
	bear := battlefieldCard(g, active, creature("Bear", "{1}{G}", 3, 3))
	token := battlefieldCard(g, active, creature("Goblin", "", 1, 1))
	battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 3, 3))
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), glutton)
	var got []string
	for _, m := range acts {
		ids := blightedIDsOf(t, m)
		if len(ids) != 1 {
			t.Fatalf("move %q blights %v, want exactly one creature", m.Label, ids)
		}
		if m.Cost == nil || m.Cost.Blight == nil || m.Cost.Blight.N != 1 ||
			m.Cost.Blight.CardID.String() != ids[0] || m.Cost.Blight.Counter != game.CounterMinusOne {
			t.Errorf("move %q: cost %+v, want the one -1/-1 counter on %s", m.Label, m.Cost, ids[0])
		}
		got = append(got, ids[0])
	}
	// The Glutton (survives, power 1), the Bear (survives, power 3), then
	// the 1/1 it would kill. Never the opponent's creature.
	want := []string{glutton.String(), bear.String(), token.String()}
	if len(got) != len(want) {
		t.Fatalf("blighted %v, want %v (%v)", got, want, labels(acts))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("move %d blights %s, want %s", i, got[i], want[i])
		}
	}
	dispatchAll(t, g, active.ID, acts)
}

// A graveyard ability blights a creature on the battlefield, and with no
// creature to blight it is not offered (CR 701.68b).
func TestBlightCostFromTheGraveyardNeedsACreature(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	lands(g, active, "Plains", "Plains", 2)
	gift := graveyardCard(active, game.Card{
		Name: "Evershrike's Gift", TypeLine: "Enchantment — Aura", OracleID: oracleEvershrikesGift, ManaCost: "{W}",
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), gift); len(acts) != 0 {
		t.Fatalf("the return is offered with no creature to blight: %v", labels(acts))
	}

	wall := battlefieldCard(g, active, creature("Wall", "{1}{W}", 0, 4))
	acts := activationsOf(legal.EnumerateFor(g, active.ID), gift)
	if len(acts) != 1 {
		t.Fatalf("moves %v, want one return blighting the Wall", labels(acts))
	}
	if ids := blightedIDsOf(t, acts[0]); len(ids) != 1 || ids[0] != wall.String() {
		t.Errorf("blights %v, want the Wall %s", ids, wall)
	}
	if b := acts[0].Cost.Blight; b == nil || b.N != 2 {
		t.Errorf("cost %+v, want two -1/-1 counters", acts[0].Cost)
	}
	dispatchAll(t, g, active.ID, acts)
}
