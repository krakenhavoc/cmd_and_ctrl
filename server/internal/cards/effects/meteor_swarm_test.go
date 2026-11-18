package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const meteorSwarmOracle = "3e7bd846-7295-4590-a419-c97cd2526bc1"

func TestMeteorSwarmDividesEightAmongXTargets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	b12PlayFromHand(t, g, "Meteor Swarm", "Sorcery", meteorSwarmOracle, game.CastSpellParams{
		XValue: 3, Targets: cardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 5, b: 2, c: 1}})
	passPriorityAroundTable(t, g)
	got := [3]int{e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked, e2Card(t, g, c).DamageMarked}
	if got != [3]int{5, 2, 1} {
		t.Errorf("damage = %v, want [5 2 1]", got)
	}
}

func TestMeteorSwarmRefusesWrongCountOrDivision(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Meteor Swarm", TypeLine: "Sorcery",
		OracleID: meteorSwarmOracle, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	cases := map[string]game.CastSpellParams{
		"X=2, one target":  {XValue: 2, Targets: cardRefs(a), Distribution: map[uuid.UUID]int{a: 8}},
		"X=1, two targets": {XValue: 1, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 4, b: 4}},
		"sum is not eight": {XValue: 2, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 3, b: 4}},
		"a zero share":     {XValue: 2, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 8, b: 0}},
	}
	for name, p := range cases {
		p.Strict = true
		if err := g.CastSpell(active.ID, id, p); !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: CastSpell = %v, want ErrInvalidParam", name, err)
		}
	}
}
