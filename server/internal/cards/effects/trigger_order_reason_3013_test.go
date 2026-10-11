package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// trigger_order_reason_3013_test.go — ADR 0018's #3013 amendment,
// sub-PR 1: the boards #2884's tests already ask on, each counted under
// the reason the check found first. These reasons are the data the
// owner decides the watcher and opacity rules on, so each must say what
// actually kept the prompt.

func promptsByReason(t *testing.T, reason game.TriggerOrderReason) float64 {
	t.Helper()
	return metricstest.Value(t, "cmdctrl_trigger_order_prompts_total", map[string]string{"reason": string(reason)})
}

func TestTriggerOrderPromptsAreCountedByTheirReason(t *testing.T) {
	cases := []struct {
		name   string
		reason game.TriggerOrderReason
		run    func(t *testing.T) *game.PendingChoice
	}{
		{
			name:   "an opponent's Ajani's Pridemate watches life change",
			reason: game.TriggerOrderReasonReactor,
			run: func(t *testing.T) *game.PendingChoice {
				g := newCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				pushCatalogPermanent(g, opp.ID, "Ajani's Pridemate", "Creature — Cat Soldier", b24AjanisPridemateOracle, false)
				pushCatalogPermanent(g, me.ID, "Guttersnipe", "Creature — Goblin Shaman", b03GuttersnipeOracle, false)
				pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
				castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
				return settleUntilOrderPrompt(t, g)
			},
		},
		{
			name:   "Ugin targets Vivi while Vivi puts a counter on herself",
			reason: game.TriggerOrderReasonObjectTie,
			run: func(t *testing.T) *game.PendingChoice {
				g := newCatalogGame(t)
				me := g.Seats[0]
				vivi := pushColoredVivi(g, me.ID)
				pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
				toMain(t, g)
				castColorlessSorcery(t, g)
				b04WaitForPick(t, g, me.ID)
				pickCard(t, g, me.ID, vivi)
				return triggerOrderPromptFor(g, me.ID)
			},
		},
		{
			name:   "Vivi's damage can end the opponent whose permanent Ugin exiles",
			reason: game.TriggerOrderReasonElimination,
			run: func(t *testing.T) *game.PendingChoice {
				g := newCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				pushColoredVivi(g, me.ID)
				pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
				victim := pushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
				toMain(t, g)
				opp.Life = 1
				castColorlessSorcery(t, g)
				b04WaitForPick(t, g, me.ID)
				pickCard(t, g, me.ID, victim)
				return triggerOrderPromptFor(g, me.ID)
			},
		},
		{
			name:   "a creature count is a closure the check cannot read",
			reason: game.TriggerOrderReasonUndescribed,
			run: func(t *testing.T) *game.PendingChoice {
				g := newCatalogGame(t)
				me := g.Seats[0]
				pushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", independenceTokenProbe, false)
				pushCatalogPermanent(g, me.ID, "Independence Counting Probe", "Creature — Construct", independenceCountingProbe, false)
				castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
				return settleUntilOrderPrompt(t, g)
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := promptsByReason(t, c.reason)
			if ch := c.run(t); ch == nil {
				t.Fatal("no trigger-order prompt")
			}
			if got := promptsByReason(t, c.reason) - before; got != 1 {
				t.Errorf("%q moved by %v, want 1", c.reason, got)
			}
		})
	}
}
