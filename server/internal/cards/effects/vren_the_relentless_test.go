package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const vrenOracle = "1216900e-93e3-41a4-b354-9c81938639c2"

func TestVrenIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(vrenOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Vren: %+v %v", spec, ok)
	}
}

func TestVrenExilesOpponentsDyingCreaturesAndMakesThatManyRatsAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dskB1Creature(g, me.ID, "Vren, the Relentless", "Legendary Creature — Rat Rogue", vrenOracle, 3, 4)
	a := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tok := fcgToken(g, opp.ID, "Their Goblin")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, opp.ID, a)
	killCreature(t, g, opp.ID, tok)
	killCreature(t, g, me.ID, mine)
	if !g.Exile.Contains(a) {
		t.Fatal("the opponent's creature did not go to exile")
	}
	if opp.Graveyard.Contains(a) {
		t.Error("the opponent's creature went to the graveyard")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own creature was exiled instead of dying")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	rats := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Rat") {
			rats++
		}
	}
	if rats != 2 {
		t.Fatalf("Rat tokens: %d, want 2 (a Bear and a token were exiled under opponents' control)", rats)
	}
	// Each Rat gets +1/+1 for each other Rat: Vren is a Rat too, so each
	// token sees Vren and the other token.
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Rat") {
			if e := c.Effective(); e.Power != 3 || e.Toughness != 3 {
				t.Errorf("Rat token is %d/%d, want 3/3 (two other Rats)", e.Power, e.Toughness)
			}
		}
	}
}

func TestVrenMakesNoRatsWhenNothingWasExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dskB1Creature(g, me.ID, "Vren, the Relentless", "Legendary Creature — Rat Rogue", vrenOracle, 3, 4)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Error("a token appeared with nothing exiled")
		}
	}
}
