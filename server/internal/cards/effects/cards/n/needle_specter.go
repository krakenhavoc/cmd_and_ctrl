package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Needle Specter — Creature — Specter {1}{B}{B}, 1/1 (Eventide, S82 OLD batch
// 520, #5590):
//
//	"Flying
//	 Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	 Whenever this creature deals combat damage to a player, that player discards
//	 that many cards."
//
// Banshee of the Dread Choir's combat-damage trigger over the damage dealt: the
// damaged player is the event's target and "that many" is the combat damage
// amount (CR 120.3, a wither creature dealing damage to a player deals it as
// ordinary damage), and they choose the cards (CR 701.9b); a hand with fewer
// cards discards what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea8acbd4-1ace-4634-9e88-7f91401d2b30",
		Name:            "Needle Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "wither"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Needle Specter — that player discards that many cards",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player: item.Trigger.Event.Target,
						Source: item.SourceCardID,
						N:      item.Trigger.Event.Amount,
					})
					return nil
				}),
		},
	})
}
