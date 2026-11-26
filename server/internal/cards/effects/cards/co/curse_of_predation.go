package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Predation — Enchantment — Aura Curse {2}{G} (Commander 2013, S82
// older sets):
//
//	"Enchant player
//	 Whenever a creature attacks enchanted player, put a +1/+1 counter on
//	 it."
//
// One trigger per attacking creature (not batched), controlled by the
// Curse's controller. The counter goes on the creature that attacked if it is
// still the same object on the battlefield as the ability resolves
// (CR 400.7, 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69ea4ce2-f749-4d5b-a392-efcfdce14a30",
		Name:         "Curse of Predation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.OldB383WheneverACreatureAttacksEnchantedPlayer("Curse of Predation — put a +1/+1 counter on the attacking creature", false,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					return effects.AddCounter{Target: ctx.Trigger().Event.CardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}
