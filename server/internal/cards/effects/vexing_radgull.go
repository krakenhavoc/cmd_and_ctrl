package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vexing Radgull — Creature — Bird Mutant {1}{U}, 1/2:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 gets two rad counters if they don't have any rad counters.
//	 Otherwise, proliferate."
//
// #2042: rad counters now do something (CR 728.1). Whether the damaged
// player has a rad counter is read as the trigger resolves, so a counter
// given or removed in response changes the branch. The proliferate is
// the controller's own choice (Proliferate{}), over every permanent and
// player with a counter, as any proliferate is.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "2b632e67-ff80-4d52-9f44-3bc4bf083ab1",
		Name:            "Vexing Radgull",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Vexing Radgull — two rad counters, or proliferate",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					victim := item.Trigger.Event.Target
					p := g.PlayerByIDForEffect(victim)
					if p == nil || p.Eliminated {
						return nil
					}
					if p.Counters[game.CounterRad] == 0 {
						return playerGetsRadCounters(g, item.Controller, victim, 2)
					}
					return Proliferate{}.Apply(NewContext(g, item))
				}),
		},
	})
}
