package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB163ScavengerOracle = "ce068b46-9301-42fb-9db1-9c0da5275d89"

func TestOldB163SilumgarScavengerExploitsACreatureAndGainsHaste(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	id := effectstest.ExploitCast(t, g, me, "Silumgar Scavenger", "Creature — Zombie Bird", oldB163ScavengerOracle)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	c := effectstest.EffectiveCardForTest(t, g, id)
	if !effectstest.EffectiveHasKeyword(&c, "haste") {
		t.Error("the Scavenger that exploited the creature has no haste")
	}
	if !effectstest.EffectiveHasKeyword(&c, "flying") {
		t.Error("the Scavenger lost flying")
	}
}

func TestOldB163SilumgarScavengerGetsACounterButNoHasteWhenItDidNotExploit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Silumgar Scavenger", "Creature — Zombie Bird", oldB163ScavengerOracle, false)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Goblin", 1, 1)
	effectstest.B18Kill(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	c := effectstest.EffectiveCardForTest(t, g, id)
	if effectstest.EffectiveHasKeyword(&c, "haste") {
		t.Error("the Scavenger gained haste without exploiting the creature")
	}
	// An opponent's creature dying is not "another creature you control".
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	effectstest.B18Kill(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after an opposing creature died, want 1", got)
	}
}
