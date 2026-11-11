package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const spectacularShowdownOracle = "b6f61b77-d571-4508-a5e8-435b4e097f3a"

func TestSpectacularShowdownIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(spectacularShowdownOracle)
	if !ok || spec.Name != "Spectacular Showdown" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].ManaCost != "{4}{R}{R}{R}" {
		t.Errorf("alternative costs: %+v", spec.AlternativeCosts)
	}
}

func TestSpectacularShowdownGivesOneCreatureDoubleStrikeAndGoadsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	target := b12Creature(g, opp.ID, "Target", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)

	tdCastFromHand(t, g, game.Card{Name: "Spectacular Showdown", TypeLine: "Sorcery", OracleID: spectacularShowdownOracle, ManaCost: "{1}{R}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	passPriorityAroundTable(t, g)

	c, _ := g.LookupCardForEffect(target)
	if countersOn(g, target, game.CounterDoubleStrike) != 1 || !game.HasKeyword(&c, "double strike") {
		t.Errorf("no double strike counter / keyword on the target")
	}
	if !c.IsGoadedBy(me.ID) {
		t.Error("the target is not goaded")
	}
	b, _ := g.LookupCardForEffect(bystander)
	if b.IsGoaded() || countersOn(g, bystander, game.CounterDoubleStrike) != 0 {
		t.Error("the bystander was affected")
	}
}

func TestSpectacularShowdownOverloadedAffectsEachCreatureIncludingYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	tdCastFromHand(t, g, game.Card{Name: "Spectacular Showdown", TypeLine: "Sorcery", OracleID: spectacularShowdownOracle, ManaCost: "{1}{R}"},
		game.CastSpellParams{AlternativeCost: "overload"})
	if pick := latestPickTarget(g, me.ID); pick != nil {
		t.Fatalf("an overloaded spell asked for a target: %+v", pick)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		c, _ := g.LookupCardForEffect(id)
		if countersOn(g, id, game.CounterDoubleStrike) != 1 || !c.IsGoadedBy(me.ID) {
			t.Errorf("%s: double strike counter / goad missing", c.Name)
		}
	}
}
