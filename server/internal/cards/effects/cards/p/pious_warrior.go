package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pious Warrior — Creature — Human Rebel Warrior {3}{W}, 2/3 (Mercadian
// Masques, #5619):
//
//	"Whenever this creature is dealt combat damage, you gain that much
//	 life."
//
// Ripjaw Raptor's shape, limited to combat damage: one trigger per damage
// instance, waiting for the combat damage to settle (AtBatchEnd) so the
// amount is the total dealt to the Warrior at once (CR 120.3, CR 510.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "163e5482-700c-4d91-b70c-9170cc8f1ad2",
		Name:         "Pious Warrior",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Combat && effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Pious Warrior — you gain that much life",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
