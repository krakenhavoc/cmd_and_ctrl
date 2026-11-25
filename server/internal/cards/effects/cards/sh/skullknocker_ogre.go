package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullknocker Ogre — Creature — Ogre {3}{R}, 4/3 (Throne of Eldraine,
// #5516):
//
//	"Whenever this creature deals damage to an opponent, that player
//	 discards a card at random. If the player does, they draw a card."
//
// Fungal Shambler's trigger (any damage to an opponent), with the engine's
// random discard. The draw follows only when a card really left the hand
// (CR 608.2c), so an empty hand draws nothing; the damaged opponent draws.
//
// No simplification.
func init() {
	const label = "Skullknocker Ogre — that player discards a card at random; if the player does, they draw a card"
	effects.Register(effects.Spec{
		OracleID:     "980a0d4d-b070-46c1-be8d-31bdda0d1c33",
		Name:         "Skullknocker Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent(label, func(g *game.Game, item *game.StackItem) error {
				return g.DiscardRandomKeyedForEffect(item.Trigger.Event.Target, 1, effects.NewContext(g, item).ItemTail(skullknockerOgreDraw))
			}),
		},
	})
}

// skullknockerOgreDraw is "if the player does, they draw a card": the damaged
// opponent draws only for a card that really was discarded (keyed, #1027).
var skullknockerOgreDraw = effects.CardTail("skullknocker-ogre/draw-if-discarded", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Item.Trigger.Event.Target, N: 1}.Apply(ctx)
})
