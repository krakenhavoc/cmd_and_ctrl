package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphere Grid — Enchantment {1}{G}:
//
//	"Whenever a creature you control deals combat damage to a player,
//	 put a +1/+1 counter on that creature.
//	 Unlock Ability — Creatures you control with +1/+1 counters on them
//	 have reach and trample."
//
// "Unlock Ability" is an ability word: it names the static and does
// nothing of its own. The trigger is one per creature that connects
// (each damage event names its source), and the counter goes on that
// creature if it is still on the battlefield. The static is Exava's
// counter-gated grant, once for reach and once for trample.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4770345-f15e-4e31-b6ef-4c2608df9644",
		Name:         "Sphere Grid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Sphere Grid — put a +1/+1 counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					id := item.Trigger.Event.Source
					if !onBattlefield(g, id) {
						return nil
					}
					return nothingIfGone(g.AddCounterByForEffect(item.Controller, id, game.CounterPlusOne, 1))
				}),
		},
		Static: []game.StaticAbility{
			CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "reach"),
			CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "trample"),
		},
	})
}
