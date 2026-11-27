package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Order of Yawgmoth — Creature — Phyrexian Zombie Knight 2/2 {2}{B}{B} (Urza's Saga, S82 OLD b723):
//
//	Fear
//	Whenever this creature deals damage to a player, that player discards a card.
//
// Fear is a printed keyword. The trigger is any damage to any player (combat or not), and the player chooses their own discard, as Blazing Specter's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c1795e9c-31c9-4892-a796-7656e09257b3",
		Name:            "Order of Yawgmoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToAPlayer, "Order of Yawgmoth — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, item.Trigger.Event.Target)
				}),
		},
	})
}
