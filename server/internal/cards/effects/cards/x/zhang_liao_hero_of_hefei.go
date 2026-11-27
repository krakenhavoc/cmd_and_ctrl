package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zhang Liao, Hero of Hefei — Legendary Creature — Human Soldier {4}{B}{B}, 3/3
// (Portal Three Kingdoms, S82 #5622):
//
//	"Whenever Zhang Liao deals damage to an opponent, that opponent discards a
//	 card."
//
// Any damage, combat or not, to one of its controller's opponents (the shared
// ThisDealtDamageToAnOpponent watch). "That opponent" is the damage event's
// target, and they choose the card to discard (CR 701.9b); one with an empty hand
// discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "244541f6-cbfc-4e89-9156-64e412f745a7",
		Name:         "Zhang Liao, Hero of Hefei",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToAnOpponent,
				"Zhang Liao — that opponent discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, effects.NewContext(g, item).Trigger().Event.Target)
				}),
		},
	})
}
