package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warren Instigator — Creature — Goblin Berserker {R}{R}, 1/1 (Secret Lair
// Drop, #5462):
//
//	"Double strike
//	 Whenever this creature deals damage to an opponent, you may put a
//	 Goblin creature card from your hand onto the battlefield."
//
// Goblin Lackey's trigger narrowed to an opponent and to Goblin CREATURE
// cards. Any damage counts, combat or not, and with double strike a creature
// that connects in both damage steps triggers twice. The put is a "you may"
// answered as the trigger resolves, a put and not a cast (CR 110.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47f75e25-8a52-4452-a3a8-bf1bd00942d1",
		Name:            "Warren Instigator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Warren Instigator — you may put a Goblin creature card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Match:    effects.And(effects.Creature(), effects.Subtype("Goblin")),
						Optional: true,
						Label:    "Warren Instigator — you may put a Goblin creature card from your hand onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
