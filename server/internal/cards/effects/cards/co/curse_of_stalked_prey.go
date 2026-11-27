package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Stalked Prey — Enchantment — Aura Curse {1}{R}:
//
//	"Enchant player
//	 Whenever a creature deals combat damage to enchanted player, put a +1/+1
//	 counter on that creature."
//
// One trigger per creature that dealt combat damage; the counter goes on the
// creature if it is still on the battlefield as the trigger resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7971a2a6-f89f-46a5-9008-96cfa49de41a",
		Name:         "Curse of Stalked Prey",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.OldB444CombatDamageToEnchantedPlayerBy,
				"Curse of Stalked Prey — put a +1/+1 counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.NothingIfGone(effects.AddCounter{Target: item.Trigger.Event.Source, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item)))
				}),
		},
	})
}
