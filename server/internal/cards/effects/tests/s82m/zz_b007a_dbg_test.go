package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestZZB007ADbg(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	snake := effectstest.B12Creature(g, me.ID, "Snake", "Creature — Snake", 1, 1)
	s82OldB007Keyword(t, g, snake, "deathtouch")
	enemy := effectstest.S82Walker(g, opp.ID, "Test Walker", "", 9)
	effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamage(g, me.ID, snake, enemy)
	t.Logf("stack %d", g.Stack.Size())
	effectstest.PassPriorityAroundTable(t, g)
	t.Logf("counters %d", effectstest.CounterCount(g, snake, game.CounterPlusOne))
}
