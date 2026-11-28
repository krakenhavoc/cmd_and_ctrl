package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cold Snap — Enchantment {2}{W} (Ice Age, S82 OLD batch 838, #5636):
//
//	"Cumulative upkeep {2}
//	 At the beginning of each player's upkeep, this enchantment deals damage to
//	 that player equal to the number of snow lands they control."
//
// Furnace Punisher's every-upkeep trigger: the upkeep's player is read off the
// event and their snow lands are counted as the trigger resolves (CR 608.2h);
// with none, no damage is dealt. Damage from a Cold Snap that has left is still
// dealt by its last-known information.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dcbcf5e-bdbe-46b3-881e-5ac4030c6e38",
		Name:         "Cold Snap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Cold Snap — cumulative upkeep {2}", "{2}"),
			effects.AtEachUpkeep("Cold Snap — damage to that player equal to the number of snow lands they control",
				func(g *game.Game, item *game.StackItem) error {
					who := item.Trigger.Event.Actor
					n := 0
					for i := range g.Battlefield.Cards {
						if c := g.Battlefield.Cards[i]; c.Controller == who && c.IsLand() && c.HasSupertype("snow") {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: who, Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
