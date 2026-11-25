package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ohran Viper — Snow Creature — Snake {1}{G}{G}, 1/3 (Commander 2015, S82 OLD
// #5546):
//
//	"Whenever this creature deals combat damage to a creature, destroy that
//	 creature at end of combat.
//	 Whenever this creature deals combat damage to a player, you may draw a
//	 card."
//
// The first trigger is Kaldra Compleat's shape: the creature dealt damage is
// recorded on the item as the object it was (CR 400.7) as the trigger goes on
// the stack, and the ability schedules a delayed trigger at the beginning of
// the end of combat step (CR 511.2) that destroys that object if it is still
// on the battlefield (Glyph of Destruction's body). The second is an optional
// draw (CR 603.5).
//
// No simplification.
func init() {
	const destroyLabel = "Ohran Viper — destroy the creature it dealt combat damage to at end of combat"
	destroy := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventDealDamage},
		AppliesTo: effects.ThisDealtCombatDamageToACreature,
		Key:       destroyLabel,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, destroyLabel)
			item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
			return item
		},
		Effect: ohranViperScheduleDestroy,
	}
	effects.Register(effects.Spec{
		OracleID:     "879fdcd8-62ea-4d39-abd4-44e6d82adf4d",
		Name:         "Ohran Viper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			destroy,
			effects.Optional(effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Ohran Viper — draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Ohran Viper — draw a card?"),
		},
	})
}

func ohranViperScheduleDestroy(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	if ref.ID == [16]byte{} {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:     game.StepEndCombat,
		Label:  "Ohran Viper — destroy that creature",
		Cards:  []uuid.UUID{ref.ID},
		Body:   effects.DestroyTheObjectBody,
		Params: game.EffectParams{Object: ref},
	}.Apply(effects.NewContext(g, item))
}
