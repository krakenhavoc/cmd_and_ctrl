package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rites of Flourishing — Enchantment {2}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"At the beginning of each player's draw step, that player draws an
//	 additional card.
//	 Each player may play an additional land on each of their turns."
//
// The trigger is Howling Mine's without the tap condition: it goes on
// the stack after the turn's draw (the ruling) and the active player,
// the event's actor, draws. The land drop is a player-scoped static the
// engine derives (AdditionalLandPlaysForEachPlayer): every player's
// allowance rises by one while the Rites is on the battlefield, whoever
// controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                         "7080fb43-5d93-41f5-87d3-bc1801805ea0",
		Name:                             "Rites of Flourishing",
		Completeness:                     CompletenessFull,
		Purpose:                          game.Purpose{ExtraLandDrops: 1},
		AdditionalLandPlays:              1,
		AdditionalLandPlaysForEachPlayer: true,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginDrawStep, AnyPlayer, "Rites of Flourishing — that player draws an additional card",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNForEffect(item.Trigger.Event.Actor, 1)
				}),
		},
	})
}
