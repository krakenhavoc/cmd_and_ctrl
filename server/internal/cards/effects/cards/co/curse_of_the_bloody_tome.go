package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of the Bloody Tome — Enchantment — Aura Curse {2}{U}:
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, that player mills two
//	 cards."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5480f855-80b7-45c7-94a9-071c7b5d5ccf",
		Name:         "Curse of the Bloody Tome",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.OldB444EnchantedPlayersUpkeepTrigger("Curse of the Bloody Tome — that player mills two cards",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.MillCards{Player: item.Trigger.Event.Actor, N: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
