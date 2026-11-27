package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noble Purpose — Enchantment {3}{W}{W} (Eighth Edition, S82 OLD batch 610,
// #5607):
//
//	"Whenever a creature you control deals combat damage, you gain that much
//	 life."
//
// One trigger per creature per damage event, to a player, creature or any
// other recipient; the life is the amount that creature dealt (the event's
// Amount, after prevention), gained by Noble Purpose's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84528204-faf2-45cb-bce0-365b8adf5339",
		Name:         "Noble Purpose",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.OldB610YourCreatureDealtCombatDamage, "Noble Purpose — you gain that much life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
