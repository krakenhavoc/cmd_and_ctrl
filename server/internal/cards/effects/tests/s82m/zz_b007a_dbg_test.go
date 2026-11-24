package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestZZB007ADbg(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	enemy := effectstest.S82Walker(g, opp.ID, "Test Walker", "", 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	tok := s82OldB007Named(g, "Assassin")[0]
	effectstest.DealCombatDamage(g, me.ID, tok.InstanceID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	t.Logf("after player hit counters %d", effectstest.CounterCount(g, tok.InstanceID, game.CounterPlusOne))
	t.Logf("vraska on bf %v", effectstest.LoyaltyOf(g, vraska))
	effectstest.DealCombatDamage(g, me.ID, tok.InstanceID, enemy)
	effectstest.PassPriorityAroundTable(t, g)
	t.Logf("counters %d", effectstest.CounterCount(g, tok.InstanceID, game.CounterPlusOne))
}
